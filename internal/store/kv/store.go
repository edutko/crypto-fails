package kv

import "io"

type Store[T any] interface {
	Get(key string) (T, error)
	Put(key string, value T) error
	Exists(key string) bool
	PutIfNotExists(key string, value T) (bool, error)
	Delete(key string) (bool, error)
	ListKeys() ([]string, error)
	ListKeysWithPrefix(prefix string) ([]string, error)
	io.Closer
}
