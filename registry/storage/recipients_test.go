package storage_test

import (
	"fmt"
	"testing"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/logging"
	"github.com/bloxapp/ssv/registry/storage"
	ssvstorage "github.com/bloxapp/ssv/storage"
	"github.com/bloxapp/ssv/storage/basedb"
)

func TestStorage_SaveAndGetRecipientData(t *testing.T) {
	logger := logging.TestLogger(t)
	storageCollection, done := newRecipientStorageForTest(logger)
	require.NotNil(t, storageCollection)
	defer done()

	recipientData := &storage.RecipientData{
		Owner: common.BytesToAddress([]byte("0x1")),
	}
	copy(recipientData.FeeRecipient[:], "0x2")

	t.Run("get non-existing recipient", func(t *testing.T) {
		nonExistingRecipient, found, err := storageCollection.GetRecipientData(recipientData.Owner)
		require.NoError(t, err)
		require.Nil(t, nonExistingRecipient)
		require.False(t, found)
	})

	t.Run("create and get recipient", func(t *testing.T) {
		rd, err := storageCollection.SaveRecipientData(recipientData)
		require.NoError(t, err)

		recipientDataFromDB, found, err := storageCollection.GetRecipientData(recipientData.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, recipientData.Owner, recipientDataFromDB.Owner)
		require.Equal(t, recipientData.FeeRecipient, recipientDataFromDB.FeeRecipient)
		require.Equal(t, recipientData.Owner, rd.Owner)
		require.Equal(t, recipientData.FeeRecipient, rd.FeeRecipient)
	})

	t.Run("create existing recipient", func(t *testing.T) {
		rdToSave := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x2")),
		}
		copy(rdToSave.FeeRecipient[:], "0x2")

		rd, err := storageCollection.SaveRecipientData(rdToSave)
		require.NoError(t, err)
		require.NotNil(t, rd)

		rdDup, err := storageCollection.SaveRecipientData(rdToSave)
		require.NoError(t, err)
		require.Nil(t, rdDup)

		rdFromDB, found, err := storageCollection.GetRecipientData(rd.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, rdFromDB)
	})

	t.Run("update existing recipient", func(t *testing.T) {
		rdToSave := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x3")),
		}
		copy(rdToSave.FeeRecipient[:], "0x2")

		rd, err := storageCollection.SaveRecipientData(rdToSave)
		require.NoError(t, err)
		require.NotNil(t, rd)

		copy(rdToSave.FeeRecipient[:], "0x3")
		rdNew, err := storageCollection.SaveRecipientData(rdToSave)
		require.NoError(t, err)
		require.NotNil(t, rdNew)

		rdFromDB, found, err := storageCollection.GetRecipientData(rd.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, rdNew.Owner, rdFromDB.Owner)
		require.Equal(t, rdNew.FeeRecipient, rdFromDB.FeeRecipient)
	})

	t.Run("delete recipient", func(t *testing.T) {
		rdToSave := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x4")),
		}
		copy(rdToSave.FeeRecipient[:], "0x2")

		rd, err := storageCollection.SaveRecipientData(rdToSave)
		require.NoError(t, err)
		require.NotNil(t, rd)

		err = storageCollection.DeleteRecipientData(rd.Owner)
		require.NoError(t, err)

		rdFromDB, found, err := storageCollection.GetRecipientData(rd.Owner)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, rdFromDB)
	})

	t.Run("create and get many recipients", func(t *testing.T) {
		var ownerAddresses []common.Address
		var savedRecipients []*storage.RecipientData
		for i := 0; i < 10; i++ {
			rd := storage.RecipientData{
				Owner: common.BytesToAddress([]byte(fmt.Sprintf("0x%d", i))),
			}
			copy(recipientData.FeeRecipient[:], fmt.Sprintf("0x%d", i))
			ownerAddresses = append(ownerAddresses, rd.Owner)

			_, err := storageCollection.SaveRecipientData(&rd)
			require.NoError(t, err)

			savedRecipients = append(savedRecipients, &rd)
		}

		recipients, err := storageCollection.GetRecipientDataMany(logger, ownerAddresses)
		require.NoError(t, err)
		require.Equal(t, len(ownerAddresses), len(recipients))

		for _, r := range savedRecipients {
			require.Equal(t, r.FeeRecipient, recipients[r.Owner])
		}
	})

	t.Run("create recipient should not initializing nonce", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11111")),
		}

		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)

		recipientDataFromDB, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, rdToCreate.Owner, recipientDataFromDB.Owner)
		require.Equal(t, rdToCreate.FeeRecipient, recipientDataFromDB.FeeRecipient)
		require.Nil(t, recipientDataFromDB.Nonce)
		require.Equal(t, rdToCreate.Owner, rd.Owner)
		require.Equal(t, rdToCreate.FeeRecipient, rd.FeeRecipient)
		require.Nil(t, rd.Nonce)
	})

	t.Run("bump nonce before fee recipient created", func(t *testing.T) {
		owner := common.BytesToAddress([]byte("0x11112"))
		var feeRecipient bellatrix.ExecutionAddress
		copy(feeRecipient[:], owner.Bytes())

		data, found, err := storageCollection.GetRecipientData(owner)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, data)

		err = storageCollection.BumpNonce(owner)
		require.NoError(t, err)

		data, found, err = storageCollection.GetRecipientData(owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Equal(t, owner, data.Owner)
		require.Equal(t, feeRecipient, data.FeeRecipient)
		require.Equal(t, storage.Nonce(0), *data.Nonce)
	})

	t.Run("bump nonce after fee recipient created", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11113")),
		}
		copy(rdToCreate.FeeRecipient[:], rdToCreate.Owner.Bytes())
		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)
		require.NotNil(t, rd)

		err = storageCollection.BumpNonce(rdToCreate.Owner)
		require.NoError(t, err)

		data, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Equal(t, storage.Nonce(0), *data.Nonce)
	})

	t.Run("bump non-zero nonce", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11114")),
		}
		nonce := storage.Nonce(0)
		copy(rdToCreate.FeeRecipient[:], rdToCreate.Owner.Bytes())
		rdToCreate.Nonce = &nonce

		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)
		require.NotNil(t, rd)

		err = storageCollection.BumpNonce(rdToCreate.Owner)
		require.NoError(t, err)

		data, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Equal(t, storage.Nonce(1), *data.Nonce)
	})

	t.Run("get next nonce before fee recipient created - should be 0", func(t *testing.T) {
		owner := common.BytesToAddress([]byte("0x11115"))
		var feeRecipient bellatrix.ExecutionAddress
		copy(feeRecipient[:], owner.Bytes())

		data, found, err := storageCollection.GetRecipientData(owner)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, data)

		nonce, err := storageCollection.GetNextNonce(owner)
		require.NoError(t, err)
		require.Equal(t, storage.Nonce(0), nonce)

		data, found, err = storageCollection.GetRecipientData(owner)
		require.NoError(t, err)
		require.False(t, found)
		require.Nil(t, data)
	})

	t.Run("get next nonce after fee recipient created - should be 0", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11116")),
		}
		copy(rdToCreate.FeeRecipient[:], rdToCreate.Owner.Bytes())

		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)
		require.NotNil(t, rd)

		nonce, err := storageCollection.GetNextNonce(rdToCreate.Owner)
		require.NoError(t, err)
		require.Equal(t, storage.Nonce(0), nonce)

		data, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Nil(t, data.Nonce)
	})

	t.Run("get next nonce before bump", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11117")),
		}
		copy(rdToCreate.FeeRecipient[:], rdToCreate.Owner.Bytes())

		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)
		require.NotNil(t, rd)

		nonce, err := storageCollection.GetNextNonce(rdToCreate.Owner)
		require.NoError(t, err)
		require.Equal(t, storage.Nonce(0), nonce)

		err = storageCollection.BumpNonce(rdToCreate.Owner)
		require.NoError(t, err)

		data, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Equal(t, storage.Nonce(0), *data.Nonce)
	})

	t.Run("get next nonce after bump", func(t *testing.T) {
		rdToCreate := &storage.RecipientData{
			Owner: common.BytesToAddress([]byte("0x11118")),
		}
		copy(rdToCreate.FeeRecipient[:], rdToCreate.Owner.Bytes())

		rd, err := storageCollection.SaveRecipientData(rdToCreate)
		require.NoError(t, err)
		require.NotNil(t, rd)

		err = storageCollection.BumpNonce(rdToCreate.Owner)
		require.NoError(t, err)

		nonce, err := storageCollection.GetNextNonce(rdToCreate.Owner)
		require.NoError(t, err)
		require.Equal(t, storage.Nonce(1), nonce)

		data, found, err := storageCollection.GetRecipientData(rdToCreate.Owner)
		require.NoError(t, err)
		require.True(t, found)
		require.NotNil(t, data)
		require.Equal(t, storage.Nonce(0), *data.Nonce)
	})
}

func newRecipientStorageForTest(logger *zap.Logger) (storage.Recipients, func()) {
	db, err := ssvstorage.GetStorageFactory(logger, basedb.Options{
		Type: "badger-memory",
		Path: "",
	})
	if err != nil {
		return nil, func() {}
	}
	s := storage.NewRecipientsStorage(db, []byte("test"))
	return s, func() {
		db.Close(logger)
	}
}
