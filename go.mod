module github.com/bloxapp/ssv

go 1.15

require (
	github.com/attestantio/go-eth2-client v0.6.30
	github.com/bloxapp/eth2-key-manager v1.1.2
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/ethereum/go-ethereum v1.9.25
	github.com/ferranbt/fastssz v0.0.0-20210905181407-59cf6761a7d5
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/herumi/bls-eth-go-binary v0.0.0-20210917013441-d37c07cfda4e
	github.com/ilyakaznacheev/cleanenv v1.2.5
	github.com/ipfs/go-ipfs-addr v0.0.1
	github.com/libp2p/go-libp2p v0.14.4
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/libp2p/go-libp2p-noise v0.2.0
	github.com/libp2p/go-libp2p-pubsub v0.5.0
	github.com/libp2p/go-tcp-transport v0.2.8
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/prysmaticlabs/eth2-types v0.0.0-20210303084904-c9735a06829d
	github.com/prysmaticlabs/ethereumapis v0.0.0-20210118163152-3569d231d255
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/go-ssz v0.0.0-20200612203617-6d5c9aa213ae
	github.com/prysmaticlabs/prysm v1.4.2-0.20211005004110-843ed50e0acc
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	github.com/wealdtech/go-eth2-util v1.6.3
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.18.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.37.0
)

replace github.com/ethereum/go-ethereum => github.com/prysmaticlabs/bazel-go-ethereum v0.0.0-20201113091623-013fd65b3791

replace github.com/google/flatbuffers => github.com/google/flatbuffers v1.11.0

replace github.com/attestantio/go-eth2-client v0.6.30 => github.com/bloxapp/go-eth2-client v0.6.31-0.20210706133239-eb1bd7a3cb25

replace github.com/prysmaticlabs/prysm => github.com/prysmaticlabs/prysm v1.4.2-0.20211005004110-843ed50e0acc
