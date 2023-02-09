package api

import (
	"encoding/hex"

	"github.com/pkg/errors"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
)

// Message represents an exporter message
type Message struct {
	// Type is the type of message
	Type MessageType `json:"type"`
	// Filter
	Filter MessageFilter `json:"filter"`
	// Values holds the results, optional as it's relevant for response
	Data interface{} `json:"data,omitempty"`
}

// NewDecidedAPIMsg creates a new message from the given message
// TODO: avoid converting to v0 once explorer is upgraded
func NewDecidedAPIMsg(msgs ...*specqbft.SignedMessage) Message {
	data, err := DecidedAPIData(msgs...)
	if err != nil {
		return Message{
			Type: TypeDecided,
			Data: []string{},
		}
	}

	identifier := specqbft.ControllerIdToMessageID(msgs[0].Message.Identifier)
	pkv := identifier.GetPubKey()
	role := identifier.GetRoleType()
	return Message{
		Type: TypeDecided,
		Filter: MessageFilter{
			PublicKey: hex.EncodeToString(pkv),
			From:      uint64(msgs[0].Message.Height),
			To:        uint64(msgs[len(msgs)-1].Message.Height),
			Role:      role.String(),
		},
		Data: data,
	}
}

// DecidedAPIData creates a new message from the given message
func DecidedAPIData(msgs ...*specqbft.SignedMessage) (interface{}, error) {
	if len(msgs) == 0 {
		return nil, errors.New("no messages")
	}
	return msgs, nil
}

// MessageFilter is a criteria for query in request messages and projection in responses
type MessageFilter struct {
	// From is the starting index of the desired data
	From uint64 `json:"from"`
	// To is the ending index of the desired data
	To uint64 `json:"to"`
	// Role is the duty type, optional as it's relevant for IBFT data
	Role string `json:"role,omitempty"`
	// PublicKey is optional, used for fetching decided messages or information about specific validator/operator
	PublicKey string `json:"publicKey,omitempty"`
}

// MessageType is the type of message being sent
type MessageType string

const (
	// TypeValidator is an enum for validator type messages
	TypeValidator MessageType = "validator"
	// TypeOperator is an enum for operator type messages
	TypeOperator MessageType = "operator"
	// TypeDecided is an enum for ibft type messages
	TypeDecided MessageType = "decided"
	// TypeError is an enum for error type messages
	TypeError MessageType = "error"
)
