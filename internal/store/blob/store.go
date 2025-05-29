package blob

import "io"

type Store interface {
	GetObject(key string) (io.ReadCloser, Metadata, error)
	PutObject(key string) (io.WriteCloser, error)
	PutObjectIfNotExists(key string) (io.WriteCloser, error)
	DeleteObject(key string) (Metadata, error)
	ListObjects(prefix string) ([]Metadata, error)
	io.Closer
}
