package controller

import (
	"fmt"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/qbft"
	"github.com/bloxapp/ssv/protocol/v2/qbft/instance"
)

// UponDecided returns decided msg if decided, nil otherwise
func (c *Controller) UponDecided(msg *specqbft.SignedMessage) (*specqbft.SignedMessage, error) {
	if err := ValidateDecided(
		c.config,
		msg,
		c.Share,
	); err != nil {
		return nil, errors.Wrap(err, "invalid decided msg")
	}

	// get decided value
	data, err := msg.Message.GetCommitData()
	if err != nil {
		return nil, errors.Wrap(err, "could not get decided data")
	}

	inst := c.InstanceForHeight(msg.Message.Height)
	prevDecided := inst != nil && inst.State.Decided
	isFutureDecided := msg.Message.Height > c.Height
	save := true

	operation := ""
	countCommitsBefore := 0
	if inst == nil {
		i := instance.NewInstance(c.GetConfig(), c.Share, c.Identifier, msg.Message.Height)
		i.State.Round = msg.Message.Round
		i.State.Decided = true
		i.State.DecidedValue = data.Data
		i.State.CommitContainer.AddMsg(msg)
		c.StoredInstances.addNewInstance(i)
		operation = "new"
	} else if decided, _ := inst.IsDecided(); !decided {
		inst.State.Decided = true
		inst.State.Round = msg.Message.Round
		inst.State.DecidedValue = data.Data
		inst.State.CommitContainer.AddMsg(msg)
		operation = "newly-decided"
		countCommitsBefore = len(inst.State.CommitContainer.Msgs)
	} else { // decide previously, add if has more signers
		countCommitsBefore = len(inst.State.CommitContainer.Msgs)
		signers, _ := inst.State.CommitContainer.LongestUniqueSignersForRoundAndValue(msg.Message.Round, msg.Message.Data)
		operation = fmt.Sprintf("previously-decided:uniqueSigners(%d)", len(signers))
		if len(msg.Signers) > len(signers) {
			inst.State.CommitContainer.AddMsg(msg)
		} else {
			save = false
		}
	}
	c.logger.Debug("UponDecidedDebug",
		zap.Uint64("height", uint64(msg.Message.Height)),
		zap.Uint64("round", uint64(msg.Message.Round)),
		zap.String("operation", operation),
		zap.Bool("prev_decided", prevDecided),
		zap.Bool("is_future_decided", isFutureDecided),
		zap.Int("instance_count_commits_before", countCommitsBefore),
		zap.Int("msg_count_signers", len(msg.Signers)),
		zap.Bool("save", save),
	)

	if save {
		// Retrieve instance from StoredInstances (in case it was created above)
		// and save it together with the decided message.
		if inst := c.StoredInstances.FindInstance(msg.Message.Height); inst != nil {
			logger := c.logger.With(
				zap.Uint64("msg_height", uint64(msg.Message.Height)),
				zap.Uint64("ctrl_height", uint64(c.Height)),
				zap.Any("signers", msg.Signers),
			)
			if err = c.SaveInstance(inst, msg); err != nil {
				logger.Debug("failed to save instance", zap.Error(err))
			} else {
				logger.Debug("saved instance upon decided", zap.Error(err))
			}
		}
	}

	if isFutureDecided {
		// sync gap
		c.GetConfig().GetNetwork().SyncDecidedByRange(spectypes.MessageIDFromBytes(c.Identifier), c.Height, msg.Message.Height)
		// bump height
		c.Height = msg.Message.Height
	}
	if c.NewDecidedHandler != nil {
		c.NewDecidedHandler(msg)
	}
	if !prevDecided {
		return msg, nil
	}
	return nil, nil
}

func ValidateDecided(
	config qbft.IConfig,
	signedDecided *specqbft.SignedMessage,
	share *spectypes.Share,
) error {
	if !IsDecidedMsg(share, signedDecided) {
		return errors.New("not a decided msg")
	}

	if err := signedDecided.Validate(); err != nil {
		return errors.Wrap(err, "invalid decided msg")
	}

	if err := instance.BaseCommitValidation(config, signedDecided, signedDecided.Message.Height, share.Committee); err != nil {
		return errors.Wrap(err, "invalid decided msg")
	}

	msgDecidedData, err := signedDecided.Message.GetCommitData()
	if err != nil {
		return errors.Wrap(err, "could not get msg decided data")
	}
	if err := msgDecidedData.Validate(); err != nil {
		return errors.Wrap(err, "invalid decided data")
	}

	return nil
}

// IsDecidedMsg returns true if signed commit has all quorum sigs
func IsDecidedMsg(share *spectypes.Share, signedDecided *specqbft.SignedMessage) bool {
	return share.HasQuorum(len(signedDecided.Signers)) && signedDecided.Message.MsgType == specqbft.CommitMsgType
}
