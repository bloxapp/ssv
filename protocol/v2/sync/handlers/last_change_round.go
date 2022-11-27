package handlers

import (
	"fmt"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/message"
	protocolp2p "github.com/bloxapp/ssv/protocol/v2/p2p"
	qbftstorage "github.com/bloxapp/ssv/protocol/v2/qbft/storage"
)

// LastChangeRoundHandler handler for last-decided protocol
// TODO: add msg validation and report scores
func LastChangeRoundHandler(plogger *zap.Logger, storeMap *qbftstorage.QBFTSyncMap, reporting protocolp2p.ValidationReporting) protocolp2p.RequestHandler {
	//plogger = plogger.With(zap.String("who", "last decided handler"))
	return func(msg *spectypes.SSVMessage) (*spectypes.SSVMessage, error) {
		logger := plogger.With(zap.String("msg_id_hex", fmt.Sprintf("%x", msg.MsgID)))
		sm := &message.SyncMessage{}
		err := sm.Decode(msg.Data)
		if err != nil {
			reporting.ReportValidation(msg, protocolp2p.ValidationRejectLow)
			sm.Status = message.StatusBadRequest
		} else if sm.Protocol != message.LastChangeRoundType {
			// not this protocol
			// TODO: remove after v0
			return nil, nil
		} else {
			store := storeMap.Get(msg.MsgID.GetRoleType())
			if store == nil {
				return nil, errors.New(fmt.Sprintf("not storage found for type %s", msg.MsgID.GetRoleType().String()))
			}
			res, err := store.GetLastChangeRoundMsg(msg.MsgID[:])
			if err != nil {
				logger.Warn("change round sync msg error", zap.Error(err))
			}
			logger.Debug("last change round handler", zap.Any("msgs", res), zap.Error(err))
			sm.UpdateResults(err, res...)
		}

		data, err := sm.Encode()
		if err != nil {
			return nil, errors.Wrap(err, "could not encode result data")
		}
		msg.Data = data

		return msg, nil
	}
}
