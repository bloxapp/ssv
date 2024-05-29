package qbft

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/herumi/bls-eth-go-binary/bls"
	genesisspecqbft "github.com/ssvlabs/ssv-spec-pre-cc/qbft"
	genesisspectypes "github.com/ssvlabs/ssv-spec-pre-cc/types"
	genesistestingutils "github.com/ssvlabs/ssv-spec-pre-cc/types/testingutils"
	"github.com/ssvlabs/ssv-spec/types/testingutils"
)

var TestingMessage = &genesisspecqbft.Message{
	MsgType:    genesisspecqbft.ProposalMsgType,
	Height:     genesisspecqbft.FirstHeight,
	Round:      genesisspecqbft.FirstRound,
	Identifier: []byte{1, 2, 3, 4},
	Root:       [32]byte{1, 2, 3, 4},
}

var TestingSignedMsg = func() *genesisspecqbft.SignedMessage {
	return SignMsg(TestingSK, 1, TestingMessage)
}()

var SignMsg = func(sk *bls.SecretKey, id genesisspectypes.OperatorID, msg *genesisspecqbft.Message) *genesisspecqbft.SignedMessage {
	domain := genesistestingutils.TestingSSVDomainType
	sigType := genesisspectypes.QBFTSignatureType

	r, _ := genesisspectypes.ComputeSigningRoot(msg, genesisspectypes.ComputeSignatureDomain(domain, sigType))
	sig := sk.SignByte(r[:])

	return &genesisspecqbft.SignedMessage{
		Message:   *msg,
		Signers:   []genesisspectypes.OperatorID{id},
		Signature: sig.Serialize(),
	}
}

var TestingSK = func() *bls.SecretKey {
	genesisspectypes.InitBLS()
	ret := &bls.SecretKey{}
	ret.SetByCSPRNG()
	return ret
}()

var testingValidatorPK = phase0.BLSPubKey{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}

var testingShare = &genesisspectypes.Share{
	OperatorID:      1,
	ValidatorPubKey: testingValidatorPK[:],
	SharePubKey:     TestingSK.GetPublicKey().Serialize(),
	DomainType:      genesistestingutils.TestingSSVDomainType,
	Quorum:          3,
	PartialQuorum:   2,
	Committee: []*genesisspectypes.Operator{
		{
			OperatorID:  1,
			SharePubKey: TestingSK.GetPublicKey().Serialize(),
		},
	},
}

var TestingInstanceStruct = &genesisspecqbft.Instance{
	State: &genesisspecqbft.State{
		Share:                           testingShare,
		ID:                              []byte{1, 2, 3, 4},
		Round:                           1,
		Height:                          1,
		LastPreparedRound:               1,
		LastPreparedValue:               []byte{1, 2, 3, 4},
		ProposalAcceptedForCurrentRound: TestingSignedMsg,
		Decided:                         false,
		DecidedValue:                    []byte{1, 2, 3, 4},

		ProposeContainer: &genesisspecqbft.MsgContainer{
			Msgs: map[genesisspecqbft.Round][]*genesisspecqbft.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		PrepareContainer: &genesisspecqbft.MsgContainer{
			Msgs: map[genesisspecqbft.Round][]*genesisspecqbft.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		CommitContainer: &genesisspecqbft.MsgContainer{
			Msgs: map[genesisspecqbft.Round][]*genesisspecqbft.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
		RoundChangeContainer: &genesisspecqbft.MsgContainer{
			Msgs: map[genesisspecqbft.Round][]*genesisspecqbft.SignedMessage{
				1: {
					TestingSignedMsg,
				},
			},
		},
	},
}

var TestingControllerStruct = &genesisspecqbft.Controller{
	Identifier:      []byte{1, 2, 3, 4},
	Height:          genesisspecqbft.Height(1),
	Share:           testingShare,
	StoredInstances: genesisspecqbft.InstanceContainer{TestingInstanceStruct},
}

type testingOperatorSigner struct {
	SSVOperatorSK *rsa.PrivateKey
	operatorID    genesisspectypes.OperatorID
}

func NewTestingOperatorSigner(keySet *testingutils.TestKeySet, operatorID genesisspectypes.OperatorID) *testingOperatorSigner {
	return &testingOperatorSigner{
		SSVOperatorSK: keySet.OperatorKeys[operatorID],
		operatorID:    operatorID,
	}
}

func (km *testingOperatorSigner) SignSSVMessage(data []byte) ([256]byte, error) {
	return SignSSVMessage(km.SSVOperatorSK, data)
}

// GetOperatorID returns the operator ID
func (km *testingOperatorSigner) GetOperatorID() genesisspectypes.OperatorID {
	return km.operatorID
}

func SignSSVMessage(sk *rsa.PrivateKey, data []byte) ([256]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, sk, crypto.SHA256, hash[:])
	if err != nil {
		return [256]byte{}, err
	}

	return [256]byte(signature), nil
}
