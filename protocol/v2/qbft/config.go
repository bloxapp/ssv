package qbft

import (
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/bloxapp/ssv-spec/types"
	spectypes "github.com/bloxapp/ssv-spec/types"

	"github.com/bloxapp/ssv/protocol/v2/qbft/roundtimer"
	qbftstorage "github.com/bloxapp/ssv/protocol/v2/qbft/storage"
)

type signing interface {
	// GetOperatorSigner returns an OperatorSigner instance
	GetOperatorSigner() spectypes.OperatorSigner
	// GetSignatureDomainType returns the Domain type used for signatures
	GetSignatureDomainType() spectypes.DomainType
}

type IConfig interface {
	signing
	// GetValueCheckF returns value check function
	GetValueCheckF() specqbft.ProposedValueCheckF
	// GetProposerF returns func used to calculate proposer
	GetProposerF() specqbft.ProposerF
	// GetNetwork returns a p2p Network instance
	GetNetwork() specqbft.Network
	// GetStorage returns a storage instance
	GetStorage() qbftstorage.QBFTStore
	// GetTimer returns round timer
	GetTimer() roundtimer.Timer
	// GetSignatureVerifier returns the signature verifier for operator signatures
	GetSignatureVerifier() spectypes.SignatureVerifier
	// GetCutOffRound returns the round cut-off
	GetCutOffRound() int
	// VerifySignatures returns if signature is checked
	VerifySignatures() bool
}

type Config struct {
	OperatorSigner        spectypes.OperatorSigner
	SigningPK             []byte
	Domain                spectypes.DomainType
	ValueCheckF           specqbft.ProposedValueCheckF
	ProposerF             specqbft.ProposerF
	Storage               qbftstorage.QBFTStore
	Network               specqbft.Network
	Timer                 roundtimer.Timer
	SignatureVerifier     types.SignatureVerifier
	CutOffRound           int
	SignatureVerification bool
}

// GetOperatorSigner returns an OperatorSigner instance
func (c *Config) GetOperatorSigner() spectypes.OperatorSigner {
	return c.OperatorSigner
}

// GetSigningPubKey returns the public key used to sign all QBFT messages
func (c *Config) GetSigningPubKey() []byte {
	return c.SigningPK
}

// GetSignatureDomainType returns the Domain type used for signatures
func (c *Config) GetSignatureDomainType() spectypes.DomainType {
	return c.Domain
}

// GetValueCheckF returns value check instance
func (c *Config) GetValueCheckF() specqbft.ProposedValueCheckF {
	return c.ValueCheckF
}

// GetProposerF returns func used to calculate proposer
func (c *Config) GetProposerF() specqbft.ProposerF {
	return c.ProposerF
}

// GetNetwork returns a p2p Network instance
func (c *Config) GetNetwork() specqbft.Network {
	return c.Network
}

// GetStorage returns a storage instance
func (c *Config) GetStorage() qbftstorage.QBFTStore {
	return c.Storage
}

// GetTimer returns round timer
func (c *Config) GetTimer() roundtimer.Timer {
	return c.Timer
}

func (c *Config) GetSignatureVerifier() spectypes.SignatureVerifier {
	return c.SignatureVerifier
}

func (c *Config) GetCutOffRound() int {
	return c.CutOffRound
}

func (c *Config) VerifySignatures() bool {
	return c.SignatureVerification
}
