package beacon

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
)

// Network is a beacon chain network.
type Network struct {
	spectypes.BeaconNetwork
	minGenesisTime uint64
}

// NewNetwork creates a new beacon chain network.
func NewNetwork(network spectypes.BeaconNetwork, minGenesisTime uint64) Network {
	return Network{network, minGenesisTime}
}

// GetSlotStartTime returns the start time for the given slot
func (n Network) GetSlotStartTime(slot phase0.Slot) time.Time {
	timeSinceGenesisStart := uint64(slot) * uint64(n.SlotDuration.Seconds())
	start := time.Unix(int64(n.MinGenesisTime()+timeSinceGenesisStart), 0)
	return start
}

func (n Network) MinGenesisTime() uint64 {
	if n.minGenesisTime > 0 {
		return n.minGenesisTime
	} else {
		return n.BeaconNetwork.MinGenesisTime
	}
}

// EstimatedCurrentSlot returns the estimation of the current slot
func (n Network) EstimatedCurrentSlot() phase0.Slot {
	return n.EstimatedSlotAtTime(time.Now().Unix())
}

// EstimatedSlotAtTime estimates slot at the given time
func (n Network) EstimatedSlotAtTime(time int64) phase0.Slot {
	genesis := int64(n.MinGenesisTime())
	if time < genesis {
		return 0
	}
	return phase0.Slot(uint64(time-genesis) / uint64(n.SlotDuration.Seconds()))
}

// EstimatedCurrentEpoch estimates the current epoch
// https://github.com/ethereum/eth2.0-specs/blob/dev/specs/phase0/beacon-chain.md#compute_start_slot_at_epoch
func (n Network) EstimatedCurrentEpoch() phase0.Epoch {
	return n.EstimatedEpochAtSlot(n.EstimatedCurrentSlot())
}

// EstimatedEpochAtSlot estimates epoch at the given slot
func (n Network) EstimatedEpochAtSlot(slot phase0.Slot) phase0.Epoch {
	return phase0.Epoch(slot / phase0.Slot(n.SlotsPerEpoch))
}

// IsFirstSlotOfEpoch estimates epoch at the given slot
func (n Network) IsFirstSlotOfEpoch(slot phase0.Slot) bool {
	return uint64(slot)%n.SlotsPerEpoch == 0
}

// GetEpochFirstSlot returns the beacon node first slot in epoch
func (n Network) GetEpochFirstSlot(epoch phase0.Epoch) phase0.Slot {
	return phase0.Slot(epoch * 32)
}

// EstimatedSyncCommitteePeriodAtEpoch estimates the current sync committee period at the given Epoch
func (n Network) EstimatedSyncCommitteePeriodAtEpoch(epoch phase0.Epoch) uint64 {
	return uint64(epoch) / 256 // EpochsPerSyncCommitteePeriod
}

// FirstEpochOfSyncPeriod calculates the first epoch of the given sync period.
func (n Network) FirstEpochOfSyncPeriod(period uint64) phase0.Epoch {
	return phase0.Epoch(period * 256) // EpochsPerSyncCommitteePeriod
}

// LastSlotOfSyncPeriod calculates the first epoch of the given sync period.
func (n Network) LastSlotOfSyncPeriod(period uint64) phase0.Slot {
	lastEpoch := n.FirstEpochOfSyncPeriod(period+1) - 1
	// If we are in the sync committee that ends at slot x we do not generate a message during slot x-1
	// as it will never be included, hence -1.
	return n.GetEpochFirstSlot(lastEpoch+1) - 2
}
