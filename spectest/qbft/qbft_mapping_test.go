package qbft

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bloxapp/ssv/protocol/v1/message"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/bloxapp/ssv-spec/qbft/spectest"
	spectests "github.com/bloxapp/ssv-spec/qbft/spectest/tests"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/commit"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/messages"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/prepare"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/proposal"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/proposer"
	"github.com/bloxapp/ssv-spec/qbft/spectest/tests/roundchange"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	protocolp2p "github.com/bloxapp/ssv/protocol/v1/p2p"
	qbftprotocol "github.com/bloxapp/ssv/protocol/v1/qbft"
	forksfactory "github.com/bloxapp/ssv/protocol/v1/qbft/controller/forks/factory"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/leader/roundrobin"
	"github.com/bloxapp/ssv/protocol/v1/qbft/instance/msgcont"
	qbftstorage "github.com/bloxapp/ssv/protocol/v1/qbft/storage"
	"github.com/bloxapp/ssv/protocol/v1/validator"
	"github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/logex"
)

// nolint
func testsToRun() map[string]struct{} {
	testList := spectest.AllTests
	testList = []spectest.SpecTest{
		proposer.FourOperators(),
		proposer.SevenOperators(),
		proposer.TenOperators(),
		proposer.ThirteenOperators(),

		messages.RoundChangeDataInvalidJustifications(),
		messages.RoundChangeDataInvalidPreparedRound(),
		messages.RoundChangeDataInvalidPreparedValue(),
		messages.RoundChangePrePreparedJustifications(),
		messages.RoundChangeNotPreparedJustifications(),
		messages.CommitDataEncoding(),
		messages.DecidedMsgEncoding(),
		messages.MsgNilIdentifier(),
		messages.MsgNonZeroIdentifier(),
		messages.MsgTypeUnknown(),
		messages.PrepareDataEncoding(),
		messages.ProposeDataEncoding(),
		messages.MsgDataNil(),
		messages.MsgDataNonZero(),
		messages.SignedMsgSigTooShort(),
		messages.SignedMsgSigTooLong(),
		messages.SignedMsgNoSigners(),
		messages.GetRoot(),
		messages.SignedMessageEncoding(),
		messages.CreateProposal(),
		messages.CreateProposalPreviouslyPrepared(),
		messages.CreateProposalNotPreviouslyPrepared(),
		messages.CreatePrepare(),
		messages.CreateCommit(),
		messages.CreateRoundChange(),
		messages.CreateRoundChangePreviouslyPrepared(),
		messages.RoundChangeDataEncoding(),

		spectests.HappyFlow(),
		spectests.SevenOperators(),
		spectests.TenOperators(),
		spectests.ThirteenOperators(),

		proposal.HappyFlow(),
		//proposal.NotPreparedPreviouslyJustification(), // TODO(nkryuchkov): failure; need to handle proposal justifications
		//proposal.PreparedPreviouslyJustification(),    // TODO(nkryuchkov): failure; handle PJ (proposal justifications) & fix "change round justification does not constitute a quorum"
		//proposal.DifferentJustifications(),            // TODO(nkryuchkov): failure; handle PJ
		//proposal.JustificationsNotHeighest(),          // TODO(nkryuchkov): failure; handle PJ
		//proposal.JustificationsValueNotJustified(),    // TODO(nkryuchkov): failure; handle PJ
		//proposal.DuplicateMsg(),                       // TODO(nkryuchkov): failure; need to check if proposal was already accepted
		proposal.FirstRoundJustification(),
		//proposal.FutureRoundNoAcceptedProposal(), // TODO(nkryuchkov): failure; need to decline proposals from future rounds if no proposal was accepted for current round
		//proposal.FutureRoundAcceptedProposal(),   // TODO(nkryuchkov): failure; need to accept proposals from future rounds if already accepted proposal for current round
		//proposal.PastRound(),                     // TODO(nkryuchkov): failure; need to decline proposals from past rounds
		proposal.ImparsableProposalData(),
		//proposal.InvalidRoundChangeJustificationPrepared(),        // TODO(nkryuchkov): failure; need to handle invalid rc justification
		//proposal.InvalidRoundChangeJustification(),                // TODO(nkryuchkov): failure; need to handle invalid rc justification
		//proposal.PreparedPreviouslyNoRCJustificationQuorum(),      // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.NoRCJustification(),                              // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.PreparedPreviouslyNoPrepareJustificationQuorum(), // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.PreparedPreviouslyDuplicatePrepareMsg(),          // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.PreparedPreviouslyDuplicateRCMsg(),               // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.DuplicateRCMsg(),                                 // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously with or without quorum of prepared or round change msgs justification
		//proposal.InvalidPrepareJustificationValue(),               // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously but one of the prepare justifications has value != highest prepared or round != highest prepared round
		//proposal.InvalidPrepareJustificationRound(),               // TODO(nkryuchkov): failure; need to check if proposal for round >1 was prepared or not prepared previously but one of the prepare justifications has value != highest prepared or round != highest prepared round
		proposal.InvalidProposalData(),
		//proposal.InvalidValueCheck(), // TODO(nkryuchkov): failure; need to check message data
		proposal.MultiSigner(),
		//proposal.PostDecided(),            // TODO(nkryuchkov): failure; need to avoid processing proposal message after instance became prepared or decided
		//proposal.PostPrepared(),           // TODO(nkryuchkov): failure; need to avoid processing proposal message after instance became prepared or decided
		//proposal.SecondProposalForRound(), // TODO(nkryuchkov): failure; need to forbid second proposal for round
		//proposal.WrongHeight(),            // TODO(nkryuchkov): failure; need to expect the same error in spec if height is wrong
		proposal.WrongProposer(),
		//proposal.WrongSignature(), // TODO(nkryuchkov): failure; fix expected errors in spec
		//
		prepare.DuplicateMsg(),
		prepare.HappyFlow(),
		prepare.ImparsableProposalData(),
		//prepare.InvalidPrepareData(), // TODO(nkryuchkov): failure; need to expect same error in spec if message is wrong
		//prepare.MultiSigner(),        // TODO(nkryuchkov): failure; need to check that message has only 1 signer
		//prepare.NoPreviousProposal(), // TODO(nkryuchkov): failure; need to fail to process message if proposal was not received
		//prepare.OldRound(),           // TODO(nkryuchkov): failure; need to fail to process message if its round is not equal to current one
		//prepare.FutureRound(),        // TODO(nkryuchkov): failure; need to fail to process message if its round is not equal to current one
		prepare.PostDecided(),
		//prepare.WrongData(),      // TODO(nkryuchkov): failure; need to check if message data is different from proposal data
		prepare.WrongHeight(),
		//prepare.WrongSignature(), // TODO(nkryuchkov): failure; fix expected errors in spec
		//
		commit.CurrentRound(),
		//commit.FutureRound(), // TODO(nkryuchkov): failure; need to fail to process message if its round is not equal to current one
		//commit.PastRound(),   // TODO(nkryuchkov): failure; need to fail to process message if its round is not equal to current one
		commit.DuplicateMsg(),
		commit.HappyFlow(),
		//commit.InvalidCommitData(), // TODO(nkryuchkov): failure; need to expect same error in spec if message is wrong
		commit.PostDecided(),
		//commit.WrongData1(),             // TODO(nkryuchkov): failure; need to check if message data is different from proposal data
		//commit.WrongData2(),             // TODO(nkryuchkov): failure; need to check if message data is different from proposal data
		//commit.MultiSignerWithOverlap(), // TODO(nkryuchkov): failure; need to fix case when multi signer commit msg which does overlap previous valid commit signers and previous valid commits
		commit.MultiSignerNoOverlap(),
		//commit.Decided(),                // TODO(nkryuchkov): failure; need to fix case when multi signer commit msg which does overlap previous valid commit signers and previous valid commits
		//commit.NoPrevAcceptedProposal(), // TODO(nkryuchkov): failure; need to fail to process message if proposal was not received
		//commit.WrongHeight(),            // TODO(nkryuchkov): failure; need to expect the same error in spec if height is wrong
		commit.ImparsableCommitData(),
		//commit.WrongSignature(), // TODO(nkryuchkov): failure; fix expected errors in spec
		//
		//roundchange.HappyFlow(),         // TODO(nkryuchkov): failure; substitution of message signature works incorrectly
		//roundchange.F1Speedup(),         // TODO(nkryuchkov): failure; data inside ProposalAcceptedForCurrentRound misses RoundChangeJustification
		//roundchange.F1SpeedupPrepared(), // TODO(nkryuchkov): failure; need to substitute identifier in justifications in mapping
		//roundchange.WrongHeight(),       // TODO(nkryuchkov): failure; need to expect the same error in spec if height is wrong
		//roundchange.WrongSig(),          // TODO(nkryuchkov): failure; fix expected errors in spec
		//roundchange.MultiSigner(),       // TODO(nkryuchkov): failure; need to check that message has only 1 signer
		roundchange.NotPrepared(),
		//roundchange.Prepared(),                     // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.PeerPrepared(),                 // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.PeerPreparedDifferentHeights(), // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		roundchange.JustificationWrongValue(),
		roundchange.JustificationWrongRound(),
		//roundchange.JustificationNoQuorum(),     // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.JustificationMultiSigners(), // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.JustificationInvalidSig(),   // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		roundchange.JustificationInvalidRound(),
		//roundchange.JustificationInvalidPrepareData(), // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.JustificationDuplicateMsg(),       // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.InvalidRoundChangeData(),          // TODO(nkryuchkov): failure; need to check message data
		//roundchange.FutureRound(),                     // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.PastRound(),                       // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.F1SpeedupDifferentRounds(),        // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.DuplicateMsgQuorum(),              // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.DuplicateMsgPartialQuorum(),       // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		//roundchange.DuplicateMsgPrepared(),            // TODO(nkryuchkov): failure; fix "change round justification does not constitute a quorum"
		roundchange.ImparsableRoundChangeData(),
	}

	result := make(map[string]struct{})
	for _, test := range testList {
		result[test.TestName()] = struct{}{}
	}

	return result
}

