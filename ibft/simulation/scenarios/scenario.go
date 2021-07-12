package scenarios

import (
	"github.com/bloxapp/ssv/ibft"
	validatorstorage "github.com/bloxapp/ssv/validator/storage"
)

// IScenario is an interface for simulator scenarios
type IScenario interface {
	// Start is a blocking call to start scenario
	Start(nodes []ibft.IBFT, shares map[uint64]*validatorstorage.Share)
}
