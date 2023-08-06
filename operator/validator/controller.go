package validator

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	specqbft "github.com/bloxapp/ssv-spec/qbft"
	specssv "github.com/bloxapp/ssv-spec/ssv"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jellydator/ttlcache/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/ibft/storage"
	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/logging/fields"
	"github.com/bloxapp/ssv/network"
	forksfactory "github.com/bloxapp/ssv/network/forks/factory"
	nodestorage "github.com/bloxapp/ssv/operator/storage"
	forksprotocol "github.com/bloxapp/ssv/protocol/forks"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v2/message"
	p2pprotocol "github.com/bloxapp/ssv/protocol/v2/p2p"
	"github.com/bloxapp/ssv/protocol/v2/qbft"
	qbftcontroller "github.com/bloxapp/ssv/protocol/v2/qbft/controller"
	"github.com/bloxapp/ssv/protocol/v2/qbft/roundtimer"
	utilsprotocol "github.com/bloxapp/ssv/protocol/v2/queue"
	"github.com/bloxapp/ssv/protocol/v2/queue/worker"
	"github.com/bloxapp/ssv/protocol/v2/ssv/queue"
	"github.com/bloxapp/ssv/protocol/v2/ssv/runner"
	"github.com/bloxapp/ssv/protocol/v2/ssv/validator"
	"github.com/bloxapp/ssv/protocol/v2/sync/handlers"
	"github.com/bloxapp/ssv/protocol/v2/types"
	ssvtypes "github.com/bloxapp/ssv/protocol/v2/types"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/tasks"
)

//go:generate mockgen -package=mocks -destination=./mocks/controller.go -source=./controller.go

const (
	metadataBatchSize        = 100
	networkRouterConcurrency = 2048
)

// ShareEncryptionKeyProvider is a function that returns the operator private key
type ShareEncryptionKeyProvider = func() (*rsa.PrivateKey, bool, error)

type GetRecipientDataFunc func(r basedb.Reader, owner common.Address) (*registrystorage.RecipientData, bool, error)

// ShareEventHandlerFunc is a function that handles event in an extended mode
type ShareEventHandlerFunc func(share *ssvtypes.SSVShare)

// ControllerOptions for creating a validator controller
type ControllerOptions struct {
	Context                    context.Context
	DB                         basedb.Database
	SignatureCollectionTimeout time.Duration `yaml:"SignatureCollectionTimeout" env:"SIGNATURE_COLLECTION_TIMEOUT" env-default:"5s" env-description:"Timeout for signature collection after consensus"`
	MetadataUpdateInterval     time.Duration `yaml:"MetadataUpdateInterval" env:"METADATA_UPDATE_INTERVAL" env-default:"12m" env-description:"Interval for updating metadata"`
	HistorySyncBatchSize       int           `yaml:"HistorySyncBatchSize" env:"HISTORY_SYNC_BATCH_SIZE" env-default:"25" env-description:"Maximum number of messages to sync in a single batch"`
	MinPeers                   int           `yaml:"MinimumPeers" env:"MINIMUM_PEERS" env-default:"2" env-description:"The required minimum peers for sync"`
	BeaconNetwork              beaconprotocol.Network
	Network                    network.P2PNetwork
	Beacon                     beaconprotocol.BeaconNode
	ShareEncryptionKeyProvider ShareEncryptionKeyProvider
	FullNode                   bool `yaml:"FullNode" env:"FULLNODE" env-default:"false" env-description:"Save decided history rather than just highest messages"`
	Exporter                   bool `yaml:"Exporter" env:"EXPORTER" env-default:"false" env-description:""`
	BuilderProposals           bool `yaml:"BuilderProposals" env:"BUILDER_PROPOSALS" env-default:"false" env-description:"Use external builders to produce blocks"`
	KeyManager                 spectypes.KeyManager
	OperatorData               *registrystorage.OperatorData
	RegistryStorage            nodestorage.Storage
	ForkVersion                forksprotocol.ForkVersion
	NewDecidedHandler          qbftcontroller.NewDecidedHandler
	DutyRoles                  []spectypes.BeaconRole
	StorageMap                 *storage.QBFTStores
	Metrics                    validatorMetrics

	// worker flags
	WorkersCount    int `yaml:"MsgWorkersCount" env:"MSG_WORKERS_COUNT" env-default:"256" env-description:"Number of goroutines to use for message workers"`
	QueueBufferSize int `yaml:"MsgWorkerBufferSize" env:"MSG_WORKER_BUFFER_SIZE" env-default:"1024" env-description:"Buffer size for message workers"`
	GasLimit        uint64
}

