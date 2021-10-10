package ibft

import (
	"github.com/bloxapp/ssv/ibft/leader"
	"github.com/bloxapp/ssv/ibft/pipeline"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/ibft/valcheck"
	"github.com/bloxapp/ssv/network"
	"github.com/bloxapp/ssv/network/msgqueue"
	"github.com/bloxapp/ssv/validator/storage"
	"go.uber.org/zap"
)

// ControllerStartInstanceOptions defines type for Controller instance options
type ControllerStartInstanceOptions struct {
	Logger         *zap.Logger
	ValueCheck     valcheck.ValueCheck
	SeqNumber      uint64
	Value          []byte
	ValidatorShare *storage.Share
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

	// NextSeqNumber returns the previous decided instance seq number + 1
	// In case it's the first instance it returns 0
	NextSeqNumber() (uint64, error)

	// GetIBFTCommittee returns a map of the iBFT committee where the key is the member's id.
	GetIBFTCommittee() map[uint64]*proto.Node

	// GetIdentifier returns ibft identifier made of public key and role (type)
	GetIdentifier() []byte
}

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
	// RequireMinPeers flag to require minimum peers before starting an instance
	// useful for tests where we want (sometimes) to avoid networking
	RequireMinPeers bool
}

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

type Pipelines interface {
	// PrePrepareMsgPipeline is the full processing msg pipeline for a pre-prepare msg
	PrePrepareMsgPipeline() pipeline.Pipeline
	// PrepareMsgPipeline is the full processing msg pipeline for a prepare msg
	PrepareMsgPipeline() pipeline.Pipeline
	// CommitMsgValidationPipeline is a msg validation ONLY pipeline
	CommitMsgValidationPipeline() pipeline.Pipeline
	// CommitMsgPipeline is the full processing msg pipeline for a commit msg
	CommitMsgPipeline() pipeline.Pipeline
	// DecidedMsgPipeline is a specific full processing pipeline for a decided msg
	DecidedMsgPipeline() pipeline.Pipeline
	// ChangeRoundMsgValidationPipeline is a msg validation ONLY pipeline for a change round msg
	ChangeRoundMsgValidationPipeline() pipeline.Pipeline
	// ChangeRoundMsgPipeline is the full processing msg pipeline for a change round msg
	ChangeRoundMsgPipeline() pipeline.Pipeline
}
