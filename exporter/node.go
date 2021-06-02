package exporter

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/ssv/eth1"
	"github.com/bloxapp/ssv/exporter/ibft"
	"github.com/bloxapp/ssv/exporter/storage"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/network"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/storage/collections"
	"github.com/bloxapp/ssv/validator"
	validatorstorage "github.com/bloxapp/ssv/validator/storage"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/big"
)

const (
	defaultOffset string = "49e08f"
)

// Exporter represents the main interface of this package
type Exporter interface {
	Start() error
	Sync() error
}

// Options contains options to create the node
type Options struct {
	APIPort int `yaml:"APIPort" env-default:"5001"`

	Logger     *zap.Logger
	ETHNetwork *core.Network

	Eth1Client eth1.Client

	Network network.Network

	DB basedb.IDb
}

// exporter is the internal implementation of Exporter interface
type exporter struct {
	store            storage.ExporterStorage
	validatorStorage validatorstorage.ICollection
	ibftStorage      collections.Iibft
	logger           *zap.Logger
	network          network.Network
	eth1Client       eth1.Client
	ibftDisptcher    ibft.Dispatcher

	httpHandlers apiHandlers
}

// New creates a new Exporter instance
func New(opts Options) Exporter {
	validatorStorage := validatorstorage.NewCollection(
		validatorstorage.CollectionOptions{
			DB:     opts.DB,
			Logger: opts.Logger,
		},
	)
	ibftStorage := collections.NewIbft(opts.DB, opts.Logger, "attestation")
	e := exporter{
		store:            storage.NewExporterStorage(opts.DB, opts.Logger),
		ibftStorage:      &ibftStorage,
		validatorStorage: validatorStorage,
		logger:           opts.Logger,
		network:          opts.Network,
		eth1Client:       opts.Eth1Client,
		ibftDisptcher:    ibft.NewDispatcher(context.TODO(), opts.Logger),
	}
	e.httpHandlers = newHTTPHandlers(&e, opts.APIPort)
	go e.HandleEth1Events()

	return &e
}

// Start starts the exporter
func (exp *exporter) Start() error {
	exp.logger.Info("exporter.Start()")
	//go func() {
	//	shares, err1 := exp.validatorStorage.GetAllValidatorsShare()
	//	if err1 != nil {
	//		exp.logger.Debug("could not read all validators shares", zap.Error(err1))
	//	}
	//	exp.logger.Debug("all validators shares", zap.Int("len", len(shares)))
	//	for _, share := range shares {
	//		exp.syncIbftData(share.PublicKey)
	//	}
	//}()
	err := exp.eth1Client.Start()
	if err != nil {
		exp.logger.Error("could not start eth1 client")
	}
	return exp.httpHandlers.Listen()
}

// Sync takes care of syncing an exporter node with:
//  1. ibft data from ssv nodes
//  2. registry data (validator/operator added) from eth1 contract
func (exp *exporter) Sync() error {
	exp.logger.Info("exporter.Sync()")
	go exp.ibftDisptcher.StartInterval()
	offset := exp.syncOffset()
	return exp.eth1Client.Sync(offset)
}

