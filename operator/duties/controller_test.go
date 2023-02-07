package duties

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/eth2-key-manager/core"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/cornelk/hashmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/operator/duties/mocks"
	"github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
)

func TestDutyController_ListenToTicker(t *testing.T) {
	var wg sync.WaitGroup

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockExecutor := mocks.NewMockDutyExecutor(mockCtrl)
	mockExecutor.EXPECT().ExecuteDuty(gomock.Any()).DoAndReturn(func(duty *spectypes.Duty) error {
		require.NotNil(t, duty)
		require.True(t, duty.Slot > 0)
		wg.Done()
		return nil
	}).AnyTimes()

	mockFetcher := mocks.NewMockDutyFetcher(mockCtrl)
	mockFetcher.EXPECT().GetDuties(gomock.Any()).DoAndReturn(func(slot phase0.Slot) ([]spectypes.Duty, error) {
		return []spectypes.Duty{{Slot: slot, PubKey: phase0.BLSPubKey{}}}, nil
	}).AnyTimes()

	dutyCtrl := &dutyController{
		logger: zap.L(), ctx: context.Background(), ethNetwork: beacon.NewNetwork(core.PraterNetwork, 0),
		executor:               mockExecutor,
		fetcher:                mockFetcher,
		syncCommitteeDutiesMap: hashmap.New[phase0.Slot, []*spectypes.Duty](),
	}

	cn := make(chan phase0.Slot)

	secPerSlot = 2
	defer func() {
		secPerSlot = 12
	}()

	currentSlot := dutyCtrl.ethNetwork.EstimatedCurrentSlot()

	go dutyCtrl.listenToTicker(cn)
	wg.Add(2)
	go func() {
		cn <- currentSlot
		time.Sleep(time.Second * time.Duration(secPerSlot))
		cn <- currentSlot + 1
	}()

	wg.Wait()
}

func TestDutyController_ShouldExecute(t *testing.T) {
	ctrl := dutyController{logger: zap.L(), ethNetwork: beacon.NewNetwork(core.PraterNetwork, 0)}
	currentSlot := uint64(ctrl.ethNetwork.EstimatedCurrentSlot())

	require.True(t, ctrl.shouldExecute(&spectypes.Duty{Slot: phase0.Slot(currentSlot), PubKey: phase0.BLSPubKey{}}))
	require.False(t, ctrl.shouldExecute(&spectypes.Duty{Slot: phase0.Slot(currentSlot - 1000), PubKey: phase0.BLSPubKey{}}))
	require.False(t, ctrl.shouldExecute(&spectypes.Duty{Slot: phase0.Slot(currentSlot + 1000), PubKey: phase0.BLSPubKey{}}))
}

func TestDutyController_GetSlotStartTime(t *testing.T) {
	d := dutyController{logger: zap.L(), ethNetwork: beacon.NewNetwork(core.PraterNetwork, 0)}

	ts := d.ethNetwork.GetSlotStartTime(646523)
	require.Equal(t, int64(1624266276), ts.Unix())
}

func TestDutyController_GetCurrentSlot(t *testing.T) {
	d := dutyController{logger: zap.L(), ethNetwork: beacon.NewNetwork(core.PraterNetwork, 0)}

	slot := d.ethNetwork.EstimatedCurrentSlot()
	require.Greater(t, slot, phase0.Slot(646855))
}

func TestDutyController_GetEpochFirstSlot(t *testing.T) {
	d := dutyController{logger: zap.L(), ethNetwork: beacon.NewNetwork(core.PraterNetwork, 0)}

	slot := d.ethNetwork.GetEpochFirstSlot(20203)
	require.EqualValues(t, 646496, slot)
}
