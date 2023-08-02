package duties

import (
	"context"
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/cornelk/hashmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/operator/duties/mocks"
)

func setupProposerDutiesMock(s *Scheduler, dutiesMap *hashmap.Map[phase0.Epoch, []*v1.ProposerDuty]) (chan struct{}, chan []*spectypes.Duty) {
	fetchDutiesCall := make(chan struct{})
	executeDutiesCall := make(chan []*spectypes.Duty)

	s.beaconNode.(*mocks.MockBeaconNode).EXPECT().ProposerDuties(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, epoch phase0.Epoch, indices []phase0.ValidatorIndex) ([]*v1.ProposerDuty, error) {
			fetchDutiesCall <- struct{}{}
			duties, _ := dutiesMap.Get(epoch)
			return duties, nil
		}).AnyTimes()

	s.validatorController.(*mocks.MockValidatorController).EXPECT().ActiveValidatorIndices(gomock.Any(), gomock.Any()).DoAndReturn(
		func(logger *zap.Logger, epoch phase0.Epoch) []phase0.ValidatorIndex {
			uniqueIndices := make(map[phase0.ValidatorIndex]bool)

			duties, _ := dutiesMap.Get(epoch)
			for _, d := range duties {
				uniqueIndices[d.ValidatorIndex] = true
			}

			indices := make([]phase0.ValidatorIndex, 0, len(uniqueIndices))
			for index := range uniqueIndices {
				indices = append(indices, index)
			}

			return indices
		}).AnyTimes()

	return fetchDutiesCall, executeDutiesCall
}

func expectedExecutedProposerDuties(handler *ProposerHandler, duties []*v1.ProposerDuty) []*spectypes.Duty {
	expectedDuties := make([]*spectypes.Duty, 0)
	for _, d := range duties {
		expectedDuties = append(expectedDuties, handler.toSpecDuty(d, spectypes.BNRoleProposer))
	}
	return expectedDuties
}

func TestScheduler_Proposer_Same_Slot(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(0))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	dutiesMap.Set(phase0.Epoch(0), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(0),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})

	// STEP 1: wait for proposer duties to be fetched and executed at the same slot
	duties, _ := dutiesMap.Get(phase0.Epoch(0))
	expected := expectedExecutedProposerDuties(handler, duties)
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}

func TestScheduler_Proposer_Diff_Slots(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(0))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	dutiesMap.Set(phase0.Epoch(0), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(2),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})

	// STEP 1: wait for proposer duties to be fetched
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 2: wait for no action to be taken
	currentSlot.SetSlot(phase0.Slot(1))
	ticker.Send(currentSlot.GetSlot())
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 3: wait for proposer duties to be executed
	currentSlot.SetSlot(phase0.Slot(2))
	duties, _ := dutiesMap.Get(phase0.Epoch(0))
	expected := expectedExecutedProposerDuties(handler, duties)
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}

// execute duty after two slots after the indices changed
func TestScheduler_Proposer_Indices_Changed(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(0))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	// STEP 1: wait for no action to be taken
	ticker.Send(currentSlot.GetSlot())
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 2: wait for no action to be taken
	currentSlot.SetSlot(phase0.Slot(1))
	ticker.Send(currentSlot.GetSlot())
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 3: trigger a change in active indices
	scheduler.indicesChg <- struct{}{}
	dutiesMap.Set(phase0.Epoch(0), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(1),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
		{
			PubKey:         phase0.BLSPubKey{1, 2, 4},
			Slot:           phase0.Slot(2),
			ValidatorIndex: phase0.ValidatorIndex(2),
		},
		{
			PubKey:         phase0.BLSPubKey{1, 2, 5},
			Slot:           phase0.Slot(3),
			ValidatorIndex: phase0.ValidatorIndex(3),
		},
	})
	// no execution should happen in slot 1
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 4: wait for proposer duties to be fetched again
	currentSlot.SetSlot(phase0.Slot(2))
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)
	// no execution should happen in slot 2
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 4: wait for proposer duties to be executed
	currentSlot.SetSlot(phase0.Slot(3))
	duties, _ := dutiesMap.Get(phase0.Epoch(0))
	expected := expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[2]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}

func TestScheduler_Proposer_Multiple_Indices_Changed_Same_Slot(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(0))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	dutiesMap.Set(phase0.Epoch(0), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(2),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})

	// STEP 1: wait for proposer duties to be fetched
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 2: trigger a change in active indices
	scheduler.indicesChg <- struct{}{}
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)
	duties, _ := dutiesMap.Get(phase0.Epoch(0))
	dutiesMap.Set(phase0.Epoch(0), append(duties, &v1.ProposerDuty{
		PubKey:         phase0.BLSPubKey{1, 2, 4},
		Slot:           phase0.Slot(3),
		ValidatorIndex: phase0.ValidatorIndex(2),
	}))

	// STEP 3: trigger a change in active indices in the same slot
	scheduler.indicesChg <- struct{}{}
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)
	duties, _ = dutiesMap.Get(phase0.Epoch(0))
	dutiesMap.Set(phase0.Epoch(0), append(duties, &v1.ProposerDuty{
		PubKey:         phase0.BLSPubKey{1, 2, 5},
		Slot:           phase0.Slot(4),
		ValidatorIndex: phase0.ValidatorIndex(3),
	}))

	// STEP 4: wait for proposer duties to be fetched again
	currentSlot.SetSlot(phase0.Slot(1))
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 5: wait for proposer duties to be executed
	currentSlot.SetSlot(phase0.Slot(2))
	duties, _ = dutiesMap.Get(phase0.Epoch(0))
	expected := expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[0]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// STEP 6: wait for proposer duties to be executed
	currentSlot.SetSlot(phase0.Slot(3))
	duties, _ = dutiesMap.Get(phase0.Epoch(0))
	expected = expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[1]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// STEP 7: wait for proposer duties to be executed
	currentSlot.SetSlot(phase0.Slot(4))
	duties, _ = dutiesMap.Get(phase0.Epoch(0))
	expected = expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[2]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}

