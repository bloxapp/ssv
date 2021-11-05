package api

import (
	"context"
	"github.com/bloxapp/ssv/pubsub"
	"github.com/bloxapp/ssv/utils/tasks"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

// WebSocketServer is responsible for managing all
type WebSocketServer interface {
	Start(addr string) error
	OutboundEmitter() pubsub.EventPublisher
	UseQueryHandler(handler QueryMessageHandler)
}

// wsServer is an implementation of WebSocketServer
type wsServer struct {
	logger *zap.Logger
	// outbound is a subject for writing messages
	outbound pubsub.Emitter

	handler QueryMessageHandler

	adapter WebSocketAdapter

	router *http.ServeMux
}

// NewWsServer creates a new instance
func NewWsServer(logger *zap.Logger, adapter WebSocketAdapter, handler QueryMessageHandler, mux *http.ServeMux) WebSocketServer {
	ws := wsServer{
		logger.With(zap.String("component", "exporter/api/server")),
		pubsub.NewEmitter(),
		handler, adapter, mux,
	}
	return &ws
}

func (ws *wsServer) UseQueryHandler(handler QueryMessageHandler) {
	ws.handler = handler
}

func (ws *wsServer) Start(addr string) error {
	if ws.adapter == nil {
		return errors.New("websocket adapter is missing")
	}
	ws.adapter.RegisterHandler(ws.router, "/query", ws.handleQuery)
	ws.adapter.RegisterHandler(ws.router, "/stream", ws.handleStream)

	ws.logger.Info("starting websocket server",
		zap.String("addr", addr),
		zap.Strings("endPoints", []string{"/query", "/stream"}))

	err := http.ListenAndServe(addr, ws.router)
	if err != nil {
		ws.logger.Warn("could not start http server", zap.Error(err))
	}
	return err
}

func (ws *wsServer) OutboundEmitter() pubsub.EventPublisher {
	return ws.outbound
}

// handleQuery receives query message and respond async
func (ws *wsServer) handleQuery(conn Connection) {
	if ws.handler == nil {
		return
	}
	cid := ConnectionID(conn)
	logger := ws.logger.With(zap.String("cid", cid))

	for {
		var nm NetworkMessage
		var incoming Message
		err := ws.adapter.Receive(conn, &incoming)
		if err != nil {
			if ws.adapter.IsCloseError(err) { // stop on any close error
				logger.Debug("failed to read message as the connection was closed", zap.Error(err))
				return
			}
			ws.logger.Warn("could not read incoming message", zap.Error(err))
			nm = NetworkMessage{incoming, err, conn}
		} else {
			nm = NetworkMessage{incoming, nil, conn}
		}
		// handler is processing the request
		ws.handler(&nm)

		err = tasks.Retry(func() error {
			return ws.adapter.Send(conn, &nm.Msg)
		}, 3)
		if err != nil {
			logger.Error("could not send message", zap.Error(err))
			break
		}
	}
}

// handleQuery receives query message and respond async
func (ws *wsServer) handleStream(conn Connection) {
	cid := ConnectionID(conn)
	logger := ws.logger.
		With(zap.String("cid", cid))
	// messages are being collected into a slice and picked up in another goroutine.
	//
	// the reason is that messages cannot be sent on a different goroutine,
	// but we can't use the same goroutine to pick up messages and send requests
	// as sending blocks the goroutine from picking up messages from the channel.
	out, outDone := ws.outbound.Channel("out")
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	var mut sync.Mutex
	running := true
	var msgs []NetworkMessage
	msgLimit := 1000

	go func() {
		defer outDone()

		for m := range out {
			if ctx.Err() != nil {
				return
			}
			nm, ok := m.(NetworkMessage)
			if !ok {
				logger.Warn("could not parse message")
				continue
			}
			mut.Lock()
			if len(msgs) < msgLimit {
				msgs = append(msgs, nm)
				mut.Unlock()
				reportStreamOutboundQueueCount(cid, true)
				continue
			}
			running = false
			mut.Unlock()
			logger.Error("queue is full, stopped listen on outbound channel", zap.Any("msg", nm.Msg))
			return
		}
	}()

	for {
		mut.Lock()
		if !running {
			mut.Unlock()
			return
		}
		if len(msgs) == 0 {
			mut.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		nm := msgs[0]
		msgs = msgs[1:]
		mut.Unlock()
		reportStreamOutboundQueueCount(cid, false)
		logger.Debug("sending outbound",
			zap.String("msg.type", string(nm.Msg.Type)), zap.Any("msg", nm.Msg))
		err := ws.send(ctx, conn, nm.Msg)
		reportStreamOutbound(cid, err)
		if err != nil {
			logger.Error("could not send message", zap.Error(err))
			return
		}
	}
}

// send takes the given message and try (3 times) to send it
// the whole operation will timeout after 3 sec
func (ws *wsServer) send(ctx context.Context, conn Connection, msg Message) error {
	sendCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	return tasks.RetryWithContext(sendCtx, func() error {
		return ws.adapter.Send(conn, &msg)
	}, 3)
}
