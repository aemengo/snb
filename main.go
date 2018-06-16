package main

import (
	"errors"
	str "github.com/aemengo/snb/store"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Spec struct {
	Steps []string
}

func main() {
	store, err := str.New(".snb")
	if err != nil {
		log.Fatal("Error: ", err, ".")
	}

	spec, err := getSpec()
	if err != nil {
		log.Fatal("Error: ", err, ".")
	}

	blobList := getBlobList()
	for index, step := range spec.Steps {
		err = executeStep(step)
		if err != nil {
			log.Fatal("Error: ", err, ".")
		}

		bl := getBlobList()
		modifiedFiles := getModifiedFiles(blobList, bl)
		store.SaveStep(step, index, modifiedFiles)
		blobList = bl
	}
}

func getModifiedFiles(oldBlobList, newBlobList map[string]int64) []string {
	var results []string

	for path, size := range newBlobList {
		if oldBlobList[path] != size {
			results = append(results, path)
		}
	}

	return results
}

func getBlobList() map[string]int64 {
	var blobList = make(map[string]int64)

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() ||
			isHiddenFile(path) ||
			isIgnoredFile(path) {
			return nil
		}

		blobList[path] = info.Size()

		return nil
	})

	return blobList
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

func isIgnoredFile(path string) bool {
	var blackList = []string{
		"ShakeAndBakeFile",
	}

	for _, entry := range blackList {
		if filepath.Base(path) == entry {
			return true
		}
	}

	return false
}

func executeStep(step string) error {
	command := exec.Command("bash", "-c", step)
	command.Env = os.Environ()

	_, err := command.Output()
	if err != nil {
		return err
	}

	return nil
}

func getSpec() (Spec, error) {
	path := "ShakeAndBakeFile"

	if !exists(path) {
		return Spec{}, errors.New("ShakeAndBakeFile not found. please execute in directory containing spec, or pass the working directory in as the only argument")
	}

	specFile, err := ioutil.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}

	stepRegex := `RUN\s(.*)`
	matches := regexp.MustCompile(stepRegex).FindAllStringSubmatch(string(specFile), -1)

	if len(matches) == 0 {
		//TODO something or no-op
	}

	var spec Spec
	for _, match := range matches {
		spec.Steps = append(spec.Steps, match[1])
	}

	return spec, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
