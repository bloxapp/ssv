package ibft

import (
	"github.com/bloxapp/ssv/ibft/eventqueue"
	msgcontinmem "github.com/bloxapp/ssv/ibft/msgcont/inmem"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/ibft/roundtimer"
	"github.com/bloxapp/ssv/network/msgqueue"
	"github.com/bloxapp/ssv/utils/dataval/bytesval"
	"github.com/bloxapp/ssv/utils/threadsafe"
	"github.com/bloxapp/ssv/validator/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

func TestChangeRoundTimer(t *testing.T) {
	secretKeys, nodes := GenerateNodes(4)
	instance := &Instance{
		MsgQueue:            msgqueue.New(),
		eventQueue:          eventqueue.New(),
		ChangeRoundMessages: msgcontinmem.New(3),
		PrepareMessages:     msgcontinmem.New(3),
		Config: &proto.InstanceConfig{
			RoundChangeDurationSeconds:   0.2,
			LeaderPreprepareDelaySeconds: 0.1,
		},
		State: &proto.State{
			Round:         threadsafe.Uint64(1),
			Stage:         threadsafe.Int32(int32(proto.RoundState_PrePrepare)),
			Lambda:        threadsafe.BytesS("Lambda"),
			SeqNumber:     threadsafe.Uint64(1),
			PreparedValue: threadsafe.Bytes(nil),
			PreparedRound: threadsafe.Uint64(0),
		},
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
			ShareKey:  secretKeys[1],
			PublicKey: secretKeys[1].GetPublicKey(),
		},
		ValueCheck: bytesval.New([]byte(time.Now().Weekday().String())),
		Logger:     zaptest.NewLogger(t),
		roundTimer: roundtimer.New(),
	}
	go instance.startRoundTimerLoop()
	instance.initialized = true
	time.Sleep(time.Millisecond * 200)

	instance.resetRoundTimer()
	time.Sleep(time.Millisecond * 500)
	instance.eventQueue.Pop()()
	require.EqualValues(t, 2, instance.State.Round.Get())
}
