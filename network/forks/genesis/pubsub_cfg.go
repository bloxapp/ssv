package genesis

import (
	"encoding/binary"

	"github.com/bloxapp/ssv/network/forks"
	"github.com/cespare/xxhash/v2"
)

// MsgID returns msg_id for the given message
func (genesis *ForkGenesis) MsgID() forks.MsgIDFunc {
	return func(msg []byte) string {
		if len(msg) == 0 {
			return ""
		}
		var b [12]byte
		binary.LittleEndian.PutUint64(b[:], xxhash.Sum64(msg))
		return string(b[:])
	}
}

// Subnets returns the subnets count for this fork
func (genesis *ForkGenesis) Subnets() int {
	return int(subnetsCount)
}
