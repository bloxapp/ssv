package beacon

// RoleType type of the validator role for a specific duty
type RoleType int

// String returns name of the role
func (r RoleType) String() string {
	switch r {
	case RoleTypeUnknown:
		return "UNKNOWN"
	case RoleTypeAttester:
		return "ATTESTER"
	case RoleTypeAggregator:
		return "AGGREGATOR"
	case RoleTypeProposer:
		return "PROPOSER"
	default:
		return "UNDEFINED"
	}
}

// List of roles
const (
	RoleTypeUnknown = iota
	RoleTypeAttester
	RoleTypeAggregator
	RoleTypeProposer
)

// ToRoleType takes a string and convert it to RoleType
func ToRoleType(role string) RoleType {
	switch role {
	case "ATTESTER":
		return RoleTypeAttester
	case "AGGREGATOR":
		return RoleTypeAggregator
	case "PROPOSER":
		return RoleTypeProposer
	default:
		return RoleTypeUnknown
	}
}