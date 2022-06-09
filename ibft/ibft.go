package ibft

import (
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/ibft/valcheck"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
)

// ControllerStartInstanceOptions defines type for Controller instance options
type ControllerStartInstanceOptions struct {
	Logger     *zap.Logger
	ValueCheck valcheck.ValueCheck
	SeqNumber  uint64
	Value      []byte
	// RequireMinPeers flag to require minimum peers before starting an instance
	// useful for tests where we want (sometimes) to avoid networking
	RequireMinPeers bool
}

// InstanceResult is a struct holding the result of a single iBFT instance
type InstanceResult struct {
	Decided bool
	Msg     *proto.SignedMessage
}

// Controller represents behavior of the Controller
type Controller interface {
	// Init should be called after creating an Controller instance to init the instance, sync it, etc.
	Init() error

	// StartInstance starts a new instance by the given options
	StartInstance(opts ControllerStartInstanceOptions) (*InstanceResult, error)

	// StopInstance stops the running instance
	StopInstance() error

	// NextSeqNumber returns the previous decided instance seq number + 1
	// In case it's the first instance it returns 0
	NextSeqNumber() (uint64, error)

	// GetIBFTCommittee returns a map of the iBFT committee where the key is the member's id.
	GetIBFTCommittee() map[uint64]*proto.Node

	// GetIdentifier returns ibft identifier made of public key and role (type)
	GetIdentifier() []byte
}

// Instance represents an iBFT instance (a single sequence number)
type Instance interface {
	Pipelines
	Init()
	Start(inputValue []byte) error
	Stop()
	State() *proto.State
	ForceDecide(msg *proto.SignedMessage)
	GetStageChan() chan proto.RoundState
	GetLastChangeRoundMsg() *proto.SignedMessage
	CommittedAggregatedMsg() (*proto.SignedMessage, error)
}

// Pipelines holds all major instance pipeline implementations
type Pipelines interface {
	// PrePrepareMsgPipeline is the full processing msg pipeline for a pre-prepare msg
	PrePrepareMsgPipeline() pipelines.SignedMessagePipeline
	// PrepareMsgPipeline is the full processing msg pipeline for a prepare msg
	PrepareMsgPipeline() pipelines.SignedMessagePipeline
	// CommitMsgValidationPipeline is a msg validation ONLY pipeline
	CommitMsgValidationPipeline() pipelines.SignedMessagePipeline
	// CommitMsgPipeline is the full processing msg pipeline for a commit msg
	CommitMsgPipeline() pipelines.SignedMessagePipeline
	// DecidedMsgPipeline is a specific full processing pipeline for a decided msg
	DecidedMsgPipeline() pipelines.SignedMessagePipeline
	// ChangeRoundMsgValidationPipeline is a msg validation ONLY pipeline for a change round msg
	ChangeRoundMsgValidationPipeline() pipelines.SignedMessagePipeline
	// ChangeRoundMsgPipeline is the full processing msg pipeline for a change round msg
	ChangeRoundMsgPipeline() pipelines.SignedMessagePipeline
}
