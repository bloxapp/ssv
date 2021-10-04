package p2p

import (
	"encoding/json"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/network"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
)

// BroadcastDecided broadcasts a decided instance with collected signatures
func (n *p2pNetwork) BroadcastDecided(topicName []byte, msg *proto.SignedMessage) error {
	msgBytes, err := json.Marshal(network.Message{
		SignedMessage: msg,
		Type:          network.NetworkMsg_DecidedType,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal message")
	}

	topic, err := n.getTopic(topicName)
	if err != nil {
		return errors.Wrap(err, "failed to get topic")
	}

	n.logger.Debug("Broadcasting decided message", zap.String("lambda", string(msg.Message.Lambda)), zap.Any("topic", topic), zap.Any("peers", topic.ListPeers()))

	// publishing on main topic as well
	go func() {
		if mainTopic, err := n.getMainTopic(); err != nil {
			n.logger.Error("failed to get main topic")
		} else if err := mainTopic.Publish(n.ctx, msgBytes[:]); err != nil {
			n.logger.Error("failed to publish on main topic")
		}
	}()

	metricsIBFTDecidedMsgsOutbound.WithLabelValues(unwrapTopicName(topic.String()),
		strconv.FormatUint(msg.Message.SeqNumber, 10)).Inc()

	return topic.Publish(n.ctx, msgBytes)
}

// ReceivedDecidedChan returns the channel for decided messages
func (n *p2pNetwork) ReceivedDecidedChan() <-chan *proto.SignedMessage {
	ls := listener{
		decidedCh: make(chan *proto.SignedMessage, MsgChanSize),
	}

	n.listenersLock.Lock()
	n.listeners = append(n.listeners, ls)
	n.listenersLock.Unlock()

	return ls.decidedCh
}
