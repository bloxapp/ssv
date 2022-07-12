package scenarios

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bloxapp/eth2-key-manager/core"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/automation/commons"
	"github.com/bloxapp/ssv/automation/qbft/runner"
	p2pv1 "github.com/bloxapp/ssv/network/p2p"
	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	"github.com/bloxapp/ssv/protocol/v1/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v1/message"
	qbftstorage "github.com/bloxapp/ssv/protocol/v1/qbft/storage"
	"github.com/bloxapp/ssv/protocol/v1/validator"
	"github.com/bloxapp/ssv/utils/rsaencryption"
)

// FullNodeScenario is the fullNode scenario name
const FullNodeScenario = "full-node"

// fullNodeScenario is the scenario with 3 regular nodes and 1 full node
type fullNodeScenario struct {
	logger     *zap.Logger
	sks        map[uint64]*bls.SecretKey
	share      *beacon.Share
	validators []validator.IValidator
}

// newFullNodeScenario creates a fullNode scenario instance
func newFullNodeScenario(logger *zap.Logger) runner.Scenario {
	return &fullNodeScenario{logger: logger}
}

func (r *fullNodeScenario) NumOfOperators() int {
	return 3
}

func (r *fullNodeScenario) NumOfBootnodes() int {
	return 0
}

func (r *fullNodeScenario) NumOfFullNodes() int {
	return 2
}

func (r *fullNodeScenario) Name() string {
	return FullNodeScenario
}

func (r *fullNodeScenario) PreExecution(ctx *runner.ScenarioContext) error {
	totalNodes := r.NumOfOperators() + r.NumOfFullNodes()
	share, sks, validators, err := createShareAndValidators(ctx.Ctx, r.logger, ctx.LocalNet, ctx.KeyManagers, ctx.Stores, totalNodes, totalNodes-1)
	if err != nil {
		return errors.Wrap(err, "could not create share")
	}

	// save all references
	r.validators = validators
	r.sks = sks
	r.share = share

	routers := make([]*runner.Router, totalNodes)

	loggerFactory := func(who string) *zap.Logger {
		logger := zap.L().With(zap.String("who", who))
		return logger
	}

	for i, node := range ctx.LocalNet.Nodes {
		routers[i] = &runner.Router{
			Logger:      loggerFactory(fmt.Sprintf("msgRouter-%d", i)),
			Controllers: r.validators[i].(*validator.Validator).Ibfts(),
		}
		node.UseMessageRouter(routers[i])
	}

	return nil
}
func (r *fullNodeScenario) Execute(ctx *runner.ScenarioContext) error {
	if len(r.sks) == 0 || r.share == nil {
		return errors.New("pre-execution failed")
	}

	var wg sync.WaitGroup
	var startErr error
	for _, val := range r.validators {
		wg.Add(1)
		go func(val validator.IValidator) {
			defer wg.Done()
			if err := val.Start(); err != nil {
				startErr = errors.Wrap(err, "could not start validator")
			}
			<-time.After(time.Second * 3)
		}(val)
	}
	wg.Wait()

	if startErr != nil {
		return errors.Wrap(startErr, "could not start validators")
	}

	const fromHeight = message.Height(0)
	const toHeight = message.Height(3)
	for height := fromHeight; height <= toHeight; height++ {
		allButOneValidators := r.validators[:r.NumOfOperators()+(r.NumOfFullNodes()-1)]
		if err := r.startInstances(height, allButOneValidators...); err != nil {
			return errors.Wrap(err, "could not start instances")
		}
	}

	<-time.After(time.Second * 3)

	return nil
}

func (r *fullNodeScenario) PostExecution(ctx *runner.ScenarioContext) error {
	msgs, err := ctx.Stores[len(ctx.Stores)-1].GetDecided(message.NewIdentifier(r.share.PublicKey.Serialize(), message.RoleTypeAttester), message.Height(0), message.Height(4))
	if err != nil {
		return err
	}

	if got, want := len(msgs), 4; got < want {
		return fmt.Errorf("node-4 didn't sync all messages, got %d, want %d", got, want)
	}
	r.logger.Debug("msgs", zap.Any("msgs", msgs))

	return nil
}

func createShareAndValidators(ctx context.Context, logger *zap.Logger, net *p2pv1.LocalNet, kms []beacon.KeyManager, stores []qbftstorage.QBFTStore, regularNodes, committeeNodes int) (*beacon.Share, map[uint64]*bls.SecretKey, []validator.IValidator, error) {
	validators := make([]validator.IValidator, 0)
	operators := make([][]byte, 0)
	for i := 0; i < len(net.NodeKeys) && i < committeeNodes; i++ {
		pub, err := rsaencryption.ExtractPublicKey(net.NodeKeys[i].OperatorKey)
		if err != nil {
			return nil, nil, nil, err
		}
		operators = append(operators, []byte(pub))
	}
	// create share
	share, sks, err := commons.CreateShare(operators)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not create share")
	}
	// add to key-managers and subscribe to topic
	for i, km := range kms {
		if i < committeeNodes {
			err = km.AddShare(sks[uint64(i+1)])
			if err != nil {
				return nil, nil, nil, err
			}
		}

		val := validator.NewValidator(&validator.Options{
			Context:     ctx,
			Logger:      logger.With(zap.String("who", fmt.Sprintf("node-%d", i))),
			IbftStorage: stores[i],
			P2pNetwork:  net.Nodes[i],
			Network:     beacon.NewNetwork(core.NetworkFromString("prater")),
			Share: &beacon.Share{
				NodeID:       message.OperatorID(i + 1),
				PublicKey:    share.PublicKey,
				Committee:    share.Committee,
				Metadata:     share.Metadata,
				OwnerAddress: share.OwnerAddress,
				Operators:    share.Operators,
			},
			ForkVersion:                forksprotocol.V0ForkVersion, // TODO need to check v1 too?
			Beacon:                     nil,
			Signer:                     km,
			SyncRateLimit:              time.Millisecond * 10,
			SignatureCollectionTimeout: time.Second * 5,
			ReadMode:                   i >= committeeNodes,
			FullNode:                   i >= regularNodes,
			DutyRoles:                  []message.RoleType{message.RoleTypeAttester},
		})
		validators = append(validators, val)
	}
	return share, sks, validators, nil
}

func (r *fullNodeScenario) startInstances(height message.Height, instances ...validator.IValidator) error {
	var wg sync.WaitGroup

	for i, instance := range instances {
		wg.Add(1)
		go func(node validator.IValidator, index int, seqNumber message.Height) {
			if err := startNode(node, seqNumber, []byte("value"), r.logger); err != nil {
				r.logger.Error("could not start node", zap.Int("node", index), zap.Error(err))
			}
			wg.Done()
		}(instance, i, height)
	}

	wg.Wait()

	return nil
}
