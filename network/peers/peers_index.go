package peers

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"strconv"
	"sync"
	"time"

	"github.com/bloxapp/ssv/network/records"
	"github.com/bloxapp/ssv/utils/rsaencryption"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	nodeInfoKey = "nodeInfo"
)

// MaxPeersProvider returns the max peers for the given topic.
// empty string means that we want to check the total max peers (for all topics).
type MaxPeersProvider func(topic string) int

// NetworkKeyProvider is a function that provides the network private key
type NetworkKeyProvider func() libp2pcrypto.PrivKey

// peersIndex implements Index interface.
type peersIndex struct {
	netKeyProvider NetworkKeyProvider
	network        libp2pnetwork.Network

	states        *nodeStates
	scoreIdx      ScoreIndex
	subnets       SubnetsIndex
	nodeInfoStore *nodeInfoStore

	selfLock *sync.RWMutex
	self     *records.NodeInfo

	maxPeers MaxPeersProvider
}

// NewPeersIndex creates a new Index
func NewPeersIndex(logger *zap.Logger, network libp2pnetwork.Network, self *records.NodeInfo, maxPeers MaxPeersProvider,
	netKeyProvider NetworkKeyProvider, subnetsCount int, pruneTTL time.Duration) Index {
	return &peersIndex{
		network:        network,
		states:         newNodeStates(pruneTTL),
		scoreIdx:       newScoreIndex(),
		subnets:        newSubnetsIndex(subnetsCount),
		nodeInfoStore:  newNodeInfoStore(network),
		self:           self,
		selfLock:       &sync.RWMutex{},
		maxPeers:       maxPeers,
		netKeyProvider: netKeyProvider,
	}
}

// IsBad returns whether the given peer is bad.
// a peer is considered to be bad if one of the following applies:
// - pruned (that was not expired)
// - bad score
func (pi *peersIndex) IsBad(logger *zap.Logger, id peer.ID) bool {
	if pi.states.pruned(id.String()) {
		logger.Debug("bad peer (pruned)")
		return true
	}
	// TODO: check scores
	threshold := -10000.0
	scores, err := pi.GetScore(id, "")
	if err != nil {
		// logger.Debug("could not read score", zap.Error(err))
		return false
	}
	for _, score := range scores {
		if score.Value < threshold {
			logger.Debug("bad peer (low score)")
			return true
		}
	}
	return false
}

func (pi *peersIndex) Connectedness(id peer.ID) libp2pnetwork.Connectedness {
	return pi.network.Connectedness(id)
}

func (pi *peersIndex) CanConnect(id peer.ID) bool {
	cntd := pi.network.Connectedness(id)
	switch cntd {
	case libp2pnetwork.Connected:
		fallthrough
	case libp2pnetwork.CannotConnect: // recently failed to connect
		return false
	default:
	}
	return true
}

func (pi *peersIndex) Limit(dir libp2pnetwork.Direction) bool {
	maxPeers := pi.maxPeers("")
	peers := pi.network.Peers()
	return len(peers) > maxPeers
}

func (pi *peersIndex) UpdateSelfRecord(newSelf *records.NodeInfo) {
	pi.selfLock.Lock()
	defer pi.selfLock.Unlock()

	pi.self = newSelf
}

func (pi *peersIndex) Self() *records.NodeInfo {
	return pi.self
}

