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

type AttesterHandler struct {
	baseHandler

	duties            *Duties[*eth2apiv1.AttesterDuty]
	fetchCurrentEpoch bool
	fetchNextEpoch    bool
}

func NewAttesterHandler() *AttesterHandler {
	h := &AttesterHandler{
		duties: NewDuties[*eth2apiv1.AttesterDuty](),
	}
	h.fetchCurrentEpoch = true
	h.fetchFirst = true
	return h
}

func (h *AttesterHandler) Name() string {
	return spectypes.BNRoleAttester.String()
}

// HandleDuties manages the duty lifecycle, handling different cases:
//
// On First Run:
//  1. Fetch duties for the current epoch.
//  2. If necessary, fetch duties for the next epoch.
//  3. Execute duties.
//
// On Re-org:
//
//	If the previous dependent root changed:
//	    1. Fetch duties for the current epoch.
//	    2. Execute duties.
//	If the current dependent root changed:
//	    1. Execute duties.
//	    2. If necessary, fetch duties for the next epoch.
//
// On Indices Change:
//  1. Execute duties.
//  2. Reset duties for the current epoch.
//  3. Fetch duties for the current epoch.
//  4. If necessary, fetch duties for the next epoch.
//
// On Ticker event:
//  1. Execute duties.
//  2. If necessary, fetch duties for the next epoch.
func (h *AttesterHandler) HandleDuties(ctx context.Context) {
	h.logger.Info("starting duty handler")

	h.fetchNextEpoch = true

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
				}
				h.processFetching(ctx, currentEpoch, slot)
			}

			slotsPerEpoch := h.network.Beacon.SlotsPerEpoch()

			// If we have reached the mid-point of the epoch, fetch the duties for the next epoch in the next slot.
			// This allows us to set them up at a time when the beacon node should be less busy.
			if uint64(slot)%slotsPerEpoch == slotsPerEpoch/2-2 {
				h.fetchNextEpoch = true
			}

			// last slot of epoch
			if uint64(slot)%slotsPerEpoch == slotsPerEpoch-1 {
				h.duties.Reset(currentEpoch)
			}

		case reorgEvent := <-h.reorg:
			currentEpoch := h.network.Beacon.EstimatedEpochAtSlot(reorgEvent.Slot)
			buildStr := fmt.Sprintf("e%v-s%v-#%v", currentEpoch, reorgEvent.Slot, reorgEvent.Slot%32+1)
			h.logger.Info("🔀 reorg event received", zap.String("epoch_slot_seq", buildStr), zap.Any("event", reorgEvent))

			// reset current epoch duties
			if reorgEvent.Previous {
				h.duties.Reset(currentEpoch)
				h.fetchFirst = true
				h.fetchCurrentEpoch = true
			} else if reorgEvent.Current {
				// reset & re-fetch next epoch duties if in appropriate slot range,
				// otherwise they will be fetched by the appropriate slot tick.
				if h.shouldFetchNexEpoch(reorgEvent.Slot) {
					h.duties.Reset(currentEpoch + 1)
					h.fetchNextEpoch = true
				}
			}

		case <-h.indicesChange:
			slot := h.network.Beacon.EstimatedCurrentSlot()
			currentEpoch := h.network.Beacon.EstimatedEpochAtSlot(slot)
			buildStr := fmt.Sprintf("e%v-s%v-#%v", currentEpoch, slot, slot%32+1)
			h.logger.Info("🔁 indices change received", zap.String("epoch_slot_seq", buildStr))

			h.indicesChanged = true
			h.fetchCurrentEpoch = true

			// reset next epoch duties if in appropriate slot range
			if h.shouldFetchNexEpoch(slot) {
				h.duties.Reset(currentEpoch + 1)
				h.fetchNextEpoch = true
			}
		}
	}
}

func (h *AttesterHandler) processFetching(ctx context.Context, epoch phase0.Epoch, slot phase0.Slot) {
	ctx, cancel := context.WithDeadline(ctx, h.network.Beacon.GetSlotStartTime(slot+1).Add(100*time.Millisecond))
	defer cancel()

	if h.fetchCurrentEpoch {
		if err := h.fetchAndProcessDuties(ctx, epoch); err != nil {
			h.logger.Error("failed to fetch duties for current epoch", zap.Error(err))
			return
		}
		h.fetchCurrentEpoch = false
	}

	if h.fetchNextEpoch && h.shouldFetchNexEpoch(slot) {
		if err := h.fetchAndProcessDuties(ctx, epoch+1); err != nil {
			h.logger.Error("failed to fetch duties for next epoch", zap.Error(err))
			return
		}
		h.fetchNextEpoch = false
	}
}

