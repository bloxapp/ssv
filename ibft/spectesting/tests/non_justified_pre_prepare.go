package tests

import (
	"github.com/bloxapp/ssv/ibft"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/ibft/spectesting"
	"github.com/bloxapp/ssv/network"
	"testing"
)

// NonJustifiedPrePrepapre tests coming to consensus after a non prepared change round
type NonJustifiedPrePrepapre struct {
	instance   *ibft.Instance
	inputValue []byte
	lambda     []byte
}

// Name returns test name
func (test *NonJustifiedPrePrepapre) Name() string {
	return "pre-prepare -> simulate round timeout -> unjustified pre-prepare"
}

// Prepare prepares the test
func (test *NonJustifiedPrePrepapre) Prepare(t *testing.T) {
	test.lambda = []byte{1, 2, 3, 4}
	test.inputValue = spectesting.TestInputValue()

	test.instance = spectesting.TestIBFTInstance(t, test.lambda)
	test.instance.State.Round = 1

	// load messages to queue
	for _, msg := range test.MessagesSequence(t) {
		test.instance.MsgQueue.AddMessage(&network.Message{
			SignedMessage: msg,
			Type:          network.NetworkMsg_IBFTType,
		})
	}
}

// MessagesSequence includes all test messages
func (test *NonJustifiedPrePrepapre) MessagesSequence(t *testing.T) []*proto.SignedMessage {
	return []*proto.SignedMessage{
		spectesting.PrePrepareMsg(t, spectesting.TestSKs()[0], test.lambda, test.inputValue, 1, 1),
		spectesting.PrePrepareMsg(t, spectesting.TestSKs()[1], test.lambda, test.inputValue, 2, 2),
	}
}

// Run runs the test
func (test *NonJustifiedPrePrepapre) Run(t *testing.T) {
	// pre-prepare
	spectesting.RequireReturnedTrueNoError(t, test.instance.ProcessMessage)
	spectesting.SimulateTimeout(test.instance, 2)

	// try to broadcast unjustified pre-prepare
	spectesting.RequireReturnedTrueWithError(t, test.instance.ProcessMessage, "pre-prepare message sender (id 2) is not the round's leader (expected 1)")
}
