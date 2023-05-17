package fee_recipient

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/operator/slot_ticker"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v2/types"
	"github.com/bloxapp/ssv/registry/storage"
)

//go:generate mockgen -package=mocks -destination=./mocks/controller.go -source=./controller.go

// RecipientController submit proposal preparation to beacon node for all committee validators
type RecipientController interface {
	Start(logger *zap.Logger)
}

// ControllerOptions holds the needed dependencies
type ControllerOptions struct {
	Ctx              context.Context
	BeaconClient     beaconprotocol.Beacon
	EthNetwork       beaconprotocol.Network
	ShareStorage     storage.Shares
	RecipientStorage storage.Recipients
	Ticker           slot_ticker.Ticker
	OperatorData     *storage.OperatorData
}

// recipientController implementation of RecipientController
type recipientController struct {
	ctx              context.Context
	beaconClient     beaconprotocol.Beacon
	ethNetwork       beaconprotocol.Network
	shareStorage     storage.Shares
	recipientStorage storage.Recipients
	ticker           slot_ticker.Ticker
	operatorData     *storage.OperatorData
}

func NewController(opts *ControllerOptions) *recipientController {
	return &recipientController{
		ctx:              opts.Ctx,
		beaconClient:     opts.BeaconClient,
		ethNetwork:       opts.EthNetwork,
		shareStorage:     opts.ShareStorage,
		recipientStorage: opts.RecipientStorage,
		ticker:           opts.Ticker,
		operatorData:     opts.OperatorData,
	}
}

func (rc *recipientController) Start(logger *zap.Logger) {
	tickerChan := make(chan phase0.Slot, 32)
	rc.ticker.Subscribe(tickerChan)
	rc.listenToTicker(logger, tickerChan)
}

// listenToTicker loop over the given slot channel
// TODO: re-think this logic, we can use validator map instead of iterating over all shares
// in addition, submitting "same data" every slot is not efficient and can overload beacon node
// instead we can subscribe to beacon node events and submit only when there is
// a new fee recipient event (or new validator) was handled or when there is a syncing issue with beacon node
func (rc *recipientController) listenToTicker(logger *zap.Logger, slots chan phase0.Slot) {
	firstTimeSubmitted := false
	for currentSlot := range slots {
		// submit if first time or if first slot in epoch
		if firstTimeSubmitted && uint64(currentSlot)%rc.ethNetwork.SlotsPerEpoch != 0 {
			continue
		}
		firstTimeSubmitted = true

		err := rc.prepareAndSubmit(logger, currentSlot)
		if err != nil {
			logger.Warn("could not submit proposal preparations", zap.Error(err))
		}
	}
}

func (rc *recipientController) prepareAndSubmit(logger *zap.Logger, slot phase0.Slot) error {
	shares, err := rc.shareStorage.GetFilteredShares(logger, storage.ByOperatorIDAndActive(rc.operatorData.ID))
	if err != nil {
		return errors.Wrap(err, "could not get shares")
	}

	const batchSize = 500
	var submitted int
	for start := 0; start < len(shares); start += batchSize {
		end := start + batchSize
		if end > len(shares) {
			end = len(shares)
		}
		batch := shares[start:end]

		count, err := rc.submit(logger, batch)
		if err != nil {
			logger.Warn("could not submit proposal preparation batch",
				zap.Int("start_index", start),
				zap.Error(err),
			)
			continue
		}
		submitted += count
	}

	logger.Debug("✅ successfully submitted proposal preparations",
		zap.Int("submitted", submitted),
		zap.Int("total", len(shares)),
	)
	return nil
}

func (rc *recipientController) submit(logger *zap.Logger, shares []*types.SSVShare) (int, error) {
	m, err := rc.toProposalPreparation(logger, shares)
	if err != nil {
		return 0, errors.Wrap(err, "could not build proposal preparation batch")
	}
	err = rc.beaconClient.SubmitProposalPreparation(m)
	if err != nil {
		return 0, errors.Wrap(err, "could not submit proposal preparation batch")
	}
	return len(m), nil
}

func (rc *recipientController) toProposalPreparation(logger *zap.Logger, shares []*types.SSVShare) (map[phase0.ValidatorIndex]bellatrix.ExecutionAddress, error) {
	// build unique owners
	keys := make(map[common.Address]bool)
	var uniq []common.Address
	for _, entry := range shares {
		if _, value := keys[entry.OwnerAddress]; !value {
			keys[entry.OwnerAddress] = true
			uniq = append(uniq, entry.OwnerAddress)
		}
	}

	// get recipients
	rds, err := rc.recipientStorage.GetRecipientDataMany(logger, uniq)
	if err != nil {
		return nil, errors.Wrap(err, "could not get recipients data")
	}

	// build proposal preparation
	m := make(map[phase0.ValidatorIndex]bellatrix.ExecutionAddress)
	for _, share := range shares {
		feeRecipient, found := rds[share.OwnerAddress]
		if !found {
			copy(feeRecipient[:], share.OwnerAddress.Bytes())
		}
		m[share.BeaconMetadata.Index] = feeRecipient
	}

	return m, nil
}