// Controller represent the validators controller,
// it takes care of bootstrapping, updating and managing existing validators and their shares
type Controller interface {
	StartValidators()
	ActiveValidatorIndices(epoch phase0.Epoch) []phase0.ValidatorIndex
	GetValidator(pubKey string) (*validator.Validator, bool)
	ExecuteDuty(logger *zap.Logger, duty *spectypes.Duty)
	UpdateValidatorMetaDataLoop()
	StartNetworkHandlers()
	GetOperatorShares() []*ssvtypes.SSVShare
	// GetValidatorStats returns stats of validators, including the following:
	//  - the amount of validators in the network
	//  - the amount of active validators (i.e. not slashed or existed)
	//  - the amount of validators assigned to this operator
	GetValidatorStats() (uint64, uint64, uint64, error)
	GetOperatorData() *registrystorage.OperatorData
	SetOperatorData(data *registrystorage.OperatorData)
	IndicesChangeChan() chan struct{}

	StartValidator(share *ssvtypes.SSVShare) error
	StopValidator(publicKey []byte) error
	LiquidateCluster(owner common.Address, operatorIDs []uint64, toLiquidate []*ssvtypes.SSVShare) error
	ReactivateCluster(owner common.Address, operatorIDs []uint64, toReactivate []*ssvtypes.SSVShare) error
	UpdateFeeRecipient(owner, recipient common.Address) error
}

type nonCommitteeValidator struct {
	*validator.NonCommitteeValidator
	sync.Mutex
}

// controller implements Controller
type controller struct {
	context context.Context

	logger  *zap.Logger
	metrics validatorMetrics

	sharesStorage     registrystorage.Shares
	operatorsStorage  registrystorage.Operators
	recipientsStorage registrystorage.Recipients
	ibftStorageMap    *storage.QBFTStores

	beacon     beaconprotocol.BeaconNode
	keyManager spectypes.KeyManager

	shareEncryptionKeyProvider ShareEncryptionKeyProvider
	operatorData               *registrystorage.OperatorData
	operatorDataMutex          sync.RWMutex

	validatorsMap    *validatorsMap
	validatorOptions *validator.Options

	metadataUpdateQueue    utilsprotocol.Queue
	metadataUpdateInterval time.Duration

	operatorsIDs         *sync.Map
	network              network.P2PNetwork
	forkVersion          forksprotocol.ForkVersion
	messageRouter        *messageRouter
	messageWorker        *worker.Worker
	historySyncBatchSize int

	// nonCommittees is a cache of initialized nonCommitteeValidator instances
	nonCommitteeValidators *ttlcache.Cache[spectypes.MessageID, *nonCommitteeValidator]
	nonCommitteeMutex      sync.Mutex

	recentlyStartedValidators atomic.Int64
	indicesChange             chan struct{}
}

// NewController creates a new validator controller instance
func NewController(logger *zap.Logger, options ControllerOptions) Controller {
	logger.Debug("setting validator controller")

	// lookup in a map that holds all relevant operators
	operatorsIDs := &sync.Map{}

	msgID := forksfactory.NewFork(options.ForkVersion).MsgID()

	workerCfg := &worker.Config{
		Ctx:          options.Context,
		WorkersCount: options.WorkersCount,
		Buffer:       options.QueueBufferSize,
	}

	validatorOptions := &validator.Options{ //TODO add vars
		Network:       options.Network,
		Beacon:        options.Beacon,
		BeaconNetwork: options.BeaconNetwork.BeaconNetwork,
		Storage:       options.StorageMap,
		//Share:   nil,  // set per validator
		Signer: options.KeyManager,
		//Mode: validator.ModeRW // set per validator
		DutyRunners:       nil, // set per validator
		NewDecidedHandler: options.NewDecidedHandler,
		FullNode:          options.FullNode,
		Exporter:          options.Exporter,
		BuilderProposals:  options.BuilderProposals,
		GasLimit:          options.GasLimit,
	}

	// If full node, increase queue size to make enough room
	// for history sync batches to be pushed whole.
	if options.FullNode {
		size := options.HistorySyncBatchSize * 2
		if size > validator.DefaultQueueSize {
			validatorOptions.QueueSize = size
		}
	}

	if options.Metrics == nil {
		options.Metrics = nopMetrics{}
	}

	ctrl := controller{
		logger:                     logger.Named(logging.NameController),
		metrics:                    options.Metrics,
		sharesStorage:              options.RegistryStorage.Shares(),
		operatorsStorage:           options.RegistryStorage,
		recipientsStorage:          options.RegistryStorage,
		ibftStorageMap:             options.StorageMap,
		context:                    options.Context,
		beacon:                     options.Beacon,
		shareEncryptionKeyProvider: options.ShareEncryptionKeyProvider,
		operatorData:               options.OperatorData,
		keyManager:                 options.KeyManager,
		network:                    options.Network,
		forkVersion:                options.ForkVersion,

		validatorsMap:    newValidatorsMap(options.Context, validatorOptions),
		validatorOptions: validatorOptions,

		metadataUpdateQueue:    tasks.NewExecutionQueue(10 * time.Millisecond),
		metadataUpdateInterval: options.MetadataUpdateInterval,

		operatorsIDs: operatorsIDs,

		messageRouter:        newMessageRouter(msgID),
		messageWorker:        worker.NewWorker(logger, workerCfg),
		historySyncBatchSize: options.HistorySyncBatchSize,

		nonCommitteeValidators: ttlcache.New(
			ttlcache.WithTTL[spectypes.MessageID, *nonCommitteeValidator](time.Minute * 13),
		),
		indicesChange: make(chan struct{}),
	}

	// Start automatic expired item deletion in nonCommitteeValidators.
	go ctrl.nonCommitteeValidators.Start()

	return &ctrl
}

