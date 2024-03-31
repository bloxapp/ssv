// go:build linux
package keys

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	"github.com/stretchr/testify/require"

	"testing"
)

func Test_VerifyRegularSigWithOpenSSL(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	msg := []byte("hello")
	hashed := sha256.Sum256(msg)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	require.NoError(t, err)

	pk := &privateKey{key, nil}
	pub := pk.Public().(*publicKey)

	require.NoError(t, VerifyRSA(pub, msg, sig))
}

func Test_VerifyOpenSSLWithOpenSSL(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	msg := []byte("hello")
	priv := &privateKey{key, nil}
	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	pub := priv.Public().(*publicKey)

	require.NoError(t, VerifyRSA(pub, msg, sig))
}

func Test_Caches(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	msg := []byte("hello")
	priv := &privateKey{key, nil}
	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	pub := priv.Public().(*publicKey)

	require.NoError(t, VerifyRSA(pub, msg, sig))

	// should sign using cache
	require.NotNil(t, priv.cachedPrivKey)

	sig2, err := priv.Sign(msg)
	require.NoError(t, err)

	require.NotNil(t, pub.cachedPubkey)

	require.NoError(t, VerifyRSA(pub, msg, sig2))
}
