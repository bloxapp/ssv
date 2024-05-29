package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	spectypes "github.com/ssvlabs/ssv-spec/types"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	beaconprotocol "github.com/ssvlabs/ssv/protocol/v2/blockchain/beacon"
	"github.com/ssvlabs/ssv/protocol/v2/types"
	"github.com/ssvlabs/ssv/storage/basedb"
)

var sharesPrefix = []byte("shares")

// SharesFilter is a function that filters shares.
type SharesFilter func(*types.SSVShare) bool

// SharesListFunc is a function that returns a filtered list of shares.
type SharesListFunc = func(filters ...SharesFilter) []*types.SSVShare

// Shares is the interface for managing shares.
type Shares interface {
	// Get returns the share for the given public key, or nil if not found.
	Get(txn basedb.Reader, pubKey []byte) *types.SSVShare

	// List returns a list of shares, filtered by the given filters (if any).
	List(txn basedb.Reader, filters ...SharesFilter) []*types.SSVShare

	// Save saves the given shares.
	Save(txn basedb.ReadWriter, shares ...*types.SSVShare) error

	// Delete deletes the share for the given public key.
	Delete(txn basedb.ReadWriter, pubKey []byte) error

	// Drop deletes all shares.
	Drop() error

	// UpdateValidatorMetadata updates validator metadata.
	UpdateValidatorMetadata(pk string, metadata *beaconprotocol.ValidatorMetadata) error
}

type sharesStorage struct {
	logger    *zap.Logger
	db        basedb.Database
	prefix    []byte
	shares    map[string]*types.SSVShare
	operators map[uint64][]byte
	mu        sync.RWMutex
}
type storageOperator struct {
	OperatorID spectypes.OperatorID
	PubKey     []byte `ssz-size:"48"`
}

// Share represents a storage share.
// The better name of the struct is storageShare,
// but we keep the name Share to avoid conflicts with gob encoding.
type Share struct {
	OperatorID            spectypes.OperatorID
	ValidatorPubKey       spectypes.ValidatorPK `ssz-size:"48"`
	SharePubKey           []byte                `ssz-size:"48"`
	Committee             []*storageOperator    `ssz-max:"13"`
	Quorum, PartialQuorum uint64
	DomainType            spectypes.DomainType `ssz-size:"4"`
	FeeRecipientAddress   [20]byte             `ssz-size:"20"`
	Graffiti              []byte               `ssz-size:"32"`
}

type storageShare struct {
	Share
	types.Metadata
}

// Encode encodes Share using gob.
func (s *storageShare) Encode() ([]byte, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(s); err != nil {
		return nil, fmt.Errorf("encode storageShare: %w", err)
	}

	return b.Bytes(), nil
}

// Decode decodes Share using gob.
func (s *storageShare) Decode(data []byte) error {
	if len(data) > types.MaxAllowedShareSize {
		return fmt.Errorf("share size is too big, got %v, max allowed %v", len(data), types.MaxAllowedShareSize)
	}

	d := gob.NewDecoder(bytes.NewReader(data))
	if err := d.Decode(s); err != nil {
		return fmt.Errorf("decode storageShare: %w", err)
	}
	s.Quorum, s.PartialQuorum = types.ComputeQuorumAndPartialQuorum(len(s.Committee))
	return nil
}

func NewSharesStorage(logger *zap.Logger, db basedb.Database, prefix []byte) (Shares, error) {
	storage := &sharesStorage{
		logger:    logger,
		shares:    make(map[string]*types.SSVShare),
		operators: make(map[uint64][]byte),
		db:        db,
		prefix:    prefix,
	}

	if err := storage.loadOperators(); err != nil {
		return nil, err
	}

	if err := storage.load(); err != nil {
		return nil, err
	}
	return storage, nil
}

// loadOperators reads all operators from db.
func (s *sharesStorage) loadOperators() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	opStorage := NewOperatorsStorage(s.logger, s.db, s.prefix)
	operators, err := opStorage.ListOperators(nil, 0, 0)
	if err != nil {
		return err
	}

	for _, op := range operators {
		s.operators[op.ID] = op.PublicKey
	}

	return nil
}

// load reads all shares from db.
func (s *sharesStorage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.GetAll(append(s.prefix, sharesPrefix...), func(i int, obj basedb.Obj) error {
		val := &storageShare{}
		if err := val.Decode(obj.Value); err != nil {
			return fmt.Errorf("failed to deserialize share: %w", err)
		}

		share, err := s.storageShareToSpecShare(val)
		if err != nil {
			return fmt.Errorf("failed to convert storage share to spec share: %w", err)
		}

		s.shares[hex.EncodeToString(val.ValidatorPubKey)] = share
		return nil
	})
}

func (s *sharesStorage) Get(_ basedb.Reader, pubKey []byte) *types.SSVShare {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.shares[hex.EncodeToString(pubKey)]
}

func (s *sharesStorage) List(_ basedb.Reader, filters ...SharesFilter) []*types.SSVShare {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(filters) == 0 {
		return maps.Values(s.shares)
	}

	var shares []*types.SSVShare
Shares:
	for _, share := range s.shares {
		for _, filter := range filters {
			if !filter(share) {
				continue Shares
			}
		}
		shares = append(shares, share)
	}
	return shares
}

