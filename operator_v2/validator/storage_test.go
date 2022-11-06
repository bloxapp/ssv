package validator

import (
	"encoding/hex"
	"testing"

	spectypes "github.com/bloxapp/ssv-spec/types"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/protocol/v2/types"
	"github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/threshold"
)

func TestValidatorSerializer(t *testing.T) {
	threshold.Init()

	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	const keysCount = 4

	splitKeys, err := threshold.Create(sk.Serialize(), keysCount-1, keysCount)
	require.NoError(t, err)

	validatorShare, _ := generateRandomValidatorShare(splitKeys)
	b, err := validatorShare.Encode()
	require.NoError(t, err)

	obj := basedb.Obj{
		Key:   validatorShare.ValidatorPubKey,
		Value: b,
	}
	v := &spectypes.Share{}
	require.NoError(t, v.Decode(obj.Value))
	require.NotNil(t, v.ValidatorPubKey)
	require.Equal(t, hex.EncodeToString(v.ValidatorPubKey), hex.EncodeToString(validatorShare.ValidatorPubKey))
	require.NotNil(t, v.Committee)
	require.NotNil(t, v.OperatorID)
}

func TestSaveAndGetValidatorStorage(t *testing.T) {
	options := basedb.Options{
		Type:   "badger-memory",
		Logger: zap.L(),
		Path:   "",
	}

	db, err := storage.GetStorageFactory(options)
	require.NoError(t, err)
	defer db.Close()

	collection := NewCollection(CollectionOptions{
		DB:     db,
		Logger: options.Logger,
	})

	threshold.Init()
	const keysCount = 4

	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	splitKeys, err := threshold.Create(sk.Serialize(), keysCount-1, keysCount)
	require.NoError(t, err)

	validatorShare, _ := generateRandomValidatorShare(splitKeys)
	require.NoError(t, collection.SaveValidatorShare(validatorShare))

	validatorShare2, _ := generateRandomValidatorShare(splitKeys)
	require.NoError(t, collection.SaveValidatorShare(validatorShare2))

	validatorShareByKey, found, err := collection.GetValidatorShare(validatorShare.ValidatorPubKey)
	require.True(t, found)
	require.NoError(t, err)
	require.EqualValues(t, hex.EncodeToString(validatorShareByKey.ValidatorPubKey), hex.EncodeToString(validatorShare.ValidatorPubKey))

	validators, err := collection.GetAllValidatorShares()
	require.NoError(t, err)
	require.EqualValues(t, 2, len(validators))

	require.NoError(t, collection.DeleteValidatorShare(validatorShare.ValidatorPubKey))
	_, found, err = collection.GetValidatorShare(validatorShare.ValidatorPubKey)
	require.NoError(t, err)
	require.False(t, found)
}

func generateRandomValidatorShare(splitKeys map[uint64]*bls.SecretKey) (*spectypes.Share, *bls.SecretKey) {
	threshold.Init()
	sk := bls.SecretKey{}
	sk.SetByCSPRNG()

	ibftCommittee := []*spectypes.Operator{
		{
			OperatorID: 1,
			PubKey:     splitKeys[1].Serialize(),
		},
		{
			OperatorID: 2,
			PubKey:     splitKeys[2].Serialize(),
		},
		{
			OperatorID: 3,
			PubKey:     splitKeys[3].Serialize(),
		},
		{
			OperatorID: 4,
			PubKey:     splitKeys[4].Serialize(),
		},
	}

	return &spectypes.Share{
		OperatorID:      1,
		ValidatorPubKey: sk.GetPublicKey().Serialize(),
		Committee:       ibftCommittee,
	}, &sk
}

func generateRandomValidatorMetadata() *types.ShareMetadata {
	return &types.ShareMetadata{
		OwnerAddress: "0xFeedB14D8b2C76FdF808C29818b06b830E8C2c0e",
	}
}
