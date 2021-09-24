package beacon

import (
	"encoding/hex"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv/utils/logex"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"strings"
	"sync/atomic"
	"testing"
)

func init() {
	logex.Build("test", zap.InfoLevel, nil)
}

func TestValidatorMetadata_Status(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		meta := &ValidatorMetadata{
			Status: v1.ValidatorStateActiveOngoing,
		}
		require.True(t, meta.Activated())
		require.False(t, meta.Exiting())
		require.False(t, meta.Slashed())
	})

	t.Run("exited", func(t *testing.T) {
		meta := &ValidatorMetadata{
			Status: v1.ValidatorStateWithdrawalPossible,
		}
		require.True(t, meta.Exiting())
		require.True(t, meta.Activated())
		require.False(t, meta.Slashed())
	})

	t.Run("exitedSlashed", func(t *testing.T) {
		meta := &ValidatorMetadata{
			Status: v1.ValidatorStateExitedSlashed,
		}
		require.True(t, meta.Slashed())
		require.True(t, meta.Exiting())
		require.True(t, meta.Activated())
	})
}

func TestUpdateValidatorsMetadata(t *testing.T) {
	var updateCount uint64
	pks := []string{
		"a17bb48a3f8f558e29d08ede97d6b7b73823d8dc2e0530fe8b747c93d7d6c2755957b7ffb94a7cec830456fd5492ba19",
		"a0cf5642ed5aa82178a5f79e00292c5b700b67fbf59630ce4f542c392495d9835a99c826aa2459a67bc80867245386c6",
		"8bafb7165f42e1179f83b7fcd8fe940e60ed5933fac176fdf75a60838c688eaa3b57717be637bde1d5cebdadb8e39865",
	}

	decodeds := make([][]byte, len(pks))
	blsPubKeys := make([]spec.BLSPubKey, len(pks))
	for i, pk := range pks {
		blsPubKey := spec.BLSPubKey{}
		decoded, _ := hex.DecodeString(pk)
		copy(blsPubKey[:], decoded[:])
		decodeds[i] = decoded
		blsPubKeys[i] = blsPubKey
	}

	data := map[spec.BLSPubKey]*v1.Validator{}
	data[blsPubKeys[0]] = &v1.Validator{
		Index:     spec.ValidatorIndex(210961),
		Status:    v1.ValidatorStateWithdrawalPossible,
		Validator: &spec.Validator{PublicKey: blsPubKeys[0]},
	}
	data[blsPubKeys[1]] = &v1.Validator{
		Index:     spec.ValidatorIndex(213820),
		Status:    v1.ValidatorStateActiveOngoing,
		Validator: &spec.Validator{PublicKey: blsPubKeys[1]},
	}
	bc := NewMockBeacon(map[uint64][]*Duty{}, data)

	storage := NewMockValidatorMetadataStorage()

	onUpdated := func(pk string, meta *ValidatorMetadata) {
		joined := strings.Join(pks, ":")
		require.True(t, strings.Contains(joined, pk))
		require.True(t, meta.Index == spec.ValidatorIndex(210961) || meta.Index == spec.ValidatorIndex(213820))
		atomic.AddUint64(&updateCount, 1)
	}
	err := UpdateValidatorsMetadata([][]byte{blsPubKeys[0][:], blsPubKeys[1][:], blsPubKeys[2][:]}, storage, bc, onUpdated)
	require.Nil(t, err)
	require.Equal(t, uint64(2), updateCount)
	require.Equal(t, 2, storage.(*mockValidatorMetadataStorage).Size())
}
