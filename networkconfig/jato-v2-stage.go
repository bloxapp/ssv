package networkconfig

import (
	"math"
	"math/big"

	spectypes "github.com/bloxapp/ssv-spec/types"

	"github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
)

var JatoV2Stage = NetworkConfig{
	Name:                 "jato-v2-stage",
	Beacon:               beacon.NewNetwork(spectypes.PraterNetwork),
	Domain:               [4]byte{0x00, 0x00, 0x30, 0x12},
	GenesisEpoch:         152834,
	RegistrySyncOffset:   new(big.Int).SetInt64(9249887),
	RegistryContractAddr: "0xd6b633304Db2DD59ce93753FA55076DA367e5b2c",
	Bootnodes: []string{
		"enr:-Li4QO86ZMZr_INMW_WQBsP2jS56yjrHnZXxAUOKJz4_qFPKD1Cr3rghQD2FtXPk2_VPnJUi8BBiMngOGVXC0wTYpJGGAYgqnGSNh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDkvpOTAAAQIP__________gmlkgnY0gmlwhArqAsGJc2VjcDI1NmsxoQKNW0Mf-xTXcevRSkZOvoN0Q0T9OkTjGZQyQeOl3bYU3YN0Y3CCE4iDdWRwgg-g",
		"enr:-Li4QBoH15fXLV78y1_nmD5sODveptALORh568iWLS_eju3SUvF2ZfGE2j-nERKU1zb2g5KlS8L70SRLdRUJ-pHH-fmGAYgvh9oGh2F0dG5ldHOIAAAAAAAAAACEZXRoMpDkvpOTAAAQIP__________gmlkgnY0gmlwhArqAsGJc2VjcDI1NmsxoQO_tV3JP75ZUZPjhOgc2VqEu_FQEMeHc4AyOz6Lz33M2IN0Y3CCE4mDdWRwgg-h",
	},
	WhitelistedOperatorKeys: []string{
		// Blox's exporter nodes.
		"LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBdEt5SUkyOHZ0WnBCMStEdTV4V0sKQkJYanZiM1ZQVUFIVUtaemg4eDhCd21PTzQxZlYvR0o3VzF4RjBOWTREajB3UkpDQ09OcnM0VVl6RjhJVUpkNAp4dHBKbnNWV2RiMFVGOFZEOFZuZm1mTXEwT2VoQnNWTE1ZbzJPbE4wS1lzWHVXTnFXS1VUbmtHWkd2VjN6SEsyCkNTK0FwcGJaVFZPU21tQVBxc3R5aFdKVnhiWTE3V0RQRitsME5UNnpFSjB2VG1ucDhwWjkrSG8rK2pMY0dFR0UKcm5VR2gyMlYrU2dLdEUwSElFTUVzOUo5eFNnL3YxZFpib1QyQ1BKbExWeVBqR29yOXh4THVXZjRWN2ptOC9CaApKWTRvRldSL2ZSR25MWFQzSHB5R25DK2YrdDZ5SnM4ejc2ZStMWm40SkpYSnZwdFZDamNTMXVSako4QlNOQXlsCnR3SURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K",
		"LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBNWhSc003RzVGaGswelhXMTd0SXYKa0JZSW1yYnpjQjRXWVhBSVNmNjhIbjVicngwWmFmcERYczRzeUt1WitaR01INTZSSWMxN0pPVU04YUF5ekZCLwpPSEszQ2thSnRmUWdCbE1jOXUvTnp5Z1FwalM5QXF4VTEzaTdrNkkwMHc4RVNSVzd2WnVBSzVxM3NCNHZMVWdFCnR3c3ZNVWtVUUJZSzdPSjdaR1Q3UklMTStPSXl4ZTh2MXhiM3lNeWo4aTk3OVZ2Q0xPL1Z2YnlSVmRvSndCbHoKSG9zcmRoN1UwT0lKYVFmTFVLTDlPUzdpQUM0NlRtalZxa3djeUFDRm90VGh0cS80L1RKcUc4eGwwWHIxTm12UQpaSGxwVDRFU0FOM21JUHY0Wko1NnYzeXFaUUNQYWw2VEkxM3ZhSldnN3krbzRoQURuQmdIdGF5TTRwU1c0Z3FECmdRSURBUUFCCi0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0K",
	},
	FinalizedCheckpointForkActivationHeight: math.MaxUint64, // inf #TODO replace
}
