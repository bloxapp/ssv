package sync

import (
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/storage/kv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestReqHandler_getHighestDecided(t *testing.T) {
	ibftStorage := ibftStorage(t)
	handler := ReqHandler{
		paginationMaxSize: 0,
		identifier:        []byte{1, 2, 3, 4},
		network:           nil,
		storage:           &ibftStorage,
		logger:            zap.L(),
	}

	t.Run("not found", func(t *testing.T) {
		h, err := handler.getHighestDecided()
		require.NoError(t, err)
		require.NotNil(t, h)
		require.Equal(t, h.Error, kv.EntryNotFoundError)
	})

	t.Run("valid", func(t *testing.T) {
		err := ibftStorage.SaveHighestDecidedInstance(&proto.SignedMessage{
			Message:   &proto.Message{
				Type:      proto.RoundState_Decided,
				Round:     1,
				Lambda:    []byte{1, 2, 3, 4},
				SeqNumber: 4,
			},
			Signature: []byte("sig"),
			SignerIds: []uint64{1, 2, 3},
		})
		require.NoError(t, err)

		h, err := handler.getHighestDecided()
		require.NoError(t, err)
		require.NotNil(t, h)
		require.Len(t, h.SignedMessages, 1)
	})

}
