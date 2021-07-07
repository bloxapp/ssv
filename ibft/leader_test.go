package ibft

import (
	"fmt"
	"github.com/bloxapp/ssv/ibft/leader/deterministic"
	msgcontinmem "github.com/bloxapp/ssv/ibft/msgcont/inmem"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/validator/storage"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestLeaderCalculation(t *testing.T) {
	// leader selection
	l, err := deterministic.New(append([]byte{1, 2, 3, 2, 5, 6, 1, 1}, []byte(strconv.FormatUint(1, 10))...), 4)
	require.NoError(t, err)

	_, nodes := GenerateNodes(4)
	instance := &Instance{
		PrepareMessages: msgcontinmem.New(3),
		Config:          proto.DefaultConsensusParams(),
		ValidatorShare:  &storage.Share{Committee: nodes, NodeID: 1},
		State: &proto.State{
			Round:         1,
			PreparedRound: 0,
			PreparedValue: nil,
		},
		LeaderSelector: l,
	}

	for i := 1; i < 50; i++ {
		t.Run(fmt.Sprintf("round %d", i), func(t *testing.T) {
			instance.BumpRound()
			require.EqualValues(t, uint64(i%4+1), instance.ThisRoundLeader())
			require.EqualValues(t, uint64(i+1), instance.State.Round)

			if instance.ThisRoundLeader() == 1 {
				require.True(t, instance.IsLeader())
			} else {
				require.False(t, instance.IsLeader())
			}
		})
	}
}
