package store_test

import (
	. "github.com/aemengo/snb/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
	"os"
)

var _ = Describe("Store", func() {
	var (
		dir string
		store *Store
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "snb-store-")
		Expect(err).NotTo(HaveOccurred())

		store, err = New(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	Describe("SaveBlob", func() {
		var testFilePath string

		BeforeEach(func() {
			testFilePath = filepath.Join(dir, "test-file")
			contents := []byte("some-sample-content")

			err := ioutil.WriteFile(testFilePath, contents, 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes blob to object directory", func() {
			store.SaveBlob(testFilePath)
			Expect(store.Err).NotTo(HaveOccurred())

			files, err := filepath.Glob(filepath.Join(dir, "objects", "*", "*"))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(1))
			Expect(contentsAt(files[0])).To(Equal("some-sample-content"))
		})
	})
})
