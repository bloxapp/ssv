package runner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/api"
	eth2apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/qbft"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/cornelk/hashmap"
	ssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/qbft/controller"
	"github.com/bloxapp/ssv/protocol/v2/types"
)

type ValidatorRegistrationRunner struct {
	BaseRunner *BaseRunner

	beacon                  specssv.BeaconNode
	network                 specssv.Network
	signer                  spectypes.KeyManager
	valCheck                qbft.ProposedValueCheckF
	lastRegistered          *hashmap.Map[string, phase0.Epoch]
	registrationSubmissions chan api.VersionedSignedValidatorRegistration
	submissionErrors        *hashmap.Map[string, chan error] // TODO: implement
}

const (
	submissionsLimit   = 500
	submissionInterval = 500 * time.Millisecond
	batchSubmission    = false // experimental
)

func NewValidatorRegistrationRunner(
	beaconNetwork spectypes.BeaconNetwork,
	share *spectypes.Share,
	qbftController *controller.Controller,
	beacon specssv.BeaconNode,
	network specssv.Network,
	signer spectypes.KeyManager,
) Runner {
	r := &ValidatorRegistrationRunner{
		BaseRunner: &BaseRunner{
			BeaconRoleType: spectypes.BNRoleValidatorRegistration,
			BeaconNetwork:  beaconNetwork,
			Share:          share,
			QBFTController: qbftController,
		},

		beacon:                  beacon,
		network:                 network,
		signer:                  signer,
		lastRegistered:          hashmap.New[string, phase0.Epoch](),
		submissionErrors:        hashmap.New[string, chan error](),
		registrationSubmissions: make(chan api.VersionedSignedValidatorRegistration, submissionsLimit),
	}

	if batchSubmission {
		go r.startRegistrationSubmitter()
	}

	return r
}

func (r *ValidatorRegistrationRunner) startRegistrationSubmitter() {
	t := time.NewTicker(submissionInterval)
	defer t.Stop()

	registrationList := make([]*api.VersionedSignedValidatorRegistration, 0, submissionsLimit)
	errorChanList := make([]chan error, 0)

	submit := func() {
		if err := r.beacon.SubmitValidatorRawRegistrations(registrationList); err != nil {
			for _, ch := range errorChanList {
				go func(ch chan<- error, err error) {
					ch <- err
					close(ch)
				}(ch, err)
			}
		}
		registrationList = make([]*api.VersionedSignedValidatorRegistration, 0, submissionsLimit)
	}

	for {
		select {
		case <-t.C:
			submit()
		case registration := <-r.registrationSubmissions:
			registrationList = append(registrationList, &registration)
			if h, err := r.hashRegistration(registration); err != nil {
				// TODO: handle error
			} else if ch, ok := r.submissionErrors.Get(string(h)); ok && ch != nil {
				errorChanList = append(errorChanList, ch)
			}

			if len(registrationList) >= submissionsLimit {
				submit()
				t.Reset(submissionInterval)
			}
		}
	}
}

func (r *ValidatorRegistrationRunner) submitRegistration(pubkey []byte, feeRecipient bellatrix.ExecutionAddress, sig phase0.BLSSignature) error {
	registration := api.VersionedSignedValidatorRegistration{
		Version: spec.BuilderVersionV1,
		V1: &eth2apiv1.SignedValidatorRegistration{
			Message: &eth2apiv1.ValidatorRegistration{
				FeeRecipient: feeRecipient,
				// TODO: This is a reasonable default, but we should probably make this configurable.
				//       Discussion here: https://github.com/ethereum/builder-specs/issues/17
				GasLimit:  30_000_000,
				Timestamp: r.beacon.GetBeaconNetwork().EpochStartTime(r.beacon.GetBeaconNetwork().EstimatedCurrentEpoch()),
				Pubkey:    *(*phase0.BLSPubKey)(pubkey),
			},
			Signature: sig,
		},
	}

	h, err := r.hashRegistration(registration)
	if err != nil {
		return fmt.Errorf("hash registration: %w", err)
	}

	errCh := make(chan error, 1)
	r.submissionErrors.Set(string(h), errCh)

	go func() {
		r.registrationSubmissions <- registration
	}()

	waitDuration := submissionInterval * 2
	t := time.NewTimer(waitDuration)
	defer t.Stop()

	select {
	case err := <-errCh:
		return err
	case <-t.C:
		return fmt.Errorf("timeout waiting for registration submission result")
	}
}

