package storage

import (
	"bytes"
	"encoding/json"
	"github.com/bloxapp/ssv/eth1"
	"github.com/bloxapp/ssv/storage/kv"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	validatorsPrefix = []byte("validators")
)

// ValidatorInformation represents a validator
type ValidatorInformation struct {
	Index     int64              `json:"index"`
	PublicKey string             `json:"publicKey"`
	Operators []OperatorNodeLink `json:"operators"`
}

// ValidatorsCollection is the interface for managing validators information
type ValidatorsCollection interface {
	GetValidatorInformation(validatorPubKey string) (*ValidatorInformation, error)
	SaveValidatorInformation(validatorInformation *ValidatorInformation) error
	ListValidators(from int64, to int64) ([]ValidatorInformation, error)
}

// OperatorNodeLink links a validator to an operator
type OperatorNodeLink struct {
	ID        uint64 `json:"nodeId"`
	PublicKey string `json:"publicKey"`
}

// ListValidators returns information of all the known validators
// when 'to' equals zero, all validators will be returned
func (es *exporterStorage) ListValidators(from int64, to int64) ([]ValidatorInformation, error) {
	objs, err := es.db.GetAllByCollection(append(storagePrefix, validatorsPrefix...))
	if err != nil {
		return nil, err
	}
	to = normalTo(to)
	var validators []ValidatorInformation
	for _, obj := range objs {
		var vi ValidatorInformation
		err = json.Unmarshal(obj.Value, &vi)
		if vi.Index >= from && vi.Index <= to {
			validators = append(validators, vi)
		}
	}
	return validators, err
}

// GetValidatorInformation returns information of the given validator by public key
func (es *exporterStorage) GetValidatorInformation(validatorPubKey string) (*ValidatorInformation, error) {
	obj, err := es.db.Get(storagePrefix, validatorKey(validatorPubKey))
	if err != nil {
		return nil, err
	}
	var vi ValidatorInformation
	err = json.Unmarshal(obj.Value, &vi)
	return &vi, err
}

// SaveValidatorInformation saves validator information by its public key
func (es *exporterStorage) SaveValidatorInformation(validatorInformation *ValidatorInformation) error {
	existing, err := es.GetValidatorInformation(validatorInformation.PublicKey)
	if err != nil && err.Error() != kv.EntryNotFoundError {
		return errors.Wrap(err, "could not read information from DB")
	}
	if existing != nil {
		es.logger.Debug("validator already exist",
			zap.String("pubKey", validatorInformation.PublicKey))
		validatorInformation.Index = existing.Index
		// TODO: update validator information (i.e. change operator)
		return nil
	}
	validatorInformation.Index, err = es.nextIndex(validatorsPrefix)
	if err != nil {
		return errors.Wrap(err, "could not calculate next validator index")
	}
	raw, err := json.Marshal(validatorInformation)
	if err != nil {
		return errors.Wrap(err, "could not marshal validator information")
	}
	return es.db.Set(storagePrefix, validatorKey(validatorInformation.PublicKey), raw)
}

// ToValidatorInformation converts raw event to ValidatorInformation
func ToValidatorInformation(validatorAddedEvent eth1.ValidatorAddedEvent) (*ValidatorInformation, error) {
	pubKey := &bls.PublicKey{}
	if err := pubKey.Deserialize(validatorAddedEvent.PublicKey); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize validator public key")
	}

	var operators []OperatorNodeLink
	for i := range validatorAddedEvent.OessList {
		oess := validatorAddedEvent.OessList[i]
		nodeID := oess.Index.Uint64() + 1
		operators = append(operators, OperatorNodeLink{
			ID: nodeID, PublicKey: string(oess.OperatorPublicKey),
		})
	}

	vi := ValidatorInformation{
		PublicKey: pubKey.SerializeToHexStr(),
		Operators: operators,
	}

	return &vi, nil
}

func validatorKey(pubKey string) []byte {
	return bytes.Join([][]byte{
		validatorsPrefix[:],
		[]byte(pubKey),
	}, []byte("/"))
}
