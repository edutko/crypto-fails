package kv

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/store"
)

func TestOpen(t *testing.T) {
	s, err := Open[string]("/tmp/n0nExist3Nt Direc70ry!/should-fail")
	assert.NotNil(t, err)
	assert.Nil(t, s)

	f := tmpFile(t)
	if err := os.WriteFile(f, []byte("not JSON"), 0755); err != nil {
		t.Failed()
	}
	s, err = Open[string](f)
	assert.NotNil(t, err)
	assert.Nil(t, s)
}

func TestOpenCloseOpen(t *testing.T) {
	f := tmpFile(t)
	s, _ := Open[string](f)
	_ = s.Put("key", "value")
	_ = s.Put("key2", "value2")
	_ = s.Put("another key", "foobar")
	_, _ = s.Delete("key2")
	assert.NoError(t, s.Close())

	s, err := Open[string](f)
	defer s.Close()

	assert.NoError(t, err)
	v, _ := s.Get("key")
	assert.Equal(t, "value", v)
	v, _ = s.Get("another key")
	assert.Equal(t, "foobar", v)
	v, err = s.Get("key2")
	assert.Equal(t, "", v)
	assert.Equal(t, store.ErrNotFound, err)
}

func TestNewMemoryStore(t *testing.T) {
	s := NewInMemoryStore[string]()
	defer s.Close()
	_ = s.Put("key", "value")

	v1, err1 := s.Get("key")
	v2, err2 := s.Get("not present")

	assert.NoError(t, err1)
	assert.Equal(t, "value", v1)
	assert.Equal(t, store.ErrNotFound, err2)
	assert.Equal(t, "", v2)
}

func TestFSStore_Get(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	v1, err1 := s.Get("key")
	v2, err2 := s.Get("not present")

	assert.NoError(t, err1)
	assert.Equal(t, "value", v1)
	assert.Equal(t, store.ErrNotFound, err2)
	assert.Equal(t, "", v2)
}

func TestFSStore_Exists(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	assert.True(t, s.Exists("key"))
	assert.False(t, s.Exists("not present"))
}

func TestFSStore_Put(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	err1 := s.Put("key", "new value")
	err2 := s.Put("new key", "foobar")

	assert.NoError(t, err1)
	v, _ := s.Get("key")
	assert.Equal(t, "new value", v)

	assert.NoError(t, err2)
	v, _ = s.Get("new key")
	assert.Equal(t, "foobar", v)
}

func TestFSStore_PutIfNotExists(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	inserted1, err1 := s.PutIfNotExists("key", "new value")
	inserted2, err2 := s.PutIfNotExists("new key", "foobar")

	assert.NoError(t, err1)
	assert.False(t, inserted1)
	v, _ := s.Get("key")
	assert.Equal(t, "value", v)

	assert.NoError(t, err2)
	assert.True(t, inserted2)
	v, _ = s.Get("new key")
	assert.Equal(t, "foobar", v)
}

func TestFSStore_Delete(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	deleted1, err1 := s.Delete("key")
	deleted2, err2 := s.Delete("nonexistent")

	assert.True(t, deleted1)
	assert.NoError(t, err1)
	v, err := s.Get("key")
	assert.Equal(t, "", v)
	assert.Equal(t, store.ErrNotFound, err)

	assert.False(t, deleted2)
	assert.NoError(t, err2)
	v, err = s.Get("nonexistent")
	assert.Equal(t, "", v)
	assert.Equal(t, store.ErrNotFound, err)
}

func TestFSStore_ListKeys(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key1", "value1")
	_ = s.Put("key3", "value3")
	_ = s.Put("key2", "value2")

	keys, err := s.ListKeys()

	assert.NoError(t, err)
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
}

func TestFSStore_ListKeysWithPrefix(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key1", "value1")
	_ = s.Put("key3", "value3")
	_ = s.Put("key2", "value2")
	_ = s.Put("another key", "foobar")

	keys, err := s.ListKeysWithPrefix("key")

	assert.NoError(t, err)
	assert.Equal(t, []string{"key1", "key2", "key3"}, keys)
}

func TestFSStore_Update(t *testing.T) {
	s, _ := Open[string](tmpFile(t))
	defer s.Close()
	_ = s.Put("key", "value")

	err := s.Update("key", func(s string) (string, error) {
		return "new value", nil
	})

	assert.NoError(t, err)
	v, _ := s.Get("key")
	assert.Equal(t, "new value", v)

	err = s.Update("nonexistent", func(s string) (string, error) {
		return "new value", nil
	})

	assert.Equal(t, err, store.ErrNotFound)
	v, err = s.Get("nonexistent")
	assert.Equal(t, "", v)
	assert.Equal(t, err, store.ErrNotFound)

	expectedErr := errors.New("test")
	err = s.Update("key", func(s string) (string, error) {
		return "oopsie", expectedErr
	})

	assert.Equal(t, err, expectedErr)
	v, err = s.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "new value", v)
}

func tmpFile(t *testing.T) string {
	return filepath.Join(t.TempDir(), "store.json")
}