func (r *ValidatorRegistrationRunner) hashRegistration(registration api.VersionedSignedValidatorRegistration) ([]byte, error) {
	bytes, err := json.Marshal(registration)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256(bytes)
	return h[:], nil
}

func (r *ValidatorRegistrationRunner) StartNewDuty(logger *zap.Logger, duty *spectypes.Duty) error {
	return r.BaseRunner.baseStartNewDuty(logger, r, duty)
}

// HasRunningDuty returns true if a duty is already running (StartNewDuty called and returned nil)
func (r *ValidatorRegistrationRunner) HasRunningDuty() bool {
	return r.BaseRunner.hasRunningDuty()
}

// TODO: define a common constant
const validatorRegistrationEpochInterval = 10

func (r *ValidatorRegistrationRunner) ProcessPreConsensus(logger *zap.Logger, signedMsg *spectypes.SignedPartialSignatureMessage) error {
	quorum, roots, err := r.BaseRunner.basePreConsensusMsgProcessing(r, signedMsg)
	if err != nil {
		return errors.Wrap(err, "failed processing validator registration message")
	}

	// quorum returns true only once (first time quorum achieved)
	if !quorum {
		return nil
	}

	currentEpoch := r.beacon.GetBeaconNetwork().EstimatedCurrentEpoch()
	pubKey := r.GetShare().ValidatorPubKey
	if prevEpoch, ok := r.lastRegistered.Get(string(pubKey)); ok && currentEpoch-prevEpoch >= validatorRegistrationEpochInterval {
		// only 1 root, verified in basePreConsensusMsgProcessing
		root := roots[0]
		// randao is relevant only for block proposals, no need to check type
		fullSig, err := r.GetState().ReconstructBeaconSig(r.GetState().PreConsensusContainer, root, pubKey)
		if err != nil {
			return errors.Wrap(err, "could not reconstruct randao sig")
		}
		specSig := phase0.BLSSignature{}
		copy(specSig[:], fullSig)

		if batchSubmission {
			if err := r.submitRegistration(r.BaseRunner.Share.ValidatorPubKey, r.BaseRunner.Share.FeeRecipientAddress, specSig); err != nil {
				return errors.Wrap(err, "could not submit batched validator registration")
			}
		} else {
			if err := r.beacon.SubmitValidatorRegistration(r.BaseRunner.Share.ValidatorPubKey, r.BaseRunner.Share.FeeRecipientAddress, specSig); err != nil {
				return errors.Wrap(err, "could not submit validator registration")
			}
		}

		// not reusing epoch in case if it was changed
		r.lastRegistered.Set(string(pubKey), r.beacon.GetBeaconNetwork().EstimatedCurrentEpoch())

		logger.Debug("validator registration submitted successfully")
	} else {
		logger.Debug("not registering validator: recently registered",
			zap.Uint64("current_epoch", uint64(currentEpoch)),
			zap.Uint64("prev_epoch", uint64(prevEpoch)),
		)
	}

	r.GetState().Finished = true
	return nil
}

func (r *ValidatorRegistrationRunner) ProcessConsensus(logger *zap.Logger, signedMsg *qbft.SignedMessage) error {
	return errors.New("no consensus phase for validator registration")
}

func (r *ValidatorRegistrationRunner) ProcessPostConsensus(logger *zap.Logger, signedMsg *spectypes.SignedPartialSignatureMessage) error {
	return errors.New("no post consensus phase for validator registration")
}

func (r *ValidatorRegistrationRunner) expectedPreConsensusRootsAndDomain() ([]ssz.HashRoot, phase0.DomainType, error) {
	vr, err := r.calculateValidatorRegistration()
	if err != nil {
		return nil, spectypes.DomainError, errors.Wrap(err, "could not calculate validator registration")
	}
	return []ssz.HashRoot{vr}, spectypes.DomainApplicationBuilder, nil
}

