package topics

import (
	"context"
	"io"
	"strconv"
	"time"

	spectypes "github.com/bloxapp/ssv-spec/types"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/network/forks"
)

//go:generate mockgen -package=mocks -destination=./mocks/controller.go -source=./controller.go

var (
	// ErrTopicNotReady happens when trying to access a topic which is not ready yet
	ErrTopicNotReady = errors.New("topic is not ready")
)

// Controller is an interface for managing pubsub topics
type Controller interface {
	// Subscribe subscribes to the given topic
	Subscribe(logger *zap.Logger, name string) error
	// Unsubscribe unsubscribes from the given topic
	Unsubscribe(logger *zap.Logger, topicName string, hard bool) error
	// Peers returns the peers subscribed to the given topic
	Peers(topicName string) ([]peer.ID, error)
	// Topics lists all the available topics
	Topics() []string
	// Broadcast publishes the message on the given topic
	Broadcast(topicName string, data []byte, timeout time.Duration) error

	io.Closer
}

// PubsubMessageHandler handles incoming messages
type PubsubMessageHandler func(string, *pubsub.Message) error

// topicsCtrl implements Controller
type topicsCtrl struct {
	ctx    context.Context
	logger *zap.Logger // struct logger to implement i.Closer
	ps     *pubsub.PubSub
	// scoreParamsFactory is a function that helps to set scoring params on topics
	scoreParamsFactory  func(string) *pubsub.TopicScoreParams
	msgValidatorFactory func(string) MsgValidatorFunc
	msgHandler          PubsubMessageHandler
	subFilter           SubFilter

	fork forks.Fork

	container *topicsContainer
}

// NewTopicsController creates an instance of Controller
func NewTopicsController(ctx context.Context, logger *zap.Logger, msgHandler PubsubMessageHandler,
	msgValidatorFactory func(string) MsgValidatorFunc, subFilter SubFilter, pubSub *pubsub.PubSub,
	fork forks.Fork, scoreParams func(string) *pubsub.TopicScoreParams) Controller {
	ctrl := &topicsCtrl{
		ctx:                 ctx,
		logger:              logger,
		ps:                  pubSub,
		scoreParamsFactory:  scoreParams,
		msgValidatorFactory: msgValidatorFactory,
		msgHandler:          msgHandler,

		subFilter: subFilter,

		fork: fork,
	}

	ctrl.container = newTopicsContainer(pubSub, ctrl.onNewTopic(logger))

	return ctrl
}

func (ctrl *topicsCtrl) onNewTopic(logger *zap.Logger) onTopicJoined {
	return func(ps *pubsub.PubSub, topic *pubsub.Topic) {
		// initial setup for the topic, should happen only once
		name := topic.String()
		if err := ctrl.setupTopicValidator(topic.String()); err != nil {
			// TODO: close topic?
			// return err
			logger.Warn("could not setup topic", zap.String("topic", name), zap.Error(err))
		}
		if ctrl.scoreParamsFactory != nil {
			if p := ctrl.scoreParamsFactory(name); p != nil {
				logger.Debug("using scoring params for topic", zap.String("topic", name), zap.Any("params", p))
				if err := topic.SetScoreParams(p); err != nil {
					// logger.Warn("could not set topic score params", zap.String("topic", name), zap.Error(err))
					logger.Warn("could not set topic score params", zap.String("topic", name), zap.Error(err))
				}
			}
		}
	}
}

// Close implements io.Closer
func (ctrl *topicsCtrl) Close() error {
	topics := ctrl.ps.GetTopics()
	for _, tp := range topics {
		_ = ctrl.Unsubscribe(ctrl.logger, ctrl.fork.GetTopicBaseName(tp), true)
		_ = ctrl.container.Leave(tp)
	}
	return nil
}

// Peers returns the peers subscribed to the given topic
func (ctrl *topicsCtrl) Peers(name string) ([]peer.ID, error) {
	name = ctrl.fork.GetTopicFullName(name)
	topic := ctrl.container.Get(name)
	if topic == nil {
		return nil, nil
	}
	return topic.ListPeers(), nil
}

// Topics lists all the available topics
func (ctrl *topicsCtrl) Topics() []string {
	topics := ctrl.ps.GetTopics()
	for i, tp := range topics {
		topics[i] = ctrl.fork.GetTopicBaseName(tp)
	}
	return topics
}

