package ibft

import (
	"encoding/hex"
	"errors"
	"github.com/bloxapp/ssv/ibft/eventqueue"
	"github.com/bloxapp/ssv/ibft/valcheck"
	"github.com/bloxapp/ssv/validator/storage"
	"sync"
	"time"

	"github.com/bloxapp/ssv/ibft/leader"

	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/msgcont"
	msgcontinmem "github.com/bloxapp/ssv/ibft/msgcont/inmem"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/network"
	"github.com/bloxapp/ssv/network/msgqueue"
)

// InstanceOptions defines option attributes for the Instance
type InstanceOptions struct {
	Logger         *zap.Logger
	ValidatorShare *storage.Share
	//Me             *proto.Node
	Network        network.Network
	Queue          *msgqueue.MessageQueue
	ValueCheck     valcheck.ValueCheck
	LeaderSelector leader.Selector
	Config         *proto.InstanceConfig
	Lambda         []byte
	SeqNumber      uint64
}

// Instance defines the instance attributes
type Instance struct {
	ValidatorShare   *storage.Share
	State            *proto.State
	network          network.Network
	ValueCheck       valcheck.ValueCheck
	LeaderSelector   leader.Selector
	Config           *proto.InstanceConfig
	roundChangeTimer *time.Timer
	Logger           *zap.Logger

	// messages
	MsgQueue            *msgqueue.MessageQueue
	PrePrepareMessages  msgcont.MessageContainer
	PrepareMessages     msgcont.MessageContainer
	CommitMessages      msgcont.MessageContainer
	ChangeRoundMessages msgcont.MessageContainer

	// event loop
	eventQueue *eventqueue.Queue

	// channels
	stageChangedChans []chan proto.RoundState

	// flags
	stopped     bool
	initialized bool

	// locks
	stopLock              sync.Mutex
	stageChangedChansLock sync.Mutex
	stageLock             sync.Mutex
}

// NewInstance is the constructor of Instance
func NewInstance(opts InstanceOptions) *Instance {
	return &Instance{
		ValidatorShare: opts.ValidatorShare,
		State: &proto.State{
			Stage:     proto.RoundState_NotStarted,
			Lambda:    opts.Lambda,
			SeqNumber: opts.SeqNumber,
		},
		network:        opts.Network,
		ValueCheck:     opts.ValueCheck,
		LeaderSelector: opts.LeaderSelector,
		Config:         opts.Config,
		Logger: opts.Logger.With(zap.Uint64("node_id", opts.ValidatorShare.NodeID),
			zap.Uint64("seq_num", opts.SeqNumber),
			zap.String("pubKey", opts.ValidatorShare.PublicKey.SerializeToHexStr())),

		MsgQueue:            opts.Queue,
		PrePrepareMessages:  msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize())),
		PrepareMessages:     msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize())),
		CommitMessages:      msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize())),
		ChangeRoundMessages: msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize())),

		eventQueue: eventqueue.New(),

		// chan
		stageChangedChans: make([]chan proto.RoundState, 0),

		// locks
		stopLock:              sync.Mutex{},
		stageLock:             sync.Mutex{},
		stageChangedChansLock: sync.Mutex{},
	}
}

// Init must be called before start can be
func (i *Instance) Init() {
	//go i.StartEventLoop()
	go i.StartMessagePipeline()
	go i.StartMainEventLoop()
	i.initialized = true
}

// Start implements the Algorithm 1 IBFT pseudocode for process pi: constants, State variables, and ancillary procedures
// procedure Start(λ, value)
// 	λi ← λ
// 	ri ← 1
// 	pri ← ⊥
// 	pvi ← ⊥
// 	inputV aluei ← value
// 	if leader(hi, ri) = pi then
// 		broadcast ⟨PRE-PREPARE, λi, ri, inputV aluei⟩ message
// 		set timer to running and expire after t(ri)
func (i *Instance) Start(inputValue []byte) error {
	if !i.initialized {
		return errors.New("can't start instance a non initialized instance")
	}
	if i.State.Lambda == nil {
		return errors.New("can't start instance with invalid Lambda")
	}

	i.Logger.Info("Node is starting iBFT instance", zap.String("Lambda", hex.EncodeToString(i.State.Lambda)))
	i.State.Round = 1 // start from 1
	i.State.InputValue = inputValue

	if i.IsLeader() {
		go func() {
			i.Logger.Info("Node is leader for round 1")
			i.SetStage(proto.RoundState_PrePrepare)

			// LeaderPreprepareDelay waits to let other nodes complete their instance start or round change.
			// Waiting will allow a more stable msg receiving for all parties.
			time.Sleep(time.Duration(i.Config.LeaderPreprepareDelay))

			msg := i.generatePrePrepareMessage(i.State.InputValue)
			//
			if err := i.SignAndBroadcast(msg); err != nil {
				i.Logger.Fatal("could not broadcast pre-prepare", zap.Error(err))
			}
		}()
	}
	i.triggerRoundChangeOnTimer()
	return nil
}

