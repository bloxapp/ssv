package storage

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	registry "github.com/bloxapp/ssv/protocol/v2/blockchain/eth1"
	registrystorage "github.com/bloxapp/ssv/registry/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/rsaencryption"
)

var (
	storagePrefix         = []byte("operator/")
	lastProcessedBlockKey = []byte("syncOffset") // TODO: temporarily left as syncOffset for compatibility, consider renaming and adding a migration for that
)

// Storage represents the interface for ssv node storage
type Storage interface {
	// TODO: de-anonymize the sub-storages, like Shares() below

	Begin() basedb.Txn
	BeginRead() basedb.ReadTxn

	SaveLastProcessedBlock(rw basedb.ReadWriter, offset *big.Int) error
	GetLastProcessedBlock(r basedb.Reader) (*big.Int, bool, error)

	registry.RegistryStore

	registrystorage.Operators
	registrystorage.Recipients
	Shares() registrystorage.Shares

	GetPrivateKey() (*rsa.PrivateKey, bool, error)
	SetupPrivateKey(operatorKeyBase64 string, generateIfNone bool) ([]byte, error)
}

type storage struct {
	logger *zap.Logger
	db     basedb.Database

	operatorStore  registrystorage.Operators
	recipientStore registrystorage.Recipients
	shareStore     registrystorage.Shares
}

// NewNodeStorage creates a new instance of Storage
func NewNodeStorage(logger *zap.Logger, db basedb.Database) (Storage, error) {
	stg := &storage{
		logger:         logger,
		db:             db,
		operatorStore:  registrystorage.NewOperatorsStorage(logger, db, storagePrefix),
		recipientStore: registrystorage.NewRecipientsStorage(logger, db, storagePrefix),
	}
	var err error
	stg.shareStore, err = registrystorage.NewSharesStorage(logger, db, storagePrefix)
	if err != nil {
		return nil, err
	}
	return stg, nil
}

func (s *storage) Begin() basedb.Txn {
	return s.db.Begin()
}

func (s *storage) BeginRead() basedb.ReadTxn {
	return s.db.BeginRead()
}

func (s *storage) Shares() registrystorage.Shares {
	return s.shareStore
}

func (s *storage) GetOperatorDataByPubKey(r basedb.Reader, operatorPubKey []byte) (*registrystorage.OperatorData, bool, error) {
	return s.operatorStore.GetOperatorDataByPubKey(r, operatorPubKey)
}

func (s *storage) GetOperatorData(r basedb.Reader, id spectypes.OperatorID) (*registrystorage.OperatorData, bool, error) {
	return s.operatorStore.GetOperatorData(r, id)
}

func (s *storage) SaveOperatorData(rw basedb.ReadWriter, operatorData *registrystorage.OperatorData) (bool, error) {
	return s.operatorStore.SaveOperatorData(rw, operatorData)
}

func (s *storage) DeleteOperatorData(rw basedb.ReadWriter, id spectypes.OperatorID) error {
	return s.operatorStore.DeleteOperatorData(rw, id)
}

func (s *storage) ListOperators(r basedb.Reader, from uint64, to uint64) ([]registrystorage.OperatorData, error) {
	return s.operatorStore.ListOperators(r, from, to)
}

func (s *storage) GetOperatorsPrefix() []byte {
	return s.operatorStore.GetOperatorsPrefix()
}

func (s *storage) GetRecipientData(r basedb.Reader, owner common.Address) (*registrystorage.RecipientData, bool, error) {
	return s.recipientStore.GetRecipientData(r, owner)
}

func (s *storage) GetRecipientDataMany(r basedb.Reader, owners []common.Address) (map[common.Address]bellatrix.ExecutionAddress, error) {
	return s.recipientStore.GetRecipientDataMany(r, owners)
}

func (s *storage) SaveRecipientData(rw basedb.ReadWriter, recipientData *registrystorage.RecipientData) (*registrystorage.RecipientData, error) {
	return s.recipientStore.SaveRecipientData(rw, recipientData)
}

func (s *storage) DeleteRecipientData(rw basedb.ReadWriter, owner common.Address) error {
	return s.recipientStore.DeleteRecipientData(rw, owner)
}

func (s *storage) GetNextNonce(r basedb.Reader, owner common.Address) (registrystorage.Nonce, error) {
	return s.recipientStore.GetNextNonce(r, owner)
}

func (s *storage) BumpNonce(rw basedb.ReadWriter, owner common.Address) error {
	return s.recipientStore.BumpNonce(rw, owner)
}

