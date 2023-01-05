package commons

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"github.com/btcsuite/btcd/btcec/v2"
	gcrypto "github.com/ethereum/go-ethereum/crypto"
	"math/big"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
)

// ConvertFromInterfacePrivKey converts crypto.PrivKey back to ecdsa.PrivateKey
func ConvertFromInterfacePrivKey(privkey crypto.PrivKey) (*ecdsa.PrivateKey, error) {
	secpKey := (privkey.(*crypto.Secp256k1PrivateKey))
	rawKey, err := secpKey.Raw()
	if err != nil {
		return nil, errors.Wrap(err, "could mot convert ecdsa.PrivateKey")
	}
	privKey := new(ecdsa.PrivateKey)
	k := new(big.Int).SetBytes(rawKey)
	privKey.D = k
	privKey.Curve = gcrypto.S256() // Temporary hack, so libp2p Secp256k1 is recognized as geth Secp256k1 in disc v5.1.
	privKey.X, privKey.Y = gcrypto.S256().ScalarBaseMult(rawKey)
	return privKey, nil
}

// ConvertToInterfacePrivkey converts ecdsa.PrivateKey to crypto.PrivKey
func ConvertToInterfacePrivkey(privkey *ecdsa.PrivateKey) (crypto.PrivKey, error) {
	privBytes := privkey.D.Bytes()
	// In the event the number of bytes outputted by the big-int are less than 32,
	// we append bytes to the start of the sequence for the missing most significant
	// bytes.
	if len(privBytes) < 32 {
		privBytes = append(make([]byte, 32-len(privBytes)), privBytes...)
	}
	return crypto.UnmarshalSecp256k1PrivateKey(privBytes)
}

// ConvertFromInterfacePubKey converts crypto.PubKey to ecdsa.PublicKey
func ConvertFromInterfacePubKey(pubKey crypto.PubKey) *ecdsa.PublicKey {
	pk := btcec.PublicKey(*(pubKey.(*crypto.Secp256k1PublicKey)))
	return pk.ToECDSA()
}

// ConvertToInterfacePubkey converts ecdsa.PublicKey to crypto.PubKey
func ConvertToInterfacePubkey(pubkey *ecdsa.PublicKey) (crypto.PubKey, error) {
	xVal, yVal := new(btcec.FieldVal), new(btcec.FieldVal)
	if xVal.SetByteSlice(pubkey.X.Bytes()) {
		return nil, errors.Errorf("X value overflows")
	}
	if yVal.SetByteSlice(pubkey.Y.Bytes()) {
		return nil, errors.Errorf("Y value overflows")
	}
	newKey := crypto.PubKey((*crypto.Secp256k1PublicKey)(btcec.NewPublicKey(xVal, yVal)))
	// Zero out temporary values.
	xVal.Zero()
	yVal.Zero()
	return newKey, nil
}

// GenNetworkKey generates a new network key
func GenNetworkKey() (*ecdsa.PrivateKey, error) {
	privInterfaceKey, _, err := crypto.GenerateSecp256k1Key(crand.Reader)
	if err != nil {
		return nil, errors.WithMessage(err, "could not generate 256k1 key")
	}
	return ConvertFromInterfacePrivKey(privInterfaceKey)
}
