package spectest

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	qbftStorage "github.com/bloxapp/ssv/ibft/storage"
	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/message"
	protocolp2p "github.com/bloxapp/ssv/protocol/v1/p2p"
	"github.com/bloxapp/ssv/protocol/v1/qbft/controller"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/msgcont"
	"github.com/bloxapp/ssv/protocol/v1/validator"
	"github.com/bloxapp/ssv/spec/qbft"
	"github.com/bloxapp/ssv/spec/ssv"
	"github.com/bloxapp/ssv/spec/ssv/spectest/tests"
	"github.com/bloxapp/ssv/spec/types"
	"github.com/bloxapp/ssv/spec/types/testingutils"
	"github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/logex"
)

func TestMappingJson(t *testing.T) {
	basedir, _ := os.Getwd()
	path := filepath.Join(basedir, "generate")
	fileName := "tests.json"
	specTests := map[string]*tests.SpecTest{}
	byteValue, err := ioutil.ReadFile(path + "/" + fileName)
	require.NoError(t, err)

	if err := json.Unmarshal(byteValue, &specTests); err != nil {
		require.NoError(t, err)
	}

	testMap := testsToRun() // TODO: remove

	for _, test := range specTests {
		test := test
		if _, ok := testMap[test.Name]; !ok {
			continue
		}

		t.Run(test.Name, func(t *testing.T) {
			runMappingTest(t, test)
		})
	}
}

func testsToRun() map[string]struct{} {
	result := make(map[string]struct{})
	for _, test := range AllTests {
		result[test.Name] = struct{}{}
	}

	return result
}

