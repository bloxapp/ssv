package testing

import (
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/herumi/bls-eth-go-binary/bls"
)

var AttesterMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleAttester)
	return ret[:]
}()

var ProposerMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleProposer)
	return ret[:]
}()
var AggregatorMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleAggregator)
	return ret[:]
}()
var SyncCommitteeMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleSyncCommittee)
	return ret[:]
}()
var SyncCommitteeContributionMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleSyncCommitteeContribution)
	return ret[:]
}()
var ValidatorRegistrationMsgID = func() []byte {
	ret := spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleValidatorRegistration)
	return ret[:]
}()

var TestAttesterConsensusData = &spectypes.ConsensusData{
	Duty:            testingutils.TestingAttesterDuty,
	AttestationData: testingutils.TestingAttestationData,
}
var TestAttesterConsensusDataByts, _ = TestAttesterConsensusData.Encode()

var TestAggregatorConsensusData = &spectypes.ConsensusData{
	Duty:              testingutils.TestingAggregatorDuty,
	AggregateAndProof: testingutils.TestingAggregateAndProof,
}
var TestAggregatorConsensusDataByts, _ = TestAggregatorConsensusData.Encode()

var TestProposerConsensusData = &spectypes.ConsensusData{
	Duty:      testingutils.TestingProposerDuty,
	BlockData: testingutils.TestingBeaconBlock,
}
var TestProposerConsensusDataByts, _ = TestProposerConsensusData.Encode()

var TestProposerBlindedBlockConsensusData = &spectypes.ConsensusData{
	Duty:             testingutils.TestingProposerDuty,
	BlindedBlockData: testingutils.TestingBlindedBeaconBlock,
}
var TestProposerBlindedBlockConsensusDataByts, _ = TestProposerBlindedBlockConsensusData.Encode()

var TestSyncCommitteeConsensusData = &spectypes.ConsensusData{
	Duty:                   testingutils.TestingSyncCommitteeDuty,
	SyncCommitteeBlockRoot: testingutils.TestingSyncCommitteeBlockRoot,
}
var TestSyncCommitteeConsensusDataByts, _ = TestSyncCommitteeConsensusData.Encode()

var TestSyncCommitteeContributionConsensusData = &spectypes.ConsensusData{
	Duty: testingutils.TestingSyncCommitteeContributionDuty,
	SyncCommitteeContribution: map[phase0.BLSSignature]*altair.SyncCommitteeContribution{
		testingutils.TestingContributionProofsSigned[0]: testingutils.TestingSyncCommitteeContributions[0],
		testingutils.TestingContributionProofsSigned[1]: testingutils.TestingSyncCommitteeContributions[1],
		testingutils.TestingContributionProofsSigned[2]: testingutils.TestingSyncCommitteeContributions[2],
	},
}
var TestSyncCommitteeContributionConsensusDataByts, _ = TestSyncCommitteeContributionConsensusData.Encode()

var TestConsensusUnkownDutyTypeData = &spectypes.ConsensusData{
	Duty:            testingutils.TestingUnknownDutyType,
	AttestationData: testingutils.TestingAttestationData,
}
var TestConsensusUnkownDutyTypeDataByts, _ = TestConsensusUnkownDutyTypeData.Encode()

var TestConsensusWrongDutyPKData = &spectypes.ConsensusData{
	Duty:            testingutils.TestingWrongDutyPK,
	AttestationData: testingutils.TestingAttestationData,
}
var TestConsensusWrongDutyPKDataByts, _ = TestConsensusWrongDutyPKData.Encode()

var SSVMsgAttester = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleAttester))
}

var SSVMsgWrongID = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingWrongValidatorPubKey[:], spectypes.BNRoleAttester))
}

var SSVMsgProposer = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleProposer))
}

var SSVMsgAggregator = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleAggregator))
}

var SSVMsgSyncCommittee = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleSyncCommittee))
}

var SSVMsgSyncCommitteeContribution = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleSyncCommitteeContribution))
}

var SSVMsgValidatorRegistration = func(qbftMsg *specqbft.SignedMessage, partialSigMsg *spectypes.SignedPartialSignatureMessage) *spectypes.SSVMessage {
	return ssvMsg(qbftMsg, partialSigMsg, spectypes.NewMsgID(testingutils.TestingValidatorPubKey[:], spectypes.BNRoleValidatorRegistration))
}

