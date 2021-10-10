package api

import (
	"github.com/bloxapp/ssv/pubsub"
	"github.com/bloxapp/ssv/utils/tasks"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
)

// WebSocketServer is responsible for managing all
type WebSocketServer interface {
	Start(addr string) error
	OutboundSubject() pubsub.Publisher
	UseQueryHandler(handler QueryMessageHandler)
}

// wsServer is an implementation of WebSocketServer
type wsServer struct {
	logger *zap.Logger
	// outbound is a subject for writing messages
	outbound pubsub.Subject

	handler QueryMessageHandler

	adapter WebSocketAdapter

	router *http.ServeMux
}

// NewWsServer creates a new instance
func NewWsServer(logger *zap.Logger, adapter WebSocketAdapter, handler QueryMessageHandler, mux *http.ServeMux) WebSocketServer {
	ws := wsServer{
		logger.With(zap.String("component", "exporter/api/server")),
		pubsub.NewSubject(logger.With(zap.String("component", "exporter/api/server/outbound-subject"))),
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

func (ws *wsServer) OutboundSubject() pubsub.Publisher {
	return ws.outbound
}

// handleQuery receives query message and respond async
func (ws *wsServer) handleQuery(conn Connection) {
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
	out, err := ws.outbound.Register(cid)
	if err != nil {
		ws.logger.Error("could not register outbound subject",
			zap.Error(err), zap.String("cid", cid))
	}
	defer ws.outbound.Deregister(cid)

	ws.processOutboundForConnection(conn, out, cid)
}

func (ws *wsServer) processOutboundForConnection(conn Connection, out pubsub.SubjectChannel, cid string) {
	logger := ws.logger.
		With(zap.String("cid", cid))

	for m := range out {
		nm, ok := m.(NetworkMessage)
		if !ok {
			logger.Warn("could not parse message")
			continue
		}
		logger.Debug("sending outbound",
			zap.String("msg.type", string(nm.Msg.Type)))
		err := tasks.Retry(func() error {
			return ws.adapter.Send(conn, &nm.Msg)
		}, 3)
		if err != nil {
			logger.Error("could not send message", zap.Error(err))
			break
		}
	}
}
