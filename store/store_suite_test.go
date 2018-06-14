package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"io/ioutil"
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