var ssvMsg = func(qbftMsg *specqbft.SignedMessage, postMsg *spectypes.SignedPartialSignatureMessage, msgID spectypes.MessageID) *spectypes.SSVMessage {
	var msgType spectypes.MsgType
	var data []byte
	if qbftMsg != nil {
		msgType = spectypes.SSVConsensusMsgType
		data, _ = qbftMsg.Encode()
	} else if postMsg != nil {
		msgType = spectypes.SSVPartialSignatureMsgType
		data, _ = postMsg.Encode()
	} else {
		panic("msg type undefined")
	}

	return &spectypes.SSVMessage{
		MsgType: msgType,
		MsgID:   msgID,
		Data:    data,
	}
}

var PostConsensusAttestationMsgWithWrongSig = func(sk *bls.SecretKey, id spectypes.OperatorID, height specqbft.Height) *spectypes.SignedPartialSignatureMessage {
	return postConsensusAttestationMsg(sk, id, height, true, false)
}

var PostConsensusAttestationMsgWithWrongRoot = func(sk *bls.SecretKey, id spectypes.OperatorID, height specqbft.Height) *spectypes.SignedPartialSignatureMessage {
	return postConsensusAttestationMsg(sk, id, height, true, false)
}

var PostConsensusAttestationMsg = func(sk *bls.SecretKey, id spectypes.OperatorID, height specqbft.Height) *spectypes.SignedPartialSignatureMessage {
	return postConsensusAttestationMsg(sk, id, height, false, false)
}

var postConsensusAttestationMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	height specqbft.Height,
	wrongRoot bool,
	wrongBeaconSig bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(testingutils.TestingAttestationData.Target.Epoch, spectypes.DomainAttester)
	signed, root, _ := signer.SignBeaconObject(testingutils.TestingAttestationData, d, sk.GetPublicKey().Serialize(), spectypes.DomainAttester)

	if wrongBeaconSig {
		signed, _, _ = signer.SignBeaconObject(testingutils.TestingAttestationData, d, testingutils.TestingWrongValidatorPubKey[:], spectypes.DomainAttester)
	}

	if wrongRoot {
		root = []byte{1, 2, 3, 4}
	}

	msgs := specssv.PartialSignatureMessages{
		Type: specssv.PostConsensusPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed,
				SigningRoot:      root,
				Signer:           id,
			},
		},
	}
	sig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: sig,
		Signer:    id,
	}
}

var PostConsensusProposerMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return postConsensusBeaconBlockMsg(sk, id, false, false)
}

var postConsensusBeaconBlockMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	wrongRoot bool,
	wrongBeaconSig bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()

	d, _ := beacon.DomainData(1, spectypes.DomainProposer) // epoch doesn't matter here, hard coded
	sig, root, _ := signer.SignBeaconObject(testingutils.TestingBeaconBlock, d, sk.GetPublicKey().Serialize(), spectypes.DomainProposer)
	blsSig := phase0.BLSSignature{}
	copy(blsSig[:], sig)

	signed := bellatrix.SignedBeaconBlock{
		Message:   testingutils.TestingBeaconBlock,
		Signature: blsSig,
	}

	if wrongBeaconSig {
		// signed, _, _ = signer.SignAttestation(testingutils.TestingAttestationData, testingutils.TestingAttesterDuty, testingutils.TestingWrongSK.GetPublicKey().Serialize())
		panic("implement")
	}

	if wrongRoot {
		root = []byte{1, 2, 3, 4}
	}

	msgs := specssv.PartialSignatureMessages{
		Type: specssv.PostConsensusPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed.Signature[:],
				SigningRoot:      root,
				Signer:           id,
			},
		},
	}
	msgSig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: msgSig,
		Signer:    id,
	}
}

var PreConsensusFailedMsg = func(
	msgSigner *bls.SecretKey,
	msgSignerID spectypes.OperatorID,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(testingutils.TestingDutyEpoch, spectypes.DomainRandao)
	signed, root, _ := signer.SignBeaconObject(spectypes.SSZUint64(testingutils.TestingDutyEpoch), d, msgSigner.GetPublicKey().Serialize(), spectypes.DomainRandao)

	msg := specssv.PartialSignatureMessages{
		Type: specssv.RandaoPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed[:],
				SigningRoot:      root,
				Signer:           msgSignerID,
			},
		},
	}
	sig, _ := signer.SignRoot(msg, spectypes.PartialSignatureType, msgSigner.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msg,
		Signature: sig,
		Signer:    msgSignerID,
	}
}

var PreConsensusRandaoMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch, testingutils.TestingDutySlot, 1)
}

// PreConsensusRandaoNextEpochMsg testing for a second duty start
var PreConsensusRandaoNextEpochMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch2, testingutils.TestingDutySlot2, 1)
}

var PreConsensusRandaoDifferentEpochMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch+1, testingutils.TestingDutySlot, 1)
}

var PreConsensusRandaoWrongSlotMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch, testingutils.TestingDutySlot+1, 1)
}

var PreConsensusRandaoMultiMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch, testingutils.TestingDutySlot, 2)
}

var PreConsensusRandaoNoMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return randaoMsg(sk, id, false, testingutils.TestingDutyEpoch, testingutils.TestingDutySlot, 0)
}

var PreConsensusRandaoDifferentSignerMsg = func(msgSigner, randaoSigner *bls.SecretKey, msgSignerID, randaoSignerID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(testingutils.TestingDutyEpoch, spectypes.DomainRandao)
	signed, root, _ := signer.SignBeaconObject(spectypes.SSZUint64(testingutils.TestingDutyEpoch), d, randaoSigner.GetPublicKey().Serialize(), spectypes.DomainRandao)

	msg := specssv.PartialSignatureMessages{
		Type: specssv.RandaoPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed[:],
				SigningRoot:      root,
				Signer:           randaoSignerID,
			},
		},
	}
	sig, _ := signer.SignRoot(msg, spectypes.PartialSignatureType, msgSigner.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msg,
		Signature: sig,
		Signer:    msgSignerID,
	}
}

var randaoMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	wrongRoot bool,
	epoch phase0.Epoch,
	slot phase0.Slot,
	msgCnt int,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(epoch, spectypes.DomainRandao)
	signed, root, _ := signer.SignBeaconObject(spectypes.SSZUint64(epoch), d, sk.GetPublicKey().Serialize(), spectypes.DomainRandao)

	msgs := specssv.PartialSignatureMessages{
		Type:     specssv.RandaoPartialSig,
		Messages: []*specssv.PartialSignatureMessage{},
	}
	for i := 0; i < msgCnt; i++ {
		msg := &specssv.PartialSignatureMessage{
			PartialSignature: signed[:],
			SigningRoot:      root,
			Signer:           id,
		}
		if wrongRoot {
			msg.SigningRoot = make([]byte, 32)
		}
		msgs.Messages = append(msgs.Messages, msg)
	}

	sig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: sig,
		Signer:    id,
	}
}

var PreConsensusSelectionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return PreConsensusCustomSlotSelectionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot)
}

var PreConsensusSelectionProofNextEpochMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return selectionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot2, testingutils.TestingDutySlot2, 1)
}

var PreConsensusMultiSelectionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return selectionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot, testingutils.TestingDutySlot, 3)
}

var PreConsensusCustomSlotSelectionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID, slot phase0.Slot) *spectypes.SignedPartialSignatureMessage {
	return selectionProofMsg(msgSK, beaconSK, msgID, beaconID, slot, testingutils.TestingDutySlot, 1)
}

var PreConsensusWrongMsgSlotSelectionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return selectionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot, testingutils.TestingDutySlot+1, 1)
}

var selectionProofMsg = func(
	sk *bls.SecretKey,
	beaconsk *bls.SecretKey,
	id spectypes.OperatorID,
	beaconid spectypes.OperatorID,
	slot phase0.Slot,
	msgSlot phase0.Slot,
	msgCnt int,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(1, spectypes.DomainSelectionProof)
	signed, root, _ := signer.SignBeaconObject(spectypes.SSZUint64(slot), d, beaconsk.GetPublicKey().Serialize(), spectypes.DomainSelectionProof)

	_msgs := make([]*specssv.PartialSignatureMessage, 0)
	for i := 0; i < msgCnt; i++ {
		_msgs = append(_msgs, &specssv.PartialSignatureMessage{
			PartialSignature: signed[:],
			SigningRoot:      root,
			Signer:           beaconid,
		})
	}

	msgs := specssv.PartialSignatureMessages{
		Type:     specssv.SelectionProofPartialSig,
		Messages: _msgs,
	}
	msgSig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: msgSig,
		Signer:    id,
	}
}

var PostConsensusAggregatorMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return postConsensusAggregatorMsg(sk, id, false, false)
}

var postConsensusAggregatorMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	wrongRoot bool,
	wrongBeaconSig bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(1, spectypes.DomainAggregateAndProof)
	signed, root, _ := signer.SignBeaconObject(testingutils.TestingAggregateAndProof, d, sk.GetPublicKey().Serialize(), spectypes.DomainAggregateAndProof)

	if wrongBeaconSig {
		// signed, _, _ = signer.SignAttestation(testingutils.TestingAttestationData, testingutils.TestingAttesterDuty, testingutils.TestingWrongSK.GetPublicKey().Serialize())
		panic("implement")
	}

	if wrongRoot {
		root = []byte{1, 2, 3, 4}
	}

	msgs := specssv.PartialSignatureMessages{
		Type: specssv.PostConsensusPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed,
				SigningRoot:      root,
				Signer:           id,
			},
		},
	}
	sig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: sig,
		Signer:    id,
	}
}

var PostConsensusSyncCommitteeMsg = func(sk *bls.SecretKey, id spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return postConsensusSyncCommitteeMsg(sk, id, false, false)
}

var postConsensusSyncCommitteeMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	wrongRoot bool,
	wrongBeaconSig bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(1, spectypes.DomainSyncCommittee)
	signed, root, _ := signer.SignBeaconObject(spectypes.SSZBytes(testingutils.TestingSyncCommitteeBlockRoot[:]), d, sk.GetPublicKey().Serialize(), spectypes.DomainSyncCommittee)

	if wrongBeaconSig {
		// signedAtt, _, _ = signer.SignAttestation(testingutils.TestingAttestationData, testingutils.TestingAttesterDuty, testingutils.TestingWrongSK.GetPublicKey().Serialize())
		panic("implement")
	}

	if wrongRoot {
		root = []byte{1, 2, 3, 4}
	}

	msgs := specssv.PartialSignatureMessages{
		Type: specssv.PostConsensusPartialSig,
		Messages: []*specssv.PartialSignatureMessage{
			{
				PartialSignature: signed,
				SigningRoot:      root,
				Signer:           id,
			},
		},
	}
	sig, _ := signer.SignRoot(msgs, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   msgs,
		Signature: sig,
		Signer:    id,
	}
}

var PreConsensusContributionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return PreConsensusCustomSlotContributionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot)
}

var PreConsensusContributionProofNextEpochMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return contributionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot2, false, false)
}

var PreConsensusCustomSlotContributionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID, slot phase0.Slot) *spectypes.SignedPartialSignatureMessage {
	return contributionProofMsg(msgSK, beaconSK, msgID, beaconID, slot, false, false)
}

var PreConsensusWrongMsgSlotContributionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return contributionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot, false, false)
}

var PreConsensusWrongOrderContributionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return contributionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot, true, false)
}

var PreConsensusWrongCountContributionProofMsg = func(msgSK, beaconSK *bls.SecretKey, msgID, beaconID spectypes.OperatorID) *spectypes.SignedPartialSignatureMessage {
	return contributionProofMsg(msgSK, beaconSK, msgID, beaconID, testingutils.TestingDutySlot, false, true)
}

var contributionProofMsg = func(
	sk, beaconsk *bls.SecretKey,
	id, beaconid spectypes.OperatorID,
	slot phase0.Slot,
	wrongMsgOrder bool,
	dropLastMsg bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	d, _ := beacon.DomainData(1, spectypes.DomainSyncCommitteeSelectionProof)

	msgs := make([]*specssv.PartialSignatureMessage, 0)
	for index := range testingutils.TestingContributionProofIndexes {
		subnet, _ := beacon.SyncCommitteeSubnetID(phase0.CommitteeIndex(index))
		data := &altair.SyncAggregatorSelectionData{
			Slot:              slot,
			SubcommitteeIndex: subnet,
		}
		sig, root, _ := signer.SignBeaconObject(data, d, beaconsk.GetPublicKey().Serialize(), spectypes.DomainSyncCommitteeSelectionProof)
		msg := &specssv.PartialSignatureMessage{
			PartialSignature: sig[:],
			SigningRoot:      ensureRoot(root),
			Signer:           beaconid,
		}

		if dropLastMsg && index == len(testingutils.TestingContributionProofIndexes)-1 {
			break
		}
		msgs = append(msgs, msg)
	}

	if wrongMsgOrder {
		m := msgs[0]
		msgs[0] = msgs[1]
		msgs[1] = m
	}

	msg := &specssv.PartialSignatureMessages{
		Type:     specssv.ContributionProofs,
		Messages: msgs,
	}

	msgSig, _ := signer.SignRoot(msg, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   *msg,
		Signature: msgSig,
		Signer:    id,
	}
}

