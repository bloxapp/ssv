package api

import (
	"encoding/hex"
	"fmt"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/storage"
	"github.com/bloxapp/ssv/protocol/v2/message"
)

const (
	unknownError = "unknown error"
)

// HandleDecidedQuery handles TypeDecided queries.
func HandleDecidedQuery(logger *zap.Logger, qbftStorage *storage.QBFTStores, nm *NetworkMessage) {
	logger.Debug("handles decided request",
		zap.Uint64("from", nm.Msg.Filter.From),
		zap.Uint64("to", nm.Msg.Filter.To),
		zap.String("pk", nm.Msg.Filter.PublicKey),
		zap.String("role", string(nm.Msg.Filter.Role)))
	res := Message{
		Type:   nm.Msg.Type,
		Filter: nm.Msg.Filter,
	}

	pkRaw, err := hex.DecodeString(nm.Msg.Filter.PublicKey)
	if err != nil {
		logger.Warn("failed to decode validator public key", zap.Error(err))
		res.Data = []string{"internal error - could not read validator key"}
		nm.Msg = res
		return
	}

	beaconRole, err := message.BeaconRoleFromString(string(nm.Msg.Filter.Role))
	if err != nil {
		logger.Warn("failed to parse role", zap.Error(err))
		res.Data = []string{"malformed role"}
		nm.Msg = res
		return
	}

	msgID := spectypes.NewMsgID(pkRaw, beaconRole)
	from := specqbft.Height(nm.Msg.Filter.From)
	to := specqbft.Height(nm.Msg.Filter.To)
	instances, err := qbftStorage.Get(beaconRole).GetInstancesInRange(msgID[:], from, to)
	if err != nil {
		logger.Warn("failed to get instances", zap.Error(err))
		res.Data = []string{"internal error - could not get decided messages"}
	} else {
		msgs := make([]*specqbft.SignedMessage, 0, len(instances))
		for _, instance := range instances {
			msgs = append(msgs, instance.DecidedMessage)
		}
		data, err := DecidedAPIData(msgs...)
		if err != nil {
			res.Data = []string{}
		} else {
			res.Data = data
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
