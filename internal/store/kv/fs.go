package kv

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/edutko/crypto-fails/internal/store"
)

type FSStore[T any] struct {
	m    map[string]T
	f    string
	lock sync.RWMutex
}

func Open[T any](file string) (*FSStore[T], error) {
	m, err := load[T](file)
	if err != nil {
		return nil, err
	}
	// ensure the file can be written to catch obvious errors early
	if err = save(m, file); err != nil {
		return nil, err
	}
	return &FSStore[T]{
		m: m,
		f: file,
	}, nil
}

func NewInMemoryStore[T any]() *FSStore[T] {
	if s, err := Open[T](""); err != nil {
		panic(err)
	} else {
		return s
	}
}

func (s *FSStore[T]) Exists(key string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, present := s.m[key]
	return present
}

func (s *FSStore[T]) Get(key string) (T, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, present := s.m[key]
	if !present {
		return v, store.ErrNotFound
	}
	return s.m[key], nil
}

func (s *FSStore[T]) Put(key string, value T) error {
	m := make(map[string]T, len(s.m))

	s.lock.Lock()
	s.m[key] = value
	maps.Copy(m, s.m)
	s.lock.Unlock()

	return save(m, s.f)
}

func (s *FSStore[T]) PutIfNotExists(key string, value T) (bool, error) {
	m := make(map[string]T, len(s.m))

	s.lock.Lock()
	_, present := s.m[key]
	if !present {
		s.m[key] = value
		maps.Copy(m, s.m)
	}
	s.lock.Unlock()

	if !present {
		err := save(m, s.f)
		return true, err
	}

	return false, nil
}

func (s *FSStore[T]) Delete(key string) (bool, error) {
	m := make(map[string]T, len(s.m))

	s.lock.Lock()
	_, present := s.m[key]
	delete(s.m, key)
	maps.Copy(m, s.m)
	s.lock.Unlock()

	return present, save(m, s.f)
}

func (s *FSStore[T]) ListKeys() ([]string, error) {
	return s.ListKeysWithPrefix("")
}

func (s *FSStore[T]) ListKeysWithPrefix(prefix string) ([]string, error) {
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

func (s *FSStore[T]) Update(key string, mutate func(v T) (T, error)) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, present := s.m[key]
	if !present {
		return store.ErrNotFound
	}

	if updated, err := mutate(v); err != nil {
		return err
	} else {
		s.m[key] = updated
		return save(s.m, s.f)
	}
}

func (s *FSStore[T]) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := save(s.m, s.f)
	s.m = make(map[string]T)
	return err
}

func load[T any](f string) (map[string]T, error) {
	if f != "" {
		b, err := os.ReadFile(f)
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]T), nil
		}
		if err != nil {
			return nil, fmt.Errorf("os.ReadFile %q: %w", f, err)
		}

		var m map[string]T
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		return m, nil
	}
	return make(map[string]T), nil
}

func save[T any](m map[string]T, f string) error {
	if f != "" {
		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}

		err = os.WriteFile(f, b, 0600)
		if err != nil {
			return fmt.Errorf("os.WriteFile %q: %w", f, err)
		}
	}
	return nil
}
