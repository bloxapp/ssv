package controller

import (
	"fmt"
	"testing"

	"github.com/bloxapp/ssv/ibft"
	instance "github.com/bloxapp/ssv/ibft/instance"
	"github.com/bloxapp/ssv/ibft/proto"
	"github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/storage/collections"
	"github.com/bloxapp/ssv/utils/threadsafe"
	validatorstorage "github.com/bloxapp/ssv/validator/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testIBFTInstance(t *testing.T) *Controller {
	ret := &Controller{
		Identifier:   []byte("lambda_11"),
		initHandlers: threadsafe.NewSafeBool(),
		initSynced:   threadsafe.NewSafeBool(),
		// instances: make([]*Instance, 0),
	}

	ret.fork = testFork(ret)
	return ret
}

func TestCanStartNewInstance(t *testing.T) {
	sks, nodes := GenerateNodes(4)

	tests := []struct {
		name            string
		opts            ibft.ControllerStartInstanceOptions
		share           *validatorstorage.Share
		storage         collections.Iibft
		initFinished    bool
		initSynced      bool
		currentInstance ibft.Instance
		expectedError   string
	}{
		{
			"valid next instance start",
			ibft.ControllerStartInstanceOptions{
				SeqNumber: 11,
			},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			populatedStorage(t, sks, 10),
			true,
			true,
			nil,
			"",
		},
		{
			"valid first instance",
			ibft.ControllerStartInstanceOptions{
				SeqNumber: 0,
			},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			nil,
			true,
			true,
			nil,
			"",
		},
		{
			"didn't finish initialization",
			ibft.ControllerStartInstanceOptions{},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			nil,
			false,
			false,
			nil,
			"iBFT hasn't initialized yet",
		},
		{
			"didn't finish sync",
			ibft.ControllerStartInstanceOptions{},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			nil,
			true,
			false,
			nil,
			"iBFT hasn't initialized yet",
		},
		{
			"sequence skips",
			ibft.ControllerStartInstanceOptions{
				SeqNumber: 12,
			},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			populatedStorage(t, sks, 10),
			true,
			true,
			nil,
			"instance seq invalid",
		},
		{
			"past instance",
			ibft.ControllerStartInstanceOptions{
				SeqNumber: 10,
			},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			populatedStorage(t, sks, 10),
			true,
			true,
			nil,
			"instance seq invalid",
		},
		{
			"didn't finish current instance",
			ibft.ControllerStartInstanceOptions{
				SeqNumber: 11,
			},
			&validatorstorage.Share{
				NodeID:    1,
				PublicKey: validatorPK(sks),
				Committee: nodes,
			},
			populatedStorage(t, sks, 10),
			true,
			true,
			instance.NewInstanceWithState(&proto.State{
				SeqNumber: threadsafe.Uint64(10),
			}),
			fmt.Sprintf("current instance (%d) is still running", 10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i := testIBFTInstance(t)
			i.initHandlers.Set(test.initFinished)
			i.initSynced.Set(test.initSynced)
			if test.currentInstance != nil {
				i.currentInstance = test.currentInstance
			}
			if test.storage != nil {
				i.ibftStorage = test.storage
			} else {
				options := basedb.Options{
					Type:   "badger-memory",
					Logger: zap.L(),
					Path:   "",
				}
				// creating new db instance each time to get cleared one (without no data)
				db, err := storage.GetStorageFactory(options)
				require.NoError(t, err)
				s := collections.NewIbft(db, options.Logger, "attestation")
				i.ibftStorage = &s
			}

			i.ValidatorShare = test.share
			i.instanceConfig = proto.DefaultConsensusParams()
			// i.instances = test.prevInstances
			instanceOpts, err := i.instanceOptionsFromStartOptions(test.opts)
			require.NoError(t, err)
			// instanceOpts.SeqNumber = test.seqNumber
			err = i.canStartNewInstance(*instanceOpts)

			if len(test.expectedError) > 0 {
				require.EqualError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
