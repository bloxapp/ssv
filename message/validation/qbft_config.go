package validation

import (
	specqbft "github.com/ssvlabs/ssv-spec/qbft"
	spectypes "github.com/ssvlabs/ssv-spec/types"

	"github.com/ssvlabs/ssv/protocol/v2/qbft/roundtimer"
	qbftstorage "github.com/ssvlabs/ssv/protocol/v2/qbft/storage"
)

// qbftConfig is used in message validation and has no signature verification.
type qbftConfig struct {
	domain spectypes.DomainType
}

func newQBFTConfig(domain spectypes.DomainType) qbftConfig {
	return qbftConfig{
		domain: domain,
	}
}

func (q qbftConfig) GetShareSigner() spectypes.ShareSigner {
	panic("should not be called")
}

func (q qbftConfig) GetOperatorSigner() spectypes.OperatorSigner {
	panic("should not be called")
}

func (q qbftConfig) GetSignatureDomainType() spectypes.DomainType {
	return q.domain
}

func (q qbftConfig) GetValueCheckF() specqbft.ProposedValueCheckF {
	panic("should not be called")
}

func (q qbftConfig) GetProposerF() specqbft.ProposerF {
	panic("should not be called")
}

func (q qbftConfig) GetNetwork() specqbft.Network {
	panic("should not be called")
}

func (q qbftConfig) GetStorage() qbftstorage.QBFTStore {
	panic("should not be called")
}

func (q qbftConfig) GetTimer() roundtimer.Timer {
	panic("should not be called")
}

func (q qbftConfig) VerifySignatures() bool {
	return false
}
