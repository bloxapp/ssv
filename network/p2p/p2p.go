package p2p

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"github.com/bloxapp/ssv/network/forks"
	"github.com/bloxapp/ssv/utils/rsaencryption"
	"github.com/bloxapp/ssv/utils/tasks"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prysmaticlabs/prysm/async"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/libp2p/go-libp2p"
	p2pHost "github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/p2p/peers"

	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/network"
)

const (
	// DiscoveryInterval is how often we re-publish our mDNS records.
	DiscoveryInterval = time.Second

	// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
	DiscoveryServiceTag = "bloxstaking.ssv"

	// MsgChanSize is the buffer size of the message channel
	MsgChanSize = 128

	topicPrefix = "bloxstaking.ssv"

	syncStreamProtocol = "/sync/0.0.1"
)

type listener struct {
	msgCh     chan *proto.SignedMessage
	sigCh     chan *proto.SignedMessage
	decidedCh chan *proto.SignedMessage
	syncCh    chan *network.SyncChanObj
}

// p2pNetwork implements network.Network interface using P2P
type p2pNetwork struct {
	ctx             context.Context
	cfg             *Config
	listenersLock   sync.Locker
	dv5Listener     discv5Listener
	eNode           *enode.LocalNode
	listeners       []listener
	logger          *zap.Logger
	privKey         *ecdsa.PrivateKey
	peers           *peers.Status
	host            p2pHost.Host
	pubsub          *pubsub.PubSub
	peersIndex      PeersIndex
	operatorPrivKey *rsa.PrivateKey
	fork            forks.Fork

	psSubscribedTopics map[string]bool
	psTopicsLock       *sync.RWMutex

	reportLastMsg bool
}

// New is the constructor of p2pNetworker
func New(ctx context.Context, logger *zap.Logger, cfg *Config) (network.Network, error) {
	// init empty topics map
	cfg.Topics = make(map[string]*pubsub.Topic)

	logger = logger.With(zap.String("component", "p2p"))

	n := &p2pNetwork{
		ctx:                ctx,
		cfg:                cfg,
		listenersLock:      &sync.Mutex{},
		logger:             logger,
		operatorPrivKey:    cfg.OperatorPrivateKey,
		psSubscribedTopics: make(map[string]bool),
		psTopicsLock:       &sync.RWMutex{},
		reportLastMsg:      cfg.ReportLastMsg,
		fork:               cfg.Fork,
	}

	if err := n.withNetworkKey(cfg.NetworkPrivateKey); err != nil {
		return nil, errors.Wrap(err, "Failed to generate p2p private key")
	}

	opts, err := n.buildOptions(cfg)
	if err != nil {
		logger.Fatal("could not build libp2p options", zap.Error(err))
	}
	host, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create p2p host")
	}
	n.host = host
	n.cfg.HostID = host.ID()

	if len(cfg.Enr) > 0 {
		n.cfg.Discv5BootStrapAddr = n.parseBootStrapAddrs(TransformEnr(n.cfg.Enr))
	}

	n.logger = logger.With(zap.String("id", n.host.ID().String()))
	n.logger.Info("listening on port", zap.String("port", n.host.Addrs()[0].String()))
	n.host.Network().Notify(n.networkNotifiee(cfg.TryReconnect))
	ps, err := n.newGossipPubsub(cfg)
	if err != nil {
		n.logger.Error("failed to start pubsub", zap.Error(err))
		return nil, errors.Wrap(err, "failed to start pubsub")
	}
	n.pubsub = ps

	var ids *identify.IDService

	if cfg.DiscoveryType == "mdns" {
		// Setup Local mDNS discovery
		if err := setupMdnsDiscovery(ctx, logger, n.host); err != nil {
			return nil, errors.Wrap(err, "failed to setup discovery")
		}
	} else if cfg.DiscoveryType == "discv5" {
		//host.RemoveStreamHandler(identify.IDDelta)
		ids, err = identify.NewIDService(host, identify.UserAgent(n.getUserAgent()))
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create p2p ID Service")
		}
		if err := setupDiscV5(ctx, n); err != nil {
			logger.Error("could not setup discv5", zap.Error(err))
			return nil, err
		}
		_ = n.verifyHostAddress()
	}
	n.peersIndex = NewPeersIndex(n.host, ids, n.logger)

	n.handleStream()

	n.watchPeers()

	return n, nil
}

