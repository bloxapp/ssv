package p2pv1

import (
	"strconv"

	spectypes "github.com/bloxapp/ssv-spec/types"

	"github.com/bloxapp/ssv/logging/fields"
	"github.com/bloxapp/ssv/network/discovery"
	"github.com/bloxapp/ssv/network/peers"
	"github.com/bloxapp/ssv/network/peers/connections"
	"github.com/bloxapp/ssv/network/streams"
	"github.com/bloxapp/ssv/network/topics"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/utils/format"
)

type Metrics interface {
	connections.Metrics
	topics.Metrics
	peers.Metrics
	streams.Metrics
	discovery.Metrics
	AllConnectedPeers(count int)
	ConnectedTopicPeers(topic string, count int)
	PeersIdentity(opPKHash, opID, nodeVersion, pid, nodeType string)
	RouterIncoming(msgType spectypes.MsgType)
}

var unknown = "unknown"

func (n *p2pNetwork) reportAllPeers(logger *zap.Logger) func() {
	return func() {
		pids := n.host.Network().Peers()
		logger.Debug("connected peers status", fields.Count(len(pids)))
		n.metrics.AllConnectedPeers(len(pids))
	}
}

func (n *p2pNetwork) reportPeerIdentities(logger *zap.Logger) func() {
	return func() {
		pids := n.host.Network().Peers()
		for _, pid := range pids {
			n.reportPeerIdentity(logger, pid)
		}
	}
}

func (n *p2pNetwork) reportTopics(logger *zap.Logger) func() {
	return func() {
		topics := n.topicsCtrl.Topics()
		nTopics := len(topics)
		logger.Debug("connected topics", fields.Count(nTopics))
		for _, name := range topics {
			n.reportTopicPeers(logger, name)
		}
	}
}

func (n *p2pNetwork) reportTopicPeers(logger *zap.Logger, name string) {
	peers, err := n.topicsCtrl.Peers(name)
	if err != nil {
		logger.Warn("could not get topic peers", fields.Topic(name), zap.Error(err))
		return
	}
	logger.Debug("topic peers status", fields.Topic(name), fields.Count(len(peers)), zap.Any("peers", peers))
	n.metrics.ConnectedTopicPeers(name, len(peers))
}

func (n *p2pNetwork) reportPeerIdentity(logger *zap.Logger, pid peer.ID) {
	opPKHash, opID, nodeVersion, nodeType := unknown, unknown, unknown, unknown
	ni := n.idx.NodeInfo(pid)
	if ni != nil {
		if ni.Metadata != nil {
			nodeVersion = ni.Metadata.NodeVersion
		}
		nodeType = "operator"
		if len(opPKHash) == 0 && nodeVersion != unknown {
			nodeType = "exporter"
		}
	}

	if pubKey, ok := n.operatorPKHashToPKCache.Get(opPKHash); ok {
		operatorData, found, opDataErr := n.nodeStorage.GetOperatorDataByPubKey(nil, pubKey)
		if opDataErr == nil && found {
			opID = strconv.FormatUint(operatorData.ID, 10)
		}
	} else {
		operators, err := n.nodeStorage.ListOperators(nil, 0, 0)
		if err != nil {
			logger.Warn("failed to get all operators for reporting", zap.Error(err))
		}

		for _, operator := range operators {
			pubKeyHash := format.OperatorID(operator.PublicKey)
			n.operatorPKHashToPKCache.Set(pubKeyHash, operator.PublicKey)
			if pubKeyHash == opPKHash {
				opID = strconv.FormatUint(operator.ID, 10)
			}
		}
	}

	state := n.idx.State(pid)
	logger.Debug("peer identity",
		fields.PeerID(pid),
		zap.String("node_version", nodeVersion),
		zap.String("operator_id", opID),
		zap.String("state", state.String()),
	)
	n.metrics.PeersIdentity(opPKHash, opID, nodeVersion, pid.String(), nodeType)
}

//
// func reportLastMsg(pid string) {
//	MetricsPeerLastMsg.WithLabelValues(pid).Set(float64(timestamp()))
//}
//
// func timestamp() int64 {
//	return time.Now().UnixNano() / int64(time.Millisecond)
//}
