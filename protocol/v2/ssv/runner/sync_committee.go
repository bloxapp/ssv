package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	ssz "github.com/ferranbt/fastssz"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/logging/fields"
	"github.com/bloxapp/ssv/protocol/v2/qbft/controller"
	"github.com/bloxapp/ssv/protocol/v2/ssv/runner/metrics"
)

type SyncCommitteeRunner struct {
	BaseRunner     *BaseRunner
	started        time.Time
	beacon         specssv.BeaconNode
	network        specssv.Network
	signer         spectypes.KeyManager
	operatorSigner spectypes.OperatorSigner
	valCheck       specqbft.ProposedValueCheckF

	metrics metrics.ConsensusMetrics
}

func NewSyncCommitteeRunner(
	beaconNetwork spectypes.BeaconNetwork,
	share *spectypes.Share,
	qbftController *controller.Controller,
	beacon specssv.BeaconNode,
	network specssv.Network,
	signer spectypes.KeyManager,
	operatorSigner spectypes.OperatorSigner,
	valCheck specqbft.ProposedValueCheckF,
	highestDecidedSlot phase0.Slot,
) Runner {
	return &SyncCommitteeRunner{
		BaseRunner: &BaseRunner{
			BeaconRoleType:     spectypes.BNRoleSyncCommittee,
			BeaconNetwork:      beaconNetwork,
			Share:              share,
			QBFTController:     qbftController,
			highestDecidedSlot: highestDecidedSlot,
		},

		beacon:         beacon,
		network:        network,
		signer:         signer,
		valCheck:       valCheck,
		operatorSigner: operatorSigner,

		metrics: metrics.NewConsensusMetrics(spectypes.BNRoleSyncCommittee),
	}
}

func (r *SyncCommitteeRunner) StartNewDuty(logger *zap.Logger, duty *spectypes.Duty) error {
	return r.BaseRunner.baseStartNewDuty(logger, r, duty)
}

// HasRunningDuty returns true if a duty is already running (StartNewDuty called and returned nil)
func (r *SyncCommitteeRunner) HasRunningDuty() bool {
	return r.BaseRunner.hasRunningDuty()
}

// executeDuty steps:
// 1) get sync block root from BN
// 2) start consensus on duty + block root data
// 3) Once consensus decides, sign partial block root and broadcast
// 4) collect 2f+1 partial sigs, reconstruct and broadcast valid sync committee sig to the BN
func (r *SyncCommitteeRunner) executeDuty(logger *zap.Logger, duty *spectypes.Duty) error {
	// TODO - waitOneThirdOrValidBlock
	r.started = time.Now()

	root, ver, err := r.GetBeaconNode().GetSyncMessageBlockRoot(duty.Slot)
	if err != nil {
		return errors.Wrap(err, "failed to get sync committee block root")
	}
	logger.Info("🧊 got sync message blockRoot", fields.BlockRootTime(time.Since(r.started)))

	r.metrics.StartDutyFullFlow()
	r.metrics.StartConsensus()

	input := &spectypes.ConsensusData{
		Duty:    *duty,
		Version: ver,
		DataSSZ: root[:],
	}

	if err := r.BaseRunner.decide(logger, r, input); err != nil {
		return errors.Wrap(err, "can't start new duty runner instance for duty")
	}
	return nil
}

func (r *SyncCommitteeRunner) ProcessPreConsensus(logger *zap.Logger, signedMsg *spectypes.SignedPartialSignatureMessage) error {
	return errors.New("no pre consensus sigs required for sync committee role")
}