func TestQBFTMapping(t *testing.T) {
	path, _ := os.Getwd()
	fileName := "tests.json"
	filePath := path + "/" + fileName
	jsonTests, err := ioutil.ReadFile(filePath)
	if err != nil {
		resp, err := http.Get("https://raw.githubusercontent.com/bloxapp/ssv-spec/V0.2/qbft/spectest/generate/tests.json")
		require.NoError(t, err)

		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

		jsonTests, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.NoError(t, ioutil.WriteFile(filePath, jsonTests, 0644))
	}

	untypedTests := map[string]interface{}{}
	if err := json.Unmarshal(jsonTests, &untypedTests); err != nil {
		panic(err.Error())
	}

	origDomain := message.GetDefaultDomain()
	message.SetDefaultDomain(spectypes.PrimusTestnet)
	defer func() {
		message.SetDefaultDomain(origDomain)
	}()

	testMap := testsToRun() // TODO(nkryuchkov): remove

	tests := make(map[string]spectest.SpecTest)
	for name, test := range untypedTests {
		name, test := name, test

		testName := strings.Split(name, "_")[1]
		testType := strings.Split(name, "_")[0]

		if _, ok := testMap[testName]; !ok {
			continue
		}

		switch testType {
		case reflect.TypeOf(&spectests.MsgProcessingSpecTest{}).String():
			byts, err := json.Marshal(test)
			require.NoError(t, err)
			typedTest := &spectests.MsgProcessingSpecTest{}
			require.NoError(t, json.Unmarshal(byts, &typedTest))

			t.Run(typedTest.TestName(), func(t *testing.T) {
				runMsgProcessingSpecTest(t, typedTest)
			})
		case reflect.TypeOf(&spectests.MsgSpecTest{}).String():
			byts, err := json.Marshal(test)
			require.NoError(t, err)
			typedTest := &spectests.MsgSpecTest{}
			require.NoError(t, json.Unmarshal(byts, &typedTest))

			tests[testName] = typedTest
			t.Run(typedTest.TestName(), func(t *testing.T) {
				runMsgSpecTest(t, typedTest)
			})
		}
	}
}

