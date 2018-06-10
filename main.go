package main

import (
	"fmt"
	"os"
	"errors"
	"log"
	"io/ioutil"
	"regexp"
	"os/exec"
)

type Spec struct {
	Steps []string
}

func main() {
	spec, err := getSpec()
	if err != nil {
		log.Fatal("Error: ", err, ".")
	}

	for _, step := range spec.Steps {
		executeStep(step)
	}

	fmt.Println("Finished")
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
	if os.IsNotExist(err) { return false }
	return true
}