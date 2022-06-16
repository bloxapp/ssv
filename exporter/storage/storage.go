package storage

import (
	eth12 "github.com/bloxapp/ssv/protocol/v1/blockchain/eth1"
	"sync"

	"github.com/bloxapp/ssv/eth1"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"go.uber.org/zap"
)

func storagePrefix() []byte {
	return []byte("exporter/")
}

// Storage represents the interface of exporter storage
type Storage interface {
	eth1.SyncOffsetStorage
	registrystorage.OperatorsCollection
	ValidatorsCollection
	eth12.RegistryStore
}

type storage struct {
	db     basedb.IDb
	logger *zap.Logger

	validatorsLock sync.RWMutex

	operatorStore registrystorage.OperatorsCollection
}

func (s *storage) GetOperatorDataByPubKey(operatorPubKey string) (*registrystorage.OperatorData, bool, error) {
	return s.operatorStore.GetOperatorDataByPubKey(operatorPubKey)
}

func (s *storage) GetOperatorData(index uint64) (*registrystorage.OperatorData, bool, error) {
	return s.operatorStore.GetOperatorData(index)
}

func (s *storage) SaveOperatorData(operatorData *registrystorage.OperatorData) error {
	return s.operatorStore.SaveOperatorData(operatorData)
}

func (s *storage) ListOperators(from uint64, to uint64) ([]registrystorage.OperatorData, error) {
	return s.operatorStore.ListOperators(from, to)
}

func (s *storage) GetOperatorsPrefix() []byte {
	return s.operatorStore.GetOperatorsPrefix()
}

// NewExporterStorage creates a new instance of Storage
func NewExporterStorage(db basedb.IDb, logger *zap.Logger) Storage {
	return &storage{
		db:            db,
		logger:        logger.With(zap.String("component", "exporter/storage")),
		operatorStore: registrystorage.NewOperatorsStorage(db, logger, storagePrefix()),
	}
}

// CleanRegistryData clears storage registry data
func (s *storage) CleanRegistryData() error {
	return s.db.RemoveAllByCollection(storagePrefix())
}

// nextIndex returns the next index for operator
// TODO: change to uint64
func (s *storage) nextIndex(prefix []byte) (int64, error) {
	n, err := s.db.CountByCollection(append(storagePrefix(), prefix...))
	if err != nil {
		return 0, err
	}
	return n, err
}