// setupNetworkHandlers registers all the required handlers for sync protocols
func (c *controller) setupNetworkHandlers() error {
	syncHandlers := []*p2pprotocol.SyncHandler{
		p2pprotocol.WithHandler(
			p2pprotocol.LastDecidedProtocol,
			handlers.LastDecidedHandler(c.logger, c.ibftStorageMap, c.network),
		),
	}
	if c.validatorOptions.FullNode {
		syncHandlers = append(
			syncHandlers,
			p2pprotocol.WithHandler(
				p2pprotocol.DecidedHistoryProtocol,
				// TODO: extract maxBatch to config
				handlers.HistoryHandler(c.logger, c.ibftStorageMap, c.network, c.historySyncBatchSize),
			),
		)
	}
	c.logger.Debug("setting up network handlers",
		zap.Int("count", len(syncHandlers)),
		zap.Bool("full_node", c.validatorOptions.FullNode),
		zap.Bool("exporter", c.validatorOptions.Exporter),
		zap.Int("queue_size", c.validatorOptions.QueueSize))
	c.network.RegisterHandlers(c.logger, syncHandlers...)
	return nil
}

func (c *controller) GetOperatorShares() []*ssvtypes.SSVShare {
	return c.sharesStorage.List(nil, registrystorage.ByOperatorID(c.GetOperatorData().ID), registrystorage.ByActiveValidator())
}

func (c *controller) GetOperatorData() *registrystorage.OperatorData {
	c.operatorDataMutex.RLock()
	defer c.operatorDataMutex.RUnlock()
	return c.operatorData
}

func (c *controller) SetOperatorData(data *registrystorage.OperatorData) {
	c.operatorDataMutex.Lock()
	defer c.operatorDataMutex.Unlock()
	c.operatorData = data
}

func (c *controller) IndicesChangeChan() chan struct{} {
	return c.indicesChange
}

func (c *controller) GetValidatorStats() (uint64, uint64, uint64, error) {
	allShares := c.sharesStorage.List(nil)
	operatorShares := uint64(0)
	active := uint64(0)
	for _, s := range allShares {
		if ok := s.BelongsToOperator(c.GetOperatorData().ID); ok {
			operatorShares++
		}
		if s.HasBeaconMetadata() && s.BeaconMetadata.IsAttesting() {
			active++
		}
	}
	return uint64(len(allShares)), active, operatorShares, nil
}

func (c *controller) handleRouterMessages() {
	ctx, cancel := context.WithCancel(c.context)
	defer cancel()
	ch := c.messageRouter.GetMessageChan()

	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("router message handler stopped")
			return
		case msg := <-ch:
			// TODO temp solution to prevent getting event msgs from network. need to to add validation in p2p
			if msg.MsgType == message.SSVEventMsgType {
				continue
			}

			pk := msg.GetID().GetPubKey()
			hexPK := hex.EncodeToString(pk)
			if v, ok := c.validatorsMap.GetValidator(hexPK); ok {
				v.HandleMessage(c.logger, &msg)
			} else {
				if msg.MsgType != spectypes.SSVConsensusMsgType {
					continue // not supporting other types
				}
				if !c.messageWorker.TryEnqueue(&msg) { // start to save non committee decided messages only post fork
					c.logger.Warn("Failed to enqueue post consensus message: buffer is full")
				}
			}
		}
	}
}