func (pi *peersIndex) SelfSealed(sender, recipient peer.ID, operatorPrivateKey *rsa.PrivateKey) ([]byte, error) {
	pi.selfLock.Lock()
	defer pi.selfLock.Unlock()

	publicKey, err := rsaencryption.ExtractPublicKey(operatorPrivateKey)
	if err != nil {
		return nil, err
	}

	handshakeData := records.HandshakeData{
		SenderPeerID:    sender,
		RecipientPeerID: recipient,
		Timestamp:       time.Now().Round(30 * time.Second),
		SenderPubKey:    publicKey,
	}
	hash := handshakeData.Hash()

	signature, err := rsa.SignPKCS1v15(rand.Reader, operatorPrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	sealed, err := pi.self.Seal(pi.netKeyProvider(), handshakeData, signature)
	if err != nil {
		return nil, err
	}

	return sealed, nil
}

// AddNodeInfo adds a new node info
func (pi *peersIndex) AddNodeInfo(logger *zap.Logger, id peer.ID, nodeInfo *records.NodeInfo) (bool, error) {
	switch pi.states.State(id) {
	case StateReady:
		return true, nil
	case StateIndexing:
		// TODO: handle
		return true, nil
	case StatePruned:
		return false, ErrWasPruned
	case StateUnknown:
	default:
	}
	pid := id.String()
	pi.states.setState(pid, StateIndexing)
	added, err := pi.nodeInfoStore.Add(logger, id, nodeInfo)
	if err != nil || !added {
		pi.states.setState(pid, StateUnknown)
	} else {
		pi.states.setState(pid, StateReady)
	}
	return added, err
}

// GetNodeInfo returns the node info of the given peer
func (pi *peersIndex) GetNodeInfo(id peer.ID) (*records.NodeInfo, error) {
	switch pi.states.State(id) {
	case StateIndexing:
		return nil, ErrIndexingInProcess
	case StatePruned:
		return nil, ErrWasPruned
	case StateUnknown:
		return nil, ErrNotFound
	default:
	}
	// if in good state -> get node info
	ni, err := pi.nodeInfoStore.Get(id)
	if err == peerstore.ErrNotFound {
		return nil, ErrNotFound
	}

	return ni, err
}

func (pi *peersIndex) State(id peer.ID) NodeState {
	return pi.states.State(id)
}

// Score adds score to the given peer
func (pi *peersIndex) Score(id peer.ID, scores ...*NodeScore) error {
	return pi.scoreIdx.Score(id, scores...)
}

// GetScore returns the desired score for the given peer
func (pi *peersIndex) GetScore(id peer.ID, names ...string) ([]NodeScore, error) {
	var scores []NodeScore
	switch pi.states.State(id) {
	case StateIndexing:
		// TODO: handle
		return scores, nil
	case StatePruned:
		return nil, ErrWasPruned
	case StateUnknown:
		return nil, ErrNotFound
	default:
	}

	return pi.scoreIdx.GetScore(id, names...)
}

// Prune set prune state for the given peer
func (pi *peersIndex) Prune(id peer.ID) error {
	return pi.states.Prune(id)
}

// EvictPruned changes to ready state instead of pruned
func (pi *peersIndex) EvictPruned(id peer.ID) {
	pi.states.EvictPruned(id)
}

// GC does garbage collection on current peers and states
func (pi *peersIndex) GC() {
	pi.states.GC()
}

func (pi *peersIndex) UpdatePeerSubnets(id peer.ID, s records.Subnets) bool {
	return pi.subnets.UpdatePeerSubnets(id, s)
}

func (pi *peersIndex) GetSubnetPeers(subnet int) []peer.ID {
	return pi.subnets.GetSubnetPeers(subnet)
}

func (pi *peersIndex) GetPeerSubnets(id peer.ID) records.Subnets {
	return pi.subnets.GetPeerSubnets(id)
}

func (pi *peersIndex) GetSubnetsStats() *SubnetsStats {
	mySubnets, err := records.Subnets{}.FromString(pi.self.Metadata.Subnets)
	if err != nil {
		mySubnets, _ = records.Subnets{}.FromString(records.ZeroSubnets)
	}
	stats := pi.subnets.GetSubnetsStats()
	if stats == nil {
		return nil
	}
	stats.Connected = make([]int, len(stats.PeersCount))
	var sumConnected int
	for subnet, count := range stats.PeersCount {
		metricsSubnetsKnownPeers.WithLabelValues(strconv.Itoa(subnet)).Set(float64(count))
		metricsMySubnets.WithLabelValues(strconv.Itoa(subnet)).Set(float64(mySubnets[subnet]))
		peers := pi.subnets.GetSubnetPeers(subnet)
		connectedCount := 0
		for _, p := range peers {
			if pi.Connectedness(p) == libp2pnetwork.Connected {
				connectedCount++
			}
		}
		stats.Connected[subnet] = connectedCount
		sumConnected += connectedCount
		metricsSubnetsConnectedPeers.WithLabelValues(strconv.Itoa(subnet)).Set(float64(connectedCount))
	}
	if len(stats.PeersCount) > 0 {
		stats.AvgConnected = sumConnected / len(stats.PeersCount)
	}

	return stats
}

// Close closes peer index
func (pi *peersIndex) Close() error {
	_ = pi.states.Close()
	if err := pi.network.Peerstore().Close(); err != nil {
		return errors.Wrap(err, "could not close peerstore")
	}
	return nil
}
