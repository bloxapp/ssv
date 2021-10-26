package ibft

import (
	"github.com/bloxapp/ssv/ibft/leader/constant"
	"github.com/bloxapp/ssv/ibft/leader/deterministic"
	"github.com/bloxapp/ssv/network/local"
	"github.com/bloxapp/ssv/utils/threadsafe"
	"github.com/bloxapp/ssv/validator/storage"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	msgcontinmem "github.com/bloxapp/ssv/ibft/instance/msgcont/inmem"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/utils/dataval/bytesval"
)

func TestJustifyPrePrepareAfterChangeRoundPrepared(t *testing.T) {
	secretKeys, nodes := GenerateNodes(4)
	value := []byte(time.Now().Weekday().String())
	instance := &Instance{
		PrePrepareMessages:  msgcontinmem.New(3, 2),
		ChangeRoundMessages: msgcontinmem.New(3, 2),
		Config:              proto.DefaultConsensusParams(),
		state: &proto.State{
			Round:         threadsafe.Uint64(1),
			Lambda:        threadsafe.BytesS("Lambda"),
			PreparedRound: threadsafe.Uint64(0),
			PreparedValue: threadsafe.Bytes(nil),
		},
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
		},
		ValueCheck: bytesval.NewNotEqualBytes(value),
		Logger:     zaptest.NewLogger(t),
	}

	t.Run("not quorum, not justified", func(t *testing.T) {
		// change round no quorum
		msg := SignMsg(t, 1, secretKeys[1], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value: changeRoundDataToBytes(&proto.ChangeRoundData{
				PreparedRound: 1,
				PreparedValue: value,
			}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)

		// no quorum achieved, err
		msg = SignMsg(t, 1, secretKeys[1], &proto.Message{
			Type:   proto.RoundState_PrePrepare,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value:  value,
		})
		instance.PrePrepareMessages.AddMessage(msg)
		err := instance.JustifyPrePrepare(2, value)
		require.EqualError(t, err, "no change round quorum")
	})

	t.Run("change round quorum, justified", func(t *testing.T) {
		// test justified change round
		msg := SignMsg(t, 2, secretKeys[2], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value: changeRoundDataToBytes(&proto.ChangeRoundData{
				PreparedRound: 1,
				PreparedValue: value,
			}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)
		msg = SignMsg(t, 3, secretKeys[3], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value: changeRoundDataToBytes(&proto.ChangeRoundData{
				PreparedRound: 1,
				PreparedValue: value,
			}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)

		err := instance.JustifyPrePrepare(2, value)
		require.NoError(t, err)
	})

	t.Run("wrong value, unjustified", func(t *testing.T) {
		err := instance.JustifyPrePrepare(2, []byte("wrong value"))
		require.EqualError(t, err, "preparedValue different than highest prepared")
	})
}

func TestJustifyPrePrepareAfterChangeRoundNoPrepare(t *testing.T) {
	secretKeys, nodes := GenerateNodes(4)
	value := []byte(time.Now().Weekday().String())
	instance := &Instance{
		PrePrepareMessages:  msgcontinmem.New(3, 2),
		PrepareMessages:     msgcontinmem.New(3, 2),
		ChangeRoundMessages: msgcontinmem.New(3, 2),
		Config:              proto.DefaultConsensusParams(),
		state: &proto.State{
			Round:         threadsafe.Uint64(1),
			Lambda:        threadsafe.BytesS("Lambda"),
			PreparedRound: threadsafe.Uint64(0),
			PreparedValue: threadsafe.Bytes(nil),
		},
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
		},
		ValueCheck: bytesval.NewNotEqualBytes(value),
		Logger:     zaptest.NewLogger(t),
	}

	t.Run("no change round quorum, not justified", func(t *testing.T) {
		// change round no quorum
		msg := SignMsg(t, 1, secretKeys[1], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)

		msg = SignMsg(t, 2, secretKeys[2], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)

		// no quorum achieved, can't justify
		err := instance.JustifyPrePrepare(2, nil)
		require.EqualError(t, err, "no change round quorum")
	})

	t.Run("change round quorum, justified", func(t *testing.T) {
		// test justified change round
		msg := SignMsg(t, 3, secretKeys[3], &proto.Message{
			Type:   proto.RoundState_ChangeRound,
			Round:  2,
			Lambda: []byte("Lambda"),
			Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
		})
		instance.ChangeRoundMessages.AddMessage(msg)

		// quorum achieved, can justify
		err := instance.JustifyPrePrepare(2, nil)
		require.NoError(t, err)
	})

	t.Run("any value can be in pre-prepare", func(t *testing.T) {
		require.NoError(t, instance.JustifyPrePrepare(2, []byte("wrong value")))
	})
}

func TestUponPrePrepareHappyFlow(t *testing.T) {
	secretKeys, nodes := GenerateNodes(4)
	leader, err := deterministic.New(append([]byte{1, 2, 3, 2, 5, 6, 1, 1}, []byte(strconv.FormatUint(1, 10))...), 4)
	require.NoError(t, err)
	instance := &Instance{
		PrePrepareMessages: msgcontinmem.New(3, 2),
		PrepareMessages:    msgcontinmem.New(3, 2),
		Config:             proto.DefaultConsensusParams(),
		state: &proto.State{
			Round:         threadsafe.Uint64(1),
			Lambda:        threadsafe.BytesS("Lambda"),
			PreparedRound: threadsafe.Uint64(0),
			PreparedValue: threadsafe.Bytes(nil),
			SeqNumber:     threadsafe.Uint64(0),
			Stage:         threadsafe.Int32(int32(proto.RoundState_NotStarted)),
		},
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
			PublicKey: secretKeys[1].GetPublicKey(),
		},
		ValueCheck:     bytesval.NewEqualBytes([]byte(time.Now().Weekday().String())),
		Logger:         zaptest.NewLogger(t),
		network:        local.NewLocalNetwork(),
		LeaderSelector: leader,
		signer:         newTestSigner(),
	}

	// test happy flow
	msg := SignMsg(t, 1, secretKeys[1], &proto.Message{
		Type:   proto.RoundState_PrePrepare,
		Round:  1,
		Lambda: []byte("Lambda"),
		Value:  []byte(time.Now().Weekday().String()),
	})
	err = instance.PrePrepareMsgPipelineV0().Run(msg)
	require.NoError(t, err)
	msgs := instance.PrePrepareMessages.ReadOnlyMessagesByRound(1)
	require.Len(t, msgs, 1)
	require.NotNil(t, msgs[0])
	require.True(t, instance.State().Stage.Get() == int32(proto.RoundState_PrePrepare))

	// return nil if another pre-prepare received.
	err = instance.UponPrePrepareMsg().Run(msg)
	require.NoError(t, err)
}

