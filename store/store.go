package store

import (
	"os"
	"path/filepath"
	"crypto/sha1"
	"io"
	"fmt"
)

type Store struct {
	root string
	Err error
}

func New(rootDir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(rootDir, "objects"), os.ModePerm); err != nil {
		return nil, err
	}

	return &Store{
		root: rootDir,
	}, nil
}

func (s *Store) SaveBlob(srcPath string) {
	if s.Err != nil {
		return
	}

	dir, name, err := s.objectPath(srcPath)
	if err != nil {
		s.Err = err
		return
	}

	if s.Err = os.MkdirAll(filepath.Join(s.root, "objects", dir), os.ModePerm); s.Err != nil {
		return
	}

	s.Err = os.Link(srcPath, filepath.Join(s.root, "objects", dir, name))
}

func (s *Store) objectPath(srcPath string) (string, string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", "", err
	}

	str := fmt.Sprintf("%x", h.Sum(nil))
	return str[:2], str[2:], nil
}