// Subscribe subscribes to the given topic, it can handle multiple concurrent calls.
// it will create a single goroutine and channel for every topic
func (ctrl *topicsCtrl) Subscribe(logger *zap.Logger, name string) error {
	name = ctrl.fork.GetTopicFullName(name)
	ctrl.subFilter.(Whitelist).Register(name)
	sub, err := ctrl.container.Subscribe(name)
	defer logger.Debug("subscribing to topic", zap.String("topic", name), zap.Bool("already_subscribed", sub == nil), zap.Error(err))
	if err != nil {
		return err
	}
	if sub == nil { // already subscribed
		return nil
	}
	go ctrl.start(logger, name, sub)

	return nil
}

// Broadcast publishes the message on the given topic
func (ctrl *topicsCtrl) Broadcast(name string, data []byte, timeout time.Duration) error {
	name = ctrl.fork.GetTopicFullName(name)

	topic, err := ctrl.container.Join(name)
	if err != nil {
		return err
	}

	go func() {
		ctx, done := context.WithTimeout(ctrl.ctx, timeout)
		defer done()

		err := topic.Publish(ctx, data)
		if err == nil {
			metricPubsubOutbound.WithLabelValues(name).Inc()
		}
	}()

	return err
}

// Unsubscribe unsubscribes from the given topic, only if there are no other subscribers of the given topic
// if hard is true, we will unsubscribe the topic even if there are more subscribers.
func (ctrl *topicsCtrl) Unsubscribe(logger *zap.Logger, name string, hard bool) error {
	ctrl.container.Unsubscribe(name)

	if ctrl.msgValidatorFactory != nil {
		err := ctrl.ps.UnregisterTopicValidator(name)
		if err != nil {
			logger.Debug("could not unregister msg validator", zap.String("topic", name), zap.Error(err))
		}
	}
	ctrl.subFilter.(Whitelist).Deregister(name)

	return nil
}

// start will listen to *pubsub.Subscription,
// if some error happened we try to leave and rejoin the topic
// the loop stops once a topic is unsubscribed and therefore not listed
func (ctrl *topicsCtrl) start(logger *zap.Logger, name string, sub *pubsub.Subscription) {
	for ctrl.ctx.Err() == nil {
		err := ctrl.listen(logger, sub)
		if err == nil {
			return
		}
		// rejoin in case failed
		logger.Debug("could not listen to topic", zap.String("topic", name), zap.Error(err))
		ctrl.container.Unsubscribe(name)
		_ = ctrl.container.Leave(name)
		sub, err = ctrl.container.Subscribe(name)
		if err == nil {
			continue
		}
		logger.Debug("could not rejoin topic", zap.String("topic", name), zap.Error(err))
	}
}

// listen handles incoming messages from the topic
func (ctrl *topicsCtrl) listen(logger *zap.Logger, sub *pubsub.Subscription) error {
	ctx, cancel := context.WithCancel(ctrl.ctx)
	defer cancel()
	topicName := sub.Topic()
	logger = logger.With(zap.String("topic", topicName))
	logger.Debug("start listening to topic")
	for ctx.Err() == nil {
		msg, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Debug("stop listening to topic: context is done")
				return nil
			} else if errors.Is(err, pubsub.ErrSubscriptionCancelled) || errors.Is(err, pubsub.ErrTopicClosed) {
				logger.Debug("stop listening to topic", zap.Error(err))
				return nil
			}
			logger.Warn("could not read message from subscription", zap.Error(err))
			continue
		}
		if msg == nil || msg.Data == nil {
			logger.Warn("got empty message from subscription")
			continue
		}

		if ssvMsg, ok := msg.ValidatorData.(spectypes.SSVMessage); ok {
			metricPubsubInbound.WithLabelValues(
				ctrl.fork.GetTopicBaseName(topicName),
				strconv.FormatUint(uint64(ssvMsg.MsgType), 10),
			).Inc()
		}

		if err := ctrl.msgHandler(topicName, msg); err != nil {
			logger.Debug("could not handle msg", zap.Error(err))
		}
	}
	return nil
}

// setupTopicValidator registers the topic validator
func (ctrl *topicsCtrl) setupTopicValidator(name string) error {
	if ctrl.msgValidatorFactory != nil {
		// first try to unregister in case there is already a msg validator for that topic (e.g. fork scenario)
		_ = ctrl.ps.UnregisterTopicValidator(name)

		var opts []pubsub.ValidatorOpt
		// Optional: set a timeout for message validation
		// opts = append(opts, pubsub.WithValidatorTimeout(time.Second))

		err := ctrl.ps.RegisterTopicValidator(name, ctrl.msgValidatorFactory(name), opts...)
		if err != nil {
			return errors.Wrap(err, "could not register topic validator")
		}
	}
	return nil
}
