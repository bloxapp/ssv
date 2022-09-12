package instance

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/message"
	protcolp2p "github.com/bloxapp/ssv/protocol/v1/p2p"
	"github.com/bloxapp/ssv/protocol/v1/qbft"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/forks"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/leader"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/msgcont"
	msgcontinmem "github.com/bloxapp/ssv/protocol/v1/qbft/instance/msgcont/inmem"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/roundtimer"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
	qbftstorage "github.com/bloxapp/ssv/protocol/v1/qbft/storage"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/signedmsg"
)

// Options defines option attributes for the Instance
type Options struct {
	Logger         *zap.Logger
	ValidatorShare *beaconprotocol.Share
	Network        protcolp2p.Network
	LeaderSelector leader.Selector
	Config         *qbft.InstanceConfig
	Identifier     []byte
	Height         specqbft.Height
	// RequireMinPeers flag to require minimum peers before starting an instance
	// useful for tests where we want (sometimes) to avoid networking
	RequireMinPeers bool
	// Fork sets the current fork to apply on instance
	Fork             forks.Fork
	SSVSigner        spectypes.SSVSigner
	ChangeRoundStore qbftstorage.ChangeRoundStore
}

// Instance defines the instance attributes
type Instance struct {
	ValidatorShare *beaconprotocol.Share
	State          *qbft.State
	network        protcolp2p.Network
	LeaderSelector leader.Selector
	Config         *qbft.InstanceConfig
	roundTimer     *roundtimer.RoundTimer
	Logger         *zap.Logger
	fork           forks.Fork
	SsvSigner      spectypes.SSVSigner

	// messages
	ContainersMap map[specqbft.MessageType]msgcont.MessageContainer
	decidedMsg    *specqbft.SignedMessage

	// channels
	stageChangedChan chan qbft.RoundState

	// flags
	stopped     atomic.Bool
	initialized bool

	// locks
	runInitOnce                  *sync.Once
	runStopOnce                  *sync.Once
	processChangeRoundQuorumOnce *sync.Once
	processPrepareQuorumOnce     *sync.Once
	processCommitQuorumOnce      *sync.Once
	lastChangeRoundMsgLock       sync.RWMutex
	stageChanCloseChan           sync.Mutex

	changeRoundStore qbftstorage.ChangeRoundStore
	ctx              context.Context
	cancelCtx        context.CancelFunc
}

// NewInstanceWithState used for testing, not PROD!
func NewInstanceWithState(state *qbft.State) Instancer {
	return &Instance{
		State: state,
	}
}

// NewInstance is the constructor of Instance
func NewInstance(opts *Options) Instancer {
	messageID := message.ToMessageID(opts.Identifier)
	metricsIBFTStage.WithLabelValues(messageID.GetRoleType().String(), hex.EncodeToString(messageID.GetPubKey())).Set(float64(qbft.RoundStateNotStarted))
	logger := opts.Logger.With(zap.Uint64("instance height", uint64(opts.Height)))
	ctx, cancelCtx := context.WithCancel(context.Background())

	ret := &Instance{
		ctx:            ctx,
		cancelCtx:      cancelCtx,
		ValidatorShare: opts.ValidatorShare,
		State:          GenerateState(opts),
		network:        opts.Network,
		LeaderSelector: opts.LeaderSelector,
		Config:         opts.Config,
		Logger:         logger,
		SsvSigner:      opts.SSVSigner,

		roundTimer: roundtimer.New(ctx, logger.With(zap.String("who", "RoundTimer"))),

		// locks
		runInitOnce:                  &sync.Once{},
		runStopOnce:                  &sync.Once{},
		processChangeRoundQuorumOnce: &sync.Once{},
		processPrepareQuorumOnce:     &sync.Once{},
		processCommitQuorumOnce:      &sync.Once{},
		lastChangeRoundMsgLock:       sync.RWMutex{},
		stageChanCloseChan:           sync.Mutex{},

		changeRoundStore: opts.ChangeRoundStore,

		stopped: *atomic.NewBool(false),
	}

	ret.ContainersMap = map[specqbft.MessageType]msgcont.MessageContainer{
		specqbft.ProposalMsgType:    msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize()), uint64(opts.ValidatorShare.PartialThresholdSize())),
		specqbft.PrepareMsgType:     msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize()), uint64(opts.ValidatorShare.PartialThresholdSize())),
		specqbft.CommitMsgType:      msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize()), uint64(opts.ValidatorShare.PartialThresholdSize())),
		specqbft.RoundChangeMsgType: msgcontinmem.New(uint64(opts.ValidatorShare.ThresholdSize()), uint64(opts.ValidatorShare.PartialThresholdSize())),
	}

	ret.setFork(opts.Fork)

	return ret
}