// reorg current dependent root changed
func TestScheduler_Proposer_Reorg_Current(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(34))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	dutiesMap.Set(phase0.Epoch(1), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(36),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})

	// STEP 1: wait for proposer duties to be fetched
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 2: trigger head event
	e := &v1.Event{
		Data: &v1.HeadEvent{
			Slot:                     currentSlot.GetSlot(),
			CurrentDutyDependentRoot: phase0.Root{0x01},
		},
	}
	scheduler.HandleHeadEvent(logger)(e)
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 3: Ticker with no action
	currentSlot.SetSlot(phase0.Slot(35))
	ticker.Send(currentSlot.GetSlot())
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 4: trigger reorg
	e = &v1.Event{
		Data: &v1.HeadEvent{
			Slot:                     currentSlot.GetSlot(),
			CurrentDutyDependentRoot: phase0.Root{0x02},
		},
	}
	dutiesMap.Set(phase0.Epoch(1), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(37),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})
	scheduler.HandleHeadEvent(logger)(e)
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 5: wait for proposer duties to be fetched again for the current epoch.
	// The first assigned duty should not be executed
	currentSlot.SetSlot(phase0.Slot(36))
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 7: The second assigned duty should be executed
	currentSlot.SetSlot(phase0.Slot(37))
	duties, _ := dutiesMap.Get(phase0.Epoch(1))
	expected := expectedExecutedProposerDuties(handler, duties)
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}

// reorg current dependent root changed
func TestScheduler_Proposer_Reorg_Current_Indices_Changed(t *testing.T) {
	var (
		handler     = NewProposerHandler()
		currentSlot = &SlotValue{}
		dutiesMap   = hashmap.New[phase0.Epoch, []*v1.ProposerDuty]()
	)
	currentSlot.SetSlot(phase0.Slot(34))
	scheduler, logger, ticker, timeout, cancel, schedulerPool := setupSchedulerAndMocks(t, handler, currentSlot)
	fetchDutiesCall, executeDutiesCall := setupProposerDutiesMock(scheduler, dutiesMap)

	dutiesMap.Set(phase0.Epoch(1), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(36),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})

	// STEP 1: wait for proposer duties to be fetched
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 2: trigger head event
	e := &v1.Event{
		Data: &v1.HeadEvent{
			Slot:                     currentSlot.GetSlot(),
			CurrentDutyDependentRoot: phase0.Root{0x01},
		},
	}
	scheduler.HandleHeadEvent(logger)(e)
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 3: Ticker with no action
	currentSlot.SetSlot(phase0.Slot(35))
	ticker.Send(currentSlot.GetSlot())
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 4: trigger reorg
	e = &v1.Event{
		Data: &v1.HeadEvent{
			Slot:                     currentSlot.GetSlot(),
			CurrentDutyDependentRoot: phase0.Root{0x02},
		},
	}
	dutiesMap.Set(phase0.Epoch(1), []*v1.ProposerDuty{
		{
			PubKey:         phase0.BLSPubKey{1, 2, 3},
			Slot:           phase0.Slot(37),
			ValidatorIndex: phase0.ValidatorIndex(1),
		},
	})
	scheduler.HandleHeadEvent(logger)(e)
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 5: trigger a change in active indices in the same slot
	scheduler.indicesChg <- struct{}{}
	duties, _ := dutiesMap.Get(phase0.Epoch(1))
	dutiesMap.Set(phase0.Epoch(1), append(duties, &v1.ProposerDuty{
		PubKey:         phase0.BLSPubKey{1, 2, 4},
		Slot:           phase0.Slot(38),
		ValidatorIndex: phase0.ValidatorIndex(2),
	}))
	waitForNoAction(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 6: wait for proposer duties to be fetched again for the current epoch.
	// The first assigned duty should not be executed
	currentSlot.SetSlot(phase0.Slot(36))
	ticker.Send(currentSlot.GetSlot())
	waitForDutiesFetch(t, logger, fetchDutiesCall, executeDutiesCall, timeout)

	// STEP 7: The second assigned duty should be executed
	currentSlot.SetSlot(phase0.Slot(37))
	duties, _ = dutiesMap.Get(phase0.Epoch(1))
	expected := expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[0]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// STEP 8: The second assigned duty should be executed
	currentSlot.SetSlot(phase0.Slot(38))
	duties, _ = dutiesMap.Get(phase0.Epoch(1))
	expected = expectedExecutedProposerDuties(handler, []*v1.ProposerDuty{duties[1]})
	setExecuteDutyFunc(scheduler, executeDutiesCall, len(expected))

	ticker.Send(currentSlot.GetSlot())
	waitForDutiesExecution(t, logger, fetchDutiesCall, executeDutiesCall, timeout, expected)

	// Stop scheduler & wait for graceful exit.
	cancel()
	require.NoError(t, schedulerPool.Wait())
}
