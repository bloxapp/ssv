package types

import (
	"testing"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/stretchr/testify/require"

	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
)

// TODO: enhance these tests

func TestShareMetadata_BelongsToOperatorID(t *testing.T) {
	metadata := &SSVShare{
		Share: spectypes.Share{
			Committee: []*spectypes.Operator{
				{
					OperatorID: 1,
					PubKey:     []byte{1, 1, 1, 1},
				},
				{
					OperatorID: 2,
					PubKey:     []byte{2, 2, 2, 2},
				},
				{
					OperatorID: 3,
					PubKey:     []byte{3, 3, 3, 3},
				},
				{
					OperatorID: 4,
					PubKey:     []byte{4, 4, 4, 4},
				},
			},
		},
	}

	require.True(t, metadata.BelongsToOperatorID(1))
	require.False(t, metadata.BelongsToOperatorID(10))
}

func TestShare_IsOperatorShare(t *testing.T) {
	metadata := &SSVShare{
		Metadata: Metadata{
			Operators: [][]byte{
				{1, 1, 1, 1},
				{2, 2, 2, 2},
			},
		},
	}

	require.True(t, metadata.BelongsToOperator(string([]byte{1, 1, 1, 1})))
	require.False(t, metadata.BelongsToOperator(string([]byte{1, 2, 3, 4})))
}

func TestShare_HasStats(t *testing.T) {
	tt := []struct {
		Name          string
		ShareMetadata *SSVShare
		Result        bool
	}{
		{
			Name:          "Metadata == nil",
			ShareMetadata: nil,
			Result:        false,
		},
		{
			Name: "BeaconMetadata == nil",
			ShareMetadata: &SSVShare{
				Metadata: Metadata{
					BeaconMetadata: nil,
				},
			},
			Result: false,
		},
		{
			Name: "BeaconMetadata != nil",
			ShareMetadata: &SSVShare{
				Metadata: Metadata{
					BeaconMetadata: &beaconprotocol.ValidatorMetadata{},
				},
			},
			Result: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Result, tc.ShareMetadata.HasBeaconMetadata())
		})
	}
}
