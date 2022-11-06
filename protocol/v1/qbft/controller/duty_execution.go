package controller

import (
	"encoding/hex"
	"time"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v1/message"
	"github.com/bloxapp/ssv/protocol/v1/qbft/msgqueue"
)

// ProcessPostConsensusMessage aggregates partial signature messages and broadcasting when quorum achieved
func (c *Controller) ProcessPostConsensusMessage(msg *specssv.SignedPartialSignatureMessage) error {
	if c.SignatureState.getState() != StateRunning {
		c.Logger.Warn(
			"trying to process post consensus signature message but timer state is not running. can't process message.",
			zap.String("state", c.SignatureState.getState().toString()),
		)
		return nil
	}

	// adjust share committee to the spec
	committee := make([]*spectypes.Operator, 0)
	for _, node := range c.ValidatorShare.Committee {
		committee = append(committee, &spectypes.Operator{
			OperatorID: spectypes.OperatorID(node.IbftID),
			PubKey:     node.Pk,
		})
	}

	if err := message.ValidatePartialSigMsg(msg, committee, c.SignatureState.duty.Slot); err != nil {
		return errors.WithMessage(err, "could not validate partial signature message")
	}
	logger := c.Logger.With(zap.Uint64("signer_id", uint64(msg.GetSigners()[0])))
	logger.Info("received valid partial signature message",
		zap.String("msg signature", hex.EncodeToString(msg.GetSignature())),
		zap.String("msg beacon signature", hex.EncodeToString(msg.Messages[0].PartialSignature)),
		zap.Any("msg", msg),
	)

	//	check if already exist, if so, ignore
	if _, found := c.SignatureState.signatures[msg.GetSigners()[0]]; found {
		c.Logger.Debug("sig already known, skip")
		return nil
	}

	c.SignatureState.signatures[msg.GetSigners()[0]] = msg.Messages[0].PartialSignature
	if len(c.SignatureState.signatures) >= c.SignatureState.sigCount {
		c.Logger.Info("collected enough signature to reconstruct",
			zap.Int("signatures", len(c.SignatureState.signatures)),
		)
		c.SignatureState.stopTimer()

		// clean queue consensus & default messages that is <= c.signatureState.duty.Slot, we don't need them anymore
		c.Q.Clean(
			msgqueue.SignedPostConsensusMsgCleaner(message.ToMessageID(c.Identifier), c.SignatureState.duty.Slot),
		)

		err := c.broadcastSignature()
		c.SignatureState.clear()
		return err
	}
	return nil
}

// broadcastSignature reconstruct sigs and broadcast to network
func (c *Controller) broadcastSignature() error {
	// Reconstruct signatures
	if err := c.reconstructAndBroadcastSignature(c.SignatureState.signatures, c.SignatureState.root, c.SignatureState.valueStruct, c.SignatureState.duty); err != nil {
		return errors.Wrap(err, "failed to reconstruct and broadcast signature")
	}
	c.Logger.Info("Successfully submitted role!")

	metricsTimeFullSubmissionFlow.WithLabelValues(c.ValidatorShare.PublicKey.SerializeToHexStr()).
		Set(time.Since(c.instanceStartTime).Seconds())

	return nil
}

// PostConsensusDutyExecution signs the eth2 duty after iBFT came to consensus and start signature state
func (c *Controller) PostConsensusDutyExecution(logger *zap.Logger, decidedValue []byte, signaturesCount int, role spectypes.BeaconRole) error {
	c.postConsensusStartTime = time.Now()

	// sign input value and broadcast
	sig, root, valueStruct, duty, err := c.signDuty(logger, decidedValue, role)
	if err != nil {
		return errors.Wrap(err, "failed to sign input data")
	}
	psm, err := c.generatePartialSignatureMessage(sig, root, duty.Slot)
	if err != nil {
		return errors.Wrap(err, "failed to generate sig message")
	}
	if err := c.signAndBroadcast(logger, psm); err != nil {
		return errors.Wrap(err, "failed to sign and broadcast post consensus")
	}

	//	start timer, clear new map and set var's
	c.SignatureState.start(c.Logger, signaturesCount, root, valueStruct, duty)
	return nil
}

// signAndBroadcast checks and adds the signed message to the appropriate round state type
func (c *Controller) signAndBroadcast(logger *zap.Logger, psm specssv.PartialSignatureMessages) error {
	pk, err := c.ValidatorShare.OperatorSharePubKey()
	if err != nil {
		return errors.Wrap(err, "failed to get operator share pubkey")
	}
	signature, err := c.KeyManager.SignRoot(psm, spectypes.PartialSignatureType, pk.Serialize())
	if err != nil {
		return errors.Wrap(err, "failed to sign message")
	}

	signedMsg := &specssv.SignedPartialSignatureMessage{
		Type:      specssv.PostConsensusPartialSig,
		Messages:  psm,
		Signature: signature,
		Signers:   []spectypes.OperatorID{c.ValidatorShare.NodeID},
	}

	encodedSignedMsg, err := signedMsg.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode signed message")
	}
	ssvMsg := spectypes.SSVMessage{
		MsgType: spectypes.SSVPartialSignatureMsgType,
		MsgID:   message.ToMessageID(c.GetIdentifier()),
		Data:    encodedSignedMsg,
	}

	if err := c.Network.Broadcast(ssvMsg); err != nil {
		return errors.Wrap(err, "failed to broadcast signature")
	}
	logger.Info("broadcasting partial signature post consensus")
	return nil
}

// generatePartialSignatureMessage returns a PartialSignatureMessage struct
func (c *Controller) generatePartialSignatureMessage(sig []byte, root []byte, slot spec.Slot) (specssv.PartialSignatureMessages, error) {
	signers := []spectypes.OperatorID{c.ValidatorShare.NodeID}
	return specssv.PartialSignatureMessages{
		&specssv.PartialSignatureMessage{
			Slot:             slot,
			PartialSignature: sig,
			SigningRoot:      root,
			Signers:          signers,
		},
	}, nil
}
