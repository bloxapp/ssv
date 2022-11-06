package validator

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	beaconprotocol "github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
)

var (
	allMetrics = []prometheus.Collector{
		metricsCurrentSlot,
		metricsValidatorStatus,
		metricsTimeAttestationData,
		metricsTimeConsensus,
	}
	metricsCurrentSlot = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:validator:ibft_current_slot1",
		Help: "Current running slot",
	}, []string{"pubKey"})
	metricsValidatorStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:validator:status1",
		Help: "Validator status",
	}, []string{"pubKey"})
	metricsTimeAttestationData = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:validator:time_attestation_data",
		Help: "Validator attestation data time (seconds)",
	}, []string{"pubKey"})
	metricsTimeConsensus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ssv:validator:time_consensus",
		Help: "Validator consensus time (seconds)",
	}, []string{"pubKey"})
)

func init() {
	for _, c := range allMetrics {
		if err := prometheus.Register(c); err != nil {
			log.Println("could not register prometheus collector")
		}
	}
}

// ReportValidatorStatus reports the current status of validator
func ReportValidatorStatus(pk string, meta *beaconprotocol.ValidatorMetadata, logger *zap.Logger) {
	logger = logger.With(zap.String("pubKey", pk), zap.String("who", "ReportValidatorStatus"),
		zap.Any("metadata", meta))
	if meta == nil {
		logger.Debug("validator metadata not found")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusNotFound))
	} else if meta.IsActive() {
		logger.Debug("validator is ready")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusReady))
	} else if meta.Slashed() {
		logger.Debug("validator slashed")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusSlashed))
	} else if meta.Exiting() {
		logger.Debug("validator exiting / exited")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusExiting))
	} else if !meta.Activated() {
		logger.Debug("validator not activated")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusNotActivated))
	} else if meta.Pending() {
		logger.Debug("validator pending")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusPending))
	} else if meta.Index == 0 {
		logger.Debug("validator index not found")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusNoIndex))
	} else {
		logger.Debug("validator is unknown")
		metricsValidatorStatus.WithLabelValues(pk).Set(float64(validatorStatusUnknown))
	}
}

type validatorStatus int32

var (
	//validatorStatusInactive     validatorStatus = 0
	validatorStatusNoIndex validatorStatus = 1
	//validatorStatusError        validatorStatus = 2
	validatorStatusReady        validatorStatus = 3
	validatorStatusNotActivated validatorStatus = 4
	validatorStatusExiting      validatorStatus = 5
	validatorStatusSlashed      validatorStatus = 6
	validatorStatusNotFound     validatorStatus = 7
	validatorStatusPending      validatorStatus = 8
	validatorStatusUnknown      validatorStatus = 9
)