func runMsgProcessingSpecTest(t *testing.T, test *spectests.MsgProcessingSpecTest) {
	ctx := context.TODO()
	logger := logex.Build(test.Name, zapcore.DebugLevel, nil)

	forkVersion := forksprotocol.GenesisForkVersion
	pi, _ := protocolp2p.GenPeerID()
	p2pNet := protocolp2p.NewMockNetwork(logger, pi, 10)
	beacon := validator.NewTestBeacon(t)

	var keysSet *testingutils.TestKeySet
	switch len(test.Pre.State.Share.Committee) {
	case 4:
		keysSet = testingutils.Testing4SharesSet()
	case 7:
		keysSet = testingutils.Testing7SharesSet()
	case 10:
		keysSet = testingutils.Testing10SharesSet()
	case 13:
		keysSet = testingutils.Testing13SharesSet()
	default:
		t.Error("unknown key set length")
	}

	db, err := storage.GetStorageFactory(basedb.Options{
		Type:   "badger-memory",
		Logger: logger,
		Ctx:    ctx,
	})
	require.NoError(t, err)

	qbftStorage := qbftstorage.NewQBFTStore(db, logger, spectypes.BNRoleAttester.String())

	share := &beaconprotocol.Share{
		NodeID:    1,
		PublicKey: keysSet.ValidatorPK,
		Committee: make(map[spectypes.OperatorID]*beaconprotocol.Node),
	}

	var mappedCommittee []*spectypes.Operator
	for i := range test.Pre.State.Share.Committee {
		operatorID := spectypes.OperatorID(i) + 1
		require.NoError(t, beacon.AddShare(keysSet.Shares[operatorID]))

		pk := keysSet.Shares[operatorID].GetPublicKey().Serialize()
		share.Committee[spectypes.OperatorID(i)+1] = &beaconprotocol.Node{
			IbftID: uint64(i) + 1,
			Pk:     pk,
		}

		share.OperatorIds = append(share.OperatorIds, uint64(i)+1)

		mappedCommittee = append(mappedCommittee, &spectypes.Operator{
			OperatorID: operatorID,
			PubKey:     pk,
		})
	}

	mappedShare := &spectypes.Share{
		OperatorID:      share.NodeID,
		ValidatorPubKey: share.PublicKey.Serialize(),
		SharePubKey:     share.Committee[share.NodeID].Pk,
		Committee:       mappedCommittee,
		Quorum:          keysSet.Threshold,
		PartialQuorum:   keysSet.PartialThreshold,
		DomainType:      message.GetDefaultDomain(),
		Graffiti:        nil,
	}

	qbftInstance := newQbftInstance(logger, qbftStorage, p2pNet, beacon, share, mappedShare, test.Pre.StartValue, forkVersion)
	qbftInstance.Init()
	qbftInstance.State().InputValue.Store(test.Pre.StartValue)
	qbftInstance.State().Round.Store(test.Pre.State.Round)
	qbftInstance.State().Height.Store(test.Pre.State.Height)

	var lastErr error
	for _, msg := range test.InputMessages {
		if _, err := qbftInstance.ProcessMsg(msg); err != nil {
			lastErr = err
		}
	}

	mappedInstance := new(specqbft.Instance)
	if qbftInstance != nil {
		preparedValue := qbftInstance.State().GetPreparedValue()
		if len(preparedValue) == 0 {
			preparedValue = nil
		}
		round := qbftInstance.State().GetRound()

		_, decidedErr := qbftInstance.CommittedAggregatedMsg()
		var decidedValue []byte
		if decidedErr == nil && len(test.InputMessages) != 0 {
			decidedValue = test.InputMessages[0].Message.Identifier
		}

		mappedInstance.State = &specqbft.State{
			Share:                           mappedShare,
			ID:                              test.Pre.State.ID,
			Round:                           round,
			Height:                          qbftInstance.State().GetHeight(),
			LastPreparedRound:               qbftInstance.State().GetPreparedRound(),
			LastPreparedValue:               preparedValue,
			ProposalAcceptedForCurrentRound: test.Pre.State.ProposalAcceptedForCurrentRound,
			Decided:                         decidedErr == nil,
			DecidedValue:                    decidedValue,
			ProposeContainer:                convertToSpecContainer(t, qbftInstance.Containers()[specqbft.ProposalMsgType]),
			PrepareContainer:                convertToSpecContainer(t, qbftInstance.Containers()[specqbft.PrepareMsgType]),
			CommitContainer:                 convertToSpecContainer(t, qbftInstance.Containers()[specqbft.CommitMsgType]),
			RoundChangeContainer:            convertToSpecContainer(t, qbftInstance.Containers()[specqbft.RoundChangeMsgType]),
		}

		allMessages := mappedInstance.State.ProposeContainer.AllMessaged()
		if len(allMessages) != 0 {
			var highestRound specqbft.Round
			var highestMessage *specqbft.SignedMessage
			for _, message := range allMessages {
				if message.Message.Round > highestRound {
					highestRound = message.Message.Round
					highestMessage = message
				}
			}
			mappedInstance.State.ProposalAcceptedForCurrentRound = highestMessage
		}

		mappedInstance.StartValue = qbftInstance.State().GetInputValue()
	}

	if len(test.ExpectedError) != 0 {
		require.EqualError(t, lastErr, test.ExpectedError)
	} else {
		require.NoError(t, lastErr)
	}

	mappedRoot, err := mappedInstance.State.GetRoot()
	require.NoError(t, err)
	require.Equal(t, test.PostRoot, hex.EncodeToString(mappedRoot))

	type BroadcastMessagesGetter interface {
		GetBroadcastMessages() []spectypes.SSVMessage
	}

	outputMessages := p2pNet.(BroadcastMessagesGetter).GetBroadcastMessages()

	require.Equal(t, len(test.OutputMessages), len(outputMessages))

	for i, outputMessage := range outputMessages {
		r1, _ := test.OutputMessages[i].GetRoot()

		msg2 := &specqbft.SignedMessage{}
		require.NoError(t, msg2.Decode(outputMessage.Data))

		r2, _ := msg2.GetRoot()
		require.EqualValues(t, r1, r2, fmt.Sprintf("output msg %d roots not equal", i))
	}

	db.Close()
}