func runMappingTest(t *testing.T, test *tests.SpecTest) {
	ctx := context.TODO()
	logger := logex.Build(test.Name, zapcore.DebugLevel, nil)

	forkVersion := forksprotocol.V0ForkVersion
	pi, _ := protocolp2p.GenPeerID()
	beacon := validator.NewTestBeacon(t)

	keysSet := testingutils.Testing4SharesSet()

	beaconNetwork := core.NetworkFromString(string(test.Runner.BeaconNetwork))
	if beaconNetwork == "" {
		beaconNetwork = core.PraterNetwork
	}

	db, err := storage.GetStorageFactory(basedb.Options{
		Type:   "badger-memory",
		Logger: logger,
		Ctx:    ctx,
	})
	require.NoError(t, err)

	beaconRoleType := convertFromSpecRole(test.Runner.BeaconRoleType)
	require.Equalf(t, message.RoleTypeAttester, beaconRoleType, "only attester role is supported now")

	ibftStorage := qbftStorage.New(db, logger, beaconRoleType.String(), forkVersion)
	require.NoError(t, beacon.AddShare(keysSet.Shares[1]))

	v := validator.NewValidator(&validator.Options{
		Context:                    ctx,
		Logger:                     logger,
		IbftStorage:                ibftStorage,
		Network:                    beaconprotocol.NewNetwork(beaconNetwork),
		P2pNetwork:                 protocolp2p.NewMockNetwork(logger, pi, 10),
		Beacon:                     beacon,
		Share:                      convertShare(t, test.Runner.Share),
		ForkVersion:                forkVersion,
		Signer:                     beacon,
		SyncRateLimit:              time.Second * 5,
		SignatureCollectionTimeout: time.Second * 5,
		ReadMode:                   false,
		FullNode:                   false,
	})

	qbftCtrl := v.(*validator.Validator).Ibfts()[message.RoleTypeAttester].(*controller.Controller)
	qbftCtrl.State = controller.Ready
	go qbftCtrl.StartQueueConsumer(qbftCtrl.MessageHandler)
	require.NoError(t, qbftCtrl.Init())
	go v.ExecuteDuty(12, convertDuty(test.Duty))

	for _, msg := range test.Messages {
		require.NoError(t, v.ProcessMsg(convertSSVMessage(t, msg)))
	}

	time.Sleep(time.Second * 3) // 3s round

	currentInstance := qbftCtrl.GetCurrentInstance()
	decided, err := ibftStorage.GetLastDecided(qbftCtrl.GetIdentifier())
	require.NoError(t, err)
	decidedValue := []byte("")
	if decided != nil {
		cd, err := decided.Message.GetCommitData()
		require.NoError(t, err)
		decidedValue = cd.Data
	}

	mappedInstance := new(qbft.Instance)
	if currentInstance != nil {
		mappedInstance.State = &qbft.State{
			Share:                           test.Runner.Share,
			ID:                              qbftCtrl.GetIdentifier(),
			Round:                           qbft.Round(currentInstance.State().GetRound()),
			Height:                          qbft.Height(currentInstance.State().GetHeight()),
			LastPreparedRound:               qbft.Round(currentInstance.State().GetPreparedRound()),
			LastPreparedValue:               currentInstance.State().GetPreparedValue(),
			ProposalAcceptedForCurrentRound: nil,
			Decided:                         decided != nil && decided.Message.Height == currentInstance.State().GetHeight(), // TODO might need to add this flag to qbftCtrl
			DecidedValue:                    decidedValue,                                                                    // TODO allow a way to get it
			ProposeContainer:                convertToSpecContainer(t, currentInstance.Containers()[qbft.ProposalMsgType]),
			PrepareContainer:                convertToSpecContainer(t, currentInstance.Containers()[qbft.PrepareMsgType]),
			CommitContainer:                 convertToSpecContainer(t, currentInstance.Containers()[qbft.CommitMsgType]),
			RoundChangeContainer:            convertToSpecContainer(t, currentInstance.Containers()[qbft.RoundChangeMsgType]),
		}
		mappedInstance.StartValue = currentInstance.State().GetInputValue()
	}

	mappedDecidedValue := &types.ConsensusData{
		Duty: &types.Duty{
			Type:                    0,
			PubKey:                  phase0.BLSPubKey{},
			Slot:                    0,
			ValidatorIndex:          0,
			CommitteeIndex:          0,
			CommitteeLength:         0,
			CommitteesAtSlot:        0,
			ValidatorCommitteeIndex: 0,
		},
		AttestationData:           nil,
		BlockData:                 nil,
		AggregateAndProof:         nil,
		SyncCommitteeBlockRoot:    phase0.Root{},
		SyncCommitteeContribution: nil,
	}

	mappedSignedAtts := &phase0.Attestation{
		AggregationBits: nil,
		Data:            nil,
		Signature:       phase0.BLSSignature{},
	}

	resState := ssv.NewDutyExecutionState(3)
	resState.RunningInstance = mappedInstance
	resState.DecidedValue = mappedDecidedValue
	resState.SignedAttestation = mappedSignedAtts
	resState.Finished = true // TODO need to set real value

	root, err := resState.GetRoot()
	require.NoError(t, err)

	expectedRoot, err := hex.DecodeString(test.PostDutyRunnerStateRoot)
	require.NoError(t, err)
	require.Equal(t, expectedRoot, root)

	require.NoError(t, v.Close())
	db.Close()
}

func convertDuty(duty *types.Duty) *beaconprotocol.Duty {
	return &beaconprotocol.Duty{
		Type:                    convertFromSpecRole(duty.Type),
		PubKey:                  duty.PubKey,
		Slot:                    duty.Slot,
		ValidatorIndex:          duty.ValidatorIndex,
		CommitteeIndex:          duty.CommitteeIndex,
		CommitteeLength:         duty.CommitteeLength,
		CommitteesAtSlot:        duty.CommitteesAtSlot,
		ValidatorCommitteeIndex: duty.ValidatorCommitteeIndex,
	}
}