func (exp *exporter) HandleEth1Events() {
	go func() {
		cn, err := exp.eth1Client.EventsSubject().Register("ExporterObserver")
		if err != nil {
			exp.logger.Error("could not register for eth1 events channel", zap.Error(err))
			return
		}
		offsetRaw := exp.syncOffset()
		offset := offsetRaw.Uint64()
		var errs []error
		for e := range cn {
			exp.logger.Debug("got new eth1 event in exporter")
			if event, ok := e.(eth1.Event); ok {
				var err error
				if validatorAddedEvent, ok := event.Data.(eth1.ValidatorAddedEvent); ok {
					err = exp.handleValidatorAddedEvent(validatorAddedEvent)
				} else if opertaorAddedEvent, ok := event.Data.(eth1.OperatorAddedEvent); ok {
					err = exp.handleOperatorAddedEvent(opertaorAddedEvent)
				} else if _, ok := event.Data.(eth1.SyncEndedEvent); ok && len(errs) == 0 {
					// upgrade the sync-offset (in DB) once sync is over
					// if some errors happened, avoid updating the local offset variable with new values
					// this will make sure that the offset is not being upgraded if there are some missing events
					if offset > offsetRaw.Uint64() {
						exp.logger.Debug("upgrading sync offset", zap.Uint64("offset", offset))
						offsetRaw.SetUint64(offset)
						if err := exp.store.SaveSyncOffset(offsetRaw); err != nil {
							exp.logger.Error("could not upgrade sync offset", zap.Error(err))
						}
					}
					break
				}
				// If things went well - check for new offset
				if err == nil {
					blockNumber := event.Log.BlockNumber
					if blockNumber > offset {
						offset = blockNumber
					}
				} else {
					exp.logger.Debug("could not handle event", zap.Error(err))
					errs = append(errs, err)
				}
			}
		}
		exp.logger.Debug("done reading messages from channel")
		if len(errs) > 0 {
			exp.logger.Error("could not handle all events", zap.Int("numberOfFailures", len(errs)))
		}
	}()
}

func (exp *exporter) syncOffset() *big.Int {
	offset, err := exp.store.GetSyncOffset()
	if err != nil {
		offset = new(big.Int)
		exp.logger.Debug("could not get sync offset, using default offset")
		offset.SetString(defaultOffset, 16)
	}
	return offset
}

// handleValidatorAddedEvent parses the given event and sync the ibft-data of the validator
func (exp *exporter) handleValidatorAddedEvent(event eth1.ValidatorAddedEvent) error {
	pubKeyHex := hex.EncodeToString(event.PublicKey)
	exp.logger.Info("validator added event", zap.String("pubKey", pubKeyHex))
	validatorShare, err := validator.ShareFromValidatorAddedEvent(event, true)
	if err != nil {
		return errors.Wrap(err, "could not create a share from ValidatorAddedEvent")
	}
	if err := exp.validatorStorage.SaveValidatorShare(validatorShare); err != nil {
		return errors.Wrap(err, "failed to save validator share")
	}
	exp.logger.Debug("validator share was saved", zap.String("pubKey", pubKeyHex))
	// triggers a sync for the given validator
	exp.syncIbftData(validatorShare.PublicKey)

	return nil
}

func (exp *exporter) handleOperatorAddedEvent(event eth1.OperatorAddedEvent) error {
	exp.logger.Info(fmt.Sprintf("operator added: %x", event.PublicKey))
	// TODO: save operator data
	return nil
}

func (exp *exporter) syncIbftData(validatorPubKey *bls.PublicKey) error {
	validatorShare, err := exp.validatorStorage.GetValidatorsShare(validatorPubKey.Serialize())
	if err != nil {
		return errors.Wrap(err, "could not get validator share")
	}
	exp.logger.Info("syncing ibft data for validator", zap.String("pubKey", validatorPubKey.GetHexString()))
	ibftInstance := ibft.NewIbftReadOnly(ibft.IbftReadOnlyOptions{
		Logger:      exp.logger,
		IbftStorage: exp.ibftStorage,
		Network:     exp.network,
		Params: &proto.InstanceParams{
			ConsensusParams: proto.DefaultConsensusParams(),
			IbftCommittee:   validatorShare.Committee,
		},
		ValidatorShare: validatorShare,
	})
	// subscribe to topic so we could find relevant nodes (required by Network::AllPeers)
	err = exp.network.SubscribeToValidatorNetwork(validatorPubKey)
	if err != nil {
		return errors.Wrap(err, "could not subscribe to validator channel")
	}
	exp.ibftDisptcher.Queue(ibftInstance)

	return nil
}
