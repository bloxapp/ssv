package eventsyncer

import (
	"go.uber.org/zap"
)

// Option defines EventSyncer configuration option.
type Option func(*EventSyncer)

// WithLogger enables logging.
func WithLogger(logger *zap.Logger) Option {
	return func(es *EventSyncer) {
		es.logger = logger.Named("EventSyncer")
	}
}

// WithMetrics enables reporting metrics.
func WithMetrics(metrics metrics) Option {
	return func(es *EventSyncer) {
		es.metrics = metrics
	}
}