func (n *p2pNetwork) withNetworkKey(priv *ecdsa.PrivateKey) error {
	if priv != nil {
		n.privKey = priv
	} else {
		privKey, err := privKey()
		if err != nil {
			return errors.Wrap(err, "Failed to generate p2p private key")
		}
		n.privKey = privKey
	}
	return nil
}

// reconnect tries to connect to a lost connection
func (n *p2pNetwork) reconnect(logger *zap.Logger, ai peer.AddrInfo) {
	limit := 128 * time.Second
	logger = logger.With(zap.String("who", "reconnect"))
	tasks.ExecWithInterval(func(lastTick time.Duration) (stop bool, cont bool) {
		if err := n.connectWithPeer(context.Background(), ai); err != nil {
			// stop after reaching limit
			if lastTick >= limit {
				logger.Debug("could not reconnect with peer, aborting")
				return true, false
			}
			logger.Debug("could not connect with peer, trying again")
			return false, false
		}
		logger.Debug("managed to reconnect with peer")
		return true, false
	}, 8*time.Second, limit)
}

func (n *p2pNetwork) watchPeers() {
	async.RunEvery(n.ctx, 1*time.Minute, func() {
		// index all peers and report
		go func() {
			n.peersIndex.Run()
			reportAllConnections(n)
		}()

		// topics peers
		n.psTopicsLock.RLock()
		defer n.psTopicsLock.RUnlock()
		for name, topic := range n.cfg.Topics {
			reportTopicPeers(n, name, topic)
		}
	})
}

func (n *p2pNetwork) SubscribeToValidatorNetwork(validatorPk *bls.PublicKey) error {
	n.psTopicsLock.Lock()
	defer n.psTopicsLock.Unlock()

	pubKey := validatorPk.SerializeToHexStr()

	if _, ok := n.cfg.Topics[pubKey]; !ok {
		if err := n.joinTopic(pubKey); err != nil {
			return errors.Wrap(err, "failed to join to topic")
		}
	}

	if !n.psSubscribedTopics[pubKey] {
		sub, err := n.cfg.Topics[pubKey].Subscribe()
		if err != nil {
			if err != pubsub.ErrTopicClosed {
				return errors.Wrap(err, "failed to subscribe on Topic")
			}
			// rejoin a topic in case it was closed, and trying to subscribe again
			if err := n.joinTopic(pubKey); err != nil {
				return errors.Wrap(err, "failed to join to topic")
			}
			sub, err = n.cfg.Topics[pubKey].Subscribe()
			if err != nil {
				return errors.Wrap(err, "failed to subscribe on Topic")
			}
		}
		n.psSubscribedTopics[pubKey] = true
		go func() {
			n.listen(sub)
			// mark topic as not subscribed
			n.psTopicsLock.Lock()
			defer n.psTopicsLock.Unlock()
			n.psSubscribedTopics[pubKey] = false
		}()
	}

	return nil
}

// joinTopic joins to the given topic and mark it in topics map
// this method is not thread-safe - should be called after psTopicsLock was acquired
func (n *p2pNetwork) joinTopic(pubKey string) error {
	topic, err := n.pubsub.Join(getTopicName(pubKey))
	if err != nil {
		return errors.Wrap(err, "failed to join to topic")
	}
	n.cfg.Topics[pubKey] = topic
	return nil
}

// closeTopic closes the given topic
func (n *p2pNetwork) closeTopic(topicName string) error {
	n.psTopicsLock.RLock()
	defer n.psTopicsLock.RUnlock()

	pk := unwrapTopicName(topicName)
	if t, ok := n.cfg.Topics[pk]; ok {
		delete(n.cfg.Topics, pk)
		return t.Close()
	}
	return nil
}

// listen listens to some validator's topic
func (n *p2pNetwork) listen(sub *pubsub.Subscription) {
	t := sub.Topic()
	n.logger.Info("start listen to topic", zap.String("topic", t))
	for {
		select {
		case <-n.ctx.Done():
			sub.Cancel()
			if err := n.closeTopic(t); err != nil {
				n.logger.Error("failed to close topic", zap.String("topic", t), zap.Error(err))
			}
			n.logger.Info("closed topic", zap.String("topic", t))
		default:
			msg, err := sub.Next(n.ctx)
			if err != nil {
				n.logger.Error("failed to get message from subscription Topics", zap.Error(err))
				return
			}

			// For debugging
			n.logger.Debug("received raw network msg", zap.ByteString("network.Message bytes", msg.Data))

			cm, err := n.fork.DecodeNetworkMsg(msg.Data)
			if err != nil {
				n.logger.Error("failed to un-marshal message", zap.Error(err))
				continue
			}
			if n.reportLastMsg && len(msg.ReceivedFrom) > 0 {
				reportLastMsg(msg.ReceivedFrom.String())
			}
			n.propagateSignedMsg(cm)
		}
	}
}

