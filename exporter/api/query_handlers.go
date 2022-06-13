package api

import (
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"

	"github.com/bloxapp/ssv/operator/storage"
	"github.com/bloxapp/ssv/protocol/v1/message"
	qbftstorage "github.com/bloxapp/ssv/protocol/v1/qbft/storage"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/utils/format"
)

const (
	unknownError = "unknown error"
)

// HandleOperatorsQuery handles TypeOperator queries.
func HandleOperatorsQuery(logger *zap.Logger, storage registrystorage.OperatorsCollection, nm *NetworkMessage) {
	logger.Debug("handles operators request",
		zap.Uint64("from", nm.Msg.Filter.From),
		zap.Uint64("to", nm.Msg.Filter.To),
		zap.String("pk", nm.Msg.Filter.PublicKey))
	operators, err := getOperators(storage, nm.Msg.Filter)
	res := Message{
		Type:   nm.Msg.Type,
		Filter: nm.Msg.Filter,
	}
	if err != nil {
		logger.Error("could not get operators", zap.Error(err))
		res.Data = []string{"internal error - could not get operators"}
	} else {
		res.Data = operators
	}
	nm.Msg = res
}

// HandleValidatorsQuery handles TypeValidator queries.
func HandleValidatorsQuery(logger *zap.Logger, s storage.ValidatorsCollection, nm *NetworkMessage) {
	logger.Debug("handles validators request",
		zap.Uint64("from", nm.Msg.Filter.From),
		zap.Uint64("to", nm.Msg.Filter.To),
		zap.String("pk", nm.Msg.Filter.PublicKey))
	res := Message{
		Type:   nm.Msg.Type,
		Filter: nm.Msg.Filter,
	}
	validators, err := getValidators(s, nm.Msg.Filter)
	if err != nil {
		logger.Warn("failed to get validators", zap.Error(err))
		res.Data = []string{"internal error - could not get validators"}
	} else {
		res.Data = validators
	}
	nm.Msg = res
}

// HandleDecidedQuery handles TypeDecided queries.
// TODO: un-lint
//nolint
func HandleDecidedQuery(logger *zap.Logger, validatorStorage storage.ValidatorsCollection, qbftStorage qbftstorage.QBFTStore, nm *NetworkMessage) {
	logger.Debug("handles decided request",
		zap.Uint64("from", nm.Msg.Filter.From),
		zap.Uint64("to", nm.Msg.Filter.To),
		zap.String("pk", nm.Msg.Filter.PublicKey),
		zap.String("role", string(nm.Msg.Filter.Role)))
	res := Message{
		Type:   nm.Msg.Type,
		Filter: nm.Msg.Filter,
	}
	v, found, err := validatorStorage.GetValidatorInformation(nm.Msg.Filter.PublicKey)
	if err != nil {
		logger.Warn("failed to get validators", zap.Error(err))
		res.Data = []string{"internal error - could not get validator"}
	} else if !found {
		logger.Warn("validator not found")
		res.Data = []string{"internal error - could not find validator"}
	} else {
		pkRaw, err := hex.DecodeString(v.PublicKey)
		if err != nil {
			logger.Warn("failed to decode validator public key", zap.Error(err))
			res.Data = []string{"internal error - could not read validator key"}
		} else {
			identifier := format.IdentifierFormat(pkRaw, string(nm.Msg.Filter.Role))
			from := message.Height(nm.Msg.Filter.From)
			to := message.Height(nm.Msg.Filter.To)
			msgs, err := qbftStorage.GetDecided([]byte(identifier), from, to)
			if err != nil {
				logger.Warn("failed to get decided messages", zap.Error(err))
				res.Data = []string{"internal error - could not get decided messages"}
			} else {
				res.Data = msgs
			}
		}
	}
	nm.Msg = res
}

// HandleErrorQuery handles TypeError queries.
func HandleErrorQuery(logger *zap.Logger, nm *NetworkMessage) {
	logger.Warn("handles error message")
	if _, ok := nm.Msg.Data.([]string); !ok {
		nm.Msg.Data = []string{}
	}
	errs := nm.Msg.Data.([]string)
	if nm.Err != nil {
		errs = append(errs, nm.Err.Error())
	}
	if len(errs) == 0 {
		errs = append(errs, unknownError)
	}
	nm.Msg = Message{
		Type: TypeError,
		Data: errs,
	}
}

// HandleUnknownQuery handles unknown queries.
func HandleUnknownQuery(logger *zap.Logger, nm *NetworkMessage) {
	logger.Warn("unknown message type", zap.String("messageType", string(nm.Msg.Type)))
	nm.Msg = Message{
		Type: TypeError,
		Data: []string{fmt.Sprintf("bad request - unknown message type '%s'", nm.Msg.Type)},
	}
}