func (r *SyncCommitteeRunner) ProcessConsensus(logger *zap.Logger, signedMsg *specqbft.SignedMessage) error {
	decided, decidedValue, err := r.BaseRunner.baseConsensusMsgProcessing(logger, r, signedMsg)
	if err != nil {
		return errors.Wrap(err, "failed processing consensus message")
	}

	// Decided returns true only once so if it is true it must be for the current running instance
	if !decided {
		return nil
	}

	duty := r.GetState().StartingDuty
	logger = logger.With(fields.Slot(duty.Slot))

	r.metrics.EndConsensus()
	logger.Debug("🧩 got decided", fields.QuorumTime(r.metrics.GetConsensusTime()))
	r.metrics.StartPostConsensus()
	startTime := time.Now()
	// specific duty sig
	root, err := decidedValue.GetSyncCommitteeBlockRoot()
	if err != nil {
		return errors.Wrap(err, "could not get sync committee block root")
	}
	msg, err := r.BaseRunner.signBeaconObject(r, spectypes.SSZBytes(root[:]), decidedValue.Duty.Slot, spectypes.DomainSyncCommittee)
	if err != nil {
		return errors.Wrap(err, "failed signing attestation data")
	}
	postConsensusMsg := &spectypes.PartialSignatureMessages{
		Type:     spectypes.PostConsensusPartialSig,
		Slot:     decidedValue.Duty.Slot,
		Messages: []*spectypes.PartialSignatureMessage{msg},
	}

	postSignedMsg, err := r.BaseRunner.signPostConsensusMsg(r, postConsensusMsg)
	if err != nil {
		return errors.Wrap(err, "could not sign post consensus msg")
	}

	data, err := postSignedMsg.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode post consensus signature msg")
	}

	ssvMsg := &spectypes.SSVMessage{
		MsgType: spectypes.SSVPartialSignatureMsgType,
		MsgID:   spectypes.NewMsgID(r.GetShare().DomainType, r.GetShare().ValidatorPubKey, r.BaseRunner.BeaconRoleType),
		Data:    data,
	}

	msgToBroadcast, err := spectypes.SSVMessageToSignedSSVMessage(ssvMsg, r.BaseRunner.Share.OperatorID, r.operatorSigner.SignSSVMessage)
	if err != nil {
		return errors.Wrap(err, "could not create SignedSSVMessage from SSVMessage")
	}

	if err := r.GetNetwork().Broadcast(ssvMsg.GetID(), msgToBroadcast); err != nil {
		return errors.Wrap(err, "can't broadcast partial post consensus sig")
	}
	logger.Info("✅ partial post consensus sig broadcast successfully", fields.BroadcastTime(time.Since(startTime)))
	return nil
}