func convertFromSpecRole(role types.BeaconRole) message.RoleType {
	switch role {
	case types.BNRoleAttester:
		return message.RoleTypeAttester
	case types.BNRoleAggregator:
		return message.RoleTypeAggregator
	case types.BNRoleProposer:
		return message.RoleTypeProposer
	case types.BNRoleSyncCommittee, types.BNRoleSyncCommitteeContribution:
		return message.RoleTypeUnknown
	default:
		panic(fmt.Sprintf("unknown role type! (%s)", role.String()))
	}
	return 0
}

func convertSSVMessage(t *testing.T, msg *types.SSVMessage) *message.SSVMessage {
	data := msg.Data

	if msg.MsgType == types.SSVPartialSignatureMsgType {
		sps := new(ssv.SignedPartialSignatureMessage)
		require.NoError(t, sps.Decode(msg.Data))
		spsm := sps.Messages[0]
		spcm := &message.SignedPostConsensusMessage{
			Message: &message.PostConsensusMessage{
				Height:          0, // TODO need to get height fom ssv.SignedPartialSignatureMessage
				DutySignature:   spsm.PartialSignature,
				DutySigningRoot: spsm.SigningRoot,
				Signers:         convertSingers(spsm.Signers),
			},
			Signature: message.Signature(sps.Signature),
			Signers:   convertSingers(sps.Signers),
		}

		encoded, err := spcm.Encode()
		require.NoError(t, err)
		data = encoded
	}

	var msgType message.MsgType
	switch msg.GetType() {
	case types.SSVConsensusMsgType:
		msgType = message.SSVConsensusMsgType
	case types.SSVDecidedMsgType:
		msgType = message.SSVDecidedMsgType
	case types.SSVPartialSignatureMsgType:
		msgType = message.SSVPostConsensusMsgType
	case types.DKGMsgType:
		panic("type not supported yet")
	}
	return &message.SSVMessage{
		MsgType: msgType,
		ID:      message.NewIdentifier(msg.MsgID[:], message.RoleTypeAttester),
		Data:    data,
	}
}

func convertShare(t *testing.T, share *types.Share) *beaconprotocol.Share {
	committee := make(map[message.OperatorID]*beaconprotocol.Node)
	for i, operator := range share.Committee {
		if operator == nil {
			continue
		}

		committee[message.OperatorID(operator.OperatorID)] = &beaconprotocol.Node{
			IbftID: uint64(i),
			Pk:     operator.PubKey,
		}
	}

	return &beaconprotocol.Share{
		NodeID:    message.OperatorID(share.OperatorID),
		PublicKey: bytesToBlsPubKey(t, share.SharePubKey),
		Committee: committee,
	}
}

func bytesToBlsPubKey(t *testing.T, pubKeyBytes []byte) *bls.PublicKey {
	pubKey := &bls.PublicKey{}
	require.NoError(t, pubKey.Deserialize(pubKeyBytes))
	return pubKey
}

func convertToSpecContainer(t *testing.T, container msgcont.MessageContainer) *qbft.MsgContainer {
	c := qbft.NewMsgContainer()
	container.AllMessaged(func(round message.Round, msg *message.SignedMessage) {
		var signers []types.OperatorID
		for _, s := range msg.GetSigners() {
			signers = append(signers, types.OperatorID(s))
		}

		// TODO need to use one of the message type (spec/protocol)
		ok, err := c.AddIfDoesntExist(&qbft.SignedMessage{
			Signature: types.Signature(msg.Signature),
			Signers:   signers,
			Message: &qbft.Message{
				MsgType:    qbft.MessageType(msg.Message.MsgType),
				Height:     qbft.Height(msg.Message.Height),
				Round:      qbft.Round(msg.Message.Round),
				Identifier: msg.Message.Identifier,
				Data:       msg.Message.Data,
			},
		})
		require.NoError(t, err)
		require.True(t, ok)
	})
	return c
}

func convertSingers(specSigners []types.OperatorID) []message.OperatorID {
	var signers []message.OperatorID
	for _, s := range specSigners {
		signers = append(signers, message.OperatorID(s))
	}
	return signers
}
