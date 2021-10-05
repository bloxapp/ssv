package p2p

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"time"
)

// Config - describe the config options for p2p network
type Config struct {
	// yaml/env arguments
	Enr              string        `yaml:"Enr" env:"ENR_KEY" env-description:"enr used in discovery" env-default:""`
	DiscoveryType    string        `yaml:"DiscoveryType" env:"DISCOVERY_TYPE_KEY" env-description:"Method to use in discovery" env-default:"discv5"`
	TCPPort          int           `yaml:"TcpPort" env:"TCP_PORT" env-default:"13000"`
	UDPPort          int           `yaml:"UdpPort" env:"UDP_PORT" env-default:"12000"`
	HostAddress      string        `yaml:"HostAddress" env:"HOST_ADDRESS" env-required:"true" env-description:"External ip node is exposed for discovery"`
	HostDNS          string        `yaml:"HostDNS" env:"HOST_DNS" env-description:"External DNS node is exposed for discovery"`
	RequestTimeout   time.Duration `yaml:"RequestTimeout" env:"P2P_REQUEST_TIMEOUT"  env-default:"5s"`
	MaxBatchResponse uint64        `yaml:"MaxBatchResponse" env:"P2P_MAX_BATCH_RESPONSE" env-default:"50" env-description:"maximum number of returned objects in a batch"`
	PubSubTraceOut   string        `yaml:"PubSubTraceOut" env:"PUBSUB_TRACE_OUT" env-description:"File path to hold collected pubsub traces"`
	//PubSubTracer     string        `yaml:"PubSubTracer" env:"PUBSUB_TRACER" env-description:"A remote tracer that collects pubsub traces"`

	// objects / instances
	HostID              peer.ID
	Topics              map[string]*pubsub.Topic
	Discv5BootStrapAddr []string

	// NetworkPrivateKey is used for network identity
	NetworkPrivateKey *ecdsa.PrivateKey
	// OperatorPrivateKey is used for operator identity
	OperatorPrivateKey *rsa.PrivateKey
}

// TransformEnr converts defaults enr value and convert it to slice
func TransformEnr(enr string) []string {
	if len(enr) == 0 {
		// stage enr
		//enr = "enr:-LK4QHVq6HEA2KVnAw593SRMqUOvMGlkP8Jb-qHn4yPLHx--cStvWc38Or2xLcWgDPynVxXPT9NWIEXRzrBUsLmcFkUBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhDbUHcyJc2VjcDI1NmsxoQO8KQz5L1UEXzEr-CXFFq1th0eG6gopbdul2OQVMuxfMoN0Y3CCE4iDdWRwgg-g"

		// prod enr
		//internal ip
		//enr = "enr:-LK4QPbCB0Mw_8ji7D02OwXmqSRZe9wTmitle_cQnECIl-5GBPH9PH__eUpdeiI_t122inm62uTgO9CptbGNLKNId7gBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhArsBGGJc2VjcDI1NmsxoQO8KQz5L1UEXzEr-CXFFq1th0eG6gopbdul2OQVMuxfMoN0Y3CCE4iDdWRwgg-g"
		//external ip
		enr = "enr:-LK4QMmL9hLJ1csDN4rQoSjlJGE2SvsXOETfcLH8uAVrxlHaELF0u3NeKCTY2eO_X1zy5eEKcHruyaAsGNiyyG4QWUQBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpD1pf1CAAAAAP__________gmlkgnY0gmlwhCLdu_SJc2VjcDI1NmsxoQO8KQz5L1UEXzEr-CXFFq1th0eG6gopbdul2OQVMuxfMoN0Y3CCE4iDdWRwgg-g"
	}
	return []string{
		enr,
	}
}
