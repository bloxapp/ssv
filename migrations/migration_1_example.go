package migrations

import (
	"context"
	"go.uber.org/zap"

	"github.com/bloxapp/ssv/storage/basedb"
	"github.com/pkg/errors"
)

// This migration is an Example of atomic
// View/Update transactions usage
var migrationExample2 = Migration{
	Name: "migration_1_example",
	Run: func(ctx context.Context, logger *zap.Logger, opt Options, key []byte) error {
		return opt.Db.Update(func(txn basedb.Txn) error {
			var (
				testPrefix = []byte("test_prefix/")
				testKey    = []byte("test_key")
				testValue  = []byte("test_value")
			)
			err := txn.Set(testPrefix, testKey, testValue)
			if err != nil {
				return err
			}
			obj, found, err := txn.Get(testPrefix, testKey)
			if err != nil {
				return err
			}
			if !found {
				return errors.Errorf("the key %s is not found", string(obj.Key))
			}
			logger.Debug("migration_1_example: key found", zap.String("key", string(obj.Key)), zap.String("value", string(obj.Value)))
			return txn.Set(migrationsPrefix, key, migrationCompleted)
		})
	},
}
