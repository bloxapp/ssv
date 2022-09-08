package message

import spectypes "github.com/bloxapp/ssv-spec/types"

// SSVSyncMsgType extension for spec msg type
const (
	SSVSyncMsgType spectypes.MsgType = 4
)

// MsgTypeToString extension for spec msg type. convert spec msg type to string
func MsgTypeToString(mt spectypes.MsgType) string {
	switch mt {
	case spectypes.SSVConsensusMsgType:
		return "consensus"
	case spectypes.SSVDecidedMsgType:
		return "decided"
	case spectypes.SSVPartialSignatureMsgType:
		return "partialSignature"
	case spectypes.DKGMsgType:
		return "dkg"
	default:
		return ""
	}
}

// ToMessageID extension for spec msg id, returns spec messageID
func ToMessageID(b []byte) spectypes.MessageID {
	ret := spectypes.MessageID{}
	copy(ret[0:52], b[:])
	return ret
}
