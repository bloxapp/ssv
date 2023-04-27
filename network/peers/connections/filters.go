package connections

import (
	"crypto"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/bloxapp/ssv/network/records"
	"github.com/bloxapp/ssv/operator/storage"
	"github.com/bloxapp/ssv/utils/rsaencryption"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var TimestampAllowedDifference = 30 * time.Second

// NetworkIDFilter determines whether we will connect to the given node by the network ID
func NetworkIDFilter(networkID string) HandshakeFilter {
	return func(sender peer.ID, sni *records.SignedNodeInfo) (bool, error) {
		if networkID != sni.NetworkID {
			return false, errors.Errorf("networkID '%s' instead of '%s'", sni.NetworkID, networkID)
		}
		return true, nil
	}
}

func SenderRecipientIPsCheckFilter(me peer.ID) HandshakeFilter {
	return func(sender peer.ID, sni *records.SignedNodeInfo) (bool, error) {
		if sni.HandshakeData.RecipientPeerID != me {
			return false, errors.Errorf("recepient peer ID '%s' instead of '%s'", sni.HandshakeData.RecipientPeerID, me)
		}

		if sni.HandshakeData.SenderPeerID != sender {
			return false, errors.Errorf("sender peer ID '%s' instead of '%s'", sni.HandshakeData.SenderPeerID, sender)
		}

		return true, nil
	}
}

func SignatureCheckFilter() HandshakeFilter {
	return func(sender peer.ID, sni *records.SignedNodeInfo) (bool, error) {
		publicKey, err := rsaencryption.ConvertPemToPublicKey(sni.HandshakeData.SenderPubKey)
		if err != nil {
			return false, err
		}

		hashed := sni.HandshakeData.Hash()
		if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sni.Signature); err != nil {
			return false, err
		}

		if !sni.HandshakeData.Timestamp.Round(TimestampAllowedDifference).Equal(time.Now().Round(TimestampAllowedDifference)) {
			return false, fmt.Errorf("signature was made more than %f seconds ago", TimestampAllowedDifference.Seconds())
		}

		return true, nil
	}
}

func RegisteredOperatorsFilter(logger *zap.Logger, nodeStorage storage.Storage) HandshakeFilter { //operator is not registered means operator not whitelisted
	return func(sender peer.ID, sni *records.SignedNodeInfo) (bool, error) {
		_, found, err := nodeStorage.GetOperatorDataByPubKey(logger, []byte(sni.HandshakeData.SenderPubKey))
		if !found {
			return false, errors.Wrap(err, "operator wasn't found, probably not registered to a contract")
		}

		return true, nil
	}
}
