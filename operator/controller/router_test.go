package controller

import (
	"context"
	"fmt"
	"sync"
	"testing"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/stretchr/testify/require"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/networkconfig"
	"github.com/bloxapp/ssv/protocol/v2/ssv/queue"
)

func TestRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.TestLogger(t)

	router := newMessageRouter(logger)

	expectedCount := 1000
	count := 0

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		cn := router.GetMessageChan()
		for msg := range cn {
			require.NotNil(t, msg)
			count++
			if ctx.Err() != nil || count >= expectedCount {
				return
			}
		}
	}()

	for i := 0; i < expectedCount; i++ {
		msg := &queue.DecodedSSVMessage{
			SSVMessage: &spectypes.SSVMessage{
				MsgType: spectypes.MsgType(i % 3),
				MsgID:   spectypes.NewMsgID(networkconfig.TestNetwork.Domain, []byte{1, 1, 1, 1, 1}, spectypes.BNRoleAttester),
				Data:    []byte(fmt.Sprintf("data-%d", i)),
			},
		}

		router.Route(context.TODO(), msg)
		if i%2 == 0 {
			go router.Route(context.TODO(), msg)
		}
	}

	wg.Wait()

	require.Equal(t, count, expectedCount)
}
