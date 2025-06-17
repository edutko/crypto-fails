package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsecureSignASN1(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	h := sha256.Sum256([]byte("squeamish ossifrage"))

	sig, err := InsecureSignASN1(sk, h[:])

	assert.NoError(t, err)
	assert.NotNil(t, sig)
	assert.True(t, ecdsa.VerifyASN1(&sk.PublicKey, h[:], sig))

	sig2, err := InsecureSignASN1(sk, h[:])

	assert.NoError(t, err)
	assert.Equal(t, sig, sig2)
}
