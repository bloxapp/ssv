package validation

// partial_validation.go contains methods for validating partial signature messages

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/herumi/bls-eth-go-binary/bls"
	"golang.org/x/exp/slices"

	ssvtypes "github.com/bloxapp/ssv/protocol/v2/types"
)

func (mv *messageValidator) validatePartialSignatureMessage(
	share *ssvtypes.SSVShare,
	signedMsg *spectypes.SignedPartialSignatureMessage,
	msgID spectypes.MessageID,
) (phase0.Slot, error) {
	if mv.inCommittee(share) {
		mv.metrics.InCommitteeMessage(spectypes.SSVPartialSignatureMsgType, false)
	} else {
		mv.metrics.NonCommitteeMessage(spectypes.SSVPartialSignatureMsgType, false)
	}

	msgSlot := signedMsg.Message.Slot

	if !mv.validPartialSigMsgType(signedMsg.Message.Type) {
		e := ErrUnknownPartialMessageType
		e.got = signedMsg.Message.Type
		return msgSlot, e
	}

	role := msgID.GetRoleType()
	if !mv.partialSignatureTypeMatchesRole(signedMsg.Message.Type, role) {
		return msgSlot, ErrPartialSignatureTypeRoleMismatch
	}

	if err := mv.validatePartialMessages(share, signedMsg); err != nil {
		return msgSlot, err
	}

	consensusState := mv.consensusState(msgID)

	state := mv.consensusState(msgID)
	signerState := consensusState.GetSignerState(signedMsg.Signer)
	if signerState != nil {
		if err := mv.validateSignerBehaviorPartial(state, signedMsg.Signer, share, msgID, signedMsg); err != nil {
			return msgSlot, err
		}
	}

	if err := mv.validateSignatureFormat(signedMsg.Signature); err != nil {
		return msgSlot, err
	}

	if mv.checkSignature {
		if err := mv.validPartialSignatures(share, signedMsg); err != nil {
			return msgSlot, err
		}
	}

	if signerState == nil {
		signerState = consensusState.CreateSignerState(signedMsg.Signer)
	}

	if msgSlot > signerState.Slot {
		newEpoch := mv.netCfg.Beacon.EstimatedEpochAtSlot(msgSlot) > mv.netCfg.Beacon.EstimatedEpochAtSlot(signerState.Slot)
		signerState.ResetSlot(msgSlot, specqbft.FirstRound, newEpoch)
	}

	signerState.MessageCounts.RecordPartialSignatureMessage(signedMsg)

	return msgSlot, nil
}

func (mv *messageValidator) inCommittee(share *ssvtypes.SSVShare) bool {
	return slices.ContainsFunc(share.Committee, func(operator *spectypes.Operator) bool {
		return operator.OperatorID == mv.ownOperatorID
	})
}

func (mv *messageValidator) validPartialSigMsgType(msgType spectypes.PartialSigMsgType) bool {
	switch msgType {
	case spectypes.PostConsensusPartialSig,
		spectypes.RandaoPartialSig,
		spectypes.SelectionProofPartialSig,
		spectypes.ContributionProofs,
		spectypes.ValidatorRegistrationPartialSig:
		return true
	default:
		return false
	}
}

func (mv *messageValidator) partialSignatureTypeMatchesRole(msgType spectypes.PartialSigMsgType, role spectypes.BeaconRole) bool {
	switch role {
	case spectypes.BNRoleAttester:
		return msgType == spectypes.PostConsensusPartialSig
	case spectypes.BNRoleAggregator:
		return msgType == spectypes.PostConsensusPartialSig || msgType == spectypes.SelectionProofPartialSig
	case spectypes.BNRoleProposer:
		return msgType == spectypes.PostConsensusPartialSig || msgType == spectypes.RandaoPartialSig
	case spectypes.BNRoleSyncCommittee:
		return msgType == spectypes.PostConsensusPartialSig
	case spectypes.BNRoleSyncCommitteeContribution:
		return msgType == spectypes.PostConsensusPartialSig || msgType == spectypes.ContributionProofs
	case spectypes.BNRoleValidatorRegistration:
		return msgType == spectypes.ValidatorRegistrationPartialSig
	default:
		panic("invalid role") // role validity should be checked before
	}
}

