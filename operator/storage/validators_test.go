package storage

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	ssvstorage "github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/bloxapp/ssv/utils/rsaencryption"
)

func TestStorage_SaveAndGetValidatorInformation(t *testing.T) {
	s, done := newStorageForTest()
	require.NotNil(t, s)
	defer done()

	validatorInfo := ValidatorInformation{
		PublicKey: "kds6E6tCimycIOcQRIjLaWGr6rYOVs9LoZnu07X2587WcOywZslwTcL6kxM3kjgc",
		Operators: []OperatorNodeLink{
			{
				ID:        1,
				PublicKey: hex.EncodeToString([]byte{2, 2, 2, 2}),
			},
			{
				ID:        2,
				PublicKey: hex.EncodeToString([]byte{2, 2, 2, 2}),
			},
			{
				ID:        3,
				PublicKey: hex.EncodeToString([]byte{3, 3, 3, 3}),
			},
			{
				ID:        4,
				PublicKey: hex.EncodeToString([]byte{4, 4, 4, 4}),
			},
		},
	}

	t.Run("get non-existing validator", func(t *testing.T) {
		nonExistingOperator, found, _ := s.GetValidatorInformation("dummyPK")
		require.Nil(t, nonExistingOperator)
		require.False(t, found)
	})

	t.Run("create and get validator", func(t *testing.T) {
		err := s.SaveValidatorInformation(&validatorInfo)
		require.NoError(t, err)
		validatorInfoFromDB, _, err := s.GetValidatorInformation(validatorInfo.PublicKey)
		require.NoError(t, err)
		require.Equal(t, "kds6E6tCimycIOcQRIjLaWGr6rYOVs9LoZnu07X2587WcOywZslwTcL6kxM3kjgc",
			validatorInfoFromDB.PublicKey)
		require.Equal(t, int64(0), validatorInfoFromDB.Index)
		require.Equal(t, 4, len(validatorInfoFromDB.Operators))
	})

	t.Run("create existing validator", func(t *testing.T) {
		vi := ValidatorInformation{
			PublicKey: "82e9b36feb8147d3f82c1a03ba246d4a63ac1ce0b1dabbb6991940a06401ab46fb4afbf971a3c145fdad2d4bddd30e12",
			Operators: validatorInfo.Operators[:],
		}
		err := s.SaveValidatorInformation(&vi)
		require.NoError(t, err)
		viDup := ValidatorInformation{
			PublicKey: vi.PublicKey,
			Operators: validatorInfo.Operators[1:],
		}
		err = s.SaveValidatorInformation(&viDup)
		require.NoError(t, err)
		require.Equal(t, viDup.Index, vi.Index)
	})

	t.Run("create and get multiple validators", func(t *testing.T) {
		i, err := s.(*storage).nextIndex(validatorsPrefix())
		require.NoError(t, err)

		vis := []ValidatorInformation{
			{
				PublicKey: "8111b36feb8147d3f82c1a0",
				Operators: validatorInfo.Operators[:],
			}, {
				PublicKey: "8222b36feb8147d3f82c1a0",
				Operators: validatorInfo.Operators[:],
			}, {
				PublicKey: "8333b36feb8147d3f82c1a0",
				Operators: validatorInfo.Operators[:],
			},
		}
		for _, vi := range vis {
			err = s.SaveValidatorInformation(&vi)
			require.NoError(t, err)
		}

		for _, vi := range vis {
			validatorInfoFromDB, _, err := s.GetValidatorInformation(vi.PublicKey)
			require.NoError(t, err)
			require.Equal(t, i, validatorInfoFromDB.Index)
			require.Equal(t, validatorInfoFromDB.PublicKey, vi.PublicKey)
			i++
		}
	})
}

func TestStorage_ListValidators(t *testing.T) {
	storage, done := newStorageForTest()
	require.NotNil(t, storage)
	defer done()

	n := 5
	for i := 0; i < n; i++ {
		pk, _, err := rsaencryption.GenerateKeys()
		require.NoError(t, err)
		validator := ValidatorInformation{
			PublicKey: hex.EncodeToString(pk),
			Operators: []OperatorNodeLink{},
		}
		err = storage.SaveValidatorInformation(&validator)
		require.NoError(t, err)
	}

	validators, err := storage.ListValidators(0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(validators))
}

func newStorageForTest() (Storage, func()) {
	logger := zap.L()
	db, err := ssvstorage.GetStorageFactory(basedb.Options{
		Type:   "badger-memory",
		Logger: logger,
		Path:   "",
	})
	if err != nil {
		return nil, func() {}
	}
	s := NewNodeStorage(db, logger)
	return s, func() {
		db.Close()
	}
}
