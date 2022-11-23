package eth1

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/bloxapp/ssv/eth1/abiparser"
)

//go:generate mockgen -package=eth1 -destination=./mock_client.go -source=./client.go

type eventData interface {
	toEventData() (interface{}, error)
}

type operatorRegistrationEventYAML struct {
	Id           uint32 `yaml:"Id"`
	Name         string `yaml:"Name"`
	OwnerAddress string `yaml:"OwnerAddress"`
	PublicKey    string `yaml:"PublicKey"`
}

type operatorRemovalEventYAML struct {
	OperatorId   uint32 `yaml:"Id"`
	OwnerAddress string `yaml:"OwnerAddress"`
}

type validatorRegistrationEventYAML struct {
	PublicKey        string   `yaml:"PublicKey"`
	OwnerAddress     string   `yaml:"OwnerAddress"`
	OperatorIds      []uint32 `yaml:"OperatorIds"`
	SharesPublicKeys []string `yaml:"SharesPublicKeys"`
	EncryptedKeys    []string `yaml:"EncryptedKeys"`
}

type validatorRemovalEventYAML struct {
	OwnerAddress string `yaml:"OwnerAddress"`
	PublicKey    string `yaml:"PublicKey"`
}

type accountLiquidationEventYAML struct {
	OwnerAddress string `yaml:"OwnerAddress"`
}

type accountEnableEventYAML struct {
	OwnerAddress string `yaml:"OwnerAddress"`
}

func (e *operatorRegistrationEventYAML) toEventData() (interface{}, error) {
	return abiparser.OperatorRegistrationEvent{
		Id:           e.Id,
		Name:         e.Name,
		OwnerAddress: common.HexToAddress(e.OwnerAddress),
		PublicKey:    []byte(e.PublicKey),
	}, nil
}

func (e *operatorRemovalEventYAML) toEventData() (interface{}, error) {
	return abiparser.OperatorRemovalEvent{
		OperatorId:   e.OperatorId,
		OwnerAddress: common.HexToAddress(e.OwnerAddress),
	}, nil
}

func (e *validatorRegistrationEventYAML) toEventData() (interface{}, error) {
	var toByteArr = func(orig []string) ([][]byte, error) {
		res := make([][]byte, len(orig))
		for i, v := range orig {
			d, err := hex.DecodeString(strings.TrimPrefix(v, "0x"))
			if err != nil {
				return nil, err
			}
			res[i] = d
		}
		return res, nil
	}
	var toByteArr2 = func(orig []string) ([][]byte, error) {
		outAbi, err := abiparser.GetOutAbi()
		if err != nil {
			return nil, err
		}
		res := make([][]byte, len(orig))
		for i, v := range orig {
			d, err := hex.DecodeString(strings.TrimPrefix(v, "0x"))
			if err != nil {
				return nil, err
			}
			unpackedRaw, err := outAbi.Unpack("method", d)
			if err != nil {
				return nil, err
			}
			unpacked, ok := unpackedRaw[0].(string)
			if !ok {
				return nil, errors.New("could not cast to string")
			}
			res[i] = []byte(unpacked)
		}
		return res, nil
	}
	pubKey, err := hex.DecodeString(strings.TrimPrefix(e.PublicKey, "0x"))
	if err != nil {
		return nil, err
	}
	sharePubKeys, err := toByteArr(e.SharesPublicKeys)
	if err != nil {
		return nil, err
	}
	encryptedKeys, err := toByteArr2(e.EncryptedKeys)
	if err != nil {
		return nil, err
	}
	return abiparser.ValidatorRegistrationEvent{
		PublicKey:        pubKey,
		OwnerAddress:     common.HexToAddress(e.OwnerAddress),
		OperatorIds:      e.OperatorIds,
		SharesPublicKeys: sharePubKeys,
		EncryptedKeys:    encryptedKeys,
	}, nil
}

func (e *validatorRemovalEventYAML) toEventData() (interface{}, error) {
	return abiparser.ValidatorRemovalEvent{
		OwnerAddress: common.HexToAddress(e.OwnerAddress),
		PublicKey:    []byte(strings.TrimPrefix(e.PublicKey, "0x")),
	}, nil
}

func (e *accountLiquidationEventYAML) toEventData() (interface{}, error) {
	return abiparser.AccountLiquidationEvent{
		OwnerAddress: common.HexToAddress(e.OwnerAddress),
	}, nil
}

func (e *accountEnableEventYAML) toEventData() (interface{}, error) {
	return abiparser.AccountEnableEvent{
		OwnerAddress: common.HexToAddress(e.OwnerAddress),
	}, nil
}

type eventDataUnmarshaler struct {
	name string
	data eventData
}

func (u *eventDataUnmarshaler) UnmarshalYAML(value *yaml.Node) error {
	var err error
	switch u.name {
	case "OperatorRegistration":
		var v operatorRegistrationEventYAML
		err = value.Decode(&v)
		u.data = &v
	case "OperatorRemoval":
		var v operatorRemovalEventYAML
		err = value.Decode(&v)
		u.data = &v
	case "ValidatorRegistration":
		var v validatorRegistrationEventYAML
		err = value.Decode(&v)
		u.data = &v
	case "ValidatorRemoval":
		var v validatorRemovalEventYAML
		err = value.Decode(&v)
		u.data = &v
	case "AccountLiquidation":
		var v accountLiquidationEventYAML
		err = value.Decode(&v)
		u.data = &v
	case "AccountEnable":
		var v accountEnableEventYAML
		err = value.Decode(&v)
		u.data = &v
	default:
		return errors.New("event unknown")
	}

	return err
}

func (e *Event) UnmarshalYAML(value *yaml.Node) error {
	var evName struct {
		Name string `yaml:"Name"`
	}
	err := value.Decode(&evName)
	if err != nil {
		return err
	}
	var ev struct {
		Data eventDataUnmarshaler `yaml:"Data"`
	}
	ev.Data.name = evName.Name

	if err := value.Decode(&ev); err != nil {
		return err
	}
	e.Name = ev.Data.name
	data, err := ev.Data.data.toEventData()
	if err != nil {
		return err
	}
	e.Data = data

	return nil
}
