package validator

import (
	"context"
	"crypto/sha256"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	spectypes "github.com/bloxapp/ssv-spec/types"
	ibftstorage "github.com/bloxapp/ssv/ibft/storage"
	"github.com/bloxapp/ssv/networkconfig"
	"github.com/bloxapp/ssv/protocol/v2/blockchain/beacon"
	"github.com/bloxapp/ssv/protocol/v2/ssv/runner"
	"github.com/bloxapp/ssv/protocol/v2/ssv/validator"
	"github.com/bloxapp/ssv/protocol/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func TestController_LiquidateCluster(t *testing.T) {
	storageMap := ibftstorage.NewStores()

	operatorData := buildOperatorData(1, "67Ce5c69260bd819B4e0AD13f4b873074D479811")
	secretKey := &bls.SecretKey{}
	secretKey2 := &bls.SecretKey{}
	require.NoError(t, secretKey.SetHexString(sk1Str))
	require.NoError(t, secretKey2.SetHexString(sk2Str))

	firstValidator := &validator.Validator{
		DutyRunners: runner.DutyRunners{},
		Storage:     ibftstorage.NewStores(),
		Share: &types.SSVShare{
			Share: spectypes.Share{
				ValidatorPubKey: secretKey.GetPublicKey().Serialize(),
			},
		},
	}

	ctrl, logger, sharesStorage, network, km, recipientStorage, bc := setupCommonTestComponents(t)
	defer ctrl.Finish()
	mockValidatorsMap := &validatorsMap{
		optsTemplate: &validator.Options{
			Signer:  km,
			Network: network,
			Storage: storageMap,
		},
		ctx:  context.Background(),
		lock: sync.RWMutex{},
		validatorsMap: map[string]*validator.Validator{
			secretKey.GetPublicKey().SerializeToHexStr(): firstValidator,
		},
	}

	validatorStartFunc := func(validator *validator.Validator) (bool, error) {
		return true, nil
	}
	controllerOptions := MockControllerOptions{
		beacon:              bc,
		network:             network,
		operatorData:        operatorData,
		sharesStorage:       sharesStorage,
		recipientsStorage:   recipientStorage,
		validatorsMap:       mockValidatorsMap,
		validatorOptions:    validator.Options{},
		metrics:             validator.NopMetrics{},
		metadataLastUpdated: map[string]time.Time{},
	}
	ctr := setupController(logger, controllerOptions)
	ctr.validatorStartFunc = validatorStartFunc

	require.Len(t, mockValidatorsMap.validatorsMap, 1)
	err := ctr.LiquidateCluster(common.HexToAddress("123"), []uint64{1, 2, 3, 4}, []*types.SSVShare{{Share: spectypes.Share{
		ValidatorPubKey: secretKey.GetPublicKey().Serialize(),
	}}})

	require.NoError(t, err)
	require.Len(t, mockValidatorsMap.validatorsMap, 0)
}

type signable struct {
}

func (signable) GetRoot() ([32]byte, error) {
	return sha256.Sum256([]byte("justdata")), nil
}

func TestController_StopValidator(t *testing.T) {
	storageMap := ibftstorage.NewStores()

	operatorData := buildOperatorData(1, "67Ce5c69260bd819B4e0AD13f4b873074D479811")
	secretKey := &bls.SecretKey{}
	secretKey2 := &bls.SecretKey{}
	require.NoError(t, secretKey.SetHexString(sk1Str))
	require.NoError(t, secretKey2.SetHexString(sk2Str))

	firstValidator := &validator.Validator{
		DutyRunners: runner.DutyRunners{},
		Storage:     ibftstorage.NewStores(),
		Share: &types.SSVShare{
			Share: spectypes.Share{
				ValidatorPubKey: secretKey.GetPublicKey().Serialize(),
			},
		},
	}

	ctrl, logger, sharesStorage, network, km, recipientStorage, bc := setupCommonTestComponents(t)
	defer ctrl.Finish()
	mockValidatorsMap := &validatorsMap{
		optsTemplate: &validator.Options{
			Signer:  km,
			Network: network,
			Storage: storageMap,
		},
		ctx:  context.Background(),
		lock: sync.RWMutex{},
		validatorsMap: map[string]*validator.Validator{
			secretKey.GetPublicKey().SerializeToHexStr(): firstValidator,
		},
	}

	validatorStartFunc := func(validator *validator.Validator) (bool, error) {
		return true, nil
	}
	controllerOptions := MockControllerOptions{
		beacon:              bc,
		network:             network,
		operatorData:        operatorData,
		sharesStorage:       sharesStorage,
		recipientsStorage:   recipientStorage,
		validatorsMap:       mockValidatorsMap,
		validatorOptions:    validator.Options{},
		metrics:             validator.NopMetrics{},
		metadataLastUpdated: map[string]time.Time{},
		keyManager:          km,
	}
	ctr := setupController(logger, controllerOptions)
	ctr.validatorStartFunc = validatorStartFunc

	require.NoError(t, km.AddShare(secretKey))
	_, err := km.SignRoot(signable{}, [4]byte{0, 0, 0, 0}, secretKey.GetPublicKey().Serialize())
	require.NoError(t, err)

	require.Len(t, mockValidatorsMap.validatorsMap, 1)
	err = ctr.StopValidator(secretKey.GetPublicKey().Serialize())
	require.NoError(t, err)
	require.Len(t, mockValidatorsMap.validatorsMap, 0)
}