var nonCommitteeValidatorTTLs = map[spectypes.BeaconRole]phase0.Slot{
	spectypes.BNRoleAttester:                  64,
	spectypes.BNRoleProposer:                  4,
	spectypes.BNRoleAggregator:                4,
	spectypes.BNRoleSyncCommittee:             4,
	spectypes.BNRoleSyncCommitteeContribution: 4,
}

func (c *controller) handleWorkerMessages(msg *spectypes.SSVMessage) error {
	// Get or create a nonCommitteeValidator for this MessageID, and lock it to prevent
	// other handlers from processing
	var ncv *nonCommitteeValidator
	err := func() error {
		c.nonCommitteeMutex.Lock()
		defer c.nonCommitteeMutex.Unlock()

		item := c.nonCommitteeValidators.Get(msg.GetID())
		if item != nil {
			ncv = item.Value()
		} else {
			// Create a new nonCommitteeValidator and cache it.
			share := c.sharesStorage.Get(nil, msg.GetID().GetPubKey())
			if share == nil {
				return errors.Errorf("could not find validator [%s]", hex.EncodeToString(msg.GetID().GetPubKey()))
			}

			opts := *c.validatorOptions
			opts.SSVShare = share
			ncv = &nonCommitteeValidator{
				NonCommitteeValidator: validator.NewNonCommitteeValidator(c.logger, msg.GetID(), opts),
			}

			ttlSlots := nonCommitteeValidatorTTLs[msg.MsgID.GetRoleType()]
			c.nonCommitteeValidators.Set(
				msg.GetID(),
				ncv,
				time.Duration(ttlSlots)*c.beacon.GetBeaconNetwork().SlotDurationSec(),
			)
		}

		ncv.Lock()
		return nil
	}()
	if err != nil {
		return err
	}

	// Process the message.
	defer ncv.Unlock()
	ncv.ProcessMessage(c.logger, msg)

	return nil
}

// StartValidators loads all persisted shares and setup the corresponding validators
func (c *controller) StartValidators() {
	if c.validatorOptions.Exporter {
		c.setupNonCommitteeValidators()
		return
	}

	shares := c.sharesStorage.List(nil, registrystorage.ByOperatorID(c.GetOperatorData().ID), registrystorage.ByNotLiquidated())
	if len(shares) == 0 {
		c.logger.Info("could not find validators")
		return
	}
	c.setupValidators(shares)
}

// setupValidators setup and starts validators from the given shares.
// shares w/o validator's metadata won't start, but the metadata will be fetched and the validator will start afterwards
func (c *controller) setupValidators(shares []*ssvtypes.SSVShare) {
	c.logger.Info("starting validators setup...", zap.Int("shares count", len(shares)))
	var started int
	var errs []error
	var fetchMetadata [][]byte
	for _, validatorShare := range shares {
		isStarted, err := c.onShareStart(validatorShare)
		if err != nil {
			c.logger.Warn("could not start validator", fields.PubKey(validatorShare.ValidatorPubKey), zap.Error(err))
			errs = append(errs, err)
		}
		if !isStarted && err == nil {
			// Fetch metadata, if needed.
			fetchMetadata = append(fetchMetadata, validatorShare.ValidatorPubKey)
		}
		if isStarted {
			started++
		}
	}
	c.logger.Info("setup validators done", zap.Int("map size", c.validatorsMap.Size()),
		zap.Int("failures", len(errs)), zap.Int("missing_metadata", len(fetchMetadata)),
		zap.Int("shares", len(shares)), zap.Int("started", started))

	// Try to fetch metadata once for validators that don't have it.
	if len(fetchMetadata) > 0 {
		start := time.Now()
		err := beaconprotocol.UpdateValidatorsMetadata(c.logger, fetchMetadata, c, c.beacon, c.onMetadataUpdated)
		if err != nil {
			c.logger.Error("failed to update validators metadata",
				zap.Int("shares", len(fetchMetadata)),
				fields.Took(time.Since(start)),
				zap.Error(err))
		} else {
			c.logger.Debug("updated validators metadata",
				zap.Int("shares", len(fetchMetadata)),
				fields.Took(time.Since(start)))
		}
	}
}

