package peers

import (
	"sync"

	"github.com/bloxapp/ssv/network/records"
	"github.com/libp2p/go-libp2p/core/peer"
)

// subnetsIndex implements SubnetsIndex
type subnetsIndex struct {
	subnets     [][]peer.ID
	peerSubnets map[peer.ID]records.Subnets

	lock *sync.RWMutex
}

func newSubnetsIndex(count int) SubnetsIndex {
	return &subnetsIndex{
		subnets:     make([][]peer.ID, count),
		peerSubnets: map[peer.ID]records.Subnets{},
		lock:        &sync.RWMutex{},
	}
}

func (si *subnetsIndex) UpdatePeerSubnets(id peer.ID, s records.Subnets) bool {
	si.lock.Lock()
	defer si.lock.Unlock()

	existing, ok := si.peerSubnets[id]
	if !ok {
		existing = make([]byte, 0)
	}
	diff := records.DiffSubnets(existing, s)
	if len(diff) == 0 {
		return false
	}
	si.peerSubnets[id] = s

diffLoop:
	for subnet, val := range diff {
		if subnet >= len(si.subnets) { // out of range
			continue
		}
		peers := si.subnets[subnet]
		if len(peers) == 0 {
			peers = make([]peer.ID, 0)
		}
		for i, p := range peers {
			if p == id {
				// skip if peer is already listed in a subnet to be added
				if val > byte(0) {
					continue diffLoop
				}
				// otherwise, remove peer from the subnet
				if i == 0 {
					if len(peers) == 1 {
						si.subnets[subnet] = make([]peer.ID, 0)
					} else {
						si.subnets[subnet] = peers[1:]
					}
					continue diffLoop
				}
				si.subnets[subnet] = append(peers[:i], peers[i:]...)
				continue diffLoop
			}
		}
		if val > byte(0) {
			si.subnets[subnet] = append(peers, id)
		}
	}
	return true
}

func (si *subnetsIndex) GetSubnetPeers(subnet int) []peer.ID {
	si.lock.RLock()
	defer si.lock.RUnlock()

	peers := si.subnets[subnet]
	if len(peers) == 0 {
		return nil
	}
	cp := make([]peer.ID, len(peers))
	copy(cp, peers)
	return cp
}

// GetSubnetsStats collects and returns subnets stats
func (si *subnetsIndex) GetSubnetsStats() *SubnetsStats {
	si.lock.RLock()
	defer si.lock.RUnlock()

	stats := &SubnetsStats{
		PeersCount: make([]int, len(si.subnets)),
	}
	for subnet, peers := range si.subnets {
		stats.PeersCount[subnet] = len(peers)
	}

	return stats
}

func (si *subnetsIndex) GetPeerSubnets(id peer.ID) records.Subnets {
	si.lock.RLock()
	defer si.lock.RUnlock()

	subnets, ok := si.peerSubnets[id]
	if !ok {
		return nil
	}
	cp := make(records.Subnets, len(subnets))
	copy(cp, subnets)
	return cp
}

// GetSubnetsDistributionScores returns current subnets scores based on peers distribution.
// subnets with low peer count would get higher score, and overloaded subnets gets a lower score.
func GetSubnetsDistributionScores(stats *SubnetsStats, minPerSubnet int, mySubnets records.Subnets, topicMaxPeers int) []float64 {
	allSubs, _ := records.Subnets{}.FromString(records.AllSubnets)
	activeSubnets := records.SharedSubnets(allSubs, mySubnets, 0)

	scores := make([]float64, len(allSubs))
	for _, s := range activeSubnets {
		var connected int
		if s < len(stats.Connected) {
			connected = stats.Connected[s]
		}
		scores[s] = scoreSubnet(connected, minPerSubnet, topicMaxPeers)
	}
	return scores
}

func scoreSubnet(connected, min, max int) float64 {
	// scarcityFactor is the factor by which the score is increased for
	// subnets with fewer than the desired minimum number of peers.
	const scarcityFactor = 2.0

	if connected <= 0 {
		return 2.0 * scarcityFactor
	}

	if connected > max {
		// Linear scaling when connected is above the desired maximum.
		return -1.0 * (float64(connected-max) / float64(2*(max-min)))
	}

	if connected < min {
		// Proportional scaling when connected is less than the desired minimum.
		return 1.0 + (float64(min-connected)/float64(min))*scarcityFactor
	}

	// Linear scaling when connected is between min and max.
	proportion := float64(connected-min) / float64(max-min)
	return 1 - proportion
}
