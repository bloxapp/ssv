package tests

import (
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/bloxapp/ssv/automation/qbft/runner"
	"github.com/bloxapp/ssv/automation/qbft/scenarios"
	"github.com/bloxapp/ssv/utils/logex"
)

func Test_Automation_QBFTScenarios(t *testing.T) {
	logger := logex.Build("simulation", zapcore.DebugLevel, nil)
	scenariosToRun := []string{
		scenarios.RegularScenario,
	}

	for _, s := range scenariosToRun {
		scenario := scenarios.NewScenario(s, logger)
		runner.Start(t, logger, scenario, scenarios.QBFTScenarioBootstrapper())
	}
}
