package testing

import (
	"context"

	spectypes "github.com/ssvlabs/ssv-spec/types"
	spectestingutils "github.com/ssvlabs/ssv-spec/types/testingutils"
	"go.uber.org/zap"

	"github.com/ssvlabs/ssv/networkconfig"
	"github.com/ssvlabs/ssv/protocol/v2/qbft/testing"
	"github.com/ssvlabs/ssv/protocol/v2/ssv/runner"
	"github.com/ssvlabs/ssv/protocol/v2/ssv/validator"
	"github.com/ssvlabs/ssv/protocol/v2/types"
)

var BaseValidator = func(logger *zap.Logger, keySet *spectestingutils.TestKeySet) *validator.Validator {
	ctx, cancel := context.WithCancel(context.TODO())

	return validator.NewValidator(
		ctx,
		cancel,
		validator.Options{
			Network:       spectestingutils.NewTestingNetwork(1, keySet.OperatorKeys[1]),
			Beacon:        spectestingutils.NewTestingBeaconNode(),
			BeaconNetwork: networkconfig.TestNetwork.Beacon,
			Storage:       testing.TestingStores(logger),
			SSVShare: &types.SSVShare{
				Share: *spectestingutils.TestingShare(keySet),
			},
			Signer:            spectestingutils.NewTestingKeyManager(),
			OperatorSigner:    spectestingutils.NewTestingOperatorSigner(keySet, 1),
			SignatureVerifier: spectestingutils.NewTestingVerifier(),
			DutyRunners: map[spectypes.BeaconRole]runner.Runner{
				spectypes.BNRoleAttester:                  AttesterRunner(logger, keySet),
				spectypes.BNRoleProposer:                  ProposerRunner(logger, keySet),
				spectypes.BNRoleAggregator:                AggregatorRunner(logger, keySet),
				spectypes.BNRoleSyncCommittee:             SyncCommitteeRunner(logger, keySet),
				spectypes.BNRoleSyncCommitteeContribution: SyncCommitteeContributionRunner(logger, keySet),
				spectypes.BNRoleValidatorRegistration:     ValidatorRegistrationRunner(logger, keySet),
				spectypes.BNRoleVoluntaryExit:             VoluntaryExitRunner(logger, keySet),
			},
		},
	)
}
