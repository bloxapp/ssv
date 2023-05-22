package connections

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/bloxapp/ssv/network/records"
	"github.com/bloxapp/ssv/operator/storage"
	"github.com/bloxapp/ssv/utils/rsaencryption"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var AllowedDifference = 30 * time.Second

// NetworkIDFilter determines whether we will connect to the given node by the network ID
func NetworkIDFilter(networkID string) HandshakeFilter {
	return func(sender peer.ID, ani records.AnyNodeInfo) error {
		nid := ani.GetNodeInfo().NetworkID
		if networkID != nid {
			return errors.Errorf("networkID '%s' instead of '%s'", nid, networkID)
		}
		return nil
	}
}

func SenderRecipientIPsCheckFilter(me peer.ID) HandshakeFilter { // for some reason we're loosing 'me' value
	return func(sender peer.ID, ani records.AnyNodeInfo) error {
		sni, ok := ani.(*records.SignedNodeInfo)
		if !ok {
			return fmt.Errorf("wrong format nodeinfo sent")
		}
		if sni.HandshakeData.RecipientPeerID != me {
			return errors.Errorf("recepient peer ID '%s' instead of '%s'", sni.HandshakeData.RecipientPeerID, me)
		}

		if sni.HandshakeData.SenderPeerID != sender {
			return errors.Errorf("sender peer ID '%s' instead of '%s'", sni.HandshakeData.SenderPeerID, sender)
		}

		return nil
	}
}

func SignatureCheckFilter() HandshakeFilter {
	return func(sender peer.ID, ani records.AnyNodeInfo) error {
		sni, ok := ani.(*records.SignedNodeInfo)
		if !ok {
			return fmt.Errorf("wrong format nodeinfo sent")
		}
		decodedPublicKey, err := base64.StdEncoding.DecodeString(string(sni.HandshakeData.SenderPubicKey))
		if err != nil {
			return errors.Wrap(err, "failed to decode sender public key from signed node info")
		}

		publicKey, err := rsaencryption.ConvertPemToPublicKey(decodedPublicKey)
		if err != nil {
			return err
		}

		hashed := sni.HandshakeData.Hash()
		if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sni.Signature); err != nil {
			return err
		}

		if difference := time.Since(sni.HandshakeData.Timestamp); difference > AllowedDifference {
			return fmt.Errorf("signature made %f seconds ago, should no more than %f seconds ago", difference.Seconds(), AllowedDifference.Seconds())
		}

		return nil
	}
}

func RegisteredOperatorsFilter(logger *zap.Logger, nodeStorage storage.Storage, keysConfigWhitelist []string) HandshakeFilter {
	return func(sender peer.ID, ani records.AnyNodeInfo) error {
		sni, ok := ani.(*records.SignedNodeInfo)
		if !ok {
			return fmt.Errorf("wrong format nodeinfo sent")
		}
		if len(sni.HandshakeData.SenderPubicKey) == 0 {
			return errors.New("empty SenderPubicKey")
		}

		for _, key := range keysConfigWhitelist {
			if key == string(sni.HandshakeData.SenderPubicKey) {
				return nil
			}
		}

		data, found, err := nodeStorage.GetOperatorDataByPubKey(logger, sni.HandshakeData.SenderPubicKey)
		if !found || data != nil {
			return errors.Wrap(err, "operator wasn't found, probably not registered to a contract")
		}

		return nil
	}
}