// expectedPostConsensusRootsAndDomain an INTERNAL function, returns the expected post-consensus roots to sign
func (r *ValidatorRegistrationRunner) expectedPostConsensusRootsAndDomain() ([]ssz.HashRoot, phase0.DomainType, error) {
	return nil, [4]byte{}, errors.New("no post consensus roots for validator registration")
}

func (r *ValidatorRegistrationRunner) executeDuty(logger *zap.Logger, duty *spectypes.Duty) error {
	vr, err := r.calculateValidatorRegistration()
	if err != nil {
		return errors.Wrap(err, "could not calculate validator registration")
	}

	// sign partial randao
	msg, err := r.BaseRunner.signBeaconObject(r, vr, duty.Slot, spectypes.DomainApplicationBuilder)
	if err != nil {
		return errors.Wrap(err, "could not sign validator registration")
	}
	msgs := spectypes.PartialSignatureMessages{
		Type:     spectypes.ValidatorRegistrationPartialSig,
		Slot:     duty.Slot,
		Messages: []*spectypes.PartialSignatureMessage{msg},
	}

	// sign msg
	signature, err := r.GetSigner().SignRoot(msgs, spectypes.PartialSignatureType, r.GetShare().SharePubKey)
	if err != nil {
		return errors.Wrap(err, "could not sign randao msg")
	}
	signedPartialMsg := &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: signature,
		Signer:    r.GetShare().OperatorID,
	}

	// broadcast
	data, err := signedPartialMsg.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode randao pre-consensus signature msg")
	}
	msgToBroadcast := &spectypes.SSVMessage{
		MsgType: spectypes.SSVPartialSignatureMsgType,
		MsgID:   spectypes.NewMsgID(types.GetDefaultDomain(), r.GetShare().ValidatorPubKey, r.BaseRunner.BeaconRoleType),
		Data:    data,
	}
	if err := r.GetNetwork().Broadcast(msgToBroadcast); err != nil {
		return errors.Wrap(err, "can't broadcast partial randao sig")
	}
	return nil
}

func (r *ValidatorRegistrationRunner) calculateValidatorRegistration() (*eth2apiv1.ValidatorRegistration, error) {
	pk := phase0.BLSPubKey{}
	copy(pk[:], r.BaseRunner.Share.ValidatorPubKey)

	epoch := r.BaseRunner.BeaconNetwork.EstimatedEpochAtSlot(r.BaseRunner.State.StartingDuty.Slot)

	return &eth2apiv1.ValidatorRegistration{
		FeeRecipient: r.BaseRunner.Share.FeeRecipientAddress,
		GasLimit:     30_000_000,
		Timestamp:    r.BaseRunner.BeaconNetwork.EpochStartTime(epoch),
		Pubkey:       pk,
	}, nil
}

func (r *ValidatorRegistrationRunner) GetBaseRunner() *BaseRunner {
	return r.BaseRunner
}

func (r *ValidatorRegistrationRunner) GetNetwork() specssv.Network {
	return r.network
}

func (r *ValidatorRegistrationRunner) GetBeaconNode() specssv.BeaconNode {
	return r.beacon
}

func (r *ValidatorRegistrationRunner) GetShare() *spectypes.Share {
	return r.BaseRunner.Share
}

func (r *ValidatorRegistrationRunner) GetState() *State {
	return r.BaseRunner.State
}

func (r *ValidatorRegistrationRunner) GetValCheckF() qbft.ProposedValueCheckF {
	return r.valCheck
}

func (r *ValidatorRegistrationRunner) GetSigner() spectypes.KeyManager {
	return r.signer
}

// Encode returns the encoded struct in bytes or error
func (r *ValidatorRegistrationRunner) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Decode returns error if decoding failed
func (r *ValidatorRegistrationRunner) Decode(data []byte) error {
	return json.Unmarshal(data, &r)
}

// GetRoot returns the root used for signing and verification
func (r *ValidatorRegistrationRunner) GetRoot() ([32]byte, error) {
	marshaledRoot, err := r.Encode()
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not encode DutyRunnerState")
	}
	ret := sha256.Sum256(marshaledRoot)
	return ret, nil
}
