package main

import (
	"bufio"
	"github.com/aemengo/snb/db"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
	"github.com/aemengo/snb/fs"
	"path/filepath"
	"fmt"
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
	workingDir = "."
)

func main() {
	startTime := time.Now()

	if len(os.Args) == 2 {
		var err error
		workingDir, err = filepath.Abs(os.Args[1])
		if err != nil {
			logFatal(err)
		}
	} else {
		workingDir, _ = filepath.Abs(workingDir)
	}

	path := filepath.Join(workingDir, "ShakeAndBakeFile")
	if !fs.Exists(path) {
		logFatal("ShakeAndBakeFile not found. please execute in directory containing spec, or pass the working directory in as the only argument")
	}

	store, err := db.New(filepath.Join(workingDir, ".snb"))
	if err != nil {
		logFatal(err)
	}
	defer store.Close()

	spec, err := getSpec(path)
	if err != nil {
		logFatal(err)
	}

	for index, step := range spec.Steps {
		boldWhite.Printf("Step %d/%d : %s\n", index+1, len(spec.Steps), step)

		srcFiles, err := fs.GetSrcFiles(step)
		if err != nil {
			logFatal(err)
		}

		ok, err := store.IsCached(step, index, srcFiles)
		if err != nil {
			logFatal(err)
		}

		if ok {
			white.Println(logPrefix + "Using cache")
			continue
		}

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

		//TODO save step again
		//bl := getBlobList()
		//modifiedFiles := getModifiedFiles(blobList, bl)
		//store.SaveStep(step, index, modifiedFiles)
		//blobList = bl
	}

	endTime := time.Now()

	boldGreen.Printf("\nBuild completed (%f seconds)\n", endTime.Sub(startTime).Seconds())
}


func logFatal(message interface{}) {
	fmt.Printf(color.RedString("Error") + ": %s.\n", message)
	os.Exit(1)
}

func executeStep(step string) error {
	white.Println(logPrefix + "Running")
	command := exec.Command("bash", "-c", step)
	command.Dir = workingDir
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

func getSpec(path string) (Spec, error) {
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
		spec.Steps = append(spec.Steps, strings.TrimSpace(match[1]))
	}

	return spec, nil
}

func report(stdout io.ReadCloser, clr *color.Color) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		clr.Println(scanner.Text())
	}
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