// Init must be called before start can be
func (i *Instance) Init() {
	i.runInitOnce.Do(func() {
		go i.startRoundTimerLoop()
		i.initialized = true
		i.Logger.Debug("iBFT instance init finished")
	})
}

// GetState returns instance state
func (i *Instance) GetState() *qbft.State {
	return i.State
}

// Containers returns map of containers
func (i *Instance) Containers() map[specqbft.MessageType]msgcont.MessageContainer {
	return i.ContainersMap
}

// Start implements the Algorithm 1 IBFTController pseudocode for process pi: constants, state variables, and ancillary procedures
// procedure Start(λ, value)
// 	λi ← λ
// 	ri ← 1
// 	pri ← ⊥
// 	pvi ← ⊥
// 	inputValue i ← value
// 	if leader(hi, ri) = pi then
// 		broadcast ⟨PROPOSAL, λi, ri, inputV aluei⟩ message
// 		set timer to running and expire after t(ri)
func (i *Instance) Start(inputValue []byte) error {
	if !i.initialized {
		return errors.New("instance not initialized")
	}
	if inputValue == nil {
		return errors.New("input value is nil")
	}

	messageID := message.ToMessageID(i.GetState().GetIdentifier())
	i.Logger.Info("Node is starting iBFT instance", zap.String("identifier", hex.EncodeToString(i.GetState().GetIdentifier())))
	i.GetState().InputValue.Store(inputValue)
	i.GetState().Round.Store(specqbft.Round(1)) // start from 1
	metricsIBFTRound.WithLabelValues(messageID.GetRoleType().String(), hex.EncodeToString(messageID.GetPubKey())).Set(1)

	i.Logger.Debug("state", zap.Uint64("round", uint64(i.GetState().GetRound())))
	if i.IsLeader() {
		go func() {
			i.Logger.Info("Node is leader for round 1")
			//i.ProcessStageChange(qbft.RoundStateProposal) we need to process the proposal msg in order to broadcast to prepare msg

			// LeaderProposalDelaySeconds waits to let other nodes complete their instance start or round change.
			// Waiting will allow a more stable msg receiving for all parties.
			time.Sleep(time.Duration(i.Config.LeaderProposalDelaySeconds))

			msg, err := i.GenerateProposalMessage(&specqbft.ProposalData{
				Data: i.GetState().GetInputValue(),
			})
			if err != nil {
				i.Logger.Warn("failed to generate proposal message", zap.Error(err))
				return
			}

			if err := i.SignAndBroadcast(&msg); err != nil {
				i.Logger.Error("could not broadcast proposal", zap.Error(err))
			}
		}()
	}
	i.ResetRoundTimer() // TODO could be race condition with message process?
	return nil
}

// Stop will trigger a stopped for the entire instance
func (i *Instance) Stop() {
	// stop can be run just once
	i.runStopOnce.Do(func() {
		i.stop()
		i.cancelCtx()
	})
}

// stop stops the instance
func (i *Instance) stop() {
	i.Logger.Info("stopping iBFT instance...")
	i.Logger.Debug("STOPPING IBFTController -> set stopped to true")
	i.stopped.Store(true)
	i.Logger.Debug("STOPPING IBFTController -> kill round timer")
	i.roundTimer.Kill()
	i.Logger.Debug("STOPPING IBFTController -> stopped round timer")
	i.ProcessStageChange(qbft.RoundStateStopped)
	i.Logger.Debug("STOPPING IBFTController -> round stage set stopped")
	// stop stage chan
	if i.stageChangedChan != nil {
		i.Logger.Debug("STOPPING IBFTController -> lock stage chan")
		i.stageChanCloseChan.Lock() // in order to prevent from sending to a close chan
		i.Logger.Debug("STOPPING IBFTController -> closing stage changed chan")
		close(i.stageChangedChan)
		i.Logger.Debug("STOPPING IBFTController -> closed stageChangedChan")
		i.stageChangedChan = nil
		i.Logger.Debug("STOPPING IBFTController -> stageChangedChan nilled")
		i.stageChanCloseChan.Unlock()
		i.Logger.Debug("STOPPING IBFTController -> stageChangedChan chan unlocked")
	}
	i.Logger.Info("stopped iBFT instance")
}

