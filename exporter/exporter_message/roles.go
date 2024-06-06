package exporter_message

type RunnerRole int32

const (
	RoleAttester RunnerRole = iota
	RoleAggregator
	RoleProposer
	RoleSyncCommitteeContribution

	RoleValidatorRegistration
	RoleVoluntaryExit
	RoleCommittee

	RoleUnknown = -1
)

// String returns name of the runner role
func (r RunnerRole) String() string {
	switch r {
	case RoleCommittee:
		return "COMMITTEE_RUNNER"
	case RoleAttester:
		return "ATTESTER_RUNNER"
	case RoleAggregator:
		return "AGGREGATOR_RUNNER"
	case RoleProposer:
		return "PROPOSER_RUNNER"
	case RoleSyncCommitteeContribution:
		return "SYNC_COMMITTEE_CONTRIBUTION_RUNNER"
	case RoleValidatorRegistration:
		return "VALIDATOR_REGISTRATION_RUNNER"
	case RoleVoluntaryExit:
		return "VOLUNTARY_EXIT_RUNNER"
	default:
		return "UNDEFINED"
	}
}