func TestController_ReactivateCluster(t *testing.T) {
	storageMap := ibftstorage.NewStores()

	operatorData := buildOperatorData(1, "67Ce5c69260bd819B4e0AD13f4b873074D479811")
	secretKey := &bls.SecretKey{}
	secretKey2 := &bls.SecretKey{}
	require.NoError(t, secretKey.SetHexString(sk1Str))
	require.NoError(t, secretKey2.SetHexString(sk2Str))

	ctrl, logger, sharesStorage, network, km, recipientStorage, bc := setupCommonTestComponents(t)
	defer ctrl.Finish()
	mockValidatorsMap := &validatorsMap{
		optsTemplate: &validator.Options{
			Signer:  km,
			Network: network,
			Storage: storageMap,
		},
		ctx:           context.Background(),
		lock:          sync.RWMutex{},
		validatorsMap: map[string]*validator.Validator{
			//secretKey.GetPublicKey().SerializeToHexStr(): firstValidator,
		},
		createValidatorFunc: func(logger *zap.Logger, share *types.SSVShare) *validator.Validator {
			return &validator.Validator{
				DutyRunners: nil,
				Network:     nil,
				Share:       share,
				Signer:      nil,
				Storage:     nil,
				Queues:      nil,
			}
		},
	}

	validatorStartFunc := func(validator *validator.Validator) (bool, error) {
		return true, nil
	}
	controllerOptions := MockControllerOptions{
		beacon:            bc,
		network:           network,
		operatorData:      operatorData,
		sharesStorage:     sharesStorage,
		recipientsStorage: recipientStorage,
		validatorsMap:     mockValidatorsMap,
		validatorOptions: validator.Options{
			Storage:       storageMap,
			BeaconNetwork: networkconfig.TestNetwork.Beacon,
		},
		metrics:             validator.NopMetrics{},
		metadataLastUpdated: map[string]time.Time{},
		keyManager:          km,
	}
	ctr := setupController(logger, controllerOptions)
	ctr.validatorStartFunc = validatorStartFunc
	ctr.indicesChange = make(chan struct{})

	//require.NoError(t, km.AddShare(secretKey))
	//_, err := km.SignRoot(signable{}, [4]byte{0, 0, 0, 0}, secretKey.GetPublicKey().Serialize())
	//require.NoError(t, err)

	require.Len(t, mockValidatorsMap.validatorsMap, 0)
	toReactivate := []*types.SSVShare{
		{
			Share: spectypes.Share{ValidatorPubKey: secretKey.GetPublicKey().Serialize()},
			Metadata: types.Metadata{
				BeaconMetadata: &beacon.ValidatorMetadata{
					Balance:         0,
					Status:          v1.ValidatorStateActiveOngoing, // ValidatorStateUnknown
					Index:           1,
					ActivationEpoch: 1,
				},
			},
		},
		{
			Share: spectypes.Share{ValidatorPubKey: secretKey2.GetPublicKey().Serialize()},
			Metadata: types.Metadata{
				BeaconMetadata: &beacon.ValidatorMetadata{
					Balance:         0,
					Status:          v1.ValidatorStateActiveOngoing, // ValidatorStateUnknown
					Index:           1,
					ActivationEpoch: 1,
				},
			},
		},
	}
	recipientData := buildFeeRecipient("67Ce5c69260bd819B4e0AD13f4b873074D479811", "45E668aba4b7fc8761331EC3CE77584B7A99A51A")
	recipientStorage.EXPECT().GetRecipientData(gomock.Any(), gomock.Any()).Times(2).Return(recipientData, true, nil)

	indiciesUpdate := make(chan struct{})
	go func() {
		<-ctr.indicesChange
		indiciesUpdate <- struct{}{}
	}()
	err := ctr.ReactivateCluster(common.HexToAddress("0x1231231"), []uint64{1, 2, 3, 4}, toReactivate)

	require.NoError(t, err)
	require.Len(t, mockValidatorsMap.validatorsMap, 2)
	select {
	case <-indiciesUpdate:
		break
	case <-time.After(1 * time.Second):
		require.Fail(t, "didn't get indicies update")
	}

}