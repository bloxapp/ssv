package prepare

import (
	ibft2 "github.com/bloxapp/ssv/ibft/instance"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/ibft/spectesting"
	"github.com/bloxapp/ssv/network"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"testing"
)

// PreparedAndDecideAfterChangeRound tests coming to consensus after preparing and then changing round.
type PreparedAndDecideAfterChangeRound struct {
	instance   *ibft2.Instance
	inputValue []byte
	lambda     []byte
}

// Name returns test name
func (test *PreparedAndDecideAfterChangeRound) Name() string {
	return "pre-prepare -> prepare -> change round -> prepare -> decide"
}

// Prepare prepares the test
func (test *PreparedAndDecideAfterChangeRound) Prepare(t *testing.T) {
	test.lambda = []byte{1, 2, 3, 4}
	test.inputValue = spectesting.TestInputValue()

	test.instance = spectesting.TestIBFTInstance(t, test.lambda)
	test.instance.State().Round.Set(1)

	// load messages to queue
	for _, msg := range test.MessagesSequence(t) {
		test.instance.MsgQueue.AddMessage(&network.Message{
			SignedMessage: msg,
			Type:          network.NetworkMsg_IBFTType,
		})
	}
}

// MessagesSequence includes test messages
func (test *PreparedAndDecideAfterChangeRound) MessagesSequence(t *testing.T) []*proto.SignedMessage {
	signersMap := map[uint64]*bls.SecretKey{
		1: spectesting.TestSKs()[0],
		2: spectesting.TestSKs()[1],
		3: spectesting.TestSKs()[2],
		4: spectesting.TestSKs()[3],
	}

	return []*proto.SignedMessage{
		spectesting.PrePrepareMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 1, 1),

		spectesting.PrepareMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 1, 1),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[1], test.lambda, test.inputValue, 1, 2),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[2], test.lambda, test.inputValue, 1, 3),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[3], test.lambda, test.inputValue, 1, 4),

		spectesting.ChangeRoundMsgWithPrepared(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, signersMap, 2, 1, 1),
		spectesting.ChangeRoundMsgWithPrepared(t, spectesting.TestSKs()[1], test.lambda, test.inputValue, signersMap, 2, 1, 2),
		spectesting.ChangeRoundMsgWithPrepared(t, spectesting.TestSKs()[2], test.lambda, test.inputValue, signersMap, 2, 1, 3),
		spectesting.ChangeRoundMsgWithPrepared(t, spectesting.TestSKs()[3], test.lambda, test.inputValue, signersMap, 2, 1, 4),

		spectesting.PrePrepareMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 2, 1),

		spectesting.PrepareMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 2, 1),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[1], test.lambda, test.inputValue, 2, 2),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[2], test.lambda, test.inputValue, 2, 3),
		spectesting.PrepareMsg(t, spectesting.TestSKs()[3], test.lambda, test.inputValue, 2, 4),

		spectesting.CommitMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 2, 1),
		spectesting.CommitMsg(t, spectesting.TestSKs()[1], test.lambda, test.inputValue, 2, 2),
		spectesting.CommitMsg(t, spectesting.TestSKs()[2], test.lambda, test.inputValue, 2, 3),
		spectesting.CommitMsg(t, spectesting.TestSKs()[3], test.lambda, test.inputValue, 2, 4),
	}
}

// Run runs the test
func (test *PreparedAndDecideAfterChangeRound) Run(t *testing.T) {
	// pre-prepare
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)

	// prepare
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.SimulateTimeout(test.instance, 2)

	// change round
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	err := test.instance.JustifyRoundChange(2)
	require.NoError(t, err)

	// justify pre-prepare
	err = test.instance.JustifyPrePrepare(2, test.inputValue)
	require.NoError(t, err)
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)

	// process all messages
	for {
		if res, _ := test.instance.ProcessMessage(); !res {
			break
		}
	}
	require.EqualValues(t, proto.RoundState_Decided, test.instance.State().Stage.Get())
}
