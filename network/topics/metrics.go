package topics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// TODO: replace with new metrics
var (
	metricPubsubTrace = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ssv:network:pubsub:trace",
		Help: "Traces of pubsub messages",
	}, []string{"type"})
	metricPubsubOutbound = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ssv:p2p:pubsub:msg:out",
		Help: "Count broadcasted messages",
	}, []string{"topic"})
	metricPubsubInbound = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ssv:p2p:pubsub:msg:in",
		Help: "Count incoming messages",
	}, []string{"topic", "msg_type"})
	metricPubsubPeerScoreInspect = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:score:inspect",
		Help: "Gauge for negative peer scores",
	}, []string{"pid"})
	metricPubsubFullMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:full_messages",
		Help: "Count FullMessages",
	}, []string{})
	metricPubsubControlMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:control_messages",
		Help: "Count ControlMessages",
	}, []string{})
	metricPubsubIHaveMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:ihave",
		Help: "Count of incoming IHAVE messages",
	}, []string{})
	metricPubsubIWantMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:iwant",
		Help: "Count of incoming IWANT messages",
	}, []string{})
	metricPubsubGraftMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:graft",
		Help: "Count of incoming GRAFT messages",
	}, []string{})
	metricPubsubPruneMsgs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:p2p:pubsub:msg:prune",
		Help: "Count of incoming PRUNE messages",
	}, []string{})
)

func init() {
	logger := zap.L()

	allMetrics := []prometheus.Collector{
		metricPubsubTrace,
		metricPubsubOutbound,
		metricPubsubInbound,
		metricPubsubPeerScoreInspect,
		metricPubsubFullMsgs,
		metricPubsubControlMsgs,
		metricPubsubIHaveMsgs,
		metricPubsubIWantMsgs,
		metricPubsubGraftMsgs,
		metricPubsubPruneMsgs,
	}

	for i, c := range allMetrics {
		if err := prometheus.Register(c); err != nil {
			// TODO: think how to print metric name
			logger.Debug("could not register prometheus collector",
				zap.Int("index", i),
				zap.Error(err),
			)
		}
	}
}
