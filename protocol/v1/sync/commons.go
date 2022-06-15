package sync

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v1/message"
	p2pprotocol "github.com/bloxapp/ssv/protocol/v1/p2p"
)

// GetHighest returns the highest message from the given collection
func GetHighest(logger *zap.Logger, remoteMsgs ...p2pprotocol.SyncResult) (highest *message.SignedMessage, sender string) {
	var height message.Height

	for _, remoteMsg := range remoteMsgs {
		sm, err := ExtractSyncMsg(remoteMsg.Msg)
		if err != nil {
			logger.Warn("bad sync message", zap.Error(err))
			continue
		}
		if sm == nil {
			continue
		}
		if len(sm.Data) == 0 {
			continue
		}
		signedMsg := sm.Data[0]
		if signedMsg != nil && signedMsg.Message != nil && signedMsg.Message.Height > height {
			highest = signedMsg
			height = highest.Message.Height
			sender = remoteMsg.Sender
		}
	}
	return
}

// ExtractSyncMsg extracts message.SyncMessage from message.SSVMessage
func ExtractSyncMsg(msg *message.SSVMessage) (*message.SyncMessage, error) {
	sm := &message.SyncMessage{}
	err := sm.Decode(msg.Data)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode sync message")
	}
	if sm.Status == message.StatusNotFound {
		return nil, nil
	}
	if sm.Status != message.StatusSuccess {
		return nil, errors.Errorf("failed to get sync message: %s", sm.Status.String())
	}
	return sm, nil
}
