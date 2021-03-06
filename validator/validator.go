package validator

import (
	"bytes"
	"context"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/bloxapp/ssv/beacon"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/storage/collections"
	"github.com/bloxapp/ssv/validator/storage"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft"
	"github.com/bloxapp/ssv/network"
	"github.com/bloxapp/ssv/network/msgqueue"
	"github.com/bloxapp/ssv/slotqueue"
)

// Options to add in validator struct creation
type Options struct {
	Context                    context.Context
	Logger                     *zap.Logger
	Share                      *storage.Share
	SignatureCollectionTimeout time.Duration
	SlotQueue                  slotqueue.Queue
	Network                    network.Network
	Beacon                     beacon.Beacon
	ETHNetwork                 *core.Network
}

// Validator struct that manages all ibft wrappers
type Validator struct {
	ctx    context.Context
	logger *zap.Logger
	Share  *storage.Share
	//db                collections.Iibft
	ethNetwork                 *core.Network
	beacon                     beacon.Beacon
	ibfts                      map[beacon.RoleType]ibft.IBFT
	msgQueue                   *msgqueue.MessageQueue
	network                    network.Network
	slotQueue                  slotqueue.Queue
	signatureCollectionTimeout time.Duration
}

// New Validator creation
func New(opt Options, db basedb.IDb) *Validator {
	logger := opt.Logger.With(zap.String("pubKey", opt.Share.PublicKey.SerializeToHexStr())).
		With(zap.Uint64("node_id", opt.Share.NodeID))

	msgQueue := msgqueue.New()
	ibfts := make(map[beacon.RoleType]ibft.IBFT)
	ibfts[beacon.RoleTypeAttester] = setupIbftController(beacon.RoleTypeAttester, logger, db, opt.Network, msgQueue, opt.Share)
	//ibfts[beacon.RoleAggregator] = setupIbftController(beacon.RoleAggregator, logger, db, opt.Network, msgQueue, opt.Share) TODO not supported for now
	//ibfts[beacon.RoleProposer] = setupIbftController(beacon.RoleProposer, logger, db, opt.Network, msgQueue, opt.Share) TODO not supported for now

	for _, ib := range ibfts { // init all ibfts
		go ib.Init()
	}

	if opt.Share.Index != nil { // in order ot update goclient map to prevent getting all network indices bug
		blsPubkey := spec.BLSPubKey{}
		copy(blsPubkey[:], opt.Share.PublicKey.Serialize())
		opt.Beacon.ExtendIndexMap(spec.ValidatorIndex(*opt.Share.Index), blsPubkey)
	}

	return &Validator{
		ctx:                        opt.Context,
		logger:                     logger,
		msgQueue:                   msgQueue,
		Share:                      opt.Share,
		signatureCollectionTimeout: opt.SignatureCollectionTimeout,
		slotQueue:                  opt.SlotQueue,
		network:                    opt.Network,
		ibfts:                      ibfts,
		ethNetwork:                 opt.ETHNetwork,
		beacon:                     opt.Beacon,
	}
}

// Start validator
func (v *Validator) Start() error {
	if v.network.IsSubscribeToValidatorNetwork(v.Share.PublicKey) {
		v.logger.Debug("already subscribed to validator's topic")
		return nil
	}
	if err := v.network.SubscribeToValidatorNetwork(v.Share.PublicKey); err != nil {
		return errors.Wrap(err, "failed to subscribe topic")
	}

	go v.startSlotQueueListener()
	go v.listenToSignatureMessages()
	return nil
}

// startSlotQueueListener starts slot queue listener
func (v *Validator) startSlotQueueListener() {
	v.logger.Info("start listening slot queue")

	ch, err := v.slotQueue.RegisterToNext(v.Share.PublicKey.Serialize())
	if err != nil {
		v.logger.Error("failed to register validator to slot queue", zap.Error(err))
		return
	}

	for e := range ch {
		if event, ok := e.(slotqueue.SlotEvent); ok {
			if !event.Ok {
				v.logger.Debug("no duties for slot scheduled")
				continue
			}
			go v.ExecuteDuty(v.ctx, event.Slot, event.Duty)
		} else {
			v.logger.Error("slot queue event is not ok")
			continue
		}
	}
}

func (v *Validator) listenToSignatureMessages() {
	sigChan := v.network.ReceivedSignatureChan()
	for sigMsg := range sigChan {
		if sigMsg == nil {
			v.logger.Debug("got nil message")
			continue
		}

		if sigMsg.Message != nil && v.oneOfIBFTIdentifiers(sigMsg.Message.Lambda) {
			v.msgQueue.AddMessage(&network.Message{
				SignedMessage: sigMsg,
				Type:          network.NetworkMsg_SignatureType,
			})
		}
	}
}

// getSlotStartTime returns the start time for the given slot  TODO: redundant func (in ssvNode) need to fix
func (v *Validator) getSlotStartTime(slot uint64) time.Time {
	timeSinceGenesisStart := slot * uint64(v.ethNetwork.SlotDurationSec().Seconds())
	start := time.Unix(int64(v.ethNetwork.MinGenesisTime()+timeSinceGenesisStart), 0)
	return start
}

func setupIbftController(role int, logger *zap.Logger, db basedb.IDb, network network.Network, msgQueue *msgqueue.MessageQueue, share *storage.Share) ibft.IBFT {
	ibftStorage := collections.NewIbft(db, logger, beacon.RoleType(role).String())
	identifier := []byte(IdentifierFormat(share.PublicKey.Serialize(), beacon.RoleType(role)))
	return ibft.New(beacon.RoleType(role), identifier, logger, &ibftStorage, network, msgQueue, proto.DefaultConsensusParams(), share)
}

// oneOfIBFTIdentifiers will return true if provided identifier matches one of the iBFT instances.
func (v *Validator) oneOfIBFTIdentifiers(toMatch []byte) bool {
	for _, i := range v.ibfts {
		if bytes.Equal(i.GetIdentifier(), toMatch) {
			return true
		}
	}
	return false
}

// GetIBFT will return the IBFT of the corresponding role
func (v *Validator) GetIBFT(r beacon.RoleType) (ibft.IBFT, bool) {
	i, exist := v.ibfts[r]
	return i, exist
}
