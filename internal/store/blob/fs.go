package blob

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/pkg/blob"
)

type FSStore struct {
	r *os.Root
}

func Open(rootDir string) (*FSStore, error) {
	if root, err := os.OpenRoot(rootDir); err != nil {
		return nil, err
	} else {
		return &FSStore{root}, nil
	}
}

func (s *FSStore) Close() error {
	return s.r.Close()
}

func (s *FSStore) PutObject(key string) (io.WriteCloser, error) {
	file, err := keyToPath(key)
	if err != nil {
		return nil, err
	}

	if err = s.ensureDirectories(file); err != nil {
		return nil, fmt.Errorf("ensureDirectories: %w", err)
	}

	if f, err := s.createFile(file); err != nil {
		return nil, fmt.Errorf("createFile: %w", err)
	} else {
		return f, nil
	}
}

func (s *FSStore) DeleteObject(key string) (blob.Metadata, error) {
	file, err := keyToPath(key)
	if err != nil {
		return blob.Metadata{}, err
	}

	m, err := s.getMetadata(file)
	if errors.Is(err, store.ErrNotFound) {
		return blob.Metadata{}, nil
	}

	if err := s.recursiveDelete(file); err != nil {
		return m, fmt.Errorf("recursiveDelete: %w", err)
	}

	if err != nil {
		return m, fmt.Errorf("getMetadata: %w", err)
	}
	return m, nil
}

func (s *FSStore) GetObject(key string) (io.ReadCloser, blob.Metadata, error) {
	file, err := keyToPath(key)
	if err != nil {
		return nil, blob.Metadata{}, err
	}

	m, err := s.getMetadata(file)
	if errors.Is(err, store.ErrNotFound) {
		return nil, blob.Metadata{}, store.ErrNotFound
	}

	f, err := s.openFile(file)
	if errors.Is(err, store.ErrNotFound) {
		return nil, blob.Metadata{}, store.ErrNotFound
	}
	if err != nil {
		return nil, blob.Metadata{}, fmt.Errorf("openFile: %w", err)
	}

	return f, m, nil
}

func (s *FSStore) ListObjects() ([]blob.Metadata, error) {
	return s.ListObjectsWithPrefix("")
}

func (s *FSStore) ListObjectsWithPrefix(prefix string) ([]blob.Metadata, error) {
	dir := realPath(path.Join(prefix, ".."))
	m, err := s.recursiveList(dir, prefix)
	if errors.Is(err, store.ErrNotFound) {
		return []blob.Metadata{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("recursiveList: %w", err)
	}
	return m, nil
}

func keyToPath(key string) (realPath, error) {
	if len(key) == 0 || strings.HasPrefix(key, "/") {
		return "", store.ErrInvalidKey
	}
	return realPath(path.Clean(key)), nil
}

func (s *FSStore) createFile(file realPath) (io.WriteCloser, error) {
	if f, err := s.r.Create(string(file)); err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	} else {
		return f, nil
	}
}

func (s *FSStore) ensureDirectories(file realPath) error {
	cleanPath := path.Dir(string(file))
	partialPath := ""
	for _, dir := range strings.Split(cleanPath, "/") {
		partialPath = filepath.Join(partialPath, dir)
		if err := s.r.Mkdir(partialPath, 0755); errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Join(s.r.Name(), partialPath), err)
		}
	}
	return nil
}

func (s *FSStore) getMetadata(file realPath) (blob.Metadata, error) {
	if fi, err := s.r.Stat(string(file)); errors.Is(err, fs.ErrNotExist) {
		return blob.Metadata{}, store.ErrNotFound
	} else if err != nil {
		return blob.Metadata{}, fmt.Errorf("stat %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	} else {
		return blob.Metadata{
			Key:      string(file),
			Size:     fi.Size(),
			Modified: fi.ModTime(),
		}, err
	}
}

func (s *FSStore) openFile(file realPath) (io.ReadCloser, error) {
	if f, err := s.r.Open(string(file)); errors.Is(err, fs.ErrNotExist) {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	} else {
		return f, nil
	}
}

func (s *FSStore) recursiveDelete(file realPath) error {
	if err := s.r.Remove(string(file)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	}

	partialPath := path.Dir(string(file))
	for partialPath != "." {
		if err := s.r.Remove(partialPath); errors.Is(err, fs.ErrNotExist) {
			// it's already gone; keep working our way up the tree
			continue

		} else if errors.Is(err, &fs.PathError{}) && errors.Is(err.(*fs.PathError).Err, syscall.ENOTEMPTY) {
			// this directory contains other files; stop here
			return nil

		} else if err != nil {
			// this directory probably wasn't removed, so there's no sense trying to remove its parent
			return err
		}

		partialPath = path.Dir(partialPath)
	}

	return nil
}

func (s *FSStore) recursiveList(dir realPath, prefix string) ([]blob.Metadata, error) {
	root := string(dir)
	if root == "" || root == ".." {
		root = "."
	}

	var ms []blob.Metadata
	err := fs.WalkDir(s.r.FS(), root, func(pth string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else if err != nil {
			return fmt.Errorf("fs.WalkDir %q: %w", filepath.Join(s.r.Name(), root), err)
		}

		p := append(filepath.SplitList(pth))
		key := path.Join(p...)
		if !strings.HasPrefix(key, prefix) {
			return nil
		}

		if !d.IsDir() {
			if fi, err := d.Info(); errors.Is(err, fs.ErrNotExist) {
				return nil
			} else if err != nil {
				return fmt.Errorf("fs.WalkDir %q: %w", filepath.Join(s.r.Name(), root), err)
			} else {
				ms = append(ms, blob.Metadata{
					Key:      key,
					Size:     fi.Size(),
					Modified: fi.ModTime(),
				})
			}
		}
		return nil
	})

	return ms, err
}

type realPath string