// setupNonCommitteeValidators trigger SyncHighestDecided for each validator
// to start consensus flow which would save the highest decided instance
// and sync any gaps (in protocol/v2/qbft/controller/decided.go).
func (c *controller) setupNonCommitteeValidators() {
	// Subscribe to all subnets.
	err := c.network.SubscribeAll(c.logger)
	if err != nil {
		c.logger.Error("failed to subscribe to all subnets", zap.Error(err))
		return
	}

	nonCommitteeShares := c.sharesStorage.List(nil, registrystorage.ByNotLiquidated())
	if len(nonCommitteeShares) == 0 {
		c.logger.Info("could not find non-committee validators")
		return
	}

	pubKeys := make([][]byte, 0, len(nonCommitteeShares))
	for _, validatorShare := range nonCommitteeShares {
		pubKeys = append(pubKeys, validatorShare.ValidatorPubKey)

		opts := *c.validatorOptions
		opts.SSVShare = validatorShare
		allRoles := []spectypes.BeaconRole{
			spectypes.BNRoleAttester,
			spectypes.BNRoleAggregator,
			spectypes.BNRoleProposer,
			spectypes.BNRoleSyncCommittee,
			spectypes.BNRoleSyncCommitteeContribution,
		}
		for _, role := range allRoles {
			messageID := spectypes.NewMsgID(ssvtypes.GetDefaultDomain(), validatorShare.ValidatorPubKey, role)
			err := c.network.SyncHighestDecided(messageID)
			if err != nil {
				c.logger.Error("failed to sync highest decided", zap.Error(err))
			}
		}
	}

	if len(pubKeys) > 0 {
		c.logger.Debug("updating metadata for non-committee validators", zap.Int("count", len(pubKeys)))
		if err := beaconprotocol.UpdateValidatorsMetadata(c.logger, pubKeys, c, c.beacon, c.onMetadataUpdated); err != nil {
			c.logger.Warn("could not update all validators", zap.Error(err))
		}
	}
}

// StartNetworkHandlers init msg worker that handles network messages
func (c *controller) StartNetworkHandlers() {
	// first, set stream handlers
	if err := c.setupNetworkHandlers(); err != nil {
		c.logger.Panic("could not register stream handlers", zap.Error(err))
	}
	c.network.UseMessageRouter(c.messageRouter)
	for i := 0; i < networkRouterConcurrency; i++ {
		go c.handleRouterMessages()
	}
	c.messageWorker.UseHandler(c.handleWorkerMessages)
}

// UpdateValidatorMetadata updates a given validator with metadata (implements ValidatorMetadataStorage)
func (c *controller) UpdateValidatorMetadata(pk string, metadata *beaconprotocol.ValidatorMetadata) error {
	if metadata == nil {
		return errors.New("could not update empty metadata")
	}

	// Save metadata to share storage.
	err := c.sharesStorage.UpdateValidatorMetadata(pk, metadata)
	if err != nil {
		return errors.Wrap(err, "could not update validator metadata")
	}

	// If this validator is not ours, don't start it.
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		return errors.Wrap(err, "could not decode public key")
	}

	share := c.sharesStorage.Get(nil, pkBytes)
	if share == nil {
		return errors.New("share was not found")
	}
	if !share.BelongsToOperator(c.GetOperatorData().ID) {
		return nil
	}

	// Start validator (if not already started).
	if v, found := c.validatorsMap.GetValidator(pk); found {
		v.Share.BeaconMetadata = metadata
		_, err := c.startValidator(v)
		if err != nil {
			c.logger.Warn("could not start validator", zap.Error(err))
		}
	} else {
		c.logger.Info("starting new validator", zap.String("pubKey", pk))

		started, err := c.onShareStart(share)
		if err != nil {
			return errors.Wrap(err, "could not start validator")
		}
		if started {
			c.logger.Debug("started share after metadata update", zap.Bool("started", started))
		}
	}
	return nil
}

// GetValidator returns a validator instance from validatorsMap
func (c *controller) GetValidator(pubKey string) (*validator.Validator, bool) {
	return c.validatorsMap.GetValidator(pubKey)
}

