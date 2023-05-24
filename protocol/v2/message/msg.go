package message

import (
	"fmt"

	spectypes "github.com/bloxapp/ssv-spec/types"
)

const (
	// SSVSyncMsgType extends spec msg type
	SSVSyncMsgType spectypes.MsgType = 100
	// SSVEventMsgType extends spec msg type
	SSVEventMsgType spectypes.MsgType = 200
)

// MsgTypeToString extension for spec msg type. convert spec msg type to string
func MsgTypeToString(mt spectypes.MsgType) string {
	switch mt {
	case spectypes.SSVConsensusMsgType:
		return "consensus"
	case spectypes.SSVPartialSignatureMsgType:
		return "partial_signature"
	case spectypes.DKGMsgType:
		return "dkg"
	case SSVSyncMsgType:
		return "sync"
	case SSVEventMsgType:
		return "event"
	default:
		return fmt.Sprintf("unknown - %d", mt)
	}
}

// BeaconRoleFromString returns BeaconRole from string
func BeaconRoleFromString(s string) (spectypes.BeaconRole, error) {
	switch s {
	case "ATTESTER":
		return spectypes.BNRoleAttester, nil
	case "AGGREGATOR":
		return spectypes.BNRoleAggregator, nil
	case "PROPOSER":
		return spectypes.BNRoleProposer, nil
	case "SYNC_COMMITTEE":
		return spectypes.BNRoleSyncCommittee, nil
	case "SYNC_COMMITTEE_CONTRIBUTION":
		return spectypes.BNRoleSyncCommitteeContribution, nil
	case "VALIDATOR_REGISTRATION":
		return spectypes.BNRoleValidatorRegistration, nil
	default:
		return 0, fmt.Errorf("unknown role: %s", s)
	}
}
