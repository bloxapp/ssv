package msgqueue

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
)

var (
	metricsMsgQSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:ibft:msgq:size",
		Help: "The amount of messages in queue",
	}, []string{"lambda"})
)

func init() {
	if err := prometheus.Register(metricsMsgQSize); err != nil {
		log.Println("could not register prometheus collector")
	}
}
