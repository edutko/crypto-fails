package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/edutko/crypto-fails/internal/crypto/random"
)

// CTREncrypter returns an io.Writer that encrypts using the provided key and
// iv and writes the ciphertext to w.
//
// If iv is null, CTREncrypter will generate a random IV and prepend it to the
// ciphertext.
func CTREncrypter(k, iv []byte, w io.Writer) io.WriteCloser {
	c, err := aes.NewCipher(k)
	if err != nil {
		panic(err)
	}

	if iv == nil {
		iv = random.Bytes(aes.BlockSize)
		_, err = w.Write(iv)
		if err != nil {
			panic(err)
		}
	}
	s := cipher.NewCTR(c, iv)

	return cipher.StreamWriter{S: s, W: w}
}

// CTRDecrypter returns an io.Reader that reads from r and decrypts using the
// provided key and iv.
//
// If iv is null, CTRDecrypter reads it from the first aes.BlockSize bytes of
// r.
func CTRDecrypter(k, iv []byte, r io.Reader) io.Reader {
	c, err := aes.NewCipher(k)
	if err != nil {
		panic(err)
	}

	if iv == nil {
		iv = make([]byte, aes.BlockSize)
		if _, err = io.ReadFull(r, iv); err != nil {
			panic(err)
		}
	}
	s := cipher.NewCTR(c, iv)

	return cipher.StreamReader{S: s, R: r}
}
