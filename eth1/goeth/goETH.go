package goeth

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/eth1"
	"github.com/bloxapp/ssv/shared/params"
	"github.com/bloxapp/ssv/storage/collections"
	"github.com/bloxapp/ssv/utils/rsaencryption"
)

type eth1GRPC struct {
	ctx             context.Context
	conn            *ethclient.Client
	logger          *zap.Logger
	contractEvent   *eth1.ContractEvent
	operatorStorage collections.IOperatorStorage
}

// New create new goEth instance
func New(ctx context.Context, logger *zap.Logger, nodeAddr string, operatorStorage collections.IOperatorStorage) (eth1.Eth1, error) {
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial(nodeAddr)
	if err != nil {
		logger.Error("Failed to connect to the Ethereum client", zap.Error(err))
	}

	e := &eth1GRPC{
		ctx:             ctx,
		conn:            conn,
		logger:          logger,
		operatorStorage: operatorStorage,
	}

	// init the instance which publishes an event when anything happens
	err = e.streamSmartContractEvents(params.SsvConfig().OperatorContractAddress)
	if err != nil {
		logger.Error("Failed to init operator contract address subject", zap.Error(err))
	}

	return e, nil
}

// Sync tries to fetch history of contract events
func (e *eth1GRPC) Sync() error {
	err := e.syncSmartContractsEvents(params.SsvConfig().OperatorContractAddress, params.SsvConfig().ContractABI)
	if err != nil {
		e.logger.Error("Failed to sync contract events", zap.Error(err))
	}
	return err
}

// syncSmartContractsEvents sync events history of the given contract
func (e *eth1GRPC) syncSmartContractsEvents(contractAddr, contractABI string) error {
	e.logger.Debug("syncing smart contract events")

	contractAbi, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return errors.Wrap(err, "failed to parse ABI interface")
	}

	contractAddress := common.HexToAddress(contractAddr)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: big.NewInt(800000),
	}
	logs, err := e.conn.FilterLogs(e.ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to get event logs")
	}
	e.logger.Debug(fmt.Sprintf("manage to get event logs, number of results: %d", len(logs)))

	for _, vLog := range logs {
		eventType, err := contractAbi.EventByID(vLog.Topics[0])
		if err != nil {
			e.logger.Error("Failed to find event type", zap.Error(err))
			continue
		}
		switch eventName := eventType.Name; eventName {
		case "OperatorAdded":
			err := e.ProcessOperatorAddedEvent(vLog.Data, contractAbi, eventName)
			if err != nil {
				e.logger.Error("Failed to process OperatorAdded event", zap.Error(err))
				continue
			}
		case "ValidatorAdded":
			err := e.ProcessValidatorAddedEvent(vLog.Data, contractAbi, eventName)
			if err != nil {
				e.logger.Error("Failed to process ValidatorAdded event", zap.Error(err))
				continue
			}
		default:
			e.logger.Debug("Unknown contract event is received")
		}
	}
	return nil
}

// streamSmartContractEvents listen to events from the given contract
func (e *eth1GRPC) streamSmartContractEvents(contractAddr string) error {
	contractAddress := common.HexToAddress(contractAddr)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs := make(chan types.Log)
	sub, err := e.conn.SubscribeFilterLogs(e.ctx, query, logs)
	if err != nil {
		e.logger.Fatal("Failed to subscribe to logs", zap.Error(err))
		return err
	}

	contractAbi, err := abi.JSON(strings.NewReader(params.SsvConfig().ContractABI))
	if err != nil {
		e.logger.Fatal("Failed to parse ABI interface", zap.Error(err))
	}

	e.contractEvent = eth1.NewContractEvent("smartContractEvent")
	go func() {
		for {
			select {
			case err := <-sub.Err():
				// TODO might fail consider reconnect
				e.logger.Error("Error from logs sub", zap.Error(err))

			case vLog := <-logs:
				eventType, err := contractAbi.EventByID(vLog.Topics[0])
				if err != nil {
					e.logger.Error("Failed to get event by topic hash", zap.Error(err))
					continue
				}

				switch eventName := eventType.Name; eventName {
				case "OperatorAdded":
					operatorAddedEvent := eth1.OperatorAddedEvent{}
					err = contractAbi.UnpackIntoInterface(&operatorAddedEvent, eventType.Name, vLog.Data)
					if err != nil {
						e.logger.Error("Failed to unpack event", zap.Error(err))
						continue
					}
					e.contractEvent.Data = operatorAddedEvent

				case "ValidatorAdded":
					err := e.ProcessValidatorAddedEvent(vLog.Data, contractAbi, eventName)
					if err != nil {
						e.logger.Error("Failed to process ValidatorAdded event", zap.Error(err))
					}

				default:
					e.logger.Debug("Unknown contract event is received")
				}
			}
		}
	}()
	return nil
}

