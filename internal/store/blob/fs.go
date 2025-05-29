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
)

type FSStore struct {
	r *os.Root
}

func Open(rootDir string) (*FSStore, error) {
	root, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}
	return &FSStore{
		r: root,
	}, nil
}

func (s *FSStore) Close() error {
	return s.r.Close()
}

func (s *FSStore) GetMetadata(key string) (Metadata, error) {
	file, err := keyToPath(key)
	if err != nil {
		return Metadata{}, err
	}

	m, err := s.getMetadata(file)
	if errors.Is(err, store.ErrNotFound) {
		return Metadata{}, store.ErrNotFound
	}
	if err != nil {
		return Metadata{}, fmt.Errorf("getMetadata: %w", err)
	}

	return m, nil
}

func (s *FSStore) PutObject(key string) (io.WriteCloser, error) {
	file, err := keyToPath(key)
	if err != nil {
		return nil, err
	}

	err = s.ensureDirectories(file)
	if err != nil {
		return nil, fmt.Errorf("ensureDirectories: %w", err)
	}

	f, err := s.createFile(file)
	if err != nil {
		return nil, fmt.Errorf("createFile: %w", err)
	}

	return f, nil
}

func (s *FSStore) PutObjectIfNotExists(key string) (io.WriteCloser, error) {
	_, err := s.GetMetadata(key)
	if errors.Is(err, store.ErrNotFound) {
		return s.PutObject(key)
	}
	return nil, err
}

func (s *FSStore) DeleteObject(key string) (Metadata, error) {
	file, err := keyToPath(key)
	if err != nil {
		return Metadata{}, err
	}

	m, err1 := s.getMetadata(file)
	if errors.Is(err1, store.ErrNotFound) {
		return Metadata{}, nil
	}

	err2 := s.recursiveDelete(file)
	if err2 != nil {
		return m, fmt.Errorf("recursiveDelete: %w", err2)
	}

	if err1 != nil {
		return m, fmt.Errorf("getMetadata: %w", err1)
	}
	return m, nil
}

func (s *FSStore) GetObject(key string) (io.ReadCloser, Metadata, error) {
	file, err := keyToPath(key)
	if err != nil {
		return nil, Metadata{}, err
	}

	m, err := s.getMetadata(file)
	if errors.Is(err, store.ErrNotFound) {
		return nil, Metadata{}, store.ErrNotFound
	}

	f, err := s.openFile(file)
	if errors.Is(err, store.ErrNotFound) {
		return nil, Metadata{}, store.ErrNotFound
	}
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("openFile: %w", err)
	}

	return f, m, nil
}

func (s *FSStore) ListObjects(prefix string) ([]Metadata, error) {
	dir := realPath(path.Clean(prefix))
	m, err := s.recursiveList(dir)
	if errors.Is(err, store.ErrNotFound) {
		return []Metadata{}, nil
	}
	if err != nil {
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
	f, err := s.r.Create(string(file))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	}
	return f, nil
}

func (s *FSStore) ensureDirectories(file realPath) error {
	cleanPath := path.Dir(string(file))
	partialPath := ""
	for _, dir := range strings.Split(cleanPath, "/") {
		partialPath = filepath.Join(partialPath, dir)
		err := s.r.Mkdir(partialPath, 0755)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Join(s.r.Name(), partialPath), err)
		}
	}
	return nil
}

func (s *FSStore) getMetadata(file realPath) (Metadata, error) {
	fi, err := s.r.Stat(string(file))
	if errors.Is(err, fs.ErrNotExist) {
		return Metadata{}, store.ErrNotFound
	}
	if err != nil {
		return Metadata{}, fmt.Errorf("stat %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	}

	return Metadata{
		Key:      string(file),
		Size:     fi.Size(),
		Modified: fi.ModTime(),
	}, err
}

func (s *FSStore) openFile(file realPath) (io.ReadCloser, error) {
	f, err := s.r.Open(string(file))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	}

	return f, nil
}

func (s *FSStore) recursiveDelete(file realPath) error {
	if err := s.r.Remove(string(file)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", filepath.Join(s.r.Name(), string(file)), err)
	}

	partialPath := path.Dir(string(file))
	for partialPath != "." {
		err := s.r.Remove(partialPath)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if errors.Is(err, &fs.PathError{}) && errors.Is(err.(*fs.PathError).Err, syscall.ENOTEMPTY) {
			return nil
		}
		if err != nil {
			return err
		}
		partialPath = path.Dir(partialPath)
	}

	return nil
}

func (s *FSStore) recursiveList(dir realPath) ([]Metadata, error) {
	root := string(dir)
	if root == "" {
		root = "."
	}

	var ms []Metadata
	err := fs.WalkDir(s.r.FS(), root, func(pth string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("fs.WalkDir %q: %w", filepath.Join(s.r.Name(), root), err)
		}
		if !d.IsDir() {
			fi, err := d.Info()
			if err != nil {
				return fmt.Errorf("fs.WalkDir %q: %w", filepath.Join(s.r.Name(), root), err)
			}
			p := append(filepath.SplitList(pth))
			ms = append(ms, Metadata{
				Key:      path.Join(p...),
				Size:     fi.Size(),
				Modified: fi.ModTime(),
			})
		}
		return nil
	})

	return ms, err
}

type realPath string