func (c *controller) ExecuteDuty(logger *zap.Logger, duty *spectypes.Duty) {
	// because we're using the same duty for more than 1 duty (e.g. attest + aggregator) there is an error in bls.Deserialize func for cgo pointer to pointer.
	// so we need to copy the pubkey val to avoid pointer
	var pk phase0.BLSPubKey
	copy(pk[:], duty.PubKey[:])

	if v, ok := c.GetValidator(hex.EncodeToString(pk[:])); ok {
		ssvMsg, err := CreateDutyExecuteMsg(duty, pk, types.GetDefaultDomain())
		if err != nil {
			logger.Error("could not create duty execute msg", zap.Error(err))
			return
		}
		dec, err := queue.DecodeSSVMessage(logger, ssvMsg)
		if err != nil {
			logger.Error("could not decode duty execute msg", zap.Error(err))
			return
		}
		if pushed := v.Queues[duty.Type].Q.TryPush(dec); !pushed {
			logger.Warn("dropping ExecuteDuty message because the queue is full")
		}
		// logger.Debug("📬 queue: pushed message", fields.MessageID(dec.MsgID), fields.MessageType(dec.MsgType))
	} else {
		logger.Warn("could not find validator")
	}
}

// CreateDutyExecuteMsg returns ssvMsg with event type of duty execute
func CreateDutyExecuteMsg(duty *spectypes.Duty, pubKey phase0.BLSPubKey, domain spectypes.DomainType) (*spectypes.SSVMessage, error) {
	executeDutyData := types.ExecuteDutyData{Duty: duty}
	edd, err := json.Marshal(executeDutyData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal execute duty data")
	}
	msg := types.EventMsg{
		Type: types.ExecuteDuty,
		Data: edd,
	}
	data, err := msg.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode event msg")
	}
	return &spectypes.SSVMessage{
		MsgType: message.SSVEventMsgType,
		MsgID:   spectypes.NewMsgID(domain, pubKey[:], duty.Type),
		Data:    data,
	}, nil
}

// ActiveValidatorIndices fetches indices of validators who are either attesting or queued and
// whose activation epoch is not greater than the passed epoch. It logs a warning if an error occurs.
func (c *controller) ActiveValidatorIndices(epoch phase0.Epoch) []phase0.ValidatorIndex {
	indices := make([]phase0.ValidatorIndex, 0, len(c.validatorsMap.validatorsMap))
	err := c.validatorsMap.ForEach(func(v *validator.Validator) error {
		// Beacon node throws error when trying to fetch duties for non-existing validators.
		if (v.Share.BeaconMetadata.IsAttesting() || v.Share.BeaconMetadata.Status == v1.ValidatorStatePendingQueued) &&
			v.Share.BeaconMetadata.ActivationEpoch <= epoch {
			indices = append(indices, v.Share.BeaconMetadata.Index)
		}
		return nil
	})
	if err != nil {
		c.logger.Warn("failed to get all validators public keys", zap.Error(err))
	}

	return indices
}

// onMetadataUpdated is called when validator's metadata was updated
func (c *controller) onMetadataUpdated(pk string, meta *beaconprotocol.ValidatorMetadata) {
	if meta == nil {
		return
	}
	logger := c.logger.With(zap.String("pk", pk))

	if v, exist := c.GetValidator(pk); exist {
		// update share object owned by the validator
		// TODO: check if this updates running validators
		if !v.Share.BeaconMetadata.Equals(meta) {
			v.Share.BeaconMetadata.Status = meta.Status
			v.Share.BeaconMetadata.Balance = meta.Balance
			v.Share.BeaconMetadata.ActivationEpoch = meta.ActivationEpoch
			logger.Debug("metadata was updated")
		}
		_, err := c.startValidator(v)
		if err != nil {
			logger.Warn("could not start validator after metadata update",
				zap.String("pk", pk), zap.Error(err), zap.Any("metadata", meta))
		}
		return
	}
}

// onShareRemove is called when a validator was removed
// TODO: think how we can make this function atomic (i.e. failing wouldn't stop the removal of the share)
func (c *controller) onShareRemove(pk string, removeSecret bool) error {
	// remove from validatorsMap
	v := c.validatorsMap.RemoveValidator(pk)

	// stop instance
	if v != nil {
		v.Stop()
	}
	// remove the share secret from key-manager
	if removeSecret {
		if err := c.keyManager.RemoveShare(pk); err != nil {
			return errors.Wrap(err, "could not remove share secret from key manager")
		}
	}

	return nil
}

