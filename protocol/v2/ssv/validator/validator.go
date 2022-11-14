package validator

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bloxapp/ssv-spec/p2p"
	"github.com/bloxapp/ssv-spec/qbft"
	"github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v1/message"
	v1types "github.com/bloxapp/ssv/protocol/v1/types"
	"github.com/bloxapp/ssv/protocol/v2/ssv/msgqueue"
	"github.com/bloxapp/ssv/protocol/v2/ssv/runner"
	"github.com/bloxapp/ssv/protocol/v2/types"
)

type Options struct {
	Logger      *zap.Logger
	Network     qbft.Network
	Beacon      ssv.BeaconNode
	Storage     qbft.Storage
	Share       *types.SSVShare
	Signer      spectypes.KeyManager
	DutyRunners runner.DutyRunners
	Mode        ValidatorMode
}

func (o *Options) defaults() {
	if o.Logger == nil {
		o.Logger = zap.L()
	}
}

// set of states for the controller
const (
	NotStarted uint32 = iota
	Started
)

// Validator represents an SSV ETH consensus validator Share assigned, coordinates duty execution and more.
// Every validator has a validatorID which is validator's public key.
// Each validator has multiple DutyRunners, for each duty type.
type Validator struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger

	DomainType spectypes.DomainType

	DutyRunners runner.DutyRunners

	Share  *types.SSVShare
	Beacon ssv.BeaconNode
	Signer spectypes.KeyManager

	Storage qbft.Storage // TODO: change?
	Network qbft.Network

	Q msgqueue.MsgQueue

	mode int32

	State uint32
}

type ValidatorMode int32

var (
	ModeRW ValidatorMode = 0
	ModeR  ValidatorMode = 1
)

func NewValidator(pctx context.Context, options Options) *Validator {
	options.defaults()
	ctx, cancel := context.WithCancel(pctx)

	var q msgqueue.MsgQueue
	if options.Mode == ModeRW {
		indexers := msgqueue.WithIndexers( /*msgqueue.DefaultMsgIndexer(), */ msgqueue.SignedMsgIndexer(), msgqueue.DecidedMsgIndexer(), msgqueue.SignedPostConsensusMsgIndexer())
		q, _ = msgqueue.New(options.Logger, indexers) // TODO: handle error
	}

	v := &Validator{
		ctx:         ctx,
		cancel:      cancel,
		logger:      options.Logger,
		DomainType:  v1types.GetDefaultDomain(),
		DutyRunners: options.DutyRunners,
		Network:     options.Network,
		Beacon:      options.Beacon,
		Storage:     options.Storage,
		Share:       options.Share,
		Signer:      options.Signer,
		Q:           q,
		mode:        int32(options.Mode),
		State:       NotStarted,
	}

	return v
}

func (v *Validator) Start() error {
	if atomic.CompareAndSwapUint32(&v.State, NotStarted, Started) {
		n, ok := v.Network.(p2p.Subscriber)
		if !ok {
			return nil
		}
		identifiers := v.DutyRunners.Identifiers()
		for _, identifier := range identifiers {
			_, err := v.loadLastHeight(identifier)
			if err != nil {
				v.logger.Warn("could not load highest", zap.String("identifier", identifier.String()), zap.Error(err))
			}
			if err := n.Subscribe(identifier.GetPubKey()); err != nil {
				return err
			}
			go v.StartQueueConsumer(identifier, v.ProcessMessage)
			go v.sync(identifier)
		}
	}
	return nil
}

func (v *Validator) sync(mid spectypes.MessageID) {
	ctx, cancel := context.WithCancel(v.ctx)
	defer cancel()

	// TODO: config?
	interval := time.Second
	retries := 3

	for ctx.Err() == nil {
		err := v.Network.SyncHighestDecided(mid)
		if err != nil {
			v.logger.Debug("could not sync highest decided", zap.String("identifier", mid.String()))
			retries--
			if retries > 0 {
				interval *= 2
				time.Sleep(interval)
				continue
			}
		}
		return
	}
}

func (v *Validator) Stop() error {
	v.cancel()
	if atomic.LoadInt32(&v.mode) == int32(ModeR) {
		return nil
	}
	// clear the msg q
	v.Q.Clean(func(index msgqueue.Index) bool {
		return true
	})
	return nil
}

func (v *Validator) HandleMessage(msg *spectypes.SSVMessage) {
	if atomic.LoadInt32(&v.mode) == int32(ModeR) {
		err := v.ProcessMessage(msg)
		if err != nil {
			v.logger.Warn("could not handle msg", zap.Error(err))
		}
		return
	}
	fields := []zap.Field{
		zap.Int("queue_len", v.Q.Len()),
		zap.String("msgType", message.MsgTypeToString(msg.MsgType)),
		zap.String("msgID", msg.MsgID.String()),
	}
	v.logger.Debug("got message, add to queue", fields...)
	v.Q.Add(msg)
}

// StartDuty starts a duty for the validator
func (v *Validator) StartDuty(duty *spectypes.Duty) error {
	dutyRunner := v.DutyRunners[duty.Type]
	if dutyRunner == nil {
		return errors.Errorf("duty type %s not supported", duty.Type.String())
	}
	return dutyRunner.StartNewDuty(duty)
}

// ProcessMessage processes Network Message of all types
func (v *Validator) ProcessMessage(msg *spectypes.SSVMessage) error {
	dutyRunner := v.DutyRunners.DutyRunnerForMsgID(msg.GetID())
	if dutyRunner == nil {
		return errors.Errorf("could not get duty runner for msg ID")
	}

	if err := v.validateMessage(dutyRunner, msg); err != nil {
		return errors.Wrap(err, "Message invalid")
	}

	switch msg.GetType() {
	case spectypes.SSVConsensusMsgType:
		signedMsg := &qbft.SignedMessage{}
		if err := signedMsg.Decode(msg.GetData()); err != nil {
			return errors.Wrap(err, "could not get consensus Message from network Message")
		}
		return dutyRunner.ProcessConsensus(signedMsg)
	case spectypes.SSVPartialSignatureMsgType:
		signedMsg := &ssv.SignedPartialSignatureMessage{}
		if err := signedMsg.Decode(msg.GetData()); err != nil {
			return errors.Wrap(err, "could not get post consensus Message from network Message")
		}

		if signedMsg.Message.Type == ssv.PostConsensusPartialSig {
			v.logger.Info("process post consensus")
			return dutyRunner.ProcessPostConsensus(signedMsg)
		}
		return dutyRunner.ProcessPreConsensus(signedMsg)
	default:
		return errors.New("unknown msg")
	}
}

func (v *Validator) validateMessage(runner runner.Runner, msg *spectypes.SSVMessage) error {
	if !v.Share.ValidatorPubKey.MessageIDBelongs(msg.GetID()) {
		return errors.New("msg ID doesn't match validator ID")
	}

	if len(msg.GetData()) == 0 {
		return errors.New("msg data is invalid")
	}

	return nil
}

func (v *Validator) loadLastHeight(identifier spectypes.MessageID) (qbft.Height, error) {
	knownDecided, err := v.Storage.GetHighestDecided(identifier[:])
	if err != nil {
		return qbft.Height(0), errors.Wrap(err, "failed to get heights decided")
	}
	if knownDecided == nil {
		return qbft.Height(0), nil
	}
	r := v.DutyRunners.DutyRunnerForMsgID(identifier)

	if r == nil || r.GetBaseRunner() == nil {
		return qbft.Height(0), errors.New("runner is nil")
	}
	r.GetBaseRunner().QBFTController.Height = knownDecided.Message.Height
	return knownDecided.Message.Height, nil
}
