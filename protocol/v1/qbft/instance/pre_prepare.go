package instance

import (
	"fmt"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v1/qbft"
	"github.com/bloxapp/ssv/protocol/v1/qbft/pipelines"
)

// PrePrepareMsgPipeline is the main pre-prepare msg pipeline
func (i *Instance) PrePrepareMsgPipeline() pipelines.SignedMessagePipeline {
	return pipelines.Combine(
		i.prePrepareMsgValidationPipeline(),
		pipelines.WrapFunc("add pre-prepare msg", func(signedMessage *specqbft.SignedMessage) error {
			i.Logger.Info("received valid pre-prepare message for round",
				zap.Any("sender_ibft_id", signedMessage.GetSigners()),
				zap.Uint64("round", uint64(signedMessage.Message.Round)))

			proposalData, err := signedMessage.Message.GetProposalData()
			if err != nil {
				return fmt.Errorf("could not get proposal data: %w", err)
			}
			i.containersMap[specqbft.ProposalMsgType].AddMessage(signedMessage, proposalData.Data)

			return nil
		}),
		i.UponPrePrepareMsg(),
	)
}

func (i *Instance) prePrepareMsgValidationPipeline() pipelines.SignedMessagePipeline {
	return i.fork.PrePrepareMsgValidationPipeline(i.ValidatorShare, i.State(), i.RoundLeader)
}

/*
UponPrePrepareMsg Algorithm 2 IBFTController pseudocode for process pi: normal case operation
upon receiving a valid ⟨PRE-PREPARE, λi, ri, value⟩ message m from leader(λi, round) such that:
	JustifyPrePrepare(m) do
		set timer i to running and expire after t(ri)
		broadcast ⟨PREPARE, λi, ri, value⟩
*/
func (i *Instance) UponPrePrepareMsg() pipelines.SignedMessagePipeline {
	return pipelines.WrapFunc("upon pre-prepare msg", func(signedMessage *specqbft.SignedMessage) error {
		i.State().ProposalAcceptedForCurrentRound.Store(signedMessage)

		newRound := signedMessage.Message.Round

		if currentRound := i.State().GetRound(); signedMessage.Message.Round > currentRound {
			i.Logger.Debug("received future justified proposal, bumping into its round and resetting timer",
				zap.Uint64("current_round", uint64(currentRound)),
				zap.Uint64("future_round", uint64(signedMessage.Message.Round)),
			)
			i.ResetRoundTimer()
			i.bumpToRound(newRound)
		}

		prepareMsg, err := signedMessage.Message.GetProposalData()
		if err != nil {
			return errors.Wrap(err, "failed to get prepare message")
		}

		// mark state
		i.ProcessStageChange(qbft.RoundStatePrePrepare)

		// broadcast prepare msg
		broadcastMsg, err := i.generatePrepareMessage(prepareMsg.Data)
		if err != nil {
			return errors.Wrap(err, "failed to generate prepare message")
		}
		if err := i.SignAndBroadcast(broadcastMsg); err != nil {
			i.Logger.Error("could not broadcast prepare message", zap.Error(err))
			return err
		}
		return nil
	})
}

func (i *Instance) generatePrePrepareMessage(proposalMsg *specqbft.ProposalData) (specqbft.Message, error) {
	proposalEncodedMsg, err := proposalMsg.Encode()
	if err != nil {
		return specqbft.Message{}, errors.Wrap(err, "failed to encoded proposal message")
	}
	identifier := i.State().GetIdentifier()
	return specqbft.Message{
		MsgType:    specqbft.ProposalMsgType,
		Height:     i.State().GetHeight(),
		Round:      i.State().GetRound(),
		Identifier: identifier[:],
		Data:       proposalEncodedMsg,
	}, nil
}

func (i *Instance) checkExistingPrePrepare(round specqbft.Round) (bool, *specqbft.SignedMessage, error) {
	msgs := i.containersMap[specqbft.ProposalMsgType].ReadOnlyMessagesByRound(round)
	if len(msgs) == 1 {
		return true, msgs[0], nil
	} else if len(msgs) > 1 {
		return false, nil, errors.New("multiple pre-preparer msgs, can't decide which one to use")
	}
	return false, nil, nil
}
