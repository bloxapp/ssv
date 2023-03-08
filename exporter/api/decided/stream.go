package decided

import (
	"encoding/hex"
	"fmt"
	"time"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/exporter/api"
	"github.com/bloxapp/ssv/protocol/v2/qbft/controller"
)

// NewStreamPublisher handles incoming newly decided messages.
// it forward messages to websocket stream, where messages are cached (1m TTL) to avoid flooding
func NewStreamPublisher(logger *zap.Logger, ws api.WebSocketServer) controller.NewDecidedHandler {
	logger = logger.Named("NewDecidedHandler")
	c := cache.New(time.Minute, time.Minute*3/2)
	feed := ws.BroadcastFeed()
	return func(msg *specqbft.SignedMessage) {
		identifier := hex.EncodeToString(msg.Message.Identifier)
		key := fmt.Sprintf("%s:%d:%d", identifier, msg.Message.Height, len(msg.Signers))
		_, ok := c.Get(key)
		if ok {
			return
		}
		c.SetDefault(key, true)
		logger.Debug("broadcast decided stream",
			zap.String("identifier", identifier),
			zap.Uint64("height", uint64(msg.Message.Height)))
		feed.Send(api.NewDecidedAPIMsg(msg))
	}
}