func (r *SyncCommitteeRunner) ProcessPostConsensus(logger *zap.Logger, signedMsg *spectypes.SignedPartialSignatureMessage) error {
	quorum, roots, err := r.BaseRunner.basePostConsensusMsgProcessing(logger, r, signedMsg)
	if err != nil {
		return errors.Wrap(err, "failed processing post consensus message")
	}

	if !quorum {
		return nil
	}
	duty := r.GetState().StartingDuty
	logger = logger.With(fields.Slot(duty.Slot))
	blockRoot, err := r.GetState().DecidedValue.GetSyncCommitteeBlockRoot()
	if err != nil {
		return errors.Wrap(err, "could not get sync committee block root")
	}

	for _, root := range roots {
		sig, err := r.GetState().ReconstructBeaconSig(r.GetState().PostConsensusContainer, root, r.GetShare().ValidatorPubKey)
		if err != nil {
			// If the reconstructed signature verification failed, fall back to verifying each partial signature
			for _, root := range roots {
				r.BaseRunner.FallBackAndVerifyEachSignature(r.GetState().PostConsensusContainer, root)
			}
			return errors.Wrap(err, "got post-consensus quorum but it has invalid signatures")
		}
		specSig := phase0.BLSSignature{}
		copy(specSig[:], sig)
		r.metrics.EndPostConsensus()
		logger.Debug("🧩 reconstructed partial signatures",
			zap.Uint64s("signers", getPostConsensusSigners(r.GetState(), root)),
			fields.QuorumTime(r.metrics.GetPostConsensusTime()))

		messageSubmissionEnd := r.metrics.StartBeaconSubmission()
		startSubmissionTime := time.Now()
		msg := &altair.SyncCommitteeMessage{
			Slot:            r.GetState().DecidedValue.Duty.Slot,
			BeaconBlockRoot: blockRoot,
			ValidatorIndex:  r.GetState().DecidedValue.Duty.ValidatorIndex,
			Signature:       specSig,
		}

		if err := r.GetBeaconNode().SubmitSyncMessage(msg); err != nil {
			r.metrics.RoleSubmissionFailed()
			logger.Error("❌ failed to submit attestation",
				fields.ConsensusTime(time.Since(r.started)),
				fields.SubmissionTime(time.Since(startSubmissionTime)),
				zap.Duration("decided_time", r.metrics.GetPostConsensusTime()),
				zap.Error(err))
			return errors.Wrap(err, "could not submit to Beacon chain reconstructed signed sync committee")
		}

		messageSubmissionEnd()
		r.metrics.EndDutyFullFlow(r.GetState().RunningInstance.State.Round)
		r.metrics.RoleSubmitted()

		logger.Info("✅ successfully submitted sync committee",
			fields.Slot(msg.Slot),
			fields.ConsensusTime(time.Since(r.started)),
			fields.SubmissionTime(time.Since(startSubmissionTime)),
			zap.String("block_root", hex.EncodeToString(msg.BeaconBlockRoot[:])),
			fields.Height(r.BaseRunner.QBFTController.Height),
			fields.Round(r.GetState().RunningInstance.State.Round))
	}
	r.GetState().Finished = true

	return nil
}

func (r *SyncCommitteeRunner) expectedPreConsensusRootsAndDomain() ([]ssz.HashRoot, phase0.DomainType, error) {
	return []ssz.HashRoot{}, spectypes.DomainError, errors.New("no expected pre consensus roots for sync committee")
}

// expectedPostConsensusRootsAndDomain an INTERNAL function, returns the expected post-consensus roots to sign
func (r *SyncCommitteeRunner) expectedPostConsensusRootsAndDomain() ([]ssz.HashRoot, phase0.DomainType, error) {
	root, err := r.GetState().DecidedValue.GetSyncCommitteeBlockRoot()
	if err != nil {
		return nil, phase0.DomainType{}, errors.Wrap(err, "could not get sync committee block root")
	}

	return []ssz.HashRoot{spectypes.SSZBytes(root[:])}, spectypes.DomainSyncCommittee, nil
}

func (r *SyncCommitteeRunner) GetBaseRunner() *BaseRunner {
	return r.BaseRunner
}

func (r *SyncCommitteeRunner) GetNetwork() specssv.Network {
	return r.network
}

func (r *SyncCommitteeRunner) GetBeaconNode() specssv.BeaconNode {
	return r.beacon
}

func (r *SyncCommitteeRunner) GetShare() *spectypes.Share {
	return r.BaseRunner.Share
}

func (r *SyncCommitteeRunner) GetState() *State {
	return r.BaseRunner.State
}

func (r *SyncCommitteeRunner) GetValCheckF() specqbft.ProposedValueCheckF {
	return r.valCheck
}

func (r *SyncCommitteeRunner) GetSigner() spectypes.KeyManager {
	return r.signer
}

// Encode returns the encoded struct in bytes or error
func (r *SyncCommitteeRunner) Encode() ([]byte, error) {
	return json.Marshal(r)
}

// Decode returns error if decoding failed
func (r *SyncCommitteeRunner) Decode(data []byte) error {
	return json.Unmarshal(data, &r)
}

// GetRoot returns the root used for signing and verification
func (r *SyncCommitteeRunner) GetRoot() ([32]byte, error) {
	marshaledRoot, err := r.Encode()
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not encode DutyRunnerState")
	}
	ret := sha256.Sum256(marshaledRoot)
	return ret, nil
}
