package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/edutko/crypto-fails/internal/crypto/random"
)

const (
	GCMNonceSize = 12
	GCMTagSize   = 16
)

// GCMEncrypter returns an io.WriteCloser that encrypts using the provided key
// and nonce and writes the ciphertext to w. The ciphertext and tag are not
// written until Close() is called.
//
// If nonce is null, GCMEncrypter will generate a random nonce and prepend it
// to the ciphertext.
func GCMEncrypter(k, nonce []byte, w io.Writer) io.WriteCloser {
	return newGCMStream(k, nonce, nil, w)
}

// GCMDecrypter returns an io.Reader that reads from r and decrypts using the
// provided key and nonce.
//
// If nonce is null, GCMDecrypter reads it from the first GCMNonceSize bytes of r.
func GCMDecrypter(k, nonce []byte, r io.Reader) io.Reader {
	return newGCMStream(k, nonce, r, nil)
}

type gcmStream struct {
	buf   bytes.Buffer
	r     io.Reader
	w     io.Writer
	gcm   cipher.AEAD
	nonce []byte
}

func newGCMStream(key, nonce []byte, r io.Reader, w io.Writer) *gcmStream {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	return &gcmStream{
		r:     r,
		w:     w,
		gcm:   gcm,
		nonce: nonce,
	}
}

func (s *gcmStream) Read(p []byte) (int, error) {
	if s.buf.Len() == 0 {
		if s.nonce == nil {
			s.nonce = make([]byte, GCMNonceSize)
			if _, err := io.ReadFull(s.r, s.nonce); err != nil {
				return 0, err
			}
		}
		ciphertext, err := io.ReadAll(s.r)
		if err != nil {
			return 0, err
		}
		plaintext, err := s.gcm.Open(nil, s.nonce, ciphertext, nil)
		if err != nil {
			return 0, err
		}
		s.buf.Write(plaintext)
	}

	return s.buf.Read(p)
}

func (s *gcmStream) Write(p []byte) (n int, err error) {
	return s.buf.Write(p)
}

func (s *gcmStream) Close() error {
	if s.w != nil {
		if s.nonce == nil {
			s.nonce = random.Bytes(GCMNonceSize)
			_, err := s.w.Write(s.nonce)
			if err != nil {
				return err
			}
		}
		ciphertext := s.gcm.Seal(nil, s.nonce, s.buf.Bytes(), nil)
		_, err := s.w.Write(ciphertext)
		return err
	}
	return nil
}
