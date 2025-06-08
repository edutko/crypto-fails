package blob

import (
	"io"

	"github.com/edutko/crypto-fails/pkg/blob"
)

type Store interface {
	GetObject(key string) (io.ReadCloser, blob.Metadata, error)
	PutObject(key string) (io.WriteCloser, error)
	DeleteObject(key string) (blob.Metadata, error)
	ListObjects() ([]blob.Metadata, error)
	ListObjectsWithPrefix(prefix string) ([]blob.Metadata, error)
	io.Closer
}
