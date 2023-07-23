package duties

import (
	"context"
	"fmt"
	"time"

	eth2apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/logging/fields"
)

type ProposerHandler struct {
	baseHandler

	duties *Duties[*eth2apiv1.ProposerDuty]
}

func NewProposerHandler() *ProposerHandler {
	return &ProposerHandler{
		duties: NewDuties[*eth2apiv1.ProposerDuty](),
		baseHandler: baseHandler{
			fetchFirst: true,
		},
	}
}

func (h *ProposerHandler) Name() string {
	return spectypes.BNRoleProposer.String()
}

// HandleDuties manages the duty lifecycle, handling different cases:
//
// On First Run:
//  1. Fetch duties for the current epoch.
//  2. Execute duties.
//
// On Re-org (current dependent root changed):
//  1. Fetch duties for the current epoch.
//  2. Execute duties.
//
// On Indices Change:
//  1. Execute duties.
//  2. Reset duties for the current epoch.
//  3. Fetch duties for the current epoch.
//
// On Ticker event:
//  1. Execute duties.
//  2. If necessary, fetch duties for the current epoch.
func (h *ProposerHandler) HandleDuties(ctx context.Context) {
	h.logger.Info("starting duty handler")

	for {
		select {
		case <-ctx.Done():
			return

		case slot := <-h.ticker:
			currentEpoch := h.network.Beacon.EstimatedEpochAtSlot(slot)
			buildStr := fmt.Sprintf("e%v-s%v-#%v", currentEpoch, slot, slot%32+1)
			h.logger.Debug("🛠 ticker event", zap.String("epoch_slot_seq", buildStr))

			if h.fetchFirst {
				h.fetchFirst = false
				h.processFetching(ctx, currentEpoch, slot)
				h.processExecution(currentEpoch, slot)
			} else {
				h.processExecution(currentEpoch, slot)
				if h.indicesChanged {
					h.duties.Reset(currentEpoch)
					h.indicesChanged = false
					h.processFetching(ctx, currentEpoch, slot)
				}
			}

			// last slot of epoch
			if uint64(slot)%h.network.Beacon.SlotsPerEpoch() == h.network.Beacon.SlotsPerEpoch()-1 {
				h.duties.Reset(currentEpoch)
				h.fetchFirst = true
			}

		case reorgEvent := <-h.reorg:
			currentEpoch := h.network.Beacon.EstimatedEpochAtSlot(reorgEvent.Slot)
			buildStr := fmt.Sprintf("e%v-s%v-#%v", currentEpoch, reorgEvent.Slot, reorgEvent.Slot%32+1)
			h.logger.Info("🔀 reorg event received", zap.String("epoch_slot_seq", buildStr), zap.Any("event", reorgEvent))

			// reset current epoch duties
			if reorgEvent.Current {
				h.duties.Reset(currentEpoch)
				h.fetchFirst = true
			}

		case <-h.indicesChange:
			slot := h.network.Beacon.EstimatedCurrentSlot()
			currentEpoch := h.network.Beacon.EstimatedEpochAtSlot(slot)
			buildStr := fmt.Sprintf("e%v-s%v-#%v", currentEpoch, slot, slot%32+1)
			h.logger.Info("🔁 indices change received", zap.String("epoch_slot_seq", buildStr))

			h.indicesChanged = true
		}
	}
}

func (h *ProposerHandler) processFetching(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) {
	ctx, cancel := context.WithDeadline(ctx, h.network.Beacon.GetSlotStartTime(slot+1).Add(100*time.Millisecond))
	defer cancel()

	if err := h.fetchDuties(ctx, epoch); err != nil {
		h.logger.Error("failed to fetch duties for current epoch", zap.Error(err))
		return
	}
}

func (h *ProposerHandler) processExecution(epoch phase0.Epoch, slot phase0.Slot) {
	// range over duties and execute
	if slotMap, ok := h.duties.m[epoch]; ok {
		if duties, ok := slotMap[slot]; ok {
			toExecute := make([]*spectypes.Duty, 0, len(duties))
			for _, d := range duties {
				if h.shouldExecute(d) {
					toExecute = append(toExecute, h.toSpecDuty(d, spectypes.BNRoleProposer))
				}
			}
			h.executeDuties(h.logger, toExecute)
		}
	}
}

func (h *ProposerHandler) fetchDuties(ctx context.Context, epoch phase0.Epoch) error {
	start := time.Now()
	indices := h.validatorController.ActiveValidatorIndices(h.logger, epoch)
	duties, err := h.beaconNode.ProposerDuties(ctx, epoch, indices)
	if err != nil {
		return fmt.Errorf("failed to fetch proposer duties: %w", err)
	}

	specDuties := make([]*spectypes.Duty, 0, len(duties))
	for _, d := range duties {
		h.duties.Add(epoch, d.Slot, d)
		specDuties = append(specDuties, h.toSpecDuty(d, spectypes.BNRoleProposer))
	}

	h.logger.Debug("📚 got duties",
		fields.Count(len(duties)),
		fields.Epoch(epoch),
		fields.Duties(epoch, specDuties),
		fields.Duration(start))

	return nil
}

func (h *ProposerHandler) toSpecDuty(duty *eth2apiv1.ProposerDuty, role spectypes.BeaconRole) *spectypes.Duty {
	return &spectypes.Duty{
		Type:           role,
		PubKey:         duty.PubKey,
		Slot:           duty.Slot,
		ValidatorIndex: duty.ValidatorIndex,
	}
}

func (h *ProposerHandler) shouldExecute(duty *eth2apiv1.ProposerDuty) bool {
	currentSlot := h.network.Beacon.EstimatedCurrentSlot()
	// execute task if slot already began and not pass 1 slot
	if currentSlot == duty.Slot {
		return true
	}
	if currentSlot+1 == duty.Slot {
		h.logger.Debug("current slot and duty slot are not aligned, "+
			"assuming diff caused by a time drift - ignoring and executing duty", zap.String("type", duty.String()))
		return true
	}
	return false
}
