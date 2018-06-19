package parser

import (
	"regexp"
	"strings"
)

type Spec struct {
	Steps []string
}

func Parse(specFile []byte) (Spec, error) {
	stepRegex := `(?msU)RUN(.*)(?:^\s*$|\z)`
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