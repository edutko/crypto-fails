package middleware

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type DirWithHiddenFiles string

func (d DirWithHiddenFiles) Open(name string) (fs.File, error) {
	p := path.Clean("/" + name)[1:]
	if p == "" {
		p = "."
	}
	if p != "." {
		for _, part := range strings.Split(p, "/") {
			if strings.HasPrefix(part, ".") {
				return nil, fs.ErrNotExist
			}
		}
	}

	p, err := filepath.Localize(p)
	if err != nil {
		return nil, errors.New("http: invalid or unsafe file path")
	}
	dir := string(d)
	if dir == "" {
		dir = "."
	}

	fullName := filepath.Join(dir, p)
	f, err := os.Open(fullName)
	if err != nil {
		return nil, mapOpenError(err, fullName, filepath.Separator, os.Stat)
	}

	return file{f}, nil
}

func mapOpenError(originalErr error, name string, sep rune, stat func(string) (fs.FileInfo, error)) error {
	if errors.Is(originalErr, fs.ErrNotExist) || errors.Is(originalErr, fs.ErrPermission) {
		return originalErr
	}

	parts := strings.Split(name, string(sep))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		fi, err := stat(strings.Join(parts[:i+1], string(sep)))
		if err != nil {
			return originalErr
		}
		if !fi.IsDir() {
			return fs.ErrNotExist
		}
	}
	return originalErr
}

type file struct {
	f *os.File
}

func (f file) Stat() (fs.FileInfo, error) {
	return f.f.Stat()
}

func (f file) Read(b []byte) (int, error) {
	return f.f.Read(b)
}

func (f file) Close() error {
	return f.f.Close()
}

func (f file) ReadDir(n int) ([]fs.DirEntry, error) {
	dirs, err := f.f.ReadDir(n)
	filtered := make([]fs.DirEntry, 0, len(dirs))
	for _, de := range dirs {
		if !strings.HasPrefix(de.Name(), ".") {
			filtered = append(filtered, de)
		}
	}

	slices.SortFunc(filtered, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	return filtered, err
}
