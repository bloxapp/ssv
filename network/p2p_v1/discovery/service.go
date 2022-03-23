package discovery

import (
	"context"
	"github.com/bloxapp/ssv/network/p2p_v1/peers"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

const (
	//udp4 = "udp4"
	//udp6 = "udp6"
	tcp = "tcp"
)

// CheckPeerLimit enables listener to check peers limit
type CheckPeerLimit func() bool

// PeerEvent is the data passed to peer handler
type PeerEvent struct {
	AddrInfo peer.AddrInfo
	Node     *enode.Node
}

// HandleNewPeer is the function interface for handling new peer
type HandleNewPeer func(e PeerEvent)

// PingHandler is used to facilitate ping requests
type PingHandler func(ctx context.Context, id peer.ID) error

// Options represents the options passed to create a service
type Options struct {
	Logger *zap.Logger

	Host       host.Host
	DiscV5Opts *DiscV5Options
	ConnIndex  peers.ConnectionIndex

	HostAddress string
	HostDNS     string

	PingHandler PingHandler
}

// Service is the interface for discovery
type Service interface {
	discovery.Discovery

	RegisterSubnets(subnets ...int) error
	DeregisterSubnets(subnets ...int) error
	Bootstrap(handler HandleNewPeer) error
}

// NewService creates new discovery.Service
func NewService(ctx context.Context, opts Options) (Service, error) {
	if opts.DiscV5Opts == nil {
		return NewLocalDiscovery(ctx, opts.Logger, opts.Host)
	}
	return newDiscV5Service(ctx, &opts)
}
