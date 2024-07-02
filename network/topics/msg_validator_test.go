package topics

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethereum/go-ethereum/common"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ps_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	pspb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/ssvlabs/ssv-spec/qbft"
	spectypes "github.com/ssvlabs/ssv-spec/types"
	spectestingutils "github.com/ssvlabs/ssv-spec/types/testingutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/ssvlabs/ssv/message/signatureverifier"
	"github.com/ssvlabs/ssv/message/validation"
	"github.com/ssvlabs/ssv/network/commons"
	"github.com/ssvlabs/ssv/networkconfig"
	"github.com/ssvlabs/ssv/operator/duties/dutystore"
	operatorstorage "github.com/ssvlabs/ssv/operator/storage"
	beaconprotocol "github.com/ssvlabs/ssv/protocol/v2/blockchain/beacon"
	ssvtypes "github.com/ssvlabs/ssv/protocol/v2/types"
	"github.com/ssvlabs/ssv/registry/storage"
	"github.com/ssvlabs/ssv/storage/basedb"
	"github.com/ssvlabs/ssv/storage/kv"
	"github.com/ssvlabs/ssv/utils/rsaencryption"
)

func TestMsgValidator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	db, err := kv.NewInMemory(logger, basedb.Options{})
	require.NoError(t, err)

	ns, err := operatorstorage.NewNodeStorage(logger, db)
	require.NoError(t, err)

	ks := spectestingutils.Testing4SharesSet()
	share := &ssvtypes.SSVShare{
		Share: *spectestingutils.TestingShare(ks, spectestingutils.TestingValidatorIndex),
		Metadata: ssvtypes.Metadata{
			BeaconMetadata: &beaconprotocol.ValidatorMetadata{
				Status: v1.ValidatorStateActiveOngoing,
			},
			Liquidated: false,
		},
	}
	require.NoError(t, ns.Shares().Save(nil, share))
	dutyStore := dutystore.New()
	ctrl := gomock.NewController(t)
	signatureVerifier := signatureverifier.NewMockSignatureVerifier(ctrl)
	signatureVerifier.EXPECT().VerifySignature(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mv := validation.New(networkconfig.TestNetwork, ns.ValidatorStore(), dutyStore, signatureVerifier)
	require.NotNil(t, mv)

	slot := networkconfig.TestNetwork.Beacon.GetBeaconNetwork().EstimatedCurrentSlot()

	t.Run("valid consensus msg", func(t *testing.T) {
		ssvMsg, err := dummySSVConsensusMsg(share.ValidatorPubKey, qbft.Height(slot))
		require.NoError(t, err)

		_, skByte, err := rsaencryption.GenerateKeys()
		require.NoError(t, err)
		operatorPrivateKey, err := rsaencryption.PemToPrivateKey(skByte)
		require.NoError(t, err)

		operatorId := uint64(1)

		operatorPubKey, err := rsaencryption.ExtractPublicKey(&operatorPrivateKey.PublicKey)
		require.NoError(t, err)

		od := &storage.OperatorData{
			PublicKey:    []byte(operatorPubKey),
			OwnerAddress: common.Address{},
			ID:           operatorId,
		}

		found, err := ns.SaveOperatorData(nil, od)
		require.False(t, found)
		require.NoError(t, err)

		encodedMsg, err := commons.EncodeNetworkMsg(ssvMsg)
		require.NoError(t, err)

		hash := sha256.Sum256(encodedMsg)
		signature, err := rsa.SignPKCS1v15(nil, operatorPrivateKey, crypto.SHA256, hash[:])
		require.NoError(t, err)

		sig := [256]byte{}
		copy(sig[:], signature)

		packedPubSubMsgPayload := &spectypes.SignedSSVMessage{
			Signatures:  [][]byte{sig[:]},
			OperatorIDs: []spectypes.OperatorID{operatorId},
			SSVMessage:  ssvMsg,
		}
		encPackedPubSubMsgPayload, err := packedPubSubMsgPayload.Encode()
		require.NoError(t, err)

		topicID := commons.ValidatorTopicID(ssvMsg.GetID().GetDutyExecutorID())

		pmsg := &pubsub.Message{
			Message: &pspb.Message{
				Topic: &topicID[0],
				Data:  encPackedPubSubMsgPayload,
			},
		}

		res := mv.Validate(context.Background(), "16Uiu2HAkyWQyCb6reWXGQeBUt9EXArk6h3aq3PsFMwLNq3pPGH1r", pmsg)
		require.Equal(t, pubsub.ValidationAccept, res)
	})

	t.Run("wrong topic", func(t *testing.T) {
		// pkHex := "b5de683dbcb3febe8320cc741948b9282d59b75a6970ed55d6f389da59f26325331b7ea0e71a2552373d0debb6048b8a"
		msg, err := dummySSVConsensusMsg(share.ValidatorPubKey, 15160)
		require.NoError(t, err)
		raw, err := msg.Encode()
		require.NoError(t, err)
		pk, err := hex.DecodeString("a297599ccf617c3b6118bbd248494d7072bb8c6c1cc342ea442a289415987d306bad34415f89469221450a2501a832ec")
		require.NoError(t, err)
		topics := commons.ValidatorTopicID(pk)
		pmsg := newPBMsg(raw, topics[0], []byte("16Uiu2HAkyWQyCb6reWXGQeBUt9EXArk6h3aq3PsFMwLNq3pPGH1r"))
		res := mv.Validate(context.Background(), "16Uiu2HAkyWQyCb6reWXGQeBUt9EXArk6h3aq3PsFMwLNq3pPGH1r", pmsg)
		require.Equal(t, res, pubsub.ValidationReject)
	})

	t.Run("empty message", func(t *testing.T) {
		pmsg := newPBMsg([]byte{}, "xxx", []byte{})
		res := mv.Validate(context.Background(), "xxxx", pmsg)
		require.Equal(t, pubsub.ValidationReject, res)
	})

	t.Run("invalid validator public key", func(t *testing.T) {
		pkHex := "b5de683dbcb3febe8320cc741948b9282d59b75a6970ed55d6f389da59f26325331b7ea0e71a2552373d0debb6048b8a"
		pk, err := hex.DecodeString(pkHex)
		require.NoError(t, err)
		msg, err := dummySSVConsensusMsg(spectypes.ValidatorPK(pk[:]), 1)
		require.NoError(t, err)
		raw, err := msg.Encode()
		require.NoError(t, err)
		pmsg := newPBMsg(raw, "xxx", []byte{})
		res := mv.Validate(context.Background(), "xxxx", pmsg)
		require.Equal(t, res, pubsub.ValidationReject)
	})
}

func newPBMsg(data []byte, topic string, from []byte) *pubsub.Message {
	pmsg := &pubsub.Message{
		Message: &ps_pb.Message{},
	}
	pmsg.Data = data
	pmsg.Topic = &topic
	pmsg.From = from
	return pmsg
}

func dummySSVConsensusMsg(pk spectypes.ValidatorPK, height qbft.Height) (*spectypes.SSVMessage, error) {
	id := spectypes.NewMsgID(networkconfig.TestNetwork.Domain, pk[:], spectypes.RunnerRole(spectypes.BNRoleAttester))
	ks := spectestingutils.Testing4SharesSet()

	validSignedMessage := spectestingutils.TestingRoundChangeMessageWithHeightAndIdentifier(ks.OperatorKeys[1], 1, height, id[:])

	return &spectypes.SSVMessage{
		MsgType: spectypes.SSVConsensusMsgType,
		MsgID:   id,
		Data:    validSignedMessage.SSVMessage.Data,
	}, nil
}