func (c *controller) onShareStart(share *ssvtypes.SSVShare) (bool, error) {
	if !share.HasBeaconMetadata() { // fetching index and status in case not exist
		c.logger.Warn("skipping validator until it becomes active", fields.PubKey(share.ValidatorPubKey))
		return false, nil
	}

	if err := c.setShareFeeRecipient(share, c.recipientsStorage.GetRecipientData); err != nil {
		return false, errors.Wrap(err, "could not set share fee recipient")
	}

	// Start a committee validator.
	v, err := c.validatorsMap.GetOrCreateValidator(c.logger.Named("validatorsMap"), share)
	if err != nil {
		return false, errors.Wrap(err, "could not get or create validator")
	}
	return c.startValidator(v)
}

func (c *controller) setShareFeeRecipient(share *ssvtypes.SSVShare, getRecipientData GetRecipientDataFunc) error {
	var feeRecipient bellatrix.ExecutionAddress
	data, found, err := getRecipientData(nil, share.OwnerAddress)
	if err != nil {
		return errors.Wrap(err, "could not get recipient data")
	}
	if !found {
		c.logger.Debug("setting fee recipient to owner address",
			fields.Validator(share.ValidatorPubKey), fields.FeeRecipient(share.OwnerAddress.Bytes()))
		copy(feeRecipient[:], share.OwnerAddress.Bytes())
	} else {
		c.logger.Debug("setting fee recipient to storage data",
			fields.Validator(share.ValidatorPubKey), fields.FeeRecipient(data.FeeRecipient[:]))
		feeRecipient = data.FeeRecipient
	}
	share.SetFeeRecipient(feeRecipient)

	return nil
}

// startValidator will start the given validator if applicable
func (c *controller) startValidator(v *validator.Validator) (bool, error) {
	c.reportValidatorStatus(v.Share.ValidatorPubKey, v.Share.BeaconMetadata)
	if v.Share.BeaconMetadata.Index == 0 {
		return false, errors.New("could not start validator: index not found")
	}
	if err := v.Start(c.logger); err != nil {
		c.metrics.ValidatorError(v.Share.ValidatorPubKey)
		return false, errors.Wrap(err, "could not start validator")
	}
	c.recentlyStartedValidators.Add(1)
	return true, nil
}

// UpdateValidatorMetaDataLoop updates metadata of validators in an interval
func (c *controller) UpdateValidatorMetaDataLoop() {
	go c.metadataUpdateQueue.Start()

	lastUpdated := make(map[string]time.Time)
	for {
		time.Sleep(time.Second * 12)

		// Prepare filters.
		filters := []registrystorage.SharesFilter{
			registrystorage.ByNotLiquidated(),

			// Filter out validators that were recently updated.
			func(s *ssvtypes.SSVShare) bool {
				last, ok := lastUpdated[string(s.ValidatorPubKey)]
				return !ok || time.Since(last) > c.metadataUpdateInterval
			},
		}

		// Filter out validators which don't belong to us.
		if !c.validatorOptions.Exporter {
			filters = append(filters, registrystorage.ByOperatorID(c.GetOperatorData().ID))
		}

		// Get the shares to fetch metadata for.
		shares := c.sharesStorage.List(nil, filters...)
		var pks [][]byte
		for _, share := range shares {
			pks = append(pks, share.ValidatorPubKey)
			lastUpdated[string(share.ValidatorPubKey)] = time.Now()
		}

		beaconprotocol.UpdateValidatorsMetadataBatch(c.logger, pks, c.metadataUpdateQueue, c,
			c.beacon, c.onMetadataUpdated, metadataBatchSize)
		started := c.recentlyStartedValidators.Load()
		c.logger.Debug("updated metadata in loop",
			zap.Int("shares", len(shares)),
			zap.Int64("started_validators", started))

		// Notify DutyScheduler of new validators.
		if started > 0 {
			c.recentlyStartedValidators.Store(0)
			c.indicesChange <- struct{}{}
		}
	}
}

