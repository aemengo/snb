package main

import (
	"bufio"
	"github.com/aemengo/snb/db"
	"github.com/fatih/color"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
	"github.com/aemengo/snb/fs"
	"path/filepath"
	"fmt"
	"github.com/aemengo/snb/parser"
	"strings"
)

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

	switch len(os.Args) {
	case 1:
		workingDir, _ = filepath.Abs(workingDir)
	case 2:
		if strings.HasPrefix(os.Args[1], "-") {
			showUsage()
		}

		var err error
		workingDir, err = filepath.Abs(os.Args[1])
		if err != nil {
			fatal(err)
		}
	default:
		showUsage()
	}

	fsClient, err := fs.New(workingDir)
	if err != nil {
		fatal(err)
	}

	if !fsClient.Exists("ShakeAndBakeFile") {
		fatal("ShakeAndBakeFile not found. please execute in directory containing spec, or pass the working directory in as the only argument")
	}

	contents, err := fsClient.Get("ShakeAndBakeFile")
	if err != nil {
		fatal(err)
	}

	spec, err := parser.Parse(contents)
	if err != nil {
		fatal(err)
	}

	dbClient, err := db.New(filepath.Join(workingDir, ".snb"))
	if err != nil {
		fatal(err)
	}
	defer dbClient.Close()

	for index, step := range spec.Steps {
		boldWhite.Printf("Step %d/%d : %s\n", index+1, len(spec.Steps), step)

		srcFiles, err := fsClient.GetSrcFiles(step)
		if err != nil {
			fatal(err)
		}

		ok, err := dbClient.IsCached(step, index, srcFiles)
		if err != nil {
			fatal(err)
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

		srcFiles, err = fsClient.GetSrcFiles(step)
		if err != nil {
			fatal(err)
		}

		err = dbClient.Save(step, index, srcFiles)
		if err != nil {
			fatal(err)
		}
	}

	endTime := time.Now()

	boldGreen.Printf("\nBuild completed (%f seconds)\n", endTime.Sub(startTime).Seconds())
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

func fatal(message interface{}) {
	fmt.Printf(color.RedString("Error") + ": %s.\n", message)
	os.Exit(1)
}

func showUsage() {
	fmt.Println(`
USAGE:	snb [PATH]

Build an image from a ShakeAndBakeFile
	`)
	os.Exit(1)
}