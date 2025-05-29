package kv

import "io"

type Store[T any] interface {
	Exists(key string) bool
	Get(key string) (T, error)
	Put(key string, value T) error
	PutIfNotExists(key string, value T) (bool, error)
	Delete(key string) (bool, error)
	ListKeys() ([]string, error)
	ListKeysWithPrefix(prefix string) ([]string, error)
	Update(key string, mutate func(v T) (T, error)) error
	io.Closer
}
