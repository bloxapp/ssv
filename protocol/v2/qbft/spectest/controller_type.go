package qbft

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectests "github.com/bloxapp/ssv-spec/qbft/spectest/tests"
	spectypes "github.com/bloxapp/ssv-spec/types"
	spectestingutils "github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/ssv/protocol/v2/ssv/spectest/utils"
)

func RunControllerSpecTest(t *testing.T, test *spectests.ControllerSpecTest) {
	identifier := spectypes.NewMsgID(spectestingutils.TestingValidatorPubKey[:], spectypes.BNRoleAttester)
	config := utils.TestingConfig(spectestingutils.Testing4SharesSet(), identifier.GetRoleType())
	contr := utils.NewTestingQBFTController(
		identifier[:],
		spectestingutils.TestingShare(spectestingutils.Testing4SharesSet()),
		config,
	)

	var lastErr error
	for _, runData := range test.RunInstanceData {
		err := contr.StartNewInstance(runData.InputValue)
		if err != nil {
			lastErr = err
		} else if runData.ExpectedTimerState != nil {
			if timer, ok := config.GetTimer().(*spectestingutils.TestQBFTTimer); ok {
				require.Equal(t, runData.ExpectedTimerState.Timeouts, timer.State.Timeouts)
				require.Equal(t, runData.ExpectedTimerState.Round, timer.State.Round)
			}
		}

		decidedCnt := 0
		for _, msg := range runData.InputMessages {
			decided, err := contr.ProcessMsg(msg)
			if err != nil {
				lastErr = err
			}
			if decided != nil {
				decidedCnt++

				data, _ := decided.Message.GetCommitData()
				require.EqualValues(t, runData.DecidedVal, data.Data)
			}
		}

		require.EqualValues(t, runData.DecidedCnt, decidedCnt)

		if runData.BroadcastedDecided != nil {
			// test broadcasted
			broadcastedMsgs := config.GetNetwork().(*spectestingutils.TestingNetwork).BroadcastedMsgs
			require.Greater(t, len(broadcastedMsgs), 0)
			found := false
			for _, msg := range broadcastedMsgs {
				if !bytes.Equal(identifier[:], msg.MsgID[:]) {
					continue
				}

				msg1 := &specqbft.SignedMessage{}
				require.NoError(t, msg1.Decode(msg.Data))
				r1, err := msg1.GetRoot()
				require.NoError(t, err)

				r2, err := runData.BroadcastedDecided.GetRoot()
				require.NoError(t, err)

				if bytes.Equal(r1, r2) &&
					reflect.DeepEqual(runData.BroadcastedDecided.Signers, msg1.Signers) &&
					reflect.DeepEqual(runData.BroadcastedDecided.Signature, msg1.Signature) {
					require.False(t, found)
					found = true
				}
			}
			require.True(t, found)
		}

		r, err := contr.GetRoot()
		require.NoError(t, err)
		require.EqualValues(t, runData.ControllerPostRoot, hex.EncodeToString(r))
	}

	if len(test.ExpectedError) != 0 {
		require.EqualError(t, lastErr, test.ExpectedError)
	} else {
		require.NoError(t, lastErr)
	}
}