func runMsgSpecTest(t *testing.T, test *spectests.MsgSpecTest) {
	var lastErr error

	for i, messageBytes := range test.EncodedMessages {
		m := &specqbft.SignedMessage{}
		if err := m.Decode(messageBytes); err != nil {
			lastErr = err
		}

		if len(test.ExpectedRoots) > 0 {
			r, err := m.GetRoot()
			if err != nil {
				lastErr = err
			}

			if !bytes.Equal(test.ExpectedRoots[i], r) {
				t.Fail()
			}
		}
	}

	for i, msg := range test.Messages {
		if err := msg.Validate(); err != nil {
			lastErr = err
		}

		switch msg.Message.MsgType {
		case specqbft.RoundChangeMsgType:
			rc := specqbft.RoundChangeData{}
			if err := rc.Decode(msg.Message.Data); err != nil {
				lastErr = err
			}
			if err := validateRoundChangeData(rc); err != nil {
				lastErr = err
			}
		}

		if len(test.Messages) > 0 {
			r1, err := msg.Encode()
			if err != nil {
				lastErr = err
			}

			r2, err := test.Messages[i].Encode()
			if err != nil {
				lastErr = err
			}
			if !bytes.Equal(r2, r1) {
				t.Fail()
			}
		}
	}

	if len(test.ExpectedError) != 0 {
		require.EqualError(t, lastErr, test.ExpectedError)
	} else {
		require.NoError(t, lastErr)
	}
}

