package storage

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"sync"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/storage/forks"
	forksfactory "github.com/bloxapp/ssv/ibft/storage/forks/factory"
	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	qbftstorage "github.com/bloxapp/ssv/protocol/v2/qbft/storage"
	"github.com/bloxapp/ssv/storage/basedb"
)

const (
	highestInstanceKey = "highest_instance"
	instanceKey        = "instance"
)

var (
	metricsHighestDecided = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:validator:ibft_highest_decided",
		Help: "The highest decided sequence number",
	}, []string{"identifier", "pubKey"})
)

func init() {
	if err := prometheus.Register(metricsHighestDecided); err != nil {
		log.Println("could not register prometheus collector")
	}
}

// ibftStorage struct
// instanceType is what separates different iBFT eth2 duty types (attestation, proposal and aggregation)
type ibftStorage struct {
	prefix   []byte
	db       basedb.IDb
	logger   *zap.Logger
	fork     forks.Fork
	forkLock *sync.RWMutex

	// history is used to determine if we should store history instances.
	// If false, we will only store the highest instances.
	history bool
}

// New create new ibft storage
func New(db basedb.IDb, logger *zap.Logger, prefix string, forkVersion forksprotocol.ForkVersion, history bool) qbftstorage.QBFTStore {
	ibft := &ibftStorage{
		prefix:   []byte(prefix),
		db:       db,
		logger:   logger,
		fork:     forksfactory.NewFork(forkVersion),
		forkLock: &sync.RWMutex{},
		history:  history,
	}
	return ibft
}

func (i *ibftStorage) OnFork(forkVersion forksprotocol.ForkVersion) error {
	i.forkLock.Lock()
	defer i.forkLock.Unlock()

	logger := i.logger.With(zap.String("where", "OnFork"))
	logger.Info("forking ibft storage")
	i.fork = forksfactory.NewFork(forkVersion)
	return nil
}

// GetHighestInstance returns the StoredInstance for the highest instance.
func (i *ibftStorage) GetHighestInstance(identifier []byte) (*qbftstorage.StoredInstance, error) {
	val, found, err := i.get(highestInstanceKey, identifier[:])
	if !found {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ret := &qbftstorage.StoredInstance{}
	if err := ret.Decode(val); err != nil {
		return nil, errors.Wrap(err, "could not decode instance")
	}
	return ret, nil
}

// SaveInstance saves historical StoredInstance.
func (i *ibftStorage) SaveInstance(instance *qbftstorage.StoredInstance) error {
	value, err := instance.Encode()
	if err != nil {
		return errors.Wrap(err, "could not encode instance")
	}

	if i.history {
		err = i.save(value, instanceKey, instance.State.ID, uInt64ToByteSlice(uint64(instance.State.Height)))
		if err != nil {
			return errors.Wrap(err, "could not save historical instance")
		} else {
			var signers []spectypes.OperatorID
			if instance.DecidedMessage != nil {
				signers = instance.DecidedMessage.Signers
			}
			i.logger.Debug("Saved historical instance",
				zap.String("identifier", hex.EncodeToString(instance.State.ID)),
				zap.Uint64("height", uint64(instance.State.Height)),
				zap.Any("signers", signers))
		}
	}

	i.forkLock.RLock()
	defer i.forkLock.RUnlock()

	// Save as highest instance if it's higher than the current highest instance.
	highest, err := i.GetHighestInstance(instance.State.ID)
	if err != nil {
		return errors.Wrap(err, "could not get highest instance")
	}
	if highest != nil && highest.State.Height >= instance.State.Height {
		return nil
	}

	err = i.save(value, highestInstanceKey, instance.State.ID)
	return errors.Wrap(err, "could not save instance")
}

// GetInstance returns historical StoredInstance for the given identifier and height.
func (i *ibftStorage) GetInstance(identifier []byte, height specqbft.Height) (*qbftstorage.StoredInstance, error) {
	i.forkLock.RLock()
	defer i.forkLock.RUnlock()

	val, found, err := i.get(instanceKey, identifier[:], uInt64ToByteSlice(uint64(height)))
	if !found {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ret := &qbftstorage.StoredInstance{}
	if err := ret.Decode(val); err != nil {
		return nil, errors.Wrap(err, "could not decode instance")
	}
	return ret, nil
}

// GetInstancesInRange returns historical StoredInstance's in the given range.
func (i *ibftStorage) GetInstancesInRange(identifier []byte, from specqbft.Height, to specqbft.Height) ([]*qbftstorage.StoredInstance, error) {
	if !i.history {
		return nil, errors.New("history is disabled")
	}

	i.forkLock.RLock()
	defer i.forkLock.RUnlock()

	instances := make([]*qbftstorage.StoredInstance, 0)

	for seq := from; seq <= to; seq++ {
		instance, err := i.GetInstance(identifier, seq)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get instance")
		}
		if instance != nil {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// CleanAllInstances removes all StoredInstance's & highest StoredInstance's for msgID.
func (i *ibftStorage) CleanAllInstances(msgID []byte) error {
	i.forkLock.RLock()
	defer i.forkLock.RUnlock()

	prefix := i.prefix
	prefix = append(prefix, msgID[:]...)
	prefix = append(prefix, []byte(instanceKey)...)
	n, err := i.db.DeleteByPrefix(prefix)
	if err != nil {
		return errors.Wrap(err, "failed to remove decided")
	}
	i.logger.Debug("removed decided", zap.Int("count", n),
		zap.String("identifier", hex.EncodeToString(msgID)))
	if err := i.delete(highestInstanceKey, msgID[:]); err != nil {
		return errors.Wrap(err, "failed to remove last decided")
	}
	return nil
}

func (i *ibftStorage) save(value []byte, id string, pk []byte, keyParams ...[]byte) error {
	prefix := append(i.prefix, pk...)
	key := i.key(id, keyParams...)
	return i.db.Set(prefix, key, value)
}

func (i *ibftStorage) get(id string, pk []byte, keyParams ...[]byte) ([]byte, bool, error) {
	prefix := append(i.prefix, pk...)
	key := i.key(id, keyParams...)
	obj, found, err := i.db.Get(prefix, key)
	if !found {
		return nil, found, nil
	}
	if err != nil {
		return nil, found, err
	}
	return obj.Value, found, nil
}

func (i *ibftStorage) delete(id string, pk []byte, keyParams ...[]byte) error {
	prefix := append(i.prefix, pk...)
	key := i.key(id, keyParams...)
	return i.db.Delete(prefix, key)
}

func (i *ibftStorage) key(id string, params ...[]byte) []byte {
	ret := []byte(id)
	for _, p := range params {
		ret = append(ret, p...)
	}
	return ret
}

func uInt64ToByteSlice(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}
