package encrypted

import (
	"crypto/aes"
	"crypto/sha256"
	"io"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/random"
	"github.com/edutko/crypto-fails/internal/store/blob"
	"github.com/edutko/crypto-fails/internal/store/kv"
)

type ObjectStore struct {
	b blob.Store
	k kv.Store[[]byte]
}

func NewObjectStore(b blob.Store, k kv.Store[[]byte]) (*ObjectStore, error) {
	return &ObjectStore{b, k}, nil
}

func (s *ObjectStore) NewEncryptingWriter(kid string, w io.WriteCloser) (io.Writer, error) {
	k, err := s.getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))
	return crypto.CTREncrypter(k, iv[:aes.BlockSize], w), err
}

func (s *ObjectStore) PutObject(key string) (io.WriteCloser, error) {
	return s.b.PutObject(key)
}

func (s *ObjectStore) PutObjectIfNotExists(key string) (io.WriteCloser, error) {
	return s.b.PutObjectIfNotExists(key)
}

func (s *ObjectStore) NewDecryptingReader(kid string, r io.ReadCloser) (io.Reader, error) {
	k, err := s.getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))
	return crypto.CTRDecrypter(k, iv[:aes.BlockSize], r), err
}

func (s *ObjectStore) GetObject(key string) (io.ReadCloser, blob.Metadata, error) {
	return s.b.GetObject(key)
}

func (s *ObjectStore) DeleteObject(key string) (blob.Metadata, error) {
	return s.b.DeleteObject(key)
}

func (s *ObjectStore) ListObjects(prefix string) ([]blob.Metadata, error) {
	return s.b.ListObjects(prefix)
}

func (s *ObjectStore) Close() error {
	_ = s.k.Close()
	return s.b.Close()
}

func (s *ObjectStore) getOrCreateKey(kid string) ([]byte, error) {
	if k, err := s.k.Get(kid); err == nil {
		return k, nil
	}

	if _, err := s.k.PutIfNotExists(kid, random.Bytes(32)); err != nil {
		return nil, err
	}

	if k, err := s.k.Get(kid); err != nil {
		return nil, err
	} else {
		return k, nil
	}
}