func validateRoundChangeData(d specqbft.RoundChangeData) error {
	if isRoundChangeDataPrepared(d) {
		if len(d.PreparedValue) == 0 {
			return errors.New("round change prepared value invalid")
		}
		if len(d.RoundChangeJustification) == 0 {
			return errors.New("round change justification invalid")
		}
	}

	return nil
}

func isRoundChangeDataPrepared(d specqbft.RoundChangeData) bool {
	return d.PreparedRound != 0 || len(d.PreparedValue) != 0
}

// nolint:unused
type roundRobinLeaderSelector struct {
	state       *qbftprotocol.State
	mappedShare *spectypes.Share
}

// nolint:unused
func (m roundRobinLeaderSelector) Calculate(round uint64) uint64 {
	specState := &specqbft.State{
		Share:  m.mappedShare,
		Height: m.state.GetHeight(),
	}
	// RoundRobinProposer returns OperatorID which starts with 1.
	// As the result will be used to index OperatorIds which starts from 0,
	// the result of Calculate should be decremented.
	return uint64(specqbft.RoundRobinProposer(specState, specqbft.Round(round))) - 1
}

func newQbftInstance(logger *zap.Logger, qbftStorage qbftstorage.QBFTStore, net protocolp2p.MockNetwork, beacon *validator.TestBeacon, share *beaconprotocol.Share, mappedShare *spectypes.Share, identifier []byte, forkVersion forksprotocol.ForkVersion) instance.Instancer {
	const height = 0
	fork := forksfactory.NewFork(forkVersion)
	newInstance := instance.NewInstance(&instance.Options{
		Logger:           logger,
		ValidatorShare:   share,
		Network:          net,
		Config:           qbftprotocol.DefaultConsensusParams(),
		Identifier:       identifier[:],
		Height:           height,
		RequireMinPeers:  false,
		Fork:             fork.InstanceFork(),
		SSVSigner:        beacon.KeyManager,
		ChangeRoundStore: qbftStorage,
	})

	newInstance.(*instance.Instance).LeaderSelector = roundrobin.New(share, newInstance.(*instance.Instance).State())

	return newInstance
}

func convertToSpecContainer(t *testing.T, container msgcont.MessageContainer) *specqbft.MsgContainer {
	c := specqbft.NewMsgContainer()
	container.AllMessaged(func(round specqbft.Round, msg *specqbft.SignedMessage) {
		ok, err := c.AddIfDoesntExist(&specqbft.SignedMessage{
			Signature: msg.Signature,
			Signers:   msg.Signers,
			Message: &specqbft.Message{
				MsgType:    msg.Message.MsgType,
				Height:     msg.Message.Height,
				Round:      msg.Message.Round,
				Identifier: msg.Message.Identifier,
				Data:       msg.Message.Data,
			},
		})
		require.NoError(t, err)
		require.True(t, ok)
	})
	return c
}