func TestInstance_JustifyPrePrepare(t *testing.T) {
	secretKeys, nodes := GenerateNodes(4)
	instance := &Instance{
		ChangeRoundMessages: msgcontinmem.New(3, 2),
		Config:              proto.DefaultConsensusParams(),
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
		},
		state: &proto.State{
			Round:         threadsafe.Uint64(1),
			PreparedRound: threadsafe.Uint64(0),
			PreparedValue: threadsafe.Bytes(nil),
		},
		network: local.NewLocalNetwork(),
	}

	err := instance.JustifyPrePrepare(1, nil)
	require.NoError(t, err)

	// try to justify round 2 without round change
	instance.State().Round.Set(2)
	err = instance.JustifyPrePrepare(2, nil)
	require.EqualError(t, err, "no change round quorum")

	// test no change round quorum
	msg := &proto.Message{
		Type:   proto.RoundState_ChangeRound,
		Round:  2,
		Lambda: []byte("lambdas"),
		Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
	}
	instance.ChangeRoundMessages.AddMessage(SignMsg(t, 1, secretKeys[1], msg))

	msg = &proto.Message{
		Type:   proto.RoundState_ChangeRound,
		Round:  2,
		Lambda: []byte("lambdas"),
		Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
	}
	instance.ChangeRoundMessages.AddMessage(SignMsg(t, 2, secretKeys[2], msg))

	err = instance.JustifyPrePrepare(2, nil)
	require.EqualError(t, err, "no change round quorum")

	// test with quorum of change round
	msg = &proto.Message{
		Type:   proto.RoundState_ChangeRound,
		Round:  2,
		Lambda: []byte("lambdas"),
		Value:  changeRoundDataToBytes(&proto.ChangeRoundData{}),
	}
	instance.ChangeRoundMessages.AddMessage(SignMsg(t, 3, secretKeys[3], msg))

	err = instance.JustifyPrePrepare(2, nil)
	require.NoError(t, err)
}

func TestPrePreparePipeline(t *testing.T) {
	sks, nodes := GenerateNodes(4)
	instance := &Instance{
		PrepareMessages: msgcontinmem.New(3, 2),
		Config:          proto.DefaultConsensusParams(),
		ValidatorShare: &storage.Share{
			Committee: nodes,
			NodeID:    1,
			PublicKey: sks[1].GetPublicKey(),
		},
		state: &proto.State{
			Round:     threadsafe.Uint64(1),
			Lambda:    threadsafe.Bytes(nil),
			SeqNumber: threadsafe.Uint64(0),
		},
		LeaderSelector: &constant.Constant{LeaderIndex: 1},
	}
	pipeline := instance.PrePrepareMsgPipelineV0()
	require.EqualValues(t, "combination of: combination of: basic msg validation, type check, lambda, sequence, authorize, validate pre-prepare, , add pre-prepare msg, if first pipeline non error, continue to second, ", pipeline.Name())
}
