package queue

import (
	"testing"

	"github.com/bloxapp/ssv-spec/qbft"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/stretchr/testify/require"
)

func TestPushAndPop(t *testing.T) {
	mockState := &State{
		HasRunningInstance: true,
		Height:             100,
		Slot:               64,
		Quorum:             4,
	}
	prioritizer := NewMessagePrioritizer(mockState)
	queue := New(prioritizer)

	// Push one.
	msg := decodeAndPush(t, queue, mockConsensusMessage{Height: 101, Type: qbft.PrepareMsgType}, mockState)
	require.Equal(t, 1, queue.Len())

	// Pop non-existing BeaconRole.
	popped := queue.Pop(types.BNRoleProposer)
	require.Nil(t, popped)

	// Pop one.
	popped = queue.Pop(msg.MsgID.GetRoleType())
	require.Equal(t, 0, queue.Len())
	require.Equal(t, msg, popped)

	// Pop nil.
	popped = queue.Pop(msg.MsgID.GetRoleType())
	require.Nil(t, popped)
}

func decodeAndPush(t *testing.T, queue Queue, msg mockMessage, state *State) *DecodedSSVMessage {
	decoded, err := DecodeSSVMessage(msg.ssvMessage(state))
	require.NoError(t, err)
	queue.Push(decoded)
	return decoded
}
