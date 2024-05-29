package instance

import (
	"testing"

	genesisspecqbft "github.com/ssvlabs/ssv-spec-pre-cc/qbft"
	genesisspectypes "github.com/ssvlabs/ssv-spec-pre-cc/types"
	spectypes "github.com/ssvlabs/ssv-spec/types"
	"github.com/stretchr/testify/require"

	genesistypes "github.com/ssvlabs/ssv/protocol/genesis/types"
)

var compactTests = []struct {
	name       string
	inputState *genesistypes.State
	inputMsg   *genesisspecqbft.SignedMessage
	expected   *genesistypes.State // if nil, expected to be equal to input
}{
	{
		name:       "empty",
		inputState: &genesistypes.State{},
		expected:   nil,
	},
	{
		name: "empty but not nil",
		inputState: &genesistypes.State{
			Round:                1,
			ProposeContainer:     &genesisspecqbft.MsgContainer{},
			PrepareContainer:     &genesisspecqbft.MsgContainer{},
			CommitContainer:      &genesisspecqbft.MsgContainer{},
			RoundChangeContainer: &genesisspecqbft.MsgContainer{},
		},
		expected: nil,
	},
	{
		name: "nothing to compact",
		inputState: &genesistypes.State{
			Round:                1,
			ProposeContainer:     mockContainer(1, 2),
			PrepareContainer:     mockContainer(1, 2),
			CommitContainer:      mockContainer(1, 2),
			RoundChangeContainer: mockContainer(1, 2),
		},
		expected: nil,
	},
	{
		name: "compact non-decided with previous rounds",
		inputState: &genesistypes.State{
			Round:                2,
			LastPreparedRound:    2,
			ProposeContainer:     mockContainer(1, 2),
			PrepareContainer:     mockContainer(1, 2),
			CommitContainer:      mockContainer(1, 2),
			RoundChangeContainer: mockContainer(1, 2),
		},
		expected: &genesistypes.State{
			Round:                2,
			LastPreparedRound:    2,
			ProposeContainer:     mockContainer(2),
			PrepareContainer:     mockContainer(2),
			CommitContainer:      mockContainer(2),
			RoundChangeContainer: mockContainer(2),
		},
	},
	{
		name: "compact non-decided with previous rounds except for prepared",
		inputState: &genesistypes.State{
			Round:                2,
			LastPreparedRound:    1,
			ProposeContainer:     mockContainer(1, 2),
			PrepareContainer:     mockContainer(1, 2),
			CommitContainer:      mockContainer(1, 2),
			RoundChangeContainer: mockContainer(1, 2),
		},
		expected: &genesistypes.State{
			Round:                2,
			LastPreparedRound:    1,
			ProposeContainer:     mockContainer(2),
			PrepareContainer:     mockContainer(1, 2),
			CommitContainer:      mockContainer(2),
			RoundChangeContainer: mockContainer(2),
		},
	},
	{
		name: "compact quorum decided with previous rounds",
		inputState: &genesistypes.State{
			Round:             3,
			LastPreparedRound: 3,
			Decided:           true,
			Share: &spectypes.Share{
				Committee: make([]*spectypes.ShareMember, 4),
			},
			ProposeContainer:     mockContainer(1, 2, 3, 4),
			PrepareContainer:     mockContainer(1, 2, 3, 4),
			CommitContainer:      mockContainer(1, 2, 3, 4),
			RoundChangeContainer: mockContainer(1, 2, 3, 4),
		},
		inputMsg: &genesisspecqbft.SignedMessage{
			Signers: []genesisspectypes.OperatorID{1, 2, 3},
		},
		expected: &genesistypes.State{
			Round:             3,
			LastPreparedRound: 3,
			Decided:           true,
			Share: &spectypes.Share{
				Committee: make([]*spectypes.ShareMember, 4),
			},
			ProposeContainer:     mockContainer(),
			PrepareContainer:     mockContainer(),
			CommitContainer:      mockContainer(3, 4),
			RoundChangeContainer: mockContainer(),
		},
	},
	{
		name: "compact whole committee decided with previous rounds",
		inputState: &genesistypes.State{
			Round:             2,
			LastPreparedRound: 2,
			Decided:           true,
			Share: &spectypes.Share{
				Committee: make([]*spectypes.ShareMember, 4),
			},
			ProposeContainer:     mockContainer(1, 2, 3, 4),
			PrepareContainer:     mockContainer(1, 2, 3, 4),
			CommitContainer:      mockContainer(1, 2, 3, 4),
			RoundChangeContainer: mockContainer(1, 2, 3, 4),
		},
		inputMsg: &genesisspecqbft.SignedMessage{
			Signers: []genesisspectypes.OperatorID{1, 2, 3, 4},
		},
		expected: &genesistypes.State{
			Round:             2,
			LastPreparedRound: 2,
			Decided:           true,
			Share: &spectypes.Share{
				Committee: make([]*spectypes.ShareMember, 4),
			},
			ProposeContainer:     mockContainer(),
			PrepareContainer:     mockContainer(),
			CommitContainer:      mockContainer(2, 3, 4),
			RoundChangeContainer: mockContainer(),
		},
	},
}

func TestCompact(t *testing.T) {
	for _, tt := range compactTests {
		t.Run(tt.name, func(t *testing.T) {
			inputStateBefore, err := tt.inputState.Encode()
			require.NoError(t, err)

			if tt.expected == nil {
				tt.expected = &genesistypes.State{}
				require.NoError(t, tt.expected.Decode(inputStateBefore))
			}

			// Test CompactCopy.
			stateCopy := CompactCopy(tt.inputState, tt.inputMsg)
			require.Equal(t, tt.expected, stateCopy)

			// Verify that input state was not modified by CompactCopy.
			inputStateAfter, err := tt.inputState.Encode()
			require.NoError(t, err)
			require.Equal(t, inputStateBefore, inputStateAfter)

			// Test Compact.
			Compact(tt.inputState, tt.inputMsg)
			require.Equal(t, tt.expected, tt.inputState)
		})
	}
}

func mockContainer(rounds ...genesisspecqbft.Round) *genesisspecqbft.MsgContainer {
	container := genesisspecqbft.NewMsgContainer()
	for _, round := range rounds {
		container.AddMsg(&genesisspecqbft.SignedMessage{
			Message: genesisspecqbft.Message{
				Round: round,
			},
		})
	}
	return container
}
