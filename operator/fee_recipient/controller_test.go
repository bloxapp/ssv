package fee_recipient

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v4/async/event"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/networkconfig"
	"github.com/bloxapp/ssv/operator/slot_ticker/mocks"
	"github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v2/types"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
)

func TestSubmitProposal(t *testing.T) {
	logger := logging.TestLogger(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	operatorData := &registrystorage.OperatorData{
		ID: 123456789,
	}

	db, shareStorage, recipientStorage := createStorage(t)
	defer db.Close()
	network := networkconfig.TestNetwork
	populateStorage(t, logger, shareStorage, operatorData)

	frCtrl := NewController(&ControllerOptions{
		Ctx:              context.TODO(),
		Network:          network,
		ShareStorage:     shareStorage,
		RecipientStorage: recipientStorage,
		OperatorData:     operatorData,
	})

	t.Run("submit first time or halfway through epoch", func(t *testing.T) {
		numberOfRequests := 4
		var wg sync.WaitGroup
		client := beacon.NewMockBeaconNode(ctrl)
		client.EXPECT().SubmitProposalPreparation(gomock.Any()).DoAndReturn(func(feeRecipients map[phase0.ValidatorIndex]bellatrix.ExecutionAddress) error {
			wg.Done()
			return nil
		}).MinTimes(numberOfRequests).MaxTimes(numberOfRequests) // call first time and on the halfway through epoch. each time should be 2 request as we have two batches

		ticker := mocks.NewMockTicker(ctrl)
		ticker.EXPECT().Subscribe(gomock.Any()).DoAndReturn(func(subscription chan phase0.Slot) event.Subscription {
			subscription <- 1 // first time
			time.Sleep(time.Millisecond * 500)
			subscription <- 2 // should not call submit
			time.Sleep(time.Millisecond * 500)
			subscription <- 20 // should not call submit
			time.Sleep(time.Millisecond * 500)
			subscription <- phase0.Slot(network.SlotsPerEpoch()) / 2 // halfway through epoch
			time.Sleep(time.Millisecond * 500)
			subscription <- 63 // should not call submit
			return nil
		})

		frCtrl.beaconClient = client
		frCtrl.ticker = ticker

		go frCtrl.Start(logger)
		wg.Add(numberOfRequests)
		wg.Wait()
	})

	t.Run("error handling", func(t *testing.T) {
		var wg sync.WaitGroup
		client := beacon.NewMockBeaconNode(ctrl)
		client.EXPECT().SubmitProposalPreparation(gomock.Any()).DoAndReturn(func(feeRecipients map[phase0.ValidatorIndex]bellatrix.ExecutionAddress) error {
			wg.Done()
			return errors.New("failed to submit")
		}).MinTimes(2).MaxTimes(2)

		ticker := mocks.NewMockTicker(ctrl)
		ticker.EXPECT().Subscribe(gomock.Any()).DoAndReturn(func(subscription chan phase0.Slot) event.Subscription {
			subscription <- 100 // first time
			return nil
		})

		frCtrl.beaconClient = client
		frCtrl.ticker = ticker

		go frCtrl.Start(logger)
		wg.Add(2)
		wg.Wait()
	})
}

func createStorage(t *testing.T) (basedb.IDb, registrystorage.Shares, registrystorage.Recipients) {
	logger := logging.TestLogger(t)
	options := basedb.Options{
		Type: "badger-memory",
		Path: "",
	}

	db, err := storage.GetStorageFactory(logger, options)
	require.NoError(t, err)

	shareStorage, err := registrystorage.NewSharesStorage(logger, db, []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	return db, shareStorage, registrystorage.NewRecipientsStorage(logger, db, []byte("test"))
}

func populateStorage(t *testing.T, logger *zap.Logger, storage registrystorage.Shares, operatorData *registrystorage.OperatorData) {
	createShare := func(index int, operatorID spectypes.OperatorID) *types.SSVShare {
		ownerAddr := fmt.Sprintf("%d", index)
		ownerAddrByte := [20]byte{}
		copy(ownerAddrByte[:], ownerAddr)

		return &types.SSVShare{
			Share: spectypes.Share{ValidatorPubKey: []byte(fmt.Sprintf("pk%d", index)), OperatorID: operatorID},
			Metadata: types.Metadata{
				BeaconMetadata: &beacon.ValidatorMetadata{
					Index: phase0.ValidatorIndex(index),
				},
				OwnerAddress: common.BytesToAddress(ownerAddrByte[:]),
				Liquidated:   false,
			},
		}
	}

	for i := 0; i < 1000; i++ {
		require.NoError(t, storage.Save(nil, createShare(i, operatorData.ID)))
	}

	// add none committee share
	require.NoError(t, storage.Save(nil, createShare(2000, spectypes.OperatorID(1))))

	all := storage.List(nil, registrystorage.ByOperatorID(operatorData.ID), registrystorage.ByNotLiquidated())
	require.Equal(t, 1000, len(all))
}
