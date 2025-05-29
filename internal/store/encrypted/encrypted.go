package encrypted

import (
	"crypto/aes"
	"crypto/sha256"
	"io"
	"path"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	"github.com/edutko/crypto-fails/internal/store/blob"
	"github.com/edutko/crypto-fails/internal/store/constants"
	"github.com/edutko/crypto-fails/internal/store/kv"
)

type ObjectStore struct {
	blobs blob.Store
	keys  kv.Store[[]byte]
	mode  crypto.Mode
}

func NewObjectStore(blobs blob.Store, keys kv.Store[[]byte], mode crypto.Mode) (*ObjectStore, error) {
	return &ObjectStore{blobs, keys, mode}, nil
}

func (s *ObjectStore) NewEncryptingWriter(kid string, w io.WriteCloser) (io.WriteCloser, error) {
	k, err := s.getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))

	switch s.mode {
	case crypto.ModeGCM:
		return crypto.GCMEncrypter(k, iv[:crypto.GCMNonceSize], w), err
	default:
		return crypto.CTREncrypter(k, iv[:aes.BlockSize], w), err
	}
}

func (s *ObjectStore) PutObject(key string) (io.WriteCloser, error) {
	return s.blobs.PutObject(key)
}

func (s *ObjectStore) NewDecryptingReader(kid string, r io.ReadCloser) (io.Reader, error) {
	k, err := s.getOrCreateKey(kid)
	iv := sha256.Sum256([]byte(kid))

	switch s.mode {
	case crypto.ModeGCM:
		return crypto.GCMDecrypter(k, iv[:crypto.GCMNonceSize], r), err
	default:
		return crypto.CTRDecrypter(k, iv[:aes.BlockSize], r), err
	}
}

func (s *ObjectStore) GetObject(key string) (io.ReadCloser, blob.Metadata, error) {
	return s.blobs.GetObject(key)
}

func (s *ObjectStore) DeleteObject(key string) (blob.Metadata, error) {
	return s.blobs.DeleteObject(key)
}

func (s *ObjectStore) ListObjectsWithPrefix(prefix string) ([]blob.Metadata, error) {
	m, err := s.blobs.ListObjectsWithPrefix(prefix)
	if err != nil {
		return m, err
	}

	if s.mode == crypto.ModeGCM {
		for i := range m {
			m[i].Size = m[i].Size - crypto.GCMTagSize
		}
	}
	return m, nil
}

func (s *ObjectStore) Close() error {
	_ = s.keys.Close()
	return s.blobs.Close()
}

func (s *ObjectStore) getOrCreateKey(kid string) ([]byte, error) {
	kid = path.Join(constants.BlobStoreKIDPrefix, kid)

	if k, err := s.keys.Get(kid); err == nil {
		return k, nil
	}

	if _, err := s.keys.PutIfNotExists(kid, random.Bytes(32)); err != nil {
		return nil, err
	}

	if k, err := s.keys.Get(kid); err != nil {
		return nil, err
	} else {
		return k, nil
	}
}
