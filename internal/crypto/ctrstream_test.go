package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/crypto/random"
)

func TestCTREncrypter_randomIV(t *testing.T) {
	plaintext := []byte("hello, world!")
	var w bytes.Buffer

	e := CTREncrypter(key, nil, &w)
	_, err := e.Write(plaintext)
	assert.NoError(t, err)

	assert.Len(t, w.Bytes(), len(iv)+len(plaintext))
	assert.Equal(t, plaintext, ctr(w.Bytes()[:len(iv)], w.Bytes()[len(iv):]))
}

func TestCTREncrypter_providedIV(t *testing.T) {
	plaintext := []byte("hello, world!")
	var w bytes.Buffer

	e := CTREncrypter(key, iv, &w)
	_, err := e.Write(plaintext)
	assert.NoError(t, err)

	assert.Len(t, w.Bytes(), len(plaintext))
	assert.Equal(t, plaintext, ctr(iv, w.Bytes()))
}

func TestCTRDecrypter_randomIV(t *testing.T) {
	r := bytes.NewReader(append(iv, ctr(iv, []byte("hello, world!"))...))
	plaintext := make([]byte, len("hello, world!"))

	e := CTRDecrypter(key, nil, r)
	_, err := e.Read(plaintext)

	assert.NoError(t, err)
	assert.Equal(t, "hello, world!", string(plaintext))
}

func TestCTRDecrypter_providedIV(t *testing.T) {
	r := bytes.NewReader(ctr(iv, []byte("hello, world!")))
	plaintext := make([]byte, len("hello, world!"))

	e := CTRDecrypter(key, iv, r)
	_, err := e.Read(plaintext)

	assert.NoError(t, err)
	assert.Equal(t, "hello, world!", string(plaintext))
}

var key = random.Bytes(16)
var iv = random.Bytes(aes.BlockSize)

func ctr(iv, in []byte) []byte {
	c, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	m := cipher.NewCTR(c, iv)

	out := make([]byte, len(in))
	m.XORKeyStream(out, in)

	return out
}
