package factory

import (
	"github.com/bloxapp/ssv/network/forks"
	forksv0 "github.com/bloxapp/ssv/network/forks/v0"
	forksv1 "github.com/bloxapp/ssv/network/forks/v1"
	forksv2 "github.com/bloxapp/ssv/network/forks/v2"
	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
)

// NewFork returns a new fork instance from the given version
func NewFork(forkVersion forksprotocol.ForkVersion) forks.Fork {
	switch forkVersion {
	case forksprotocol.V0ForkVersion:
		return &forksv0.ForkV0{}
	case forksprotocol.V1ForkVersion:
		return &forksv1.ForkV1{}
	case forksprotocol.V2ForkVersion:
		return &forksv2.ForkV2{}
	default:
		return &forksv0.ForkV0{}
	}
}
