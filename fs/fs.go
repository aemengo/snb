package fs

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Object struct {
	Path string
	Sha  string
}

func Exists(path string) bool {
	//TODO support globbing
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func GetSrcFiles(step string) ([]Object, error) {
	var objects []Object

	for _, item := range strings.Split(step, " ") {
		element := strings.TrimSpace(item)

		if Exists(element) {
			obj, err := objectFrom(element)
			if err != nil {
				return nil, err
			}

			objects = append(objects, obj)
			continue
		}

		goPathElement := filepath.Join("src", element)
		if Exists(goPathElement) {
			obj, err := objectFrom(element)
			if err != nil {
				return nil, err
			}

			objects = append(objects, obj)
		}
	}

	return objects, nil
}

func objectFrom(path string) (Object, error) {
	sha, err := getSha(path)
	if err != nil {
		return Object{}, err
	}

	return Object{
		Path: path,
		Sha:  sha,
	}, nil
}

func getSha(srcPath string) (string, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return shaDir(srcPath)
	} else {
		return shaFile(srcPath)
	}
}

func shaDir(srcPath string) (string, error) {
	h := sha1.New()

	filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || isHiddenFile(path) {
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

func shaFile(srcPath string) (string, error) {
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

func isHiddenFile(path string) bool {
	names := strings.Split(path, string(os.PathSeparator))
	for _, element := range names {
		if strings.HasPrefix(element, ".") {
			return true
		}
	}

	return false
}
