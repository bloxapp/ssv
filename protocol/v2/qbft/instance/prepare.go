package instance

import (
	"bytes"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/bounded"
	"github.com/bloxapp/ssv/protocol/v2/qbft"
)

// uponPrepare process prepare message
// Assumes prepare message is valid!
func (i *Instance) uponPrepare(
	logger *zap.Logger,
	signedPrepare *specqbft.SignedMessage,
	prepareMsgContainer,
	commitMsgContainer *specqbft.MsgContainer) error {

	addedMsg, err := prepareMsgContainer.AddFirstMsgForSignerAndRound(signedPrepare)
	if err != nil {
		return errors.Wrap(err, "could not add prepare msg to container")
	}
	if !addedMsg {
		return nil // uponPrepare was already called
	}

	logger.Debug("📬 got prepare message",
		zap.Uint64("round", uint64(i.State.Round)),
		zap.Any("prepare-signers", signedPrepare.Signers))

	if !specqbft.HasQuorum(i.State.Share, prepareMsgContainer.MessagesForRound(i.State.Round)) {
		return nil // no quorum yet
	}

	if didSendCommitForHeightAndRound(i.State, commitMsgContainer) {
		return nil // already moved to commit stage
	}

	proposedRoot := i.State.ProposalAcceptedForCurrentRound.Message.Root

	i.State.LastPreparedValue = i.State.ProposalAcceptedForCurrentRound.FullData
	i.State.LastPreparedRound = i.State.Round

	i.metrics.EndStagePrepare()

	commitMsg, err := CreateCommit(i.State, i.config, proposedRoot)
	if err != nil {
		return errors.Wrap(err, "could not create commit msg")
	}

	// TODO: "prepare-signers" in log should log all unique signers or all prepare messages.
	logger.Debug("📢 got prepare quorum, broadcasting commit message",
		zap.Uint64("round", uint64(i.State.Round)),
		zap.Any("prepare-signers", signedPrepare.Signers),
		zap.Any("commit-singers", commitMsg.Signers))

	if err := i.Broadcast(commitMsg); err != nil {
		return errors.Wrap(err, "failed to broadcast commit message")
	}

	return nil
}

func getRoundChangeJustification(state *specqbft.State, config qbft.IConfig, prepareMsgContainer *specqbft.MsgContainer) ([]*specqbft.SignedMessage, error) {
	if state.LastPreparedValue == nil {
		return nil, nil
	}

	r, err := specqbft.HashDataRoot(state.LastPreparedValue)
	if err != nil {
		return nil, errors.Wrap(err, "could not hash input data")
	}

	prepareMsgs := prepareMsgContainer.MessagesForRound(state.LastPreparedRound)
	ret := make([]*specqbft.SignedMessage, 0)
	for _, msg := range prepareMsgs {
		if err := validSignedPrepareForHeightRoundAndRoot(
			config,
			msg,
			state.Height,
			state.LastPreparedRound,
			r,
			state.Share.Committee,
		); err == nil {
			ret = append(ret, msg)
		}
	}
	return ret, nil
}

// validPreparesForHeightRoundAndValue returns an aggregated prepare msg for a specific Height and round
// func validPreparesForHeightRoundAndValue(
//	config IConfig,
//	prepareMessages []*SignedMessage,
//	height Height,
//	round Round,
//	value []byte,
//	operators []*types.Operator) *SignedMessage {
//	var aggregatedPrepareMsg *SignedMessage
//	for _, signedMsg := range prepareMessages {
//		if err := validSignedPrepareForHeightRoundAndValue(config, signedMsg, height, round, value, operators); err == nil {
//			if aggregatedPrepareMsg == nil {
//				aggregatedPrepareMsg = signedMsg
//			} else {
//				// TODO: check error
//				// nolint
//				aggregatedPrepareMsg.Aggregate(signedMsg)
//			}
//		}
//	}
//	return aggregatedPrepareMsg
// }

// validSignedPrepareForHeightRoundAndValue known in dafny spec as validSignedPrepareForHeightRoundAndDigest
// https://entethalliance.github.io/client-spec/qbft_spec.html#dfn-qbftspecification
func validSignedPrepareForHeightRoundAndRoot(
	config qbft.IConfig,
	signedPrepare *specqbft.SignedMessage,
	height specqbft.Height,
	round specqbft.Round,
	root [32]byte,
	operators []*spectypes.Operator) error {
	if signedPrepare.Message.MsgType != specqbft.PrepareMsgType {
		return errors.New("prepare msg type is wrong")
	}
	if signedPrepare.Message.Height != height {
		return errors.New("wrong msg height")
	}
	if signedPrepare.Message.Round != round {
		return errors.New("wrong msg round")
	}

	if err := signedPrepare.Validate(); err != nil {
		return errors.Wrap(err, "prepareData invalid")
	}

	if !bytes.Equal(signedPrepare.Message.Root[:], root[:]) {
		return errors.New("proposed data mistmatch")
	}

	if len(signedPrepare.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}

	if err := bounded.Run(func() error {
		// return signedPrepare.Signature.VerifyByOperators(signedPrepare, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators)
		return bounded.VerifyByOperators(signedPrepare.Signature, signedPrepare, config.GetSignatureDomainType(), spectypes.QBFTSignatureType, operators)
		return nil
	}); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	return nil
}

// CreatePrepare
/**
Prepare(
                    signPrepare(
                        UnsignedPrepare(
                            |current.blockchain|,
                            newRound,
                            digest(m.proposedBlock)),
                        current.id
                        )
                );
*/
func CreatePrepare(state *specqbft.State, config qbft.IConfig, newRound specqbft.Round, root [32]byte) (*specqbft.SignedMessage, error) {
	msg := &specqbft.Message{
		MsgType:    specqbft.PrepareMsgType,
		Height:     state.Height,
		Round:      newRound,
		Identifier: state.ID,

		Root: root,
	}
	sig, err := config.GetSigner().SignRoot(msg, spectypes.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing prepare msg")
	}

	signedMsg := &specqbft.SignedMessage{
		Signature: sig,
		Signers:   []spectypes.OperatorID{state.Share.OperatorID},
		Message:   *msg,
	}
	return signedMsg, nil
}
