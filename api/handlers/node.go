package handlers

import (
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/bloxapp/ssv/api"
	networkpeers "github.com/bloxapp/ssv/network/peers"
	"github.com/bloxapp/ssv/nodeprobe"
)

const (
	healthyPeerCount = 20
	healthyInbounds  = 4
)

type TopicIndex interface {
	PeersByTopic() ([]peer.ID, map[string][]peer.ID)
}

type allPeersAndTopics struct {
	AllPeers     []peer.ID    `json:"all_peers"`
	PeersByTopic []topicIndex `json:"peers_by_topic"`
}

type topicIndex struct {
	TopicName string    `json:"topic"`
	Peers     []peer.ID `json:"peers"`
}

type connection struct {
	Address   string `json:"address"`
	Direction string `json:"direction"`
}

type peerInfo struct {
	ID            peer.ID      `json:"id"`
	Addresses     []string     `json:"addresses"`
	Connections   []connection `json:"connections"`
	Connectedness string       `json:"connectedness"`
	Subnets       string       `json:"subnets"`
	Version       string       `json:"version"`
}

type nodeIdentity struct {
	PeerID    peer.ID  `json:"peer_id"`
	Addresses []string `json:"addresses"`
	Subnets   string   `json:"subnets"`
	Version   string   `json:"version"`
}

type healthStatus struct {
	err error
}

func (h healthStatus) MarshalJSON() ([]byte, error) {
	if h.err == nil {
		return json.Marshal("good")
	}
	return json.Marshal(fmt.Sprintf("bad: %s", h.err.Error()))
}

type healthCheck struct {
	P2P           healthStatus `json:"p2p"`
	BeaconNode    healthStatus `json:"beacon_node"`
	ExecutionNode healthStatus `json:"execution_node"`
	EventSyncer   healthStatus `json:"event_syncer"`
	Advanced      struct {
		Peers           int      `json:"peers"`
		InboundConns    int      `json:"inbound_conns"`
		OutboundConns   int      `json:"outbound_conns"`
		ListenAddresses []string `json:"p2p_listen_addresses"`
	} `json:"advanced"`
}

func (hc healthCheck) String() string {
	b, err := json.MarshalIndent(hc, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshalling healthCheckJSON: %s", err.Error())
	}
	return string(b)
}

type Node struct {
	ListenAddresses []string
	PeersIndex      networkpeers.Index
	TopicIndex      TopicIndex
	Network         network.Network
	NodeProber      *nodeprobe.Prober
	Signer          func(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error)
}

func (h *Node) Identity(w http.ResponseWriter, r *http.Request) error {
	nodeInfo := h.PeersIndex.Self()
	resp := nodeIdentity{
		PeerID:  h.Network.LocalPeer(),
		Subnets: nodeInfo.Metadata.Subnets,
		Version: nodeInfo.Metadata.NodeVersion,
	}
	for _, addr := range h.Network.ListenAddresses() {
		resp.Addresses = append(resp.Addresses, addr.String())
	}
	return api.Render(w, r, resp)
}

func (h *Node) Peers(w http.ResponseWriter, r *http.Request) error {
	peers := h.Network.Peers()
	resp := h.peers(peers)
	return api.Render(w, r, resp)
}

func (h *Node) Topics(w http.ResponseWriter, r *http.Request) error {
	peers, byTopic := h.TopicIndex.PeersByTopic()

	resp := allPeersAndTopics{
		AllPeers: peers,
	}
	for topic, peers := range byTopic {
		resp.PeersByTopic = append(resp.PeersByTopic, topicIndex{TopicName: topic, Peers: peers})
	}

	return api.Render(w, r, resp)
}

func (h *Node) Health(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	var resp healthCheck

	// Retrieve P2P listen addresses.
	resp.Advanced.ListenAddresses = h.ListenAddresses

	// Count peers and connections.
	peers := h.Network.Peers()
	for _, p := range h.peers(peers) {
		if p.Connectedness == network.Connected.String() {
			resp.Advanced.Peers++
		}
		for _, conn := range p.Connections {
			if conn.Direction == network.DirInbound.String() {
				resp.Advanced.InboundConns++
			} else {
				resp.Advanced.OutboundConns++
			}
		}
	}

	// Report whether P2P is healthy.
	if resp.Advanced.Peers == 0 {
		resp.P2P = healthStatus{errors.New("no peers are connected")}
	} else if resp.Advanced.Peers < healthyPeerCount {
		resp.P2P = healthStatus{errors.New("not enough connected peers")}
	} else if resp.Advanced.InboundConns < healthyInbounds {
		resp.P2P = healthStatus{errors.New("not enough inbound connections, port is likely not reachable")}
	}

	// Check the health of Ethereum nodes and EventSyncer.
	resp.BeaconNode = healthStatus{h.NodeProber.CheckBeaconNodeHealth(ctx)}
	resp.ExecutionNode = healthStatus{h.NodeProber.CheckExecutionNodeHealth(ctx)}
	resp.EventSyncer = healthStatus{(h.NodeProber.CheckEventSyncerHealth(ctx))}

	return api.Render(w, r, resp)
}

func (h *Node) peers(peers []peer.ID) []peerInfo {
	resp := make([]peerInfo, len(peers))
	for i, id := range peers {
		resp[i] = peerInfo{
			ID:            id,
			Connectedness: h.Network.Connectedness(id).String(),
			Subnets:       h.PeersIndex.GetPeerSubnets(id).String(),
		}

		for _, addr := range h.Network.Peerstore().Addrs(id) {
			resp[i].Addresses = append(resp[i].Addresses, addr.String())
		}

		conns := h.Network.ConnsToPeer(id)
		for _, conn := range conns {
			resp[i].Connections = append(resp[i].Connections, connection{
				Address:   conn.RemoteMultiaddr().String(),
				Direction: conn.Stat().Direction.String(),
			})
		}

		nodeInfo := h.PeersIndex.NodeInfo(id)
		if nodeInfo == nil {
			continue
		}
		resp[i].Version = nodeInfo.Metadata.NodeVersion
	}
	return resp
}
