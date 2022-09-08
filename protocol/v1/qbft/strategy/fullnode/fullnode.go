package fullnode

import (
	"context"
	"encoding/hex"

	spectypes "github.com/bloxapp/ssv-spec/types"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v1/message"
	p2pprotocol "github.com/bloxapp/ssv/protocol/v1/p2p"
	qbftstorage "github.com/bloxapp/ssv/protocol/v1/qbft/storage"
	"github.com/bloxapp/ssv/protocol/v1/qbft/strategy"
	"github.com/bloxapp/ssv/protocol/v1/sync/history"
	"github.com/bloxapp/ssv/protocol/v1/sync/lastdecided"
)

type fullNode struct {
	logger         *zap.Logger
	store          qbftstorage.DecidedMsgStore
	decidedFetcher lastdecided.Fetcher
	historySyncer  history.Syncer
}

// NewFullNodeStrategy creates a new instance of fullNode strategy
func NewFullNodeStrategy(logger *zap.Logger, store qbftstorage.DecidedMsgStore, syncer p2pprotocol.Syncer) strategy.Decided {
	return &fullNode{
		logger:         logger.With(zap.String("who", "FullNodeStrategy")),
		store:          store,
		decidedFetcher: lastdecided.NewLastDecidedFetcher(logger, syncer),
		historySyncer:  history.NewSyncer(logger, syncer),
	}
}

func (f *fullNode) Sync(ctx context.Context, identifier []byte, from, to *specqbft.SignedMessage) ([]*specqbft.SignedMessage, error) {
	logger := f.logger.With(zap.String("identifier", hex.EncodeToString(identifier)))
	logger.Debug("syncing decided")
	highest, sender, localHeight, err := f.decidedFetcher.GetLastDecided(ctx, message.ToMessageID(identifier), func(i spectypes.MessageID) (*specqbft.SignedMessage, error) {
		return from, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not get last decided from peers")
	}
	logger.Debug("highest decided", zap.Int64("local", int64(localHeight)),
		zap.Any("highest", highest), zap.Any("to", to))
	if highest == nil {
		logger.Debug("could not find highest decided from peers")
		if to == nil {
			return nil, nil
		}
		highest = to
	}
	if localHeight >= highest.Message.Height {
		logger.Debug("local height is equal or higher than remote")
		return nil, nil
	}

	res := make([]*specqbft.SignedMessage, 0)
	handleDecided := func(msg *specqbft.SignedMessage) error {
		res = append(res, msg)
		return nil
	}

	// a special case where no need to sync
	if localHeight+1 == highest.Message.Height {
		return []*specqbft.SignedMessage{highest}, err
	}

	if len(sender) > 0 {
		err = f.historySyncer.SyncRange(ctx, message.ToMessageID(identifier), handleDecided, localHeight, highest.Message.Height, sender)
		if err != nil {
			return nil, errors.Wrap(err, "could not complete sync")
		}
	}
	if specqbft.Height(len(res)) < highest.Message.Height-localHeight {
		logger.Warn("not all messages were saved in range",
			zap.Int("processed", len(res)),
			zap.Int64("from", int64(localHeight)),
			zap.Int64("to", int64(highest.Message.Height)))
	}

	res = append(res, highest)
	return res, err
}

func (f *fullNode) UpdateDecided(msg *specqbft.SignedMessage) (*specqbft.SignedMessage, error) {
	_, err := f.updateDecidedHistory(msg)
	if err != nil {
		f.logger.Debug("could not update decided history", zap.Error(err))
	}
	updated, err := strategy.UpdateLastDecided(f.logger, f.store, msg)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (f *fullNode) updateDecidedHistory(msg *specqbft.SignedMessage) (*specqbft.SignedMessage, error) {
	localMsgs, err := f.store.GetDecided(msg.Message.Identifier, msg.Message.Height, msg.Message.Height)
	if err != nil {
		return nil, errors.Wrap(err, "could not read decided")
	}
	if len(localMsgs) == 0 || localMsgs[0] == nil {
		// no previous decided
		return msg, f.store.SaveDecided(msg)
	}
	localMsg := localMsgs[0]
	if localMsg.Message.Height == msg.Message.Height {
		updated, ok := strategy.CheckSigners(localMsg, msg)
		if !ok {
			return nil, nil
		}
		msg = updated
	}
	if err := f.store.SaveDecided(msg); err != nil {
		return nil, errors.Wrap(err, "could not save decided history")
	}
	return msg, nil
}

func (f *fullNode) GetDecided(identifier []byte, heightRange ...specqbft.Height) ([]*specqbft.SignedMessage, error) {
	if len(heightRange) < 2 {
		return nil, errors.New("missing height range")
	}
	return f.store.GetDecided(identifier, heightRange[0], heightRange[1])
}

func (f *fullNode) GetLastDecided(identifier []byte) (*specqbft.SignedMessage, error) {
	return f.store.GetLastDecided(identifier)
}
