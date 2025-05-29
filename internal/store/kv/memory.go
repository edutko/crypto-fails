package kv

import (
	"slices"
	"strings"
	"sync"

	"github.com/edutko/crypto-fails/internal/store"
)

type MemoryStore[T any] struct {
	m    map[string]T
	lock sync.RWMutex
}

func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{m: make(map[string]T)}
}

func (s *MemoryStore[T]) Exists(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, present := s.m[key]
	return present
}

func (s *MemoryStore[T]) Get(key string) (T, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	v, present := s.m[key]
	if !present {
		return v, store.ErrNotFound
	}
	return s.m[key], nil
}

func (s *MemoryStore[T]) Put(key string, value T) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = value
	return nil
}

func (s *MemoryStore[T]) PutIfNotExists(key string, value T) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, present := s.m[key]
	if !present {
		s.m[key] = value
		return true, nil
	}
	return false, nil
}

func (s *MemoryStore[T]) Delete(key string) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, present := s.m[key]
	delete(s.m, key)
	return present, nil
}

func (s *MemoryStore[T]) ListKeys() ([]string, error) {
	return s.ListKeysWithPrefix("")
}

func (s *MemoryStore[T]) ListKeysWithPrefix(prefix string) ([]string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var ks []string
	for k := range s.m {
		if strings.HasPrefix(k, prefix) {
			ks = append(ks, k)
		}
	}
	slices.Sort(ks)

	return ks, nil
}

func (s *MemoryStore[T]) Close() error {
	return nil
}
