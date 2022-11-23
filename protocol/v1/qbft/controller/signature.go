package controller

import (
	"encoding/base64"
	"time"

	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/utils/threshold"
)

// TimerState is the state of the timer.
type TimerState int

// set of timer states.
const (
	StateSleep = iota
	StateRunning
	StateTimeout
)

func (s TimerState) toString() string {
	switch s {
	case StateSleep:
		return "sleep"
	case StateRunning:
		return "running"
	case StateTimeout:
		return "timeout"
	}
	return ""
}

// SignatureState represents the signature state.
type SignatureState struct {
	timer      *time.Timer
	state      atomic.Int32
	signatures map[spectypes.OperatorID][]byte

	SignatureCollectionTimeout time.Duration
	sigCount                   int
	root                       []byte
	valueStruct                *beaconprotocol.DutyData
	duty                       *spectypes.Duty
	lastSlot                   spec.Slot // for queue in order the know the last slot
}

func (s *SignatureState) start(logger *zap.Logger, signaturesCount int, root []byte, valueStruct *beaconprotocol.DutyData, duty *spectypes.Duty) {
	// set var's
	s.sigCount = signaturesCount
	s.root = root
	s.valueStruct = valueStruct
	s.duty = duty
	s.lastSlot = duty.Slot

	// start timer
	s.timer = time.AfterFunc(s.SignatureCollectionTimeout, func() {
		if !s.state.CAS(StateRunning, StateTimeout) {
			logger.Debug("signatures were collected before timeout", zap.Int("received", len(s.signatures)))
			return
		}
		logger.Warn("could not process post consensus signature", zap.Error(errors.Errorf("timed out waiting for post consensus signatures, received %d", len(s.signatures))))
	})
	//s.timer = time.NewTimer(s.SignatureCollectionTimeout)
	s.state.Store(StateRunning)
	// init map
	s.signatures = make(map[spectypes.OperatorID][]byte, s.sigCount)
}

// stopTimer stops timer from firing and drain the channel. also set state to sleep
func (s *SignatureState) stopTimer() {
	s.state.Store(StateSleep)
	s.timer.Stop()
}

func (s *SignatureState) clear() {
	s.sigCount = 0
	s.root = nil
	s.valueStruct = nil
	s.duty = nil
	s.state.Store(StateSleep)
	// don't reset height until new height set
}

func (s *SignatureState) getState() TimerState {
	return TimerState(s.state.Load())
}

func (c *Controller) verifyPartialSignature(signature []byte, root []byte, ibftID spectypes.OperatorID, committiee map[spectypes.OperatorID]*beaconprotocol.Node) error {
	if val, found := committiee[ibftID]; found {
		pk := &bls.PublicKey{}
		if err := pk.Deserialize(val.Pk); err != nil {
			return errors.Wrap(err, "could not deserialized pk")
		}
		sig := &bls.Sign{}
		if err := sig.Deserialize(signature); err != nil {
			return errors.Wrap(err, "could not deserialized signature")
		}

		// protect nil root
		root = ensureRoot(root)
		// verify
		if !sig.VerifyByte(pk, root) {
			return errors.Errorf("could not verify signature from iBFT member %d", ibftID)
		}
		return nil
	}
	return errors.Errorf("could not find iBFT member %d", ibftID)
}

// signDuty signs the duty after iBFT came to consensus
func (c *Controller) signDuty(logger *zap.Logger, decidedValue []byte, role spectypes.BeaconRole) ([]byte, []byte, *beaconprotocol.DutyData, *spectypes.Duty, error) {
	// get operator pk for sig
	pk, err := c.ValidatorShare.OperatorSharePubKey()
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "could not find operator pk for signing duty")
	}

	// sign input value
	var sig []byte
	var root []byte
	retValueStruct := &beaconprotocol.DutyData{}
	var duty *spectypes.Duty
	switch role {
	case spectypes.BNRoleAttester:
		s := &spectypes.ConsensusData{}
		// TODO(olegshmuelov): use SSZ decoding
		// TODO(olegshmuelov): validate the consensus data using the spec "BeaconAttestationValueCheck"
		if err := s.Decode(decidedValue); err != nil {
			logger.Warn("failed to decode consensus data", zap.Int("len", len(decidedValue)), zap.Error(err))
			return nil, nil, nil, nil, errors.Wrap(err, "failed to decode consensus data")
		}
		logger.Debug("decoded consensus data", zap.Any("data", s), zap.Int("len", len(decidedValue)))
		duty = s.Duty
		signedAttestation, r, err := c.KeyManager.SignAttestation(s.AttestationData, duty, pk.Serialize())
		if err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "failed to sign attestation")
		}

		sg := &beaconprotocol.InputValueAttestation{Attestation: signedAttestation}
		retValueStruct.SignedData = sg
		retValueStruct.GetAttestation().Signature = signedAttestation.Signature
		retValueStruct.GetAttestation().AggregationBits = signedAttestation.AggregationBits
		sig = signedAttestation.Signature[:]
		root = ensureRoot(r)
	default:
		return nil, nil, nil, nil, errors.New("unsupported role, can't sign")
	}
	return sig, root, retValueStruct, duty, err
}

// reconstructAndBroadcastSignature reconstructs the received signatures from other
// nodes and broadcasts the reconstructed signature to the beacon-chain
func (c *Controller) reconstructAndBroadcastSignature(signatures map[spectypes.OperatorID][]byte, root []byte, inputValue *beaconprotocol.DutyData, duty *spectypes.Duty) error {
	// Reconstruct signatures
	signature, err := threshold.ReconstructSignatures(signatures)
	if err != nil {
		return errors.Wrap(err, "failed to reconstruct signatures")
	}
	// verify reconstructed sig
	if res := signature.VerifyByte(c.ValidatorShare.PublicKey, root); !res {
		return errors.New("could not reconstruct a valid signature")
	}

	metricsSignatureCollectionDuration.WithLabelValues(c.ValidatorShare.PublicKey.SerializeToHexStr()).
		Observe(time.Since(c.postConsensusStartTime).Seconds())

	c.Logger.Info("signatures successfully reconstructed", zap.String("signature", base64.StdEncoding.EncodeToString(signature.Serialize())), zap.Int("signature count", len(signatures)))

	// Submit validation to beacon node
	switch duty.Type {
	case spectypes.BNRoleAttester:
		c.Logger.Debug("submitting attestation")
		blsSig := spec.BLSSignature{}
		copy(blsSig[:], signature.Serialize()[:])
		inputValue.GetAttestation().Signature = blsSig

		attestationSubmissionStartTime := time.Now()
		if err := c.Beacon.SubmitAttestation(inputValue.GetAttestation()); err != nil {
			return errors.Wrap(err, "failed to broadcast attestation")
		}
		metricsAttestationSubmissionDuration.WithLabelValues(c.ValidatorShare.PublicKey.SerializeToHexStr()).
			Observe(time.Since(attestationSubmissionStartTime).Seconds())
	default:
		return errors.New("role is undefined, can't reconstruct signature")
	}
	return nil
}

// ensureRoot ensures that root will have sufficient allocated memory
// otherwise we get panic from bls:
// github.com/herumi/bls-eth-go-binary/bls.(*Sign).VerifyByte:738
func ensureRoot(root []byte) []byte {
	n := len(root)
	if n == 0 {
		n = 1
	}
	tmp := make([]byte, n)
	copy(tmp[:], root[:])
	return tmp[:]
}
