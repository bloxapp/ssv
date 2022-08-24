package genesis

import (
	specqbft "github.com/bloxapp/ssv-spec/qbft"

	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	"github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/qbft"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/forks"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/changeround"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/commit"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/prepare"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/proposal"
	"github.com/bloxapp/ssv/protocol/v1/qbft/validation/signedmsg"
)

// ForkGenesis is the genesis fork for instances
type ForkGenesis struct {
	instance *instance.Instance
}

// New returns new ForkGenesis
func New() forks.Fork {
	return &ForkGenesis{}
}

// Apply - applies instance fork
func (g *ForkGenesis) Apply(instance *instance.Instance) {
	g.instance = instance
}

// VersionName returns version name
func (g *ForkGenesis) VersionName() string {
	return forksprotocol.GenesisForkVersion.String()
}

// ProposalMsgValidationPipeline is the validation pipeline for proposal messages
func (g *ForkGenesis) ProposalMsgValidationPipeline(share *beacon.Share, state *qbft.State, roundLeader proposal.LeaderResolver) pipelines.SignedMessagePipeline {
	identifier := state.GetIdentifier()
	return pipelines.Combine(
		signedmsg.BasicMsgValidation(),
		signedmsg.MsgTypeCheck(specqbft.ProposalMsgType),
		signedmsg.ValidateSequenceNumber(state.GetHeight()),
		signedmsg.ValidateIdentifiers(identifier[:]),
		signedmsg.AuthorizeMsg(share),
		proposal.ValidateProposalMsg(share, state, roundLeader),
	)
}

// PrepareMsgValidationPipeline is the validation pipeline for prepare messages
func (g *ForkGenesis) PrepareMsgValidationPipeline(share *beacon.Share, state *qbft.State) pipelines.SignedMessagePipeline {
	identifier := state.GetIdentifier()
	return pipelines.Combine(
		signedmsg.BasicMsgValidation(),
		signedmsg.MsgTypeCheck(specqbft.PrepareMsgType),
		signedmsg.ValidateSequenceNumber(state.GetHeight()),
		signedmsg.ValidateRound(state.GetRound()),
		signedmsg.ValidateIdentifiers(identifier[:]),
		prepare.ValidateProposal(state),
		prepare.ValidatePrepareMsgSigners(),
		signedmsg.AuthorizeMsg(share),
	)
}

// CommitMsgValidationPipeline is the validation pipeline for commit messages
func (g *ForkGenesis) CommitMsgValidationPipeline(share *beacon.Share, state *qbft.State) pipelines.SignedMessagePipeline {
	return pipelines.Combine(
		signedmsg.BasicMsgValidation(),
		signedmsg.MsgTypeCheck(specqbft.CommitMsgType),
		signedmsg.ValidateSequenceNumber(state.GetHeight()),
		signedmsg.ValidateRound(state.GetRound()),
		signedmsg.ValidateIdentifiers(state.GetIdentifier()),
		commit.ValidateProposal(state),
		signedmsg.AuthorizeMsg(share),
	)
}

// ChangeRoundMsgValidationPipeline is the validation pipeline for commit messages
func (g *ForkGenesis) ChangeRoundMsgValidationPipeline(share *beacon.Share, identifier []byte, height specqbft.Height) pipelines.SignedMessagePipeline {
	return pipelines.Combine(
		signedmsg.BasicMsgValidation(),
		signedmsg.MsgTypeCheck(specqbft.RoundChangeMsgType),
		signedmsg.ValidateIdentifiers(identifier[:]),
		signedmsg.ValidateSequenceNumber(height),
		signedmsg.AuthorizeMsg(share),
		changeround.Validate(share),
	)
}
