package validator

import (
	"fmt"
	"sort"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/storage"
	"github.com/bloxapp/ssv/logging/fields"
	"github.com/bloxapp/ssv/protocol/v2/qbft"
	qbftcontroller "github.com/bloxapp/ssv/protocol/v2/qbft/controller"
	"github.com/bloxapp/ssv/protocol/v2/ssv/queue"
	"github.com/bloxapp/ssv/protocol/v2/types"
)

type NonCommitteeValidator struct {
	Share              *types.SSVShare
	Storage            *storage.QBFTStores
	qbftController     *qbftcontroller.Controller
	commitMsgContainer *specqbft.MsgContainer
}

func NewNonCommitteeValidator(logger *zap.Logger, identifier spectypes.MessageID, opts Options) *NonCommitteeValidator {
	// currently, only need domain & storage
	config := &qbft.Config{
		Domain:                types.GetDefaultDomain(),
		Storage:               opts.Storage.Get(identifier.GetRoleType()),
		Network:               opts.Network,
		SignatureVerification: true,
	}
	ctrl := qbftcontroller.NewController(identifier[:], &opts.SSVShare.Share, config, opts.FullNode)
	ctrl.StoredInstances = make(qbftcontroller.InstanceContainer, 0, nonCommitteeInstanceContainerCapacity(opts.FullNode))
	ctrl.NewDecidedHandler = opts.NewDecidedHandler
	if _, err := ctrl.LoadHighestInstance(identifier[:]); err != nil {
		logger.Debug("❗ failed to load highest instance", zap.Error(err))
	}

	return &NonCommitteeValidator{
		Share:              opts.SSVShare,
		Storage:            opts.Storage,
		qbftController:     ctrl,
		commitMsgContainer: specqbft.NewMsgContainer(),
	}
}

func (ncv *NonCommitteeValidator) ProcessMessage(logger *zap.Logger, msg *queue.DecodedSSVMessage) {
	logger = logger.With(fields.PubKey(msg.MsgID.GetPubKey()), fields.Role(msg.MsgID.GetRoleType()))

	if err := validateMessage(ncv.Share.Share, msg); err != nil {
		logger.Warn("❌ got invalid message", zap.Error(err))
		return
	}

	switch msg.GetType() {
	case spectypes.SSVConsensusMsgType:
		signedMsg := &specqbft.SignedMessage{}
		if err := signedMsg.Decode(msg.GetData()); err != nil {
			logger.Warn("❗ failed to get consensus Message from network Message", zap.Error(err))
			return
		}
		// only supports commit msg's
		if signedMsg.Message.MsgType != specqbft.CommitMsgType {
			return
		}

		logger = logger.With(fields.Height(signedMsg.Message.Height))

		logger.Info("ncv processing message")

		//decided, err := ncv.qbftController.ProcessMsg(logger, signedMsg)
		//if err != nil {
		//	logger.Debug("❌ failed to process message",
		//		zap.Uint64("msg_height", uint64(signedMsg.Message.Height)),
		//		zap.Any("signers", signedMsg.Signers),
		//		zap.Error(err))
		//	return
		//}

		//if decided == nil {
		logger.Info("ncv add to container")
		addMsg, err := ncv.commitMsgContainer.AddFirstMsgForSignerAndRound(signedMsg)
		logger.Info("ncv add to container done")
		if err != nil {
			logger.Warn("❌ could not add commit msg to container",
				zap.Uint64("msg_height", uint64(signedMsg.Message.Height)),
				zap.Any("signers", signedMsg.Signers),
				zap.Error(err))
			return
		}
		if !addMsg {
			logger.Info("ncv didn't add commit")
			return
		}
		logger.Info("ncv added commit")

		signers, commitMsgs := ncv.commitMsgContainer.LongestUniqueSignersForRoundAndRoot(signedMsg.Message.Round, signedMsg.Message.Root)
		if !ncv.Share.HasQuorum(len(signers)) {
			logger.Info("ncv has no quorum")
			return
		}
		logger.Info("ncv has quorum")

		signedMsg, err = aggregateCommitMsgs(commitMsgs)
		if err != nil {
			logger.Warn("❌ could not add aggregate commit messages",
				zap.Uint64("msg_height", uint64(signedMsg.Message.Height)),
				zap.Any("signers", signedMsg.Signers),
				zap.Error(err))
			return
		}
		logger.Info("ncv aggregated commits")
		//}

		if inst := ncv.qbftController.StoredInstances.FindInstance(signedMsg.Message.Height); inst != nil {
			logger := logger.With(
				zap.Uint64("msg_height", uint64(signedMsg.Message.Height)),
				zap.Uint64("ctrl_height", uint64(ncv.qbftController.Height)),
				zap.Any("signers", signedMsg.Signers),
			)
			if err = ncv.qbftController.SaveInstance(inst, signedMsg); err != nil {
				logger.Warn("❗failed to save instance", zap.Error(err))
			} else {
				logger.Info("💾 saved instance")
			}
		}

		return
	}
}

// nonCommitteeInstanceContainerCapacity returns the capacity of InstanceContainer for non-committee validators
func nonCommitteeInstanceContainerCapacity(fullNode bool) int {
	if fullNode {
		// Helps full nodes reduce
		return 2
	}
	return 1
}

func aggregateCommitMsgs(msgs []*specqbft.SignedMessage) (*specqbft.SignedMessage, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("can't aggregate zero commit msgs")
	}

	var ret *specqbft.SignedMessage
	for _, m := range msgs {
		if ret == nil {
			ret = m.DeepCopy()
		} else {
			if err := ret.Aggregate(m); err != nil {
				return nil, fmt.Errorf("could not aggregate commit msg: %w", err)
			}
		}
	}

	sort.Slice(ret.Signers, func(i, j int) bool {
		return ret.Signers[i] < ret.Signers[j]
	})

	return ret, nil
}