func (s *sharesStorage) Save(rw basedb.ReadWriter, shares ...*types.SSVShare) error {
	if len(shares) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Using(rw).SetMany(s.prefix, len(shares), func(i int) (basedb.Obj, error) {
		share := specShareToStorageShare(shares[i])
		value, err := share.Encode()
		if err != nil {
			return basedb.Obj{}, fmt.Errorf("failed to serialize share: %w", err)
		}
		return basedb.Obj{Key: s.storageKey(share.ValidatorPubKey), Value: value}, nil
	})
	if err != nil {
		return err
	}

	for _, share := range shares {
		key := hex.EncodeToString(share.ValidatorPubKey)
		s.shares[key] = share
	}
	return nil
}

func specShareToStorageShare(share *types.SSVShare) *storageShare {
	committee := make([]*storageOperator, len(share.Committee))
	for i, c := range share.Committee {
		committee[i] = &storageOperator{
			OperatorID: c.OperatorID,
			PubKey:     c.SharePubKey,
		}
	}
	stShare := &storageShare{
		Share: Share{
			OperatorID:          share.OperatorID,
			ValidatorPubKey:     share.ValidatorPubKey,
			SharePubKey:         share.SharePubKey,
			Committee:           committee,
			Quorum:              share.Quorum,
			PartialQuorum:       share.PartialQuorum,
			DomainType:          share.DomainType,
			FeeRecipientAddress: share.FeeRecipientAddress,
			Graffiti:            share.Graffiti,
		},
		Metadata: share.Metadata,
	}

	return stShare
}

func (s *sharesStorage) storageShareToSpecShare(share *storageShare) (*types.SSVShare, error) {
	committee := make([]*spectypes.Operator, len(share.Committee))
	for i, c := range share.Committee {
		opPubKey, ok := s.operators[c.OperatorID]
		if !ok {
			return nil, fmt.Errorf("operator not found: %d", c.OperatorID)
		}

		committee[i] = &spectypes.Operator{
			OperatorID:        c.OperatorID,
			SharePubKey:       c.PubKey,
			SSVOperatorPubKey: opPubKey,
		}
	}
	specShare := &types.SSVShare{
		Share: spectypes.Share{
			OperatorID:          share.OperatorID,
			ValidatorPubKey:     share.ValidatorPubKey,
			SharePubKey:         share.SharePubKey,
			Committee:           committee,
			Quorum:              share.Quorum,
			PartialQuorum:       share.PartialQuorum,
			DomainType:          share.DomainType, // TODO: revert
			FeeRecipientAddress: share.FeeRecipientAddress,
			Graffiti:            share.Graffiti,
		},
		Metadata: share.Metadata,
	}

	return specShare, nil
}

func (s *sharesStorage) Delete(rw basedb.ReadWriter, pubKey []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.Using(rw).Delete(s.prefix, s.storageKey(pubKey))
	if err != nil {
		return err
	}

	delete(s.shares, hex.EncodeToString(pubKey))
	return nil
}

// UpdateValidatorMetadata updates the metadata of the given validator
func (s *sharesStorage) UpdateValidatorMetadata(pk string, metadata *beaconprotocol.ValidatorMetadata) error {
	key, err := hex.DecodeString(pk)
	if err != nil {
		return err
	}
	share := s.Get(nil, key)
	if share == nil {
		return nil
	}

	share.BeaconMetadata = metadata
	return s.Save(nil, share)
}

// Drop deletes all shares.
func (s *sharesStorage) Drop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.db.DropPrefix(bytes.Join(
		[][]byte{s.prefix, sharesPrefix, []byte("/")},
		nil,
	))
	if err != nil {
		return err
	}

	s.shares = make(map[string]*types.SSVShare)
	return nil
}

// storageKey builds share key using sharesPrefix & validator public key, e.g. "shares/0x00..01"
func (s *sharesStorage) storageKey(pk []byte) []byte {
	return bytes.Join([][]byte{sharesPrefix, pk}, []byte("/"))
}

// ByOperatorID filters by operator ID.
func ByOperatorID(operatorID spectypes.OperatorID) SharesFilter {
	return func(share *types.SSVShare) bool {
		return share.BelongsToOperator(operatorID)
	}
}

// ByNotLiquidated filters for not liquidated.
func ByNotLiquidated() SharesFilter {
	return func(share *types.SSVShare) bool {
		return !share.Liquidated
	}
}

// ByActiveValidator filters for active validators.
func ByActiveValidator() SharesFilter {
	return func(share *types.SSVShare) bool {
		return share.HasBeaconMetadata()
	}
}

// ByAttesting filters for attesting validators.
func ByAttesting(epoch phase0.Epoch) SharesFilter {
	return func(share *types.SSVShare) bool {
		return share.IsAttesting(epoch)
	}
}

// ByClusterID filters by cluster id.
func ByClusterID(clusterID []byte) SharesFilter {
	return func(share *types.SSVShare) bool {
		var operatorIDs []uint64
		for _, op := range share.Committee {
			operatorIDs = append(operatorIDs, op.OperatorID)
		}

		shareClusterID := types.ComputeClusterIDHash(share.OwnerAddress, operatorIDs)
		return bytes.Equal(shareClusterID, clusterID)
	}
}
