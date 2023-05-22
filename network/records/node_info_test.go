package records

import (
	crand "crypto/rand"
	"reflect"
	"testing"

	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/require"
)

func TestNodeInfo_Seal_Consume(t *testing.T) {
	netKey, _, err := crypto.GenerateSecp256k1Key(crand.Reader)
	require.NoError(t, err)
	ni := &NodeInfo{
		ForkVersion: forksprotocol.GenesisForkVersion,
		NetworkID:   "testnet",
		Metadata: &NodeMetadata{
			NodeVersion:   "v0.1.12",
			ExecutionNode: "geth/x",
			ConsensusNode: "prysm/x",
			OperatorID:    "xxx",
		},
	}

	data, err := ni.Seal(netKey)
	require.NoError(t, err)

	parsedRec := &NodeInfo{}
	require.NoError(t, parsedRec.Consume(data))

	require.True(t, reflect.DeepEqual(ni, parsedRec))
}

func TestNodeInfo_Marshal_Unmarshal(t *testing.T) {
	ni := &NodeInfo{
		ForkVersion: forksprotocol.GenesisForkVersion,
		NetworkID:   "testnet",
		Metadata: &NodeMetadata{
			NodeVersion:   "v0.1.12",
			ExecutionNode: "geth/x",
			ConsensusNode: "prysm/x",
			OperatorID:    "xxx",
		},
	}

	data, err := ni.MarshalRecord()
	require.NoError(t, err)

	parsedRec := &NodeInfo{}
	require.NoError(t, parsedRec.UnmarshalRecord(data))

	require.True(t, reflect.DeepEqual(ni, parsedRec))
}
