package abiparser

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math/big"
)

// Event names
const (
	OperatorAdded     = "OperatorAdded"
	ValidatorAdded    = "ValidatorAdded"
	ValidatorUpdated  = "ValidatorUpdated"
	AccountLiquidated = "AccountLiquidated"
)

// ValidatorAddedEvent struct represents event received by the smart contract
type ValidatorAddedEvent struct {
	PublicKey          []byte
	OwnerAddress       common.Address
	OperatorPublicKeys [][]byte
	OperatorIds        []*big.Int
	SharesPublicKeys   [][]byte
	EncryptedKeys      [][]byte
}

// ValidatorUpdatedEvent struct represents event received by the smart contract
type ValidatorUpdatedEvent struct {
	PublicKey          []byte
	OwnerAddress       common.Address
	OperatorPublicKeys [][]byte
	OperatorIds        []*big.Int
	SharesPublicKeys   [][]byte
	EncryptedKeys      [][]byte
}

// AccountLiquidatedEvent struct represents event received by the smart contract
type AccountLiquidatedEvent struct {
	OwnerAddress common.Address
}

// OperatorAddedEvent struct represents event received by the smart contract
type OperatorAddedEvent struct {
	Id           *big.Int //nolint
	Name         string
	OwnerAddress common.Address
	PublicKey    []byte
}

// AbiV2 parsing events from v2 abi contract
type AbiV2 struct {
}

type UnpackError struct {
	Err error
}

func (e *UnpackError) Error() string {
	return e.Err.Error()
}

// ParseOperatorAddedEvent parses an OperatorAddedEvent
func (v2 *AbiV2) ParseOperatorAddedEvent(
	logger *zap.Logger,
	data []byte,
	topics []common.Hash,
	contractAbi abi.ABI,
) (*OperatorAddedEvent, error) {
	var operatorAddedEvent OperatorAddedEvent
	err := contractAbi.UnpackIntoInterface(&operatorAddedEvent, OperatorAdded, data)
	if err != nil {
		return nil, &UnpackError{
			Err: errors.Wrap(err, "failed to unpack OperatorAdded event"),
		}
	}
	outAbi, err := getOutAbi()
	if err != nil {
		return nil, err
	}
	pubKey, err := readOperatorPubKey(operatorAddedEvent.PublicKey, outAbi)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read OperatorPublicKey")
	}
	operatorAddedEvent.PublicKey = []byte(pubKey)

	if len(topics) > 1 {
		operatorAddedEvent.OwnerAddress = common.HexToAddress(topics[1].Hex())
	} else {
		logger.Error("operator event missing topics. no owner address provided.")
	}
	return &operatorAddedEvent, nil
}

// ParseValidatorAddedEvent parses ValidatorAddedEvent
func (v2 *AbiV2) ParseValidatorAddedEvent(
	logger *zap.Logger,
	data []byte,
	contractAbi abi.ABI,
) (event *ValidatorAddedEvent, error error) {
	return v2.parseValidatorEvent(logger, data, ValidatorAdded, contractAbi)
}

// ParseValidatorUpdatedEvent parses ValidatorUpdatedEvent
func (v2 *AbiV2) ParseValidatorUpdatedEvent(
	logger *zap.Logger,
	data []byte,
	contractAbi abi.ABI,
) (*ValidatorAddedEvent, error) {
	return v2.parseValidatorEvent(logger, data, ValidatorUpdated, contractAbi)
}

func (v2 *AbiV2) ParseAccountLiquidatedEvent(logger *zap.Logger, data []byte, contractAbi abi.ABI) (*AccountLiquidatedEvent, error) {
	var accountLiquidatedEvent AccountLiquidatedEvent
	err := contractAbi.UnpackIntoInterface(&accountLiquidatedEvent, AccountLiquidated, data)
	if err != nil {
		return nil, &UnpackError{
			Err: errors.Wrap(err, "failed to unpack AccountLiquidated event"),
		}
	}

	return &accountLiquidatedEvent, nil
}

func (v2 *AbiV2) parseValidatorEvent(logger *zap.Logger, data []byte, eventName string, contractAbi abi.ABI) (*ValidatorAddedEvent, error) {
	var validatorAddedEvent ValidatorAddedEvent
	err := contractAbi.UnpackIntoInterface(&validatorAddedEvent, eventName, data)
	if err != nil {
		return nil, &UnpackError{
			Err: errors.Wrapf(err, "Failed to unpack %s event", eventName),
		}
	}

	outAbi, err := getOutAbi()
	if err != nil {
		return nil, errors.Wrap(err, "failed to define ABI")
	}

	for i, ek := range validatorAddedEvent.EncryptedKeys {
		out, err := outAbi.Unpack("method", ek)
		if err != nil {
			return nil, &UnpackError{
				Err: errors.Wrap(err, "failed to unpack EncryptedKey"),
			}
		}
		if encryptedSharePrivateKey, ok := out[0].(string); ok {
			validatorAddedEvent.EncryptedKeys[i] = []byte(encryptedSharePrivateKey)
		}
	}

	return &validatorAddedEvent, nil
}