// Stopped is stopping queue work
func (i *Instance) Stopped() bool {
	return i.stopped.Load()
}

// ProcessMsg will process the message
func (i *Instance) ProcessMsg(msg *specqbft.SignedMessage) (bool, error) {
	if err := msg.Validate(); err != nil {
		return false, errors.Wrap(err, "invalid signed message")
	}
	var p pipelines.SignedMessagePipeline

	switch msg.Message.MsgType {
	case specqbft.ProposalMsgType:
		p = i.ProposalMsgPipeline()
	case specqbft.PrepareMsgType:
		p = i.PrepareMsgPipeline()
	case specqbft.CommitMsgType:
		p = i.CommitMsgPipeline()
	case specqbft.RoundChangeMsgType:
		p = i.ChangeRoundMsgPipeline()
	default:
		i.Logger.Warn("undefined message type", zap.Any("msg", msg))
		return false, fmt.Errorf("undefined message type")
	}

	if err := p.Run(msg); err != nil {
		if errors.Is(err, signedmsg.ErrWrongRound) {
			i.Logger.Debug(fmt.Sprintf("message type %d,  round (%d) does not equal state round (%d)", msg.Message.MsgType, msg.Message.Round, i.GetState().GetRound()))
		}
		return false, err
	}

	if i.GetState().Stage.Load() == int32(qbft.RoundStateDecided) {
		return true, nil
	}
	return false, nil
}

// BumpRound is used to set bump round by 1
func (i *Instance) BumpRound() {
	i.bumpToRound(i.GetState().GetRound() + 1)
}

func (i *Instance) bumpToRound(round specqbft.Round) {
	i.processChangeRoundQuorumOnce = &sync.Once{}
	i.processPrepareQuorumOnce = &sync.Once{}
	newRound := round
	i.GetState().Round.Store(newRound)
	messageID := message.ToMessageID(i.GetState().GetIdentifier())
	metricsIBFTRound.WithLabelValues(messageID.GetRoleType().String(), hex.EncodeToString(messageID.GetPubKey())).Set(float64(newRound))
}

// ProcessStageChange set the state's round state and pushed the new state into the state channel
func (i *Instance) ProcessStageChange(stage qbft.RoundState) {
	// in order to prevent race condition between timer timeout and decided state. once decided we need to prevent any other new state
	currentStage := i.GetState().Stage.Load()
	if currentStage == int32(qbft.RoundStateStopped) {
		return
	}
	if currentStage == int32(qbft.RoundStateDecided) && stage != qbft.RoundStateStopped {
		return
	}

	messageID := message.ToMessageID(i.GetState().GetIdentifier())
	metricsIBFTStage.WithLabelValues(messageID.GetRoleType().String(), hex.EncodeToString(messageID.GetPubKey())).Set(float64(stage))

	i.GetState().Stage.Store(int32(stage))

	// blocking send to channel
	i.stageChanCloseChan.Lock()
	defer i.stageChanCloseChan.Unlock()
	if i.stageChangedChan != nil {
		i.stageChangedChan <- stage
	}
}

// GetStageChan returns a RoundState channel added to the stateChangesChans array
func (i *Instance) GetStageChan() chan qbft.RoundState {
	if i.stageChangedChan == nil {
		i.stageChangedChan = make(chan qbft.RoundState, 1) // buffer of 1 in order to support process stop stage right after decided
	}
	return i.stageChangedChan
}

