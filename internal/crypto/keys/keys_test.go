package keys

import (
	"bytes"
	"crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateECDSAKeyPair(t *testing.T) {
	k, priv, pub, err := GenerateECDSAKeyPair()

	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Implements(t, (*crypto.PrivateKey)(nil), k)
	assert.Implements(t, (*crypto.Signer)(nil), k)
	assert.True(t, bytes.HasPrefix(priv, []byte("-----BEGIN PRIVATE KEY-----")))
	assert.True(t, bytes.HasPrefix(pub, []byte("-----BEGIN PUBLIC KEY-----")))
}

func TestGenerateRSAKeyPair(t *testing.T) {
	k, priv, pub, err := GenerateRSAKeyPair()

	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Implements(t, (*crypto.PrivateKey)(nil), k)
	assert.Implements(t, (*crypto.Signer)(nil), k)
	assert.True(t, bytes.HasPrefix(priv, []byte("-----BEGIN PRIVATE KEY-----")))
	assert.True(t, bytes.HasPrefix(pub, []byte("-----BEGIN PUBLIC KEY-----")))
}

func TestParsePublicKeyPEM(t *testing.T) {
	rsaKey, _, rsaPub, err := GenerateRSAKeyPair()
	assert.NoError(t, err)
	ecKey, _, ecPub, err := GenerateRSAKeyPair()
	assert.NoError(t, err)

	pubKey, err := ParsePublicKeyPEM(rsaPub)
	assert.NoError(t, err)
	assert.Equal(t, rsaKey.Public(), pubKey)

	pubKey, err = ParsePublicKeyPEM(ecPub)
	assert.NoError(t, err)
	assert.Equal(t, ecKey.Public(), pubKey)
}
