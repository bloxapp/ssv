package validator

import (
	"context"
	"sync"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/protocol/v2/message"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	ipsflog "github.com/ipfs/go-log"
	"github.com/pkg/errors"

	"github.com/bloxapp/ssv/ibft/storage"
	"github.com/bloxapp/ssv/protocol/v2/ssv/queue"
	"github.com/bloxapp/ssv/protocol/v2/ssv/runner"
	"github.com/bloxapp/ssv/protocol/v2/types"
)

var logger = ipsflog.Logger("ssv/protocol/ssv/validator").Desugar() // TODO REVIEW

// Validator represents an SSV ETH consensus validator Share assigned, coordinates duty execution and more.
// Every validator has a validatorID which is validator's public key.
// Each validator has multiple DutyRunners, for each duty type.
type Validator struct {
	mtx    *sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	DutyRunners runner.DutyRunners
	Network     specqbft.Network
	Beacon      specssv.BeaconNode
	Share       *types.SSVShare
	Signer      spectypes.KeyManager

	Storage *storage.QBFTStores
	Queues  map[spectypes.BeaconRole]queueContainer

	state uint32
}

// NewValidator creates a new instance of Validator.
func NewValidator(pctx context.Context, cancel func(), options Options) *Validator {
	options.defaults()

	v := &Validator{
		mtx:         &sync.RWMutex{},
		ctx:         pctx,
		cancel:      cancel,
		DutyRunners: options.DutyRunners,
		Network:     options.Network,
		Beacon:      options.Beacon,
		Storage:     options.Storage,
		Share:       options.SSVShare,
		Signer:      options.Signer,
		Queues:      make(map[spectypes.BeaconRole]queueContainer),
		state:       uint32(NotStarted),
	}

	for _, dutyRunner := range options.DutyRunners {
		// Set timeout function.
		dutyRunner.GetBaseRunner().TimeoutF = v.onTimeout

		// Setup the queue.
		role := dutyRunner.GetBaseRunner().BeaconRoleType
		msgID := spectypes.NewMsgID(options.SSVShare.ValidatorPubKey, role).String()

		v.Queues[role] = queueContainer{
			Q: queue.WithMetrics(queue.New(options.QueueSize), queue.NewPrometheusMetrics(msgID)),
			queueState: &queue.State{
				HasRunningInstance: false,
				Height:             0,
				Slot:               0,
				//Quorum:             options.SSVShare.Share,// TODO
			},
		}
	}

	return v
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
func (v *Validator) ProcessMessage(msg *queue.DecodedSSVMessage) error {
	dutyRunner := v.DutyRunners.DutyRunnerForMsgID(msg.GetID())
	if dutyRunner == nil {
		return errors.Errorf("could not get duty runner for msg ID")
	}

	if err := validateMessage(v.Share.Share, msg.SSVMessage); err != nil {
		return errors.Wrap(err, "Message invalid")
	}

	logger = logger.With(logging.PubKey(v.Share.ValidatorPubKey))

	switch msg.GetType() {
	case spectypes.SSVConsensusMsgType:
		signedMsg, ok := msg.Body.(*specqbft.SignedMessage)
		if !ok {
			return errors.New("could not decode consensus message from network message")
		}
		return dutyRunner.ProcessConsensus(signedMsg)
	case spectypes.SSVPartialSignatureMsgType:
		signedMsg, ok := msg.Body.(*specssv.SignedPartialSignatureMessage)
		if !ok {
			return errors.New("could not decode post consensus message from network message")
		}
		if signedMsg.Message.Type == specssv.PostConsensusPartialSig {
			return dutyRunner.ProcessPostConsensus(logger, signedMsg)
		}
		return dutyRunner.ProcessPreConsensus(signedMsg)
	case message.SSVEventMsgType:
		return v.handleEventMessage(logger, msg, dutyRunner)
	default:
		return errors.New("unknown msg")
	}
}

func validateMessage(share spectypes.Share, msg *spectypes.SSVMessage) error {
	if !share.ValidatorPubKey.MessageIDBelongs(msg.GetID()) {
		return errors.New("msg ID doesn't match validator ID")
	}

	if len(msg.GetData()) == 0 {
		return errors.New("msg data is invalid")
	}

	return nil
}
