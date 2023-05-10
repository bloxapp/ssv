package connections

import (
	"context"
	"time"

	"github.com/bloxapp/ssv/logging/fields"
	"github.com/bloxapp/ssv/network/peers"
	"github.com/bloxapp/ssv/network/records"
	"github.com/bloxapp/ssv/utils/tasks"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ConnHandler handles new connections (inbound / outbound) using libp2pnetwork.NotifyBundle
type ConnHandler interface {
	Handle(logger *zap.Logger) *libp2pnetwork.NotifyBundle
}

// connHandler implements ConnHandler
type connHandler struct {
	ctx context.Context

	handshaker      Handshaker
	subnetsProvider SubnetsProvider
	subnetsIndex    peers.SubnetsIndex
	connIdx         peers.ConnectionIndex
}

// NewConnHandler creates a new connection handler
func NewConnHandler(ctx context.Context, handshaker Handshaker, subnetsProvider SubnetsProvider, subnetsIndex peers.SubnetsIndex, connIdx peers.ConnectionIndex) ConnHandler {
	return &connHandler{
		ctx:             ctx,
		handshaker:      handshaker,
		subnetsProvider: subnetsProvider,
		subnetsIndex:    subnetsIndex,
		connIdx:         connIdx,
	}
}

// Handle configures a network notifications handler that handshakes and tracks all p2p connections
func (ch *connHandler) Handle(logger *zap.Logger) *libp2pnetwork.NotifyBundle {

	q := tasks.NewExecutionQueue(time.Millisecond*10, tasks.WithoutErrors())

	go func() {
		c, cancel := context.WithCancel(ch.ctx)
		defer cancel()
		defer q.Stop()
		q.Start()
		<-c.Done()
	}()

	disconnect := func(net libp2pnetwork.Network, conn libp2pnetwork.Conn) {
		id := conn.RemotePeer()
		errClose := net.ClosePeer(id)
		if errClose == nil {
			metricsFilteredConnections.Inc()
		}

		logger.Debug("peer was disconnected", zap.String("selfPeer", conn.LocalPeer().String()), zap.String("otherPeer", id.String()))
	}

	onNewConnection := func(net libp2pnetwork.Network, conn libp2pnetwork.Conn) error {
		id := conn.RemotePeer()
		logger := logger.With(zap.String("targetPeer", id.String()))
		ok, err := ch.handshake(logger, conn)
		if err != nil {
			logger.Debug("could not handshake with peer", zap.Error(err))
		}
		if !ok {
			disconnect(net, conn)
			return err
		}
		if ch.connIdx.Limit(conn.Stat().Direction) {
			disconnect(net, conn)
			return errors.New("reached peers limit")
		}
		if !ch.checkSubnets(logger, conn) && conn.Stat().Direction != libp2pnetwork.DirOutbound {
			disconnect(net, conn)
			return errors.New("peer doesn't share enough subnets")
		}
		//logger.Debug("new connection is ready",
		//	zap.String("dir", conn.Stat().Direction.String()))
		metricsConnections.Inc()
		return nil
	}

	return &libp2pnetwork.NotifyBundle{
		ConnectedF: func(net libp2pnetwork.Network, conn libp2pnetwork.Conn) {
			if conn == nil || conn.RemoteMultiaddr() == nil {
				return
			}
			id := conn.RemotePeer()
			q.QueueDistinct(func() error {
				return onNewConnection(net, conn)
			}, id.String())
		},
		DisconnectedF: func(net libp2pnetwork.Network, conn libp2pnetwork.Conn) {
			if conn == nil || conn.RemoteMultiaddr() == nil {
				return
			}
			// skip if we are still connected to the peer
			if net.Connectedness(conn.RemotePeer()) == libp2pnetwork.Connected {
				return
			}
			metricsConnections.Dec()
		},
		// TODO: enable metrics
		//OpenedStreamF: func(network libp2pnetwork.Network, stream libp2pnetwork.Stream) {
		//	if conn := stream.Conn(); conn != nil {
		//		metricsStreams.WithLabelValues(string(stream.Protocol())).Inc()
		//	}
		// },
		//ClosedStreamF: func(network libp2pnetwork.Network, stream libp2pnetwork.Stream) {
		//	if conn := stream.Conn(); conn != nil {
		//		metricsStreams.WithLabelValues(string(stream.Protocol())).Dec()
		//	}
		// },
	}
}

func (ch *connHandler) handshake(logger *zap.Logger, conn libp2pnetwork.Conn) (bool, error) {
	err := ch.handshaker.Handshake(logger, conn)
	if err != nil {
		switch {
		case errors.Is(err, peers.ErrIndexingInProcess), errors.Is(err, errHandshakeInProcess), errors.Is(err, peerstore.ErrNotFound):
			// ignored errors
			return true, nil
		case errors.Is(err, errPeerWasFiltered), errors.Is(err, errUnknownUserAgent):
			// ignored errors but we still close connection
			return false, nil
		default:
		}
		return false, err
	}
	return true, nil
}

func (ch *connHandler) checkSubnets(logger *zap.Logger, conn libp2pnetwork.Conn) bool {
	pid := conn.RemotePeer()
	subnets := ch.subnetsIndex.GetPeerSubnets(pid)
	if len(subnets) == 0 {
		// no subnets for this peer
		return false
	}
	mySubnets := ch.subnetsProvider()

	logger = logger.With(fields.PeerID(pid), fields.Subnets(subnets),
		zap.String("mySubnets", mySubnets.String()))

	if mySubnets.String() == records.ZeroSubnets { // this node has no subnets
		return true
	}
	shared := records.SharedSubnets(mySubnets, subnets, 1)
	logger.Debug("checking subnets", zap.Ints("shared", shared))

	return len(shared) == 1
}