func (h *AttesterHandler) processExecution(epoch phase0.Epoch, slot phase0.Slot) {
	// range over duties and execute
	if slotMap, ok := h.duties.m[epoch]; ok {
		if duties, ok := slotMap[slot]; ok {
			toExecute := make([]*spectypes.Duty, 0, len(duties)*2)
			for _, d := range duties {
				if h.shouldExecute(d) {
					toExecute = append(toExecute, h.toSpecDuty(d, spectypes.BNRoleAttester))
					toExecute = append(toExecute, h.toSpecDuty(d, spectypes.BNRoleAggregator))
				}
			}
			h.executeDuties(h.logger, toExecute)
		}
	}
}

func (h *AttesterHandler) fetchAndProcessDuties(ctx context.Context, epoch phase0.Epoch) error {
	start := time.Now()
	indices := h.validatorController.ActiveValidatorIndices(h.logger, epoch)
	duties, err := h.beaconNode.AttesterDuties(ctx, epoch, indices)
	if err != nil {
		return fmt.Errorf("failed to fetch attester duties: %w", err)
	}

	specDuties := make([]*spectypes.Duty, 0, len(duties))
	for _, d := range duties {
		h.duties.Add(epoch, d.Slot, d)
		specDuties = append(specDuties, h.toSpecDuty(d, spectypes.BNRoleAttester))
	}

	h.logger.Debug("🗂 got duties",
		fields.Count(len(duties)),
		fields.Epoch(epoch),
		fields.Duties(epoch, specDuties),
		fields.Duration(start))

	// calculate subscriptions
	subscriptions := calculateSubscriptionInfo(duties)
	if len(subscriptions) > 0 {
		if deadline, ok := ctx.Deadline(); ok {
			go func(h *AttesterHandler, subscriptions []*eth2apiv1.BeaconCommitteeSubscription) {
				// Create a new subscription context with a deadline from parent context.
				subscriptionCtx, cancel := context.WithDeadline(context.Background(), deadline)
				defer cancel()
				if err := h.beaconNode.SubmitBeaconCommitteeSubscriptions(subscriptionCtx, subscriptions); err != nil {
					h.logger.Warn("failed to submit beacon committee subscription", zap.Error(err))
				}
			}(h, subscriptions)
		} else {
			h.logger.Warn("failed to get context deadline")
		}
	}

	return nil
}

func (h *AttesterHandler) toSpecDuty(duty *eth2apiv1.AttesterDuty, role spectypes.BeaconRole) *spectypes.Duty {
	return &spectypes.Duty{
		Type:                    role,
		PubKey:                  duty.PubKey,
		Slot:                    duty.Slot,
		ValidatorIndex:          duty.ValidatorIndex,
		CommitteeIndex:          duty.CommitteeIndex,
		CommitteeLength:         duty.CommitteeLength,
		CommitteesAtSlot:        duty.CommitteesAtSlot,
		ValidatorCommitteeIndex: duty.ValidatorCommitteeIndex,
	}
}

func (h *AttesterHandler) shouldExecute(duty *eth2apiv1.AttesterDuty) bool {
	currentSlot := h.network.Beacon.EstimatedCurrentSlot()
	// execute task if slot already began and not pass 1 epoch
	var attestationPropagationSlotRange = phase0.Slot(h.network.Beacon.SlotsPerEpoch())
	if currentSlot >= duty.Slot && currentSlot-duty.Slot <= attestationPropagationSlotRange {
		return true
	}
	if currentSlot+1 == duty.Slot {
		h.logger.Debug("current slot and duty slot are not aligned, "+
			"assuming diff caused by a time drift - ignoring and executing duty", zap.String("type", duty.String()))
		return true
	}
	return false
}

// calculateSubscriptionInfo calculates the attester subscriptions given a set of duties.
func calculateSubscriptionInfo(duties []*eth2apiv1.AttesterDuty) []*eth2apiv1.BeaconCommitteeSubscription {
	subscriptions := make([]*eth2apiv1.BeaconCommitteeSubscription, 0, len(duties)*2)
	for _, duty := range duties {
		// Append a subscription for the attester role
		subscriptions = append(subscriptions, toBeaconCommitteeSubscription(duty, spectypes.BNRoleAttester))
		// Append a subscription for the aggregator role
		subscriptions = append(subscriptions, toBeaconCommitteeSubscription(duty, spectypes.BNRoleAggregator))
	}
	return subscriptions
}

func toBeaconCommitteeSubscription(duty *eth2apiv1.AttesterDuty, role spectypes.BeaconRole) *eth2apiv1.BeaconCommitteeSubscription {
	return &eth2apiv1.BeaconCommitteeSubscription{
		ValidatorIndex:   duty.ValidatorIndex,
		Slot:             duty.Slot,
		CommitteeIndex:   duty.CommitteeIndex,
		CommitteesAtSlot: duty.CommitteesAtSlot,
		IsAggregator:     role == spectypes.BNRoleAggregator,
	}
}

func (h *AttesterHandler) shouldFetchNexEpoch(slot phase0.Slot) bool {
	return uint64(slot)%h.network.Beacon.SlotsPerEpoch() > h.network.Beacon.SlotsPerEpoch()/2-2
}