func (s *storage) GetRecipientsPrefix() []byte {
	return s.recipientStore.GetRecipientsPrefix()
}

func (s *storage) CleanRegistryData() error {
	err := s.cleanLastProcessedBlock()
	if err != nil {
		return errors.Wrap(err, "could not clean sync offset")
	}

	err = s.cleanOperators()
	if err != nil {
		return errors.Wrap(err, "could not clean operators")
	}

	err = s.cleanRecipients()
	if err != nil {
		return errors.Wrap(err, "could not clean recipients")
	}

	return nil
}

// TODO: review what's not needed anymore and delete

func (s *storage) SaveLastProcessedBlock(rw basedb.ReadWriter, offset *big.Int) error {
	return s.db.Using(rw).Set(storagePrefix, lastProcessedBlockKey, offset.Bytes())
}

func (s *storage) cleanLastProcessedBlock() error {
	return s.db.RemoveAllByCollection(append(storagePrefix, lastProcessedBlockKey...))
}

func (s *storage) cleanOperators() error {
	operatorsPrefix := s.GetOperatorsPrefix()
	return s.db.RemoveAllByCollection(append(storagePrefix, operatorsPrefix...))
}

func (s *storage) cleanRecipients() error {
	recipientsPrefix := s.GetRecipientsPrefix()
	return s.db.RemoveAllByCollection(append(storagePrefix, recipientsPrefix...))
}

// GetLastProcessedBlock returns the last processed block.
func (s *storage) GetLastProcessedBlock(r basedb.Reader) (*big.Int, bool, error) {
	obj, found, err := s.db.UsingReader(r).Get(storagePrefix, lastProcessedBlockKey)
	if !found {
		return nil, found, nil
	}
	if err != nil {
		return nil, found, err
	}

	offset := new(big.Int).SetBytes(obj.Value)
	return offset, found, nil
}

// GetPrivateKey return rsa private key
func (s *storage) GetPrivateKey() (*rsa.PrivateKey, bool, error) {
	obj, found, err := s.db.Get(storagePrefix, []byte("private-key"))
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, found, nil
	}
	sk, err := rsaencryption.ConvertPemToPrivateKey(string(obj.Value))
	if err != nil {
		return nil, false, err
	}
	return sk, found, nil
}

// SetupPrivateKey setup operator private key at the init of the node and set OperatorPublicKey config
func (s *storage) SetupPrivateKey(operatorKeyBase64 string, generateIfNone bool) ([]byte, error) {
	operatorKeyByte, err := base64.StdEncoding.DecodeString(operatorKeyBase64)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode base64")
	}

	if err := s.validateKey(generateIfNone, string(operatorKeyByte)); err != nil {
		return nil, err
	}

	sk, found, err := s.GetPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get operator private key")
	}
	if !found {
		return nil, errors.New("failed to find operator private key")
	}

	operatorPublicKey, err := rsaencryption.ExtractPublicKey(sk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract operator public key")
	}

	//TODO change the log to generated/loaded private key to indicate better on the action
	s.logger.Info("successfully setup operator keys", zap.String("pubkey", operatorPublicKey))
	return []byte(operatorPublicKey), nil
}

// validateKey validate provided and exist key. save if needed.
func (s *storage) validateKey(generateIfNone bool, operatorKey string) error {
	// check if passed new key. if so, save new key (force to always save key when provided)
	if operatorKey != "" {
		return s.savePrivateKey(operatorKey)
	}
	// new key not provided, check if key exist
	_, found, err := s.GetPrivateKey()
	if err != nil {
		return err
	}
	// if no, check if need to generate. if no, return error
	if !found {
		if !generateIfNone {
			return errors.New("key not exist or provided")
		}
		_, skByte, err := rsaencryption.GenerateKeys()
		if err != nil {
			return errors.WithMessage(err, "failed to generate new key")
		}
		return s.savePrivateKey(string(skByte))
	}

	// key exist in storage.
	return nil
}

// SavePrivateKey save operator private key
func (s *storage) savePrivateKey(operatorKey string) error {
	if err := s.db.Set(storagePrefix, []byte("private-key"), []byte(operatorKey)); err != nil {
		return err
	}
	return nil
}

func (s *storage) UpdateValidatorMetadata(pk string, metadata *beacon.ValidatorMetadata) error {
	return s.shareStore.UpdateValidatorMetadata(pk, metadata)
}