func (e *eth1GRPC) GetContractEvent() *eth1.ContractEvent {
	return e.contractEvent
}

func (e *eth1GRPC) ProcessValidatorAddedEvent(data []byte, contractAbi abi.ABI, eventName string) error {
	validatorAddedEvent := eth1.ValidatorAddedEvent{}
	err := contractAbi.UnpackIntoInterface(&validatorAddedEvent, eventName, data)
	if err != nil {
		return errors.Wrap(err, "Failed to unpack ValidatorAdded event")
	}

	isEventBelongsToOperator := false

	e.logger.Debug("ValidatorAdded Event",
		zap.String("Validator PublicKey", hex.EncodeToString(validatorAddedEvent.PublicKey)),
		zap.String("Owner Address", validatorAddedEvent.OwnerAddress.String()))
	for i := range validatorAddedEvent.OessList {
		validatorShare := &validatorAddedEvent.OessList[i]

		def := `[{ "name" : "method", "type": "function", "outputs": [{"type": "string"}]}]` //TODO need to set as var?
		outAbi, err := abi.JSON(strings.NewReader(def))
		if err != nil {
			return errors.Wrap(err, "failed to define ABI")
		}

		outOperatorPublicKey, err := outAbi.Unpack("method", validatorShare.OperatorPublicKey)
		if err != nil {
			return errors.Wrap(err, "failed to unpack OperatorPublicKey")
		}

		if operatorPublicKey, ok := outOperatorPublicKey[0].(string); ok {
			validatorShare.OperatorPublicKey = []byte(operatorPublicKey) // set for further use in code
			if strings.EqualFold(operatorPublicKey, params.SsvConfig().OperatorPublicKey) {
				sk, err := e.operatorStorage.GetPrivateKey()
				if err != nil {
					return errors.Wrap(err, "failed to get private key")
				}

				out, err := outAbi.Unpack("method", validatorShare.EncryptedKey)
				if err != nil {
					return errors.Wrap(err, "failed to unpack EncryptedKey")
				}

				if encryptedSharePrivateKey, ok := out[0].(string); ok {
					decryptedSharePrivateKey, err := rsaencryption.DecodeKey(sk, encryptedSharePrivateKey)
					decryptedSharePrivateKey = strings.Replace(decryptedSharePrivateKey, "0x", "", 1)
					if err != nil {
						return errors.Wrap(err, "failed to decrypt share private key")
					}
					validatorShare.EncryptedKey = []byte(decryptedSharePrivateKey)
					isEventBelongsToOperator = true
				}
			}
		}
	}

	if isEventBelongsToOperator {
		e.contractEvent.Data = validatorAddedEvent
		e.contractEvent.NotifyAll()
	} else {
		e.logger.Debug("ValidatorAdded Event doesn't belong to operator")
	}

	return nil
}

func (e *eth1GRPC) ProcessOperatorAddedEvent(data []byte, contractAbi abi.ABI, eventName string) error {
	operatorAddedEvent := eth1.OperatorAddedEvent{}
	err := contractAbi.UnpackIntoInterface(&operatorAddedEvent, eventName, data)
	if err != nil {
		return errors.Wrap(err, "Failed to unpack OperatorAdded event")
	}
	e.contractEvent.Data = operatorAddedEvent

	operatorPubkeyHex := hex.EncodeToString(operatorAddedEvent.Pubkey)
	e.logger.Debug("OperatorAdded Event",
		zap.String("Operator PublicKey", operatorPubkeyHex),
		zap.String("Payment Address", operatorAddedEvent.PaymentAddress.String()))

	if strings.EqualFold(operatorPubkeyHex, params.SsvConfig().OperatorPublicKey) {
		e.contractEvent.NotifyAll()
	}
	return nil
}