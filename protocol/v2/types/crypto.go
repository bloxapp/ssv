package types

import (
	"encoding/hex"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	specssv "github.com/ssvlabs/ssv-spec/ssv"
	spectypes "github.com/ssvlabs/ssv-spec/types"
	"go.uber.org/zap"
)

var (
	MetricsSignaturesVerifications = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ssv_signature_verifications_genesis",
		Help: "Number of signatures verifications (genesis",
	}, []string{})
)

func init() {
	logger := zap.L()
	if err := prometheus.Register(MetricsSignaturesVerifications); err != nil {
		logger.Debug("could not register prometheus collector")
	}
}

func ReconstructSignature(ps *specssv.PartialSigContainer, root [32]byte, validatorPubKey []byte, validatorIndex phase0.ValidatorIndex) ([]byte, error) {
	// Reconstruct signatures
	signature, err := spectypes.ReconstructSignatures(ps.Signatures[validatorIndex][rootHex(root)])
	if err != nil {
		return nil, errors.Wrap(err, "failed to reconstruct signatures")
	}
	if err := VerifyReconstructedSignature(signature, validatorPubKey, root); err != nil {
		return nil, errors.Wrap(err, "failed to verify reconstruct signature")
	}
	return signature.Serialize(), nil
}

func VerifyReconstructedSignature(sig *bls.Sign, validatorPubKey []byte, root [32]byte) error {
	MetricsSignaturesVerifications.WithLabelValues().Inc()

	pk, err := DeserializeBLSPublicKey(validatorPubKey)
	if err != nil {
		return errors.Wrap(err, "could not deserialize validator pk")
	}

	if res := sig.VerifyByte(&pk, root[:]); !res {
		return errors.New("could not reconstruct a valid signature")
	}
	return nil
}

func rootHex(r [32]byte) string {
	return hex.EncodeToString(r[:])
}
