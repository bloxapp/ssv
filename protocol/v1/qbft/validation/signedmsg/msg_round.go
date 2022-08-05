package signedmsg

import (
	"bytes"
	"fmt"

	specqbft "github.com/bloxapp/ssv-spec/qbft"

	"github.com/bloxapp/ssv/protocol/v1/qbft"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
)

// ValidateRound validates round
func ValidateRound(round specqbft.Round) pipelines.SignedMessagePipeline {
	return pipelines.WrapFunc("round", func(signedMessage *specqbft.SignedMessage) error {
		if round != signedMessage.Message.Round {
			return fmt.Errorf("round is wrong")
		}
		return nil
	})
}

// CheckProposal checks if proposal for current was received.
func CheckProposal(state *qbft.State) pipelines.SignedMessagePipeline {
	return pipelines.WrapFunc("validate proposal", func(signedMessage *specqbft.SignedMessage) error {
		proposedMsg := state.GetProposalAcceptedForCurrentRound()
		if proposedMsg == nil {
			return fmt.Errorf("did not receive proposal for this round")
		}

		return nil
	})
}

// ValidateProposal validates message against received proposal for this roundю
func ValidateProposal(state *qbft.State) pipelines.SignedMessagePipeline {
	return pipelines.WrapFunc("validate proposal", func(signedMessage *specqbft.SignedMessage) error {
		proposedMsg := state.GetProposalAcceptedForCurrentRound()
		if proposedMsg == nil {
			return fmt.Errorf("did not receive proposal for this round")
		}

		proposedCommitData, err := proposedMsg.Message.GetCommitData()
		if err != nil {
			return fmt.Errorf("could not get proposed commit data: %w", err)
		}

		msgCommitData, err := signedMessage.Message.GetCommitData()
		if err != nil {
			return fmt.Errorf("could not get msg commit data: %w", err)
		}
		if err := msgCommitData.Validate(); err != nil {
			return fmt.Errorf("msgCommitData invalid: %w", err)
		}

		if !bytes.Equal(proposedCommitData.Data, msgCommitData.Data) {
			return fmt.Errorf("proposed data different than commit msg data")
		}

		return nil
	})
}