var PostConsensusSyncCommitteeContributionMsg = func(sk *bls.SecretKey, id spectypes.OperatorID, keySet *testingutils.TestKeySet) *spectypes.SignedPartialSignatureMessage {
	return postConsensusSyncCommitteeContributionMsg(sk, id, testingutils.TestingValidatorIndex, keySet, false, false)
}

var postConsensusSyncCommitteeContributionMsg = func(
	sk *bls.SecretKey,
	id spectypes.OperatorID,
	validatorIndex phase0.ValidatorIndex,
	keySet *testingutils.TestKeySet,
	wrongRoot bool,
	wrongBeaconSig bool,
) *spectypes.SignedPartialSignatureMessage {
	signer := testingutils.NewTestingKeyManager()
	beacon := testingutils.NewTestingBeaconNode()
	dContribAndProof, _ := beacon.DomainData(1, spectypes.DomainContributionAndProof)

	msgs := make([]*specssv.PartialSignatureMessage, 0)
	for index := range testingutils.TestingSyncCommitteeContributions {
		// sign proof
		subnet, _ := beacon.SyncCommitteeSubnetID(phase0.CommitteeIndex(index))
		data := &altair.SyncAggregatorSelectionData{
			Slot:              testingutils.TestingDutySlot,
			SubcommitteeIndex: subnet,
		}
		dProof, _ := beacon.DomainData(1, spectypes.DomainSyncCommitteeSelectionProof)

		proofSig, _, _ := signer.SignBeaconObject(data, dProof, keySet.ValidatorPK.Serialize(), spectypes.DomainSyncCommitteeSelectionProof)
		blsProofSig := phase0.BLSSignature{}
		copy(blsProofSig[:], proofSig)

		// get contribution
		contribution, _ := beacon.GetSyncCommitteeContribution(testingutils.TestingDutySlot, subnet)

		// sign contrib and proof
		contribAndProof := &altair.ContributionAndProof{
			AggregatorIndex: validatorIndex,
			Contribution:    contribution,
			SelectionProof:  blsProofSig,
		}
		signed, root, _ := signer.SignBeaconObject(contribAndProof, dContribAndProof, sk.GetPublicKey().Serialize(), spectypes.DomainSyncCommitteeSelectionProof)

		if wrongRoot {
			root = []byte{1, 2, 3, 4}
		}

		msg := &specssv.PartialSignatureMessage{
			PartialSignature: signed,
			SigningRoot:      root,
			Signer:           id,
		}

		if wrongBeaconSig {
			// signedAtt, _, _ = signer.SignAttestation(testingutils.TestingAttestationData, testingutils.TestingAttesterDuty, testingutils.TestingWrongSK.GetPublicKey().Serialize())
			panic("implement")
		}

		msgs = append(msgs, msg)
	}

	msg := &specssv.PartialSignatureMessages{
		Type:     specssv.PostConsensusPartialSig,
		Messages: msgs,
	}

	sig, _ := signer.SignRoot(msg, spectypes.PartialSignatureType, sk.GetPublicKey().Serialize())
	return &spectypes.SignedPartialSignatureMessage{
		Message:   *msg,
		Signature: sig,
		Signer:    id,
	}
}

// ensureRoot ensures that SigningRoot will have sufficient allocated memory
// otherwise we get panic from bls:
// github.com/herumi/bls-eth-go-binary/bls.(*Sign).VerifyByte:738
func ensureRoot(root []byte) []byte {
	n := len(root)
	if n == 0 {
		n = 1
	}
	tmp := make([]byte, n)
	copy(tmp, root)
	return tmp
}
