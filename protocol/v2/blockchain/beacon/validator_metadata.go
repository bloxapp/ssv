package beacon

import (
	"encoding/hex"
	"math"
	"strings"
	"time"

	eth2apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/queue"
)

//go:generate mockgen -package=beacon -destination=./mock_validator_metadata.go -source=./validator_metadata.go

// ValidatorMetadataStorage interface for validator metadata
type ValidatorMetadataStorage interface {
	UpdateValidatorMetadata(logger *zap.Logger, pk string, metadata *ValidatorMetadata) error
}

// ValidatorMetadata represents validator metdata from beacon
type ValidatorMetadata struct {
	Balance phase0.Gwei              `json:"balance"`
	Status  eth2apiv1.ValidatorState `json:"status"`
	Index   phase0.ValidatorIndex    `json:"index"` // pointer in order to support nil
}

// Equals returns true if the given metadata is equal to current
func (m *ValidatorMetadata) Equals(other *ValidatorMetadata) bool {
	return other != nil &&
		m.Status == other.Status &&
		m.Index == other.Index &&
		m.Balance == other.Balance
}

// Pending returns true if the validator is pending
func (m *ValidatorMetadata) Pending() bool {
	return m.Status.IsPending()
}

// Activated returns true if the validator is not unknown. It might be pending activation or active
func (m *ValidatorMetadata) Activated() bool {
	return m.Status.HasActivated() || m.Status.IsActive() || m.Status.IsAttesting()
}

// IsActive returns true if the validator is currently active. Cant be other state
func (m *ValidatorMetadata) IsActive() bool {
	return m.Status == eth2apiv1.ValidatorStateActiveOngoing
}

// Exiting returns true if the validator is existing or exited
func (m *ValidatorMetadata) Exiting() bool {
	return m.Status.IsExited() || m.Status.HasExited()
}

// Slashed returns true if the validator is existing or exited due to slashing
func (m *ValidatorMetadata) Slashed() bool {
	return m.Status == eth2apiv1.ValidatorStateExitedSlashed || m.Status == eth2apiv1.ValidatorStateActiveSlashed
}

// OnUpdated represents a function to be called once validator's metadata was updated
type OnUpdated func(logger *zap.Logger, pk string, meta *ValidatorMetadata)

// UpdateValidatorsMetadata updates validator information for the given public keys
func UpdateValidatorsMetadata(logger *zap.Logger, pubKeys [][]byte, collection ValidatorMetadataStorage, bc Beacon, onUpdated OnUpdated) error {
	results, err := FetchValidatorsMetadata(bc, pubKeys)
	if err != nil {
		return errors.Wrap(err, "failed to get validator data from Beacon")
	}

	// TODO: importing logging/fields causes import cycle
	logger.Debug("🆕 got validators metadata", zap.Int("requested", len(pubKeys)),
		zap.Int("received", len(results)))

	var errs []error
	for pk, meta := range results {
		if err := collection.UpdateValidatorMetadata(logger, pk, meta); err != nil {
			logger.Error("❗ failed to update validator metadata",
				zap.String("validator", pk), zap.Error(err))
			errs = append(errs, err)
		}
		if onUpdated != nil {
			onUpdated(logger, pk, meta)
		}
		logger.Debug("💾️ successfully updated validator metadata",
			zap.String("pk", pk), zap.Any("metadata", meta))
	}
	if len(errs) > 0 {
		logger.Error("❌ failed to process validators returned from Beacon node",
			zap.Int("count", len(errs)), zap.Errors("errors", errs))
		return errors.Errorf("could not process %d validators returned from beacon", len(errs))
	}

	return nil
}

// FetchValidatorsMetadata is fetching validators data from beacon
func FetchValidatorsMetadata(bc Beacon, pubKeys [][]byte) (map[string]*ValidatorMetadata, error) {
	if len(pubKeys) == 0 {
		return nil, nil
	}
	var pubkeys []phase0.BLSPubKey
	for _, pk := range pubKeys {
		blsPubKey := phase0.BLSPubKey{}
		copy(blsPubKey[:], pk)
		pubkeys = append(pubkeys, blsPubKey)
	}
	validatorsIndexMap, err := bc.GetValidatorData(pubkeys)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get validators data from beacon")
	}
	ret := make(map[string]*ValidatorMetadata)
	for _, v := range validatorsIndexMap {
		pk := strings.ToLower(hex.EncodeToString(v.Validator.PublicKey[:]))
		if strings.Contains(
			"a311dcc81b1bf3c824bf1c3ceffca41af76b7ffd8c404cd0e46318153cf134519de94e9295bd35532afceffbc4c57e3e,8dad3498ba4b28b951f81e79a60a5c787e72394cb1ec2fae585dad01ae3b9e21e6a969bc77001a1367a26c31810387c5,86c953c3e43ee32fd370b61245fdc99a1bf614217028f724f83bf3c5870e62a459a03cab371ba46c3d338a6d70050cab",
			pk,
		) {
			if time.Now().Before(time.Unix(1682255157+(3*32*12), 0)) { // 3 epochs from now
				// Fake non deposited validator!
				continue
			}
		}
		meta := &ValidatorMetadata{
			Balance: v.Balance,
			Status:  v.Status,
			Index:   v.Index,
		}
		ret[pk] = meta
	}
	return ret, nil
}

// UpdateValidatorsMetadataBatch updates the given public keys in batches
func UpdateValidatorsMetadataBatch(logger *zap.Logger,
	pubKeys [][]byte,
	queue queue.Queue,
	collection ValidatorMetadataStorage,
	bc Beacon,
	onUpdated OnUpdated,
	batchSize int) {
	batch(pubKeys, queue, func(pks [][]byte) func() error {
		return func() error {
			return UpdateValidatorsMetadata(logger, pks, collection, bc, onUpdated)
		}
	}, batchSize)
}

type batchTask func(pks [][]byte) func() error

func batch(pubKeys [][]byte, queue queue.Queue, task batchTask, batchSize int) {
	n := float64(len(pubKeys))
	// in case the amount of public keys is lower than the batch size
	batchSize = int(math.Min(n, float64(batchSize)))
	batches := int(math.Ceil(n / float64(batchSize)))
	start := 0
	end := int(math.Min(n, float64(batchSize)))

	for i := 0; i < batches; i++ {
		// run task
		queue.Queue(task(pubKeys[start:end]))
		// reset start and end
		start = end
		end = int(math.Min(n, float64(start+batchSize)))
	}
}
