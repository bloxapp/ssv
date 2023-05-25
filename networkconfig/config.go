package networkconfig

import (
	"fmt"

	spectypes "github.com/bloxapp/ssv-spec/types"
)

var SupportedConfigs = map[string]spectypes.BeaconNetwork{
	Mainnet.Name:     Mainnet,
	JatoV2Stage.Name: JatoV2Stage,
}

func GetNetworkByName(name string) (spectypes.BeaconNetwork, error) {
	if network, ok := SupportedConfigs[name]; ok {
		return network, nil
	}

	return spectypes.BeaconNetwork{}, fmt.Errorf("network not supported: %v", name)
}
