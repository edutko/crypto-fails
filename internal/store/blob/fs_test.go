package blob

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edutko/crypto-fails/internal/store"
)

func TestFSStore_PutObject(t *testing.T) {
	tmpDir, s := setup(t)
	defer s.Close()

	testCases := []struct {
		name         string
		key          string
		expectedPath string
		expectedErr  error
		data         []byte
	}{
		{"empty", "zero/empty.txt", filepath.Join(tmpDir, "zero", "empty.txt"), nil, []byte{}},
		{"small file", "zero/small.dat", filepath.Join(tmpDir, "zero", "small.dat"), nil, []byte("Hello, world!")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w, err := s.PutObject(tc.key)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				_, err = w.Write(tc.data)
				assert.NoError(t, err)
				_ = w.Close()

				b, _ := os.ReadFile(tc.expectedPath)
				assert.Equal(t, tc.data, b)

			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFSStore_DeleteObject(t *testing.T) {
	tmpDir, s := setup(t)
	defer s.Close()

	testCases := []struct {
		name         string
		key          string
		expectedPath string
		data         []byte
		expectedErr  error
	}{
		{"empty", "one/nothing.md", filepath.Join(tmpDir, "one", "nothing.md"), []byte{}, nil},
		{"small file", "one/other.doc", filepath.Join(tmpDir, "one", "other.doc"), []byte("Hello, world!"), nil},
		{"nonexistent", "one/does not exist", "", nil, nil},
		{"empty name", "", "", nil, store.ErrInvalidKey},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createBlobForTest(t, s, tc.key, tc.data)
			if tc.data != nil {
				assert.True(t, fileExists(tc.expectedPath))
			}

			m, err := s.DeleteObject(tc.key)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.False(t, fileExists(tc.expectedPath))
				if tc.data != nil {
					assert.Equal(t, tc.key, m.Key)
					assert.Equal(t, int64(len(tc.data)), m.Size)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFSStore_GetObject(t *testing.T) {
	_, s := setup(t)
	defer s.Close()

	testCases := []struct {
		name        string
		key         string
		expectedErr error
		data        []byte
	}{
		{"empty", "two/empty.dat", nil, []byte{}},
		{"small file", "two/hi.txt", nil, []byte("Hello, world!")},
		{"nonexistent", "two/does not exist", store.ErrNotFound, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createBlobForTest(t, s, tc.key, tc.data)

			r, m, err := s.GetObject(tc.key)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				b, err := io.ReadAll(r)
				assert.NoError(t, err)
				r.Close()

				assert.Equal(t, tc.data, b)
				assert.Equal(t, tc.key, m.Key)
				assert.Equal(t, int64(len(tc.data)), m.Size)

			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestFSStore_ListObjects(t *testing.T) {
	_, s := setup(t)
	defer s.Close()

	createBlobForTest(t, s, "foo/a.txt", []byte{})
	createBlobForTest(t, s, "foo/b.txt", []byte{})
	createBlobForTest(t, s, "foo/bar/a.txt", []byte{})
	createBlobForTest(t, s, "foo/bar/baz/123.dat", []byte{})
	createBlobForTest(t, s, "bar/a.txt", []byte{})

	ms, err := s.ListObjects()
	assert.NoError(t, err)
	for i, p := range []string{"bar/a.txt", "foo/a.txt", "foo/b.txt", "foo/bar/a.txt", "foo/bar/baz/123.dat"} {
		assert.Equal(t, ms[i].Key, p)
	}

}

func TestFSStore_ListObjectsWithPrefix(t *testing.T) {
	_, s := setup(t)
	defer s.Close()

	createBlobForTest(t, s, "foo/a.txt", []byte{})
	createBlobForTest(t, s, "foo/b.txt", []byte{})
	createBlobForTest(t, s, "foo/bar/a.txt", []byte{})
	createBlobForTest(t, s, "foo/bar/baz/123.dat", []byte{})
	createBlobForTest(t, s, "foobar/abc.dat", []byte{})
	createBlobForTest(t, s, "bar/a.txt", []byte{})

	ms, err := s.ListObjectsWithPrefix("foo")
	assert.NoError(t, err)
	for i, p := range []string{"foo/a.txt", "foo/b.txt", "foo/bar/a.txt", "foo/bar/baz/123.dat", "foobar/abc.dat"} {
		assert.Equal(t, p, ms[i].Key)
	}
}

func setup(t *testing.T) (string, *FSStore) {
	tmpDir := t.TempDir()
	s, err := Open(tmpDir)
	assert.NoError(t, err)
	return tmpDir, s
}

func createBlobForTest(t *testing.T, s *FSStore, key string, data []byte) {
	if data == nil {
		return
	}

	w, err := s.PutObject(key)
	assert.NoError(t, err)

	_, err = w.Write(data)
	assert.NoError(t, err)
}

func fileExists(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}
