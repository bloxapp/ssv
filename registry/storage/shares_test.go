package storage

import (
	"bytes"
	"encoding/hex"
	"sort"
	"strconv"
	"testing"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/networkconfig"
	beaconprotocol "github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	ssvtypes "github.com/bloxapp/ssv/protocol/v2/types"
	ssvstorage "github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/threshold"
)

func TestValidatorSerializer(t *testing.T) {
	threshold.Init()

	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	const keysCount = 13

	splitKeys, err := threshold.Create(sk.Serialize(), keysCount-1, keysCount)
	require.NoError(t, err)

	validatorShare, _ := generateRandomValidatorShare(splitKeys)
	b, err := validatorShare.Encode()
	require.NoError(t, err)

	obj := basedb.Obj{
		Key:   validatorShare.ValidatorPubKey,
		Value: b,
	}
	v1 := &ssvtypes.SSVShare{}
	require.NoError(t, v1.Decode(obj.Value))
	require.NotNil(t, v1.ValidatorPubKey)
	require.Equal(t, hex.EncodeToString(v1.ValidatorPubKey), hex.EncodeToString(validatorShare.ValidatorPubKey))
	require.NotNil(t, v1.Committee)
	require.NotNil(t, v1.OperatorID)
	require.Equal(t, v1.BeaconMetadata, validatorShare.BeaconMetadata)
	require.Equal(t, v1.OwnerAddress, validatorShare.OwnerAddress)
	require.Equal(t, v1.Liquidated, validatorShare.Liquidated)

	tooBigEncodedShare := bytes.Repeat(obj.Value, 20)
	require.ErrorContains(t, v1.Decode(tooBigEncodedShare),
		"share size is too big, got "+strconv.Itoa(len(tooBigEncodedShare))+", max allowed "+strconv.Itoa(ssvtypes.MaxAllowedShareSize))
}

func TestMaxPossibleShareSize(t *testing.T) {
	s, err := generateMaxPossibleShare()
	require.NoError(t, err)

	b, err := s.Encode()
	require.NoError(t, err)

	require.Equal(t, ssvtypes.MaxPossibleShareSize, len(b))
}

func TestSaveAndGetValidatorStorage(t *testing.T) {
	logger := logging.TestLogger(t)
	shareStorage, done := newShareStorageForTest(logger)
	require.NotNil(t, shareStorage)
	defer done()

	threshold.Init()
	const keysCount = 4

	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	splitKeys, err := threshold.Create(sk.Serialize(), keysCount-1, keysCount)
	require.NoError(t, err)

	validatorShare, _ := generateRandomValidatorShare(splitKeys)
	require.NoError(t, shareStorage.Save(nil, validatorShare))

	validatorShare2, _ := generateRandomValidatorShare(splitKeys)
	require.NoError(t, shareStorage.Save(nil, validatorShare2))

	validatorShareByKey := shareStorage.Get(nil, validatorShare.ValidatorPubKey)
	require.NotNil(t, validatorShareByKey)
	require.NoError(t, err)
	require.EqualValues(t, hex.EncodeToString(validatorShareByKey.ValidatorPubKey), hex.EncodeToString(validatorShare.ValidatorPubKey))

	validators := shareStorage.List(nil)
	require.NoError(t, err)
	require.EqualValues(t, 2, len(validators))

	require.NoError(t, shareStorage.Delete(nil, validatorShare.ValidatorPubKey))
	share := shareStorage.Get(nil, validatorShare.ValidatorPubKey)
	require.NoError(t, err)
	require.Nil(t, share)
}

func generateRandomValidatorShare(splitKeys map[uint64]*bls.SecretKey) (*ssvtypes.SSVShare, *bls.SecretKey) {
	threshold.Init()

	sk1 := bls.SecretKey{}
	sk1.SetByCSPRNG()

	sk2 := bls.SecretKey{}
	sk2.SetByCSPRNG()

	var ibftCommittee []*spectypes.Operator
	for operatorID, sk := range splitKeys {
		ibftCommittee = append(ibftCommittee, &spectypes.Operator{
			OperatorID: operatorID,
			PubKey:     sk.Serialize(),
		})
	}
	sort.Slice(ibftCommittee, func(i, j int) bool {
		return ibftCommittee[i].OperatorID < ibftCommittee[j].OperatorID
	})

	quorum, partialQuorum := ssvtypes.ComputeQuorumAndPartialQuorum(len(splitKeys))

	return &ssvtypes.SSVShare{
		Share: spectypes.Share{
			OperatorID:          1,
			ValidatorPubKey:     sk1.GetPublicKey().Serialize(),
			SharePubKey:         sk2.GetPublicKey().Serialize(),
			Committee:           ibftCommittee,
			Quorum:              quorum,
			PartialQuorum:       partialQuorum,
			DomainType:          networkconfig.TestNetwork.Domain,
			FeeRecipientAddress: common.HexToAddress("0xFeedB14D8b2C76FdF808C29818b06b830E8C2c0e"),
			Graffiti:            bytes.Repeat([]byte{0x01}, 32),
		},
		Metadata: ssvtypes.Metadata{
			BeaconMetadata: &beaconprotocol.ValidatorMetadata{
				Balance:         1,
				Status:          2,
				Index:           3,
				ActivationEpoch: 4,
			},
			OwnerAddress: common.HexToAddress("0xFeedB14D8b2C76FdF808C29818b06b830E8C2c0e"),
			Liquidated:   true,
		},
	}, &sk1
}

func generateMaxPossibleShare() (*ssvtypes.SSVShare, error) {
	threshold.Init()

	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	const keysCount = 13

	splitKeys, err := threshold.Create(sk.Serialize(), keysCount-1, keysCount)
	if err != nil {
		return nil, err
	}

	validatorShare, _ := generateRandomValidatorShare(splitKeys)
	return validatorShare, nil
}

func newShareStorageForTest(logger *zap.Logger) (Shares, func()) {
	db, err := ssvstorage.GetStorageFactory(logger, basedb.Options{
		Type: "badger-memory",
		Path: "",
	})
	if err != nil {
		return nil, func() {}
	}
	s, err := NewSharesStorage(logger, db, []byte("test"))
	if err != nil {
		return nil, func() {}
	}
	return s, func() {
		db.Close()
	}
}