// Stop will trigger a stopped for the entire instance
func (i *Instance) Stop() {
	i.eventQueue.Add(func() {
		i.stopLock.Lock()
		defer i.stopLock.Unlock()

		i.stopped = true
		i.stopRoundChangeTimer()
		i.SetStage(proto.RoundState_Stopped)
		i.eventQueue.ClearAndStop()
		i.Logger.Info("stopped iBFT instance")
	})
	i.Logger.Info("stopping iBFT instance...")
}

// Stopped returns true if instance is stopped
func (i *Instance) Stopped() bool {
	i.stopLock.Lock()
	defer i.stopLock.Unlock()
	return i.stopped
}

// BumpRound is used to set round in the instance's MsgQueue - the message broker
func (i *Instance) BumpRound(round uint64) {
	i.State.Round = round
	i.LeaderSelector.Bump()
}

// Stage returns the instance message state
func (i *Instance) Stage() proto.RoundState {
	i.stageLock.Lock()
	defer i.stageLock.Unlock()
	return i.State.Stage
}

// SetStage set the State's round State and pushed the new State into the State channel
func (i *Instance) SetStage(stage proto.RoundState) {
	i.stageLock.Lock()
	defer i.stageLock.Unlock()

	i.State.Stage = stage

	// Delete all queue messages when decided, we do not need them anymore.
	if i.State.Stage == proto.RoundState_Decided || i.State.Stage == proto.RoundState_Stopped {
		for j := uint64(1); j <= i.State.Round; j++ {
			i.MsgQueue.PurgeIndexedMessages(msgqueue.IBFTMessageIndexKey(i.State.Lambda, i.State.SeqNumber, j))
		}
	}

	// Non blocking send to channel
	for _, ch := range i.stageChangedChans {
		select {
		case ch <- stage:
		default:
		}
	}
}

// GetStageChan returns a RoundState channel added to the stateChangesChans array
func (i *Instance) GetStageChan() chan proto.RoundState {
	ch := make(chan proto.RoundState)
	i.stageChangedChansLock.Lock()
	i.stageChangedChans = append(i.stageChangedChans, ch)
	i.stageChangedChansLock.Unlock()
	return ch
}

// SignAndBroadcast checks and adds the signed message to the appropriate round state type
func (i *Instance) SignAndBroadcast(msg *proto.Message) error {
	//sk := &bls.SecretKey{}
	//if err := sk.Deserialize(i.Me.Sk); err != nil { // TODO - cache somewhere
	//	return err
	//}

	sig, err := msg.Sign(i.ValidatorShare.ShareKey)
	if err != nil {
		return err
	}

	signedMessage := &proto.SignedMessage{
		Message:   msg,
		Signature: sig.Serialize(),
		SignerIds: []uint64{i.ValidatorShare.NodeID},
	}
	if i.network != nil {
		return i.network.Broadcast(i.ValidatorShare.PublicKey.Serialize(), signedMessage)
	}

	switch msg.Type {
	case proto.RoundState_PrePrepare:
		i.PrePrepareMessages.AddMessage(signedMessage)
	case proto.RoundState_Prepare:
		i.PrepareMessages.AddMessage(signedMessage)
	case proto.RoundState_Commit:
		i.CommitMessages.AddMessage(signedMessage)
	case proto.RoundState_ChangeRound:
		i.ChangeRoundMessages.AddMessage(signedMessage)
	}

	return nil
}