// propagateSignedMsg takes an incoming message (from validator's topic)
// and propagates it to the corresponding internal listeners
func (n *p2pNetwork) propagateSignedMsg(cm *network.Message) {
	// TODO: find a better way to deal with nil message
	// 	i.e. avoid sending nil messages in the network
	if cm == nil || cm.SignedMessage == nil {
		n.logger.Debug("could not propagate nil message")
		return
	}

	switch cm.Type {
	case network.NetworkMsg_IBFTType:
		for _, ls := range n.listeners {
			if ls.msgCh != nil {
				ls.msgCh <- cm.SignedMessage
			}
		}
	case network.NetworkMsg_SignatureType:
		for _, ls := range n.listeners {
			if ls.sigCh != nil {
				ls.sigCh <- cm.SignedMessage
			}
		}
	case network.NetworkMsg_DecidedType:
		for _, ls := range n.listeners {
			if ls.decidedCh != nil {
				ls.decidedCh <- cm.SignedMessage
			}
		}
	default:
		n.logger.Error("received unsupported message", zap.Int32("msg type", int32(cm.Type)))
	}
}

// getTopic return topic by validator public key
func (n *p2pNetwork) getTopic(validatorPK []byte) (*pubsub.Topic, error) {
	n.psTopicsLock.RLock()
	defer n.psTopicsLock.RUnlock()

	if validatorPK == nil {
		return nil, errors.New("ValidatorPk is nil")
	}
	topic := n.fork.ValidatorTopicID(validatorPK)
	if _, ok := n.cfg.Topics[topic]; !ok {
		return nil, errors.New("topic is not exist or registered")
	}
	return n.cfg.Topics[topic], nil
}

// AllPeers returns all connected peers for a validator PK (except for the validator itself)
func (n *p2pNetwork) AllPeers(validatorPk []byte) ([]string, error) {
	topic, err := n.getTopic(validatorPk)
	if err != nil {
		return nil, err
	}

	return n.allPeersOfTopic(topic), nil
}

// AllPeers returns all connected peers for a validator PK (except for the validator itself and public peers like exporter)
func (n *p2pNetwork) allPeersOfTopic(topic *pubsub.Topic) []string {
	ret := make([]string, 0)

	skippedPeers := map[string]bool{
		n.cfg.ExporterPeerID: true,
	}

	for _, p := range topic.ListPeers() {
		if s := peerToString(p); !skippedPeers[s] {
			ret = append(ret, peerToString(p))
		}
	}

	return ret
}

// getTopicName return formatted topic name
func getTopicName(pk string) string {
	return fmt.Sprintf("%s.%s", topicPrefix, pk)
}

// getTopicName return formatted topic name
func unwrapTopicName(topicName string) string {
	return strings.Replace(topicName, fmt.Sprintf("%s.", topicPrefix), "", 1)
}

// MaxBatch returns the max batch response size
func (n *p2pNetwork) MaxBatch() uint64 {
	return n.cfg.MaxBatchResponse
}

func (n *p2pNetwork) verifyHostAddress() error {
	if n.cfg.HostAddress != "" {
		a := net.JoinHostPort(n.cfg.HostAddress, fmt.Sprintf("%d", n.cfg.TCPPort))
		if err := checkAddress(a); err != nil {
			n.logger.Debug("failed to check address", zap.String("addr", a), zap.String("err", err.Error()))
			return err
		}
		n.logger.Debug("address was checked successfully", zap.String("addr", a))
	}
	return nil
}

// checkAddress checks that some address is accessible
func checkAddress(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, time.Second*10)
	if err != nil {
		return errors.Wrap(err, "IP address is not accessible")
	}
	if err := conn.Close(); err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	return nil
}

// listen for new nodes watches for new nodes in the network and adds them to the peerstore.
func (n *p2pNetwork) getOperatorPubKey() (string, error) {
	if n.operatorPrivKey != nil {
		operatorPubKey, err := rsaencryption.ExtractPublicKey(n.operatorPrivKey)
		if err != nil || len(operatorPubKey) == 0 {
			n.logger.Error("could not extract operator public key", zap.Error(err))
			return "", errors.Wrap(err, "could not extract operator public key")
		}
		return operatorPubKey, nil
	}
	return "", nil
}