func (mv *messageValidator) validPartialSignatures(share *ssvtypes.SSVShare, signedMsg *spectypes.SignedPartialSignatureMessage) error {
	if err := ssvtypes.VerifyByOperators(signedMsg.Signature, signedMsg, mv.netCfg.Domain, spectypes.PartialSignatureType, share.Committee); err != nil {
		signErr := ErrInvalidSignature
		signErr.innerErr = err
		signErr.got = fmt.Sprintf("domain %v from %v", hex.EncodeToString(mv.netCfg.Domain[:]), hex.EncodeToString(share.ValidatorPubKey))
		return signErr
	}

	for _, message := range signedMsg.Message.Messages {
		if err := mv.verifyPartialSignature(message, share); err != nil {
			return err
		}
	}

	return nil
}

func (mv *messageValidator) verifyPartialSignature(msg *spectypes.PartialSignatureMessage, share *ssvtypes.SSVShare) error {
	signer := msg.Signer
	signature := msg.PartialSignature
	root := msg.SigningRoot

	for _, n := range share.Committee {
		if n.GetID() != signer {
			continue
		}

		pk, err := ssvtypes.DeserializeBLSPublicKey(n.GetPublicKey())
		if err != nil {
			return fmt.Errorf("deserialize pk: %w", err)
		}
		sig := &bls.Sign{}
		if err := sig.Deserialize(signature); err != nil {
			return fmt.Errorf("deserialize signature: %w", err)
		}

		if !mv.aggregateVerify(sig, pk, root) {
			return ErrInvalidPartialSignature
		}

		return nil
	}

	return ErrSignerNotInCommittee
}

func (mv *messageValidator) aggregateVerify(sig *bls.Sign, pk bls.PublicKey, root [32]byte) bool {
	start := time.Now()

	valid := ssvtypes.Verifier.AggregateVerify(sig, []bls.PublicKey{pk}, root)

	sinceStart := time.Since(start)
	mv.metrics.SignatureValidationDuration(sinceStart)

	return valid
}

func (mv *messageValidator) validatePartialMessages(share *ssvtypes.SSVShare, m *spectypes.SignedPartialSignatureMessage) error {
	if err := mv.commonSignerValidation(m.Signer, share); err != nil {
		return err
	}

	if len(m.Message.Messages) == 0 {
		return ErrNoPartialMessages
	}

	seen := map[[32]byte]struct{}{}
	for _, message := range m.Message.Messages {
		if _, ok := seen[message.SigningRoot]; ok {
			return ErrDuplicatedPartialSignatureMessage
		}
		seen[message.SigningRoot] = struct{}{}

		if message.Signer != m.Signer {
			err := ErrUnexpectedSigner
			err.want = m.Signer
			err.got = message.Signer
			return err
		}

		if err := mv.commonSignerValidation(message.Signer, share); err != nil {
			return err
		}

		if err := mv.validateSignatureFormat(message.PartialSignature); err != nil {
			return err
		}
	}

	return nil
}

func (mv *messageValidator) validateSignerBehaviorPartial(
	state *ConsensusState,
	signer spectypes.OperatorID,
	share *ssvtypes.SSVShare,
	msgID spectypes.MessageID,
	signedMsg *spectypes.SignedPartialSignatureMessage,
) error {
	signerState := state.GetSignerState(signer)

	if signerState == nil {
		return nil
	}

	msgSlot := signedMsg.Message.Slot

	if msgSlot < signerState.Slot {
		// Signers aren't allowed to decrease their slot.
		// If they've sent a future message due to clock error,
		// this should be caught by the earlyMessage check.
		err := ErrSlotAlreadyAdvanced
		err.want = signerState.Slot
		err.got = msgSlot
		return err
	}

	newDutyInSameEpoch := false
	if msgSlot > signerState.Slot && mv.netCfg.Beacon.EstimatedEpochAtSlot(msgSlot) == mv.netCfg.Beacon.EstimatedEpochAtSlot(signerState.Slot) {
		newDutyInSameEpoch = true
	}

	if err := mv.validateDutyCount(signerState, msgID, newDutyInSameEpoch); err != nil {
		return err
	}

	if msgSlot <= signerState.Slot {
		limits := maxMessageCounts(len(share.Committee))
		if err := signerState.MessageCounts.ValidatePartialSignatureMessage(signedMsg, limits); err != nil {
			return err
		}
	}

	return nil
}
