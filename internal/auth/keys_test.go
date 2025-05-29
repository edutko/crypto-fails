package auth

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/store/kv"
)

func TestGetCookieEncryptionKey(t *testing.T) {
	ks := kv.NewInMemoryStore[[]byte]()
	keyStore = func() kv.Store[[]byte] { return ks }

	k := GetCookieEncryptionKey()

	assert.NotZero(t, k)
}

func TestGetDefaultJWTSigningKey(t *testing.T) {
	ks := kv.NewInMemoryStore[[]byte]()
	keyStore = func() kv.Store[[]byte] { return ks }

	k := GetDefaultJWTSigningKey()

	assert.NotZero(t, k)
}

func TestGetJWTSigningKey(t *testing.T) {
	ks := kv.NewInMemoryStore[[]byte]()
	keyStore = func() kv.Store[[]byte] { return ks }
	assert.NoError(t, InitializeKeys())

	kids, _ := keyStore().ListKeys()
	assert.Len(t, kids, 2)
	for _, kid := range kids {
		k, ok := GetJWTSigningKey(path.Base(kid))
		assert.NotZero(t, k)
		assert.True(t, ok)
	}
}

func TestGetShareLinkSecret(t *testing.T) {
	ks := kv.NewInMemoryStore[[]byte]()
	keyStore = func() kv.Store[[]byte] { return ks }

	k := GetShareLinkSecret()

	assert.NotZero(t, k)
}