// SignAndBroadcast checks and adds the signed message to the appropriate round state type
func (i *Instance) SignAndBroadcast(msg *specqbft.Message) error {
	i.Logger.Debug("broadcasting consensus msg",
		zap.Int("type", int(msg.MsgType)),
		zap.Int64("height", int64(msg.Height)),
		zap.Int64("round", int64(msg.Round)),
	)
	pk, err := i.ValidatorShare.OperatorSharePubKey()
	if err != nil {
		return errors.Wrap(err, "could not find operator pk for signing msg")
	}

	sigByts, err := i.SsvSigner.SignRoot(msg, spectypes.QBFTSignatureType, pk.Serialize())
	if err != nil {
		return err
	}

	signedMessage := &specqbft.SignedMessage{
		Message:   msg,
		Signature: sigByts,
		Signers:   []spectypes.OperatorID{i.ValidatorShare.NodeID},
	}

	// used for instance fast change round catchup
	if msg.MsgType == specqbft.RoundChangeMsgType {
		i.setLastChangeRoundMsg(signedMessage)
	}

	encodedMsg, err := signedMessage.Encode()
	if err != nil {
		return errors.New("could not encode message")
	}
	ssvMsg := spectypes.SSVMessage{
		MsgType: spectypes.SSVConsensusMsgType,
		MsgID:   message.ToMessageID(i.GetState().GetIdentifier()),
		Data:    encodedMsg,
	}
	if i.network != nil {
		return i.network.Broadcast(ssvMsg)
	}
	return errors.New("no networking, could not broadcast msg")
}

func (i *Instance) setLastChangeRoundMsg(msg *specqbft.SignedMessage) {
	_ = i.changeRoundStore.SaveLastChangeRoundMsg(msg)
}

//// GetLastChangeRoundMsg returns the latest broadcasted msg from the instance
//func (i *Instance) GetLastChangeRoundMsg() *specqbft.SignedMessage {
//	err :=  i.changeRoundStore.GetLastChangeRoundMsg()
//}

// CommittedAggregatedMsg returns a signed message for the state's committed value with the max known signatures
func (i *Instance) CommittedAggregatedMsg() (*specqbft.SignedMessage, error) {
	if i.GetState() == nil {
		return nil, errors.New("missing instance state")
	}
	if i.decidedMsg != nil {
		return i.decidedMsg, nil
	}
	return nil, errors.New("missing decided message")
}

// GetCommittedAggSSVMessage returns ssv msg with message.SSVDecidedMsgType and the agg commit signed msg
func (i *Instance) GetCommittedAggSSVMessage() (spectypes.SSVMessage, error) {
	decidedMsg, err := i.CommittedAggregatedMsg()
	if err != nil {
		return spectypes.SSVMessage{}, err
	}
	encodedAgg, err := decidedMsg.Encode()
	if err != nil {
		return spectypes.SSVMessage{}, errors.Wrap(err, "failed to encode agg message")
	}
	ssvMsg := spectypes.SSVMessage{
		MsgType: spectypes.SSVDecidedMsgType,
		MsgID:   message.ToMessageID(i.GetState().GetIdentifier()),
		Data:    encodedAgg,
	}
	return ssvMsg, nil
}

func (i *Instance) setFork(fork forks.Fork) {
	if fork == nil {
		return
	}
	i.fork = fork
	//i.fork.Apply(i)
}

// GenerateState return generated state
func GenerateState(opts *Options) *qbft.State {
	var identifier, height, round, preparedRound, preparedValue, iv, proposalReceivedForCurrentRound atomic.Value
	height.Store(opts.Height)
	round.Store(specqbft.Round(0))
	identifier.Store(opts.Identifier[:])
	preparedRound.Store(specqbft.Round(0))
	preparedValue.Store([]byte(nil))
	iv.Store([]byte{})
	proposalReceivedForCurrentRound.Store((*specqbft.SignedMessage)(nil))

	return &qbft.State{
		Stage:                           *atomic.NewInt32(int32(qbft.RoundStateNotStarted)),
		Identifier:                      identifier,
		Height:                          height,
		InputValue:                      iv,
		Round:                           round,
		PreparedRound:                   preparedRound,
		PreparedValue:                   preparedValue,
		ProposalAcceptedForCurrentRound: proposalReceivedForCurrentRound,
	}
}
