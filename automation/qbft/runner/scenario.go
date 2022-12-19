package runner

import (
	"context"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/storage"
	p2pv1 "github.com/bloxapp/ssv/network/p2p"
	"github.com/bloxapp/ssv/storage/basedb"
)

// ScenarioFactory creates Scenario instances.
type ScenarioFactory func(name string) Scenario

// ScenarioContext is the context object that is passed in execution.
type ScenarioContext struct {
	Ctx         context.Context
	LocalNet    *p2pv1.LocalNet
	Stores      []*storage.QBFTStores
	KeyManagers []spectypes.KeyManager
	DBs         []basedb.IDb
}

// Bootstrapper bootstraps the given scenario.
type Bootstrapper func(ctx context.Context, logger *zap.Logger, scenario Scenario) (*ScenarioContext, error)

type scenarioCfg interface {
	// NumOfOperators returns the desired number of operators for the test.
	NumOfOperators() int
	// NumOfBootnodes returns the desired number of bootnodes for the test.
	// zero in case we want mdns
	NumOfBootnodes() int
	// NumOfFullNodes returns the desired number of full nodes for the test.
	NumOfFullNodes() int
}

// Scenario represents a testplan for a specific scenario
type Scenario interface {
	scenarioCfg
	// Name is the name of the scenario
	Name() string
	// PreExecution is invoked prior to the scenario, used for setup
	PreExecution(ctx *ScenarioContext) error
	// Execute is the actual test scenario to run
	Execute(ctx *ScenarioContext) error
	// PostExecution is invoked after execution, used for cleanup etc.
	PostExecution(ctx *ScenarioContext) error
}
