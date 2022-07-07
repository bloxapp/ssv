package peers

import (
	"context"
	"github.com/bloxapp/ssv/network/records"
	connmgrcore "github.com/libp2p/go-libp2p-core/connmgr"
	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/zap"
)

// ConnManager is a wrapper on top of go-libp2p-core/connmgr.ConnManager.
// exposing an abstract interface so we can have the flexibility of doing some stuff manually
// rather than relaying on libp2p's connection manager.
type ConnManager interface {
	// TagBestPeers tags the best n peers from the given list, based on subnets distribution scores.
	TagBestPeers(n int, mySubnets records.Subnets, allPeers []peer.ID, topicMaxPeers int)
	// TrimPeers will trim unprotected peers.
	TrimPeers(ctx context.Context, net libp2pnetwork.Network)
}

// NewConnManager creates a new conn manager, it can be called multiple times
// but concurrency of multiple managers is not supported.
func NewConnManager(logger *zap.Logger, connMgr connmgrcore.ConnManager, subnetsIdx SubnetsIndex) ConnManager {
	return &connManager{
		logger, connMgr, subnetsIdx,
	}
}

// connManager implements ConnManager
type connManager struct {
	logger      *zap.Logger
	connManager connmgrcore.ConnManager
	subnetsIdx  SubnetsIndex
}

func (c connManager) TagBestPeers(n int, mySubnets records.Subnets, allPeers []peer.ID, topicMaxPeers int) {
	bestPeers := c.getBestPeers(n, mySubnets, topicMaxPeers, allPeers)
	c.logger.Debug("found best peers",
		zap.Int("allPeers", len(allPeers)),
		zap.Int("bestPeers", len(bestPeers)))
	if len(bestPeers) == 0 {
		return
	}
	for _, pid := range allPeers {
		if _, ok := bestPeers[pid]; ok {
			c.connManager.Protect(pid, "ssv/subnets")
			continue
		}
		c.connManager.Unprotect(pid, "ssv/subnets")
	}
}

func (c connManager) TrimPeers(ctx context.Context, net libp2pnetwork.Network) {
	//n.connManager.TrimOpenConns(ctx)
	allPeers := net.Peers()
	for _, pid := range allPeers {
		if !c.connManager.IsProtected(pid, "ssv/subnets") {
			err := net.ClosePeer(pid)
			if err != nil {
				c.logger.Debug("could not close trimmed peer",
					zap.String("pid", pid.String()), zap.Error(err))
			}
		}
	}
}

// getBestPeers loop over all the existing peers and returns the best set
// according to the number of shared subnets,
// while considering subnets with low peer count to be more important.
// it enables to distribute peers connections across subnets in a balanced way.
func (c connManager) getBestPeers(n int, mySubnets records.Subnets, topicMaxPeers int, allPeers []peer.ID) map[peer.ID]int {
	peerScores := make(map[peer.ID]int)
	if len(allPeers) < n {
		for _, p := range allPeers {
			peerScores[p] = 1
		}
		return peerScores
	}
	stats := c.subnetsIdx.GetSubnetsStats()
	minSubnetPeers := (len(allPeers) / 10) + 1
	subnetsScores := GetSubnetsDistributionScores(stats, minSubnetPeers, mySubnets, topicMaxPeers)
	for _, pid := range allPeers {
		var peerScore int
		subnets := c.subnetsIdx.GetPeerSubnets(pid)
		for subnet, val := range subnets {
			if val == byte(0) && subnetsScores[subnet] < 0 {
				peerScore -= subnetsScores[subnet]
			} else {
				peerScore += subnetsScores[subnet]
			}
		}
		// adding the number of shared subnets to the score, considering only up to 25% subnets
		shared := records.SharedSubnets(subnets, mySubnets, len(mySubnets)/4)
		peerScore += len(shared) / 2
		c.logger.Debug("peer score", zap.String("id", pid.String()), zap.Int("score", peerScore))
		peerScores[pid] = peerScore
	}

	return GetTopScores(peerScores, n)
}
