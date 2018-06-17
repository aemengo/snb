package main

import (
	"bufio"
	"errors"
	str "github.com/aemengo/snb/store"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Spec struct {
	Steps []string
}

var (
	boldWhite = color.New(color.FgWhite, color.Bold)
	boldGreen = color.New(color.FgGreen, color.Bold)
	boldRed   = color.New(color.FgRed, color.Bold)
	white     = color.New(color.FgWhite)
	red       = color.New(color.FgRed)
	logPrefix = " ---> "
)

func main() {
	startTime := time.Now()

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
		boldWhite.Printf("Step %d/%d : %s\n", index+1, len(spec.Steps), step)

		err = executeStep(step)
		if err != nil {
			exitCode, ok := exitCode(err)
			if ok {
				boldRed.Printf("\nBuild failed (exit status: %d)\n", exitCode)
			} else {
				boldRed.Printf("\nBuild failed\n")
			}
			os.Exit(exitCode)
		}

		bl := getBlobList()
		modifiedFiles := getModifiedFiles(blobList, bl)
		store.SaveStep(step, index, modifiedFiles)
		blobList = bl
	}

	endTime := time.Now()

	boldGreen.Printf("\nBuild completed (%f seconds)\n", endTime.Sub(startTime).Seconds())
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
	white.Println(logPrefix + "Running")
	command := exec.Command("bash", "-c", step)
	command.Env = os.Environ()

	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	err := command.Start()
	if err != nil {
		return err
	}

	go report(stdout, white)
	go report(stderr, red)

	return command.Wait()
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

func report(stdout io.ReadCloser, clr *color.Color) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		clr.Println(scanner.Text())
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func exitCode(err error) (int, bool) {
	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return 0, false
	}

	status, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return 0, false
	}

	return status.ExitStatus(), true
}
