package topics

import (
	"context"
	"github.com/bloxapp/ssv/network/forks"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
)

// MsgValidatorFunc represents a message validator
type MsgValidatorFunc = func(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult

// NewSSVMsgValidator creates a new msg validator that validates message structure,
// and checks that the message was sent on the right topic.
// TODO: enable post SSZ change, remove logs, break into smaller validators?
func NewSSVMsgValidator(plogger *zap.Logger, fork forks.Fork, self peer.ID) func(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
	return func(ctx context.Context, p peer.ID, pmsg *pubsub.Message) pubsub.ValidationResult {
		topic := pmsg.GetTopic()
		metricsPubsubActiveMsgValidation.WithLabelValues(topic).Inc()
		defer metricsPubsubActiveMsgValidation.WithLabelValues(topic).Dec()
		if len(pmsg.GetData()) == 0 {
			reportValidationResult(validationResultNoData)
			return pubsub.ValidationReject
		}
		msg, err := fork.DecodeNetworkMsg(pmsg.GetData())
		if err != nil {
			// can't decode message
			//logger.Debug("invalid: can't decode message", zap.Error(err))
			reportValidationResult(validationResultEncoding)
			return pubsub.ValidationReject
		}
		if msg == nil {
			reportValidationResult(validationResultEncoding)
			return pubsub.ValidationReject
		}
		pmsg.ValidatorData = *msg
		return pubsub.ValidationAccept
		// check decided topic
		//currentTopic := pmsg.GetTopic()
		//currentTopicBaseName := fork.GetTopicBaseName(currentTopic)
		//if msg.MsgType == spectypes.SSVDecidedMsgType {
		//	if decidedTopic := fork.DecidedTopic(); decidedTopic == currentTopicBaseName {
		//		return pubsub.ValidationAccept
		//	}
		//}
		//topics := fork.ValidatorTopicID(msg.GetID().GetPubKey())
		//for _, tp := range topics {
		//	if tp == currentTopicBaseName {
		//		reportValidationResult(validationResultValid)
		//		return pubsub.ValidationAccept
		//	}
		//}
		//reportValidationResult(validationResultTopic)
		//return pubsub.ValidationReject
	}
}

//// CombineMsgValidators executes multiple validators
//func CombineMsgValidators(validators ...MsgValidatorFunc) MsgValidatorFunc {
//	return func(ctx context.Context, p peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
//		res := pubsub.ValidationAccept
//		for _, v := range validators {
//			if res = v(ctx, p, msg); res == pubsub.ValidationReject {
//				break
//			}
//		}
//		return res
//	}
//}
