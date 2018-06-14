package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"io/ioutil"
	"encoding/json"
	"strings"
)

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Suite")
}

func contentsAt(path string) string {
	Expect(path).To(BeAnExistingFile())
	contents, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(contents)
}

func decodeJSONAt(path string, dest interface{}) {
	contents := contentsAt(path)
	err := json.NewDecoder(strings.NewReader(contents)).Decode(dest)
	Expect(err).NotTo(HaveOccurred())
}