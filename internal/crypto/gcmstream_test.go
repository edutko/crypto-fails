package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/crypto/random"
)

func TestGCMEncrypter_randomNonce(t *testing.T) {
	plaintext := []byte("hello, world!")
	var w bytes.Buffer

	e := GCMEncrypter(key, nil, &w)
	_, err := e.Write(plaintext)
	assert.NoError(t, err)
	err = e.Close()
	assert.NoError(t, err)

	assert.Len(t, w.Bytes(), len(plaintext)+len(nonce)+GCMTagSize)
	assert.Equal(t, plaintext, gcmUnseal(w.Bytes()[:GCMNonceSize], w.Bytes()[GCMNonceSize:]))
}

func TestGCMEncrypter_providedNonce(t *testing.T) {
	plaintext := []byte("hello, world!")
	var w bytes.Buffer

	e := GCMEncrypter(key, nonce, &w)
	_, err := e.Write(plaintext)
	assert.NoError(t, err)
	err = e.Close()
	assert.NoError(t, err)

	assert.Len(t, w.Bytes(), len(plaintext)+GCMTagSize)
	assert.Equal(t, plaintext, gcmUnseal(nonce, w.Bytes()))
}

func TestGCMDecrypter_randomNonce(t *testing.T) {
	r := bytes.NewReader(append(nonce, gcmSeal(nonce, []byte("hello, world!"))...))
	plaintext := make([]byte, len("hello, world!"))

	e := GCMDecrypter(key, nil, r)
	_, err := e.Read(plaintext)

	assert.NoError(t, err)
	assert.Equal(t, "hello, world!", string(plaintext))
}

func TestGCMDecrypter_providedIV(t *testing.T) {
	r := bytes.NewReader(gcmSeal(nonce, []byte("hello, world!")))
	plaintext := make([]byte, len("hello, world!"))

	e := GCMDecrypter(key, nonce, r)
	_, err := e.Read(plaintext)

	assert.NoError(t, err)
	assert.Equal(t, "hello, world!", string(plaintext))
}

var nonce = random.Bytes(GCMNonceSize)

func gcmSeal(nonce, in []byte) []byte {
	return newGcm().Seal(nil, nonce, in, nil)
}

func gcmUnseal(nonce, in []byte) []byte {
	out, err := newGcm().Open(nil, nonce, in, nil)
	if err != nil {
		panic(err)
	}
	return out
}

func newGcm() cipher.AEAD {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	m, err := cipher.NewGCM(c)
	if err != nil {
		panic(err)
	}

	return m
}
