package topics

import (
	"context"
	"github.com/bloxapp/ssv/logging"
	"net"
	"time"

	"github.com/bloxapp/ssv/network"
	"github.com/bloxapp/ssv/network/forks"
	"github.com/bloxapp/ssv/network/peers"
	"github.com/bloxapp/ssv/network/topics/params"
	"github.com/bloxapp/ssv/utils/async"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	// subscriptionRequestLimit sets an upper bound for the number of topic we are allowed to subscribe to.
	// 128 subnets + 1 safety buffer
	subscriptionRequestLimit = 128 + 1
)

// the following are kept in vars to allow flexibility (e.g. in tests)
var (
	// validationQueueSize is the size that we assign to the validation queue
	validationQueueSize = 512
	// outboundQueueSize is the size that we assign to the outbound message queue
	outboundQueueSize = 512
	// validateThrottle is the amount of goroutines used for pubsub msg validation
	validateThrottle = 8192
	// scoreInspectInterval is the interval for performing score inspect, which goes over all peers scores
	scoreInspectInterval = time.Minute
	// msgIDCacheTTL specifies how long a message ID will be remembered as seen, 6.4m (as ETH 2.0)
	msgIDCacheTTL = params.HeartbeatInterval * 550
)

// PububConfig is the needed config to instantiate pubsub
type PububConfig struct {
	Host        host.Host
	TraceLog    bool
	StaticPeers []peer.AddrInfo
	MsgHandler  PubsubMessageHandler
	// MsgValidatorFactory accepts the topic name and returns the corresponding msg validator
	// in case we need different validators for specific topics,
	// this should be the place to map a validator to topic
	MsgValidatorFactory func(string) MsgValidatorFunc
	ScoreIndex          peers.ScoreIndex
	Scoring             *ScoringConfig
	MsgIDHandler        MsgIDHandler
	Discovery           discovery.Discovery

	ValidateThrottle    int
	ValidationQueueSize int
	OutboundQueueSize   int
	MsgIDCacheTTL       time.Duration

	GetValidatorStats network.GetValidatorStats
}

// ScoringConfig is the configuration for peer scoring
type ScoringConfig struct {
	IPWhilelist        []*net.IPNet
	IPColocationWeight float64
	OneEpochDuration   time.Duration
}

// PubsubBundle includes the pubsub router, plus involved components
type PubsubBundle struct {
	PS         *pubsub.PubSub
	TopicsCtrl Controller
	Resolver   MsgPeersResolver
}

func (cfg *PububConfig) init() error {
	if cfg.Host == nil {
		return errors.New("bad args: missing host")
	}
	if cfg.OutboundQueueSize == 0 {
		cfg.OutboundQueueSize = outboundQueueSize
	}
	if cfg.ValidationQueueSize == 0 {
		cfg.ValidationQueueSize = validationQueueSize
	}
	if cfg.ValidateThrottle == 0 {
		cfg.ValidateThrottle = validateThrottle
	}
	if cfg.MsgIDCacheTTL == 0 {
		cfg.MsgIDCacheTTL = msgIDCacheTTL
	}
	return nil
}

// initScoring initializes scoring config
func (cfg *PububConfig) initScoring() {
	if cfg.Scoring == nil {
		cfg.Scoring = DefaultScoringConfig()
	}
}

// NewPubsub creates a new pubsub router and the necessary components
func NewPubsub(ctx context.Context, logger *zap.Logger, cfg *PububConfig, fork forks.Fork) (*pubsub.PubSub, Controller, error) {
	if err := cfg.init(); err != nil {
		return nil, nil, err
	}

	sf := newSubFilter(logger, fork, subscriptionRequestLimit)
	psOpts := []pubsub.Option{
		pubsub.WithSeenMessagesTTL(cfg.MsgIDCacheTTL),
		pubsub.WithPeerOutboundQueueSize(cfg.OutboundQueueSize),
		pubsub.WithValidateQueueSize(cfg.ValidationQueueSize),
		pubsub.WithValidateThrottle(cfg.ValidateThrottle),
		pubsub.WithSubscriptionFilter(sf),
		pubsub.WithGossipSubParams(params.GossipSubParams()),
		// pubsub.WithPeerFilter(func(pid peer.ID, topic string) bool {
		//	logger.Debug("pubsubTrace: filtering peer", zap.String("id", pid.String()), zap.String("topic", topic))
		//	return true
		// }),
	}

	if cfg.Discovery != nil {
		psOpts = append(psOpts, pubsub.WithDiscovery(cfg.Discovery))
	}

	var topicScoreFactory func(string) *pubsub.TopicScoreParams
	if cfg.ScoreIndex != nil {
		cfg.initScoring()
		inspector := scoreInspector(logger.Named(logging.NameScoreInspector), cfg.ScoreIndex)
		peerScoreParams := params.PeerScoreParams(cfg.Scoring.OneEpochDuration, cfg.MsgIDCacheTTL, cfg.Scoring.IPColocationWeight, 0, cfg.Scoring.IPWhilelist...)
		psOpts = append(psOpts, pubsub.WithPeerScore(peerScoreParams, params.PeerScoreThresholds()),
			pubsub.WithPeerScoreInspect(inspector, scoreInspectInterval))
		async.Interval(ctx, time.Hour, func() {
			// reset peer scores metric every hour because it has a label for peer ID which can grow infinitely
			metricPubsubPeerScoreInspect.Reset()
		})
		if cfg.GetValidatorStats == nil {
			cfg.GetValidatorStats = func() (uint64, uint64, uint64, error) {
				// default in case it was not injected
				return 100, 100, 10, nil
			}
		}
		topicScoreFactory = topicScoreParams(logger, cfg, fork)
	}

	if cfg.MsgIDHandler != nil {
		psOpts = append(psOpts, pubsub.WithMessageIdFn(cfg.MsgIDHandler.MsgID(logger)))
	}

	if len(cfg.StaticPeers) > 0 {
		psOpts = append(psOpts, pubsub.WithDirectPeers(cfg.StaticPeers))
	}

	psOpts = append(psOpts, pubsub.WithEventTracer(newTracer(logger, cfg.TraceLog)))

	ps, err := pubsub.NewGossipSub(ctx, cfg.Host, psOpts...)
	if err != nil {
		return nil, nil, err
	}

	ctrl := NewTopicsController(ctx, logger, cfg.MsgHandler, cfg.MsgValidatorFactory, sf, ps, fork, topicScoreFactory)

	return ps, ctrl, nil
}
