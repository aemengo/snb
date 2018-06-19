package fs

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"io/ioutil"
)

type Object struct {
	Path string
	Sha  string
}

type FS struct {
	workingDir string
}

func New(workingDir string) (*FS, error) {
	wDir, err := filepath.Abs(workingDir)
	if err != nil {
		return nil, err
	}

	return &FS{
		workingDir: wDir,
	}, nil
}

func (fs *FS) Get(path string) ([]byte, error) {
	return ioutil.ReadFile(fs.p(path))
}

func (fs *FS) Exists(path string) bool {
	_, err := os.Stat(fs.p(path))
	if os.IsNotExist(err) {
		return false
	}

	return true
}

//TODO support globbing
func  (fs *FS) GetSrcFiles(step string) ([]Object, error) {
	var objects []Object

	for _, item := range strings.Split(step, " ") {
		element := strings.TrimSpace(item)

		if fs.Exists(element) {
			obj, err := fs.objectFrom(element)
			if err != nil {
				return nil, err
			}

			objects = append(objects, obj)
			continue
		}

		goPathElement := filepath.Join("src", element)
		if fs.Exists(goPathElement) {
			obj, err := fs.objectFrom(element)
			if err != nil {
				return nil, err
			}

			objects = append(objects, obj)
		}
	}

	return objects, nil
}

func (fs *FS) objectFrom(path string) (Object, error) {
	srcPath := fs.p(path)

	sha, err := fs.getSha(srcPath)
	if err != nil {
		return Object{}, err
	}

	return Object{
		Path: path,
		Sha:  sha,
	}, nil
}

func (fs *FS) getSha(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return fs.shaDir(path)
	} else {
		return fs.shaFile(path)
	}
}

func (fs *FS) shaDir(srcPath string) (string, error) {
	h := sha1.New()

	filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || fs.isHiddenFile(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(h, f)
		return err
	})

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (fs *FS) shaFile(srcPath string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (fs *FS) isHiddenFile(path string) bool {
	names := strings.Split(path, string(os.PathSeparator))
	for _, element := range names {
		if strings.HasPrefix(element, ".") {
			return true
		}
	}

	return false
}

func (fs *FS) p(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else {
		return filepath.Join(fs.workingDir, path)
	}
}