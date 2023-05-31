package qbft

import (
	"encoding/hex"
	"testing"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectests "github.com/bloxapp/ssv-spec/qbft/spectest/tests"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func RunCreateMsg(t *testing.T, test *spectests.CreateMsgSpecTest) {
	var msg *specqbft.SignedMessage
	var err error
	switch test.CreateType {
	case spectests.CreateProposal:
		msg, err = createProposal(test)
	case spectests.CreatePrepare:
		msg, err = createPrepare(test)
	case spectests.CreateCommit:
		msg, err = createCommit(test)
	case spectests.CreateRoundChange:
		msg, err = createRoundChange(test)
	default:
		t.Fail()
	}

	if err != nil && len(test.ExpectedError) != 0 {
		require.EqualError(t, err, test.ExpectedError)
		return
	}
	require.NoError(t, err)

	r, err2 := msg.GetRoot()
	if len(test.ExpectedError) != 0 {
		require.EqualError(t, err2, test.ExpectedError)
		return
	}
	require.NoError(t, err2)
	require.EqualValues(t, test.ExpectedRoot, hex.EncodeToString(r[:]))
}

func createCommit(test *spectests.CreateMsgSpecTest) (*specqbft.SignedMessage, error) {
	ks := testingutils.Testing4SharesSet()
	state := &specqbft.State{
		Share: testingutils.TestingShare(ks),
		ID:    []byte{1, 2, 3, 4},
	}
	config := testingutils.TestingConfig(ks)

	return specqbft.CreateCommit(state, config, test.Value)
}

func createPrepare(test *spectests.CreateMsgSpecTest) (*specqbft.SignedMessage, error) {
	ks := testingutils.Testing4SharesSet()
	state := &specqbft.State{
		Share: testingutils.TestingShare(ks),
		ID:    []byte{1, 2, 3, 4},
	}
	config := testingutils.TestingConfig(ks)

	return specqbft.CreatePrepare(state, config, test.Round, test.Value)
}

func createProposal(test *spectests.CreateMsgSpecTest) (*specqbft.SignedMessage, error) {
	ks := testingutils.Testing4SharesSet()
	state := &specqbft.State{
		Share: testingutils.TestingShare(ks),
		ID:    []byte{1, 2, 3, 4},
	}
	config := testingutils.TestingConfig(ks)

	return specqbft.CreateProposal(state, config, test.Value[:], test.RoundChangeJustifications, test.PrepareJustifications)
}

func createRoundChange(test *spectests.CreateMsgSpecTest) (*specqbft.SignedMessage, error) {
	ks := testingutils.Testing4SharesSet()
	state := &specqbft.State{
		Share:            testingutils.TestingShare(ks),
		ID:               []byte{1, 2, 3, 4},
		PrepareContainer: specqbft.NewMsgContainer(),
	}
	config := testingutils.TestingConfig(ks)

	if len(test.PrepareJustifications) > 0 {
		state.LastPreparedRound = test.PrepareJustifications[0].Message.Round
		state.LastPreparedValue = test.StateValue

		for _, msg := range test.PrepareJustifications {
			_, err := state.PrepareContainer.AddFirstMsgForSignerAndRound(msg)
			if err != nil {
				return nil, errors.Wrap(err, "could not add first message for signer")
			}
		}
	}

	return specqbft.CreateRoundChange(state, config, 1, test.Value[:])
}