// SetupRunners initializes duty runners for the given validator
func SetupRunners(ctx context.Context, logger *zap.Logger, options validator.Options) runner.DutyRunners {
	if options.SSVShare == nil || options.SSVShare.BeaconMetadata == nil {
		logger.Error("missing validator metadata", zap.String("validator", hex.EncodeToString(options.SSVShare.ValidatorPubKey)))
		return runner.DutyRunners{} // TODO need to find better way to fix it
	}

	runnersType := []spectypes.BeaconRole{
		spectypes.BNRoleAttester,
		spectypes.BNRoleProposer,
		spectypes.BNRoleAggregator,
		spectypes.BNRoleSyncCommittee,
		spectypes.BNRoleSyncCommitteeContribution,
		spectypes.BNRoleValidatorRegistration,
	}

	domainType := ssvtypes.GetDefaultDomain()
	buildController := func(role spectypes.BeaconRole, valueCheckF specqbft.ProposedValueCheckF) *qbftcontroller.Controller {
		config := &qbft.Config{
			Signer:      options.Signer,
			SigningPK:   options.SSVShare.ValidatorPubKey, // TODO right val?
			Domain:      domainType,
			ValueCheckF: nil, // sets per role type
			ProposerF: func(state *specqbft.State, round specqbft.Round) spectypes.OperatorID {
				leader := specqbft.RoundRobinProposer(state, round)
				//logger.Debug("leader", zap.Int("operator_id", int(leader)))
				return leader
			},
			Storage: options.Storage.Get(role),
			Network: options.Network,
			Timer:   roundtimer.New(ctx, nil),
		}
		config.ValueCheckF = valueCheckF

		identifier := spectypes.NewMsgID(ssvtypes.GetDefaultDomain(), options.SSVShare.Share.ValidatorPubKey, role)
		qbftCtrl := qbftcontroller.NewController(identifier[:], &options.SSVShare.Share, domainType, config, options.FullNode)
		qbftCtrl.NewDecidedHandler = options.NewDecidedHandler
		return qbftCtrl
	}

	runners := runner.DutyRunners{}
	for _, role := range runnersType {
		switch role {
		case spectypes.BNRoleAttester:
			valCheck := specssv.AttesterValueCheckF(options.Signer, options.BeaconNetwork, options.SSVShare.Share.ValidatorPubKey, options.SSVShare.BeaconMetadata.Index, options.SSVShare.SharePubKey)
			qbftCtrl := buildController(spectypes.BNRoleAttester, valCheck)
			runners[role] = runner.NewAttesterRunnner(options.BeaconNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer, valCheck, 0)
		case spectypes.BNRoleProposer:
			proposedValueCheck := specssv.ProposerValueCheckF(options.Signer, options.BeaconNetwork, options.SSVShare.Share.ValidatorPubKey, options.SSVShare.BeaconMetadata.Index, options.SSVShare.SharePubKey, options.BuilderProposals)
			qbftCtrl := buildController(spectypes.BNRoleProposer, proposedValueCheck)
			runners[role] = runner.NewProposerRunner(options.BeaconNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer, proposedValueCheck, 0)
			runners[role].(*runner.ProposerRunner).ProducesBlindedBlocks = options.BuilderProposals // apply blinded block flag
		case spectypes.BNRoleAggregator:
			aggregatorValueCheckF := specssv.AggregatorValueCheckF(options.Signer, options.BeaconNetwork, options.SSVShare.Share.ValidatorPubKey, options.SSVShare.BeaconMetadata.Index)
			qbftCtrl := buildController(spectypes.BNRoleAggregator, aggregatorValueCheckF)
			runners[role] = runner.NewAggregatorRunner(options.BeaconNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer, aggregatorValueCheckF, 0)
		case spectypes.BNRoleSyncCommittee:
			syncCommitteeValueCheckF := specssv.SyncCommitteeValueCheckF(options.Signer, options.BeaconNetwork, options.SSVShare.ValidatorPubKey, options.SSVShare.BeaconMetadata.Index)
			qbftCtrl := buildController(spectypes.BNRoleSyncCommittee, syncCommitteeValueCheckF)
			runners[role] = runner.NewSyncCommitteeRunner(options.BeaconNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer, syncCommitteeValueCheckF, 0)
		case spectypes.BNRoleSyncCommitteeContribution:
			syncCommitteeContributionValueCheckF := specssv.SyncCommitteeContributionValueCheckF(options.Signer, options.BeaconNetwork, options.SSVShare.Share.ValidatorPubKey, options.SSVShare.BeaconMetadata.Index)
			qbftCtrl := buildController(spectypes.BNRoleSyncCommitteeContribution, syncCommitteeContributionValueCheckF)
			runners[role] = runner.NewSyncCommitteeAggregatorRunner(options.BeaconNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer, syncCommitteeContributionValueCheckF, 0)
		case spectypes.BNRoleValidatorRegistration:
			qbftCtrl := buildController(spectypes.BNRoleValidatorRegistration, nil)
			runners[role] = runner.NewValidatorRegistrationRunner(spectypes.PraterNetwork, &options.SSVShare.Share, qbftCtrl, options.Beacon, options.Network, options.Signer)
		}
	}
	return runners
}
