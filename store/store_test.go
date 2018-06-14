package store_test

import (
	. "github.com/aemengo/snb/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Store", func() {
	var (
		dir   string
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
		var srcFilePath string

		BeforeEach(func() {
			srcFilePath = filepath.Join(dir, "test-file")
			contents := []byte("some-sample-content")

			err := ioutil.WriteFile(srcFilePath, contents, 0600)
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes blob to object directory", func() {
			store.SaveBlob(srcFilePath)
			Expect(store.Err).NotTo(HaveOccurred())

			files, err := filepath.Glob(filepath.Join(dir, "objects", "*", "*"))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(1))
			Expect(contentsAt(files[0])).To(Equal("some-sample-content"))
		})
	})

	Describe("SaveTree", func() {
		var srcDir string

		Context("when there is one blob in the dir", func() {
			BeforeEach(func() {
				srcDir = filepath.Join(dir, "tree-dir")
				err := os.MkdirAll(srcDir, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				srcFilePath := filepath.Join(srcDir, "test-file")
				contents := []byte("some-sample-content")
				err = ioutil.WriteFile(srcFilePath, contents, 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes the tree to the object directory", func() {
				store.SaveTree(srcDir)
				Expect(store.Err).NotTo(HaveOccurred())

				files, err := filepath.Glob(filepath.Join(dir, "objects", "*", "*"))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(HaveLen(1))

				var result []struct {
					Type string `json:"type"`
					Sha  string `json:"sha"`
					Name string `json:"name"`
				}

				decodeJSONAt(files[0], &result)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Type).To(Equal("blob"))
				Expect(result[0].Name).To(Equal("test-file"))
				Expect(result[0].Sha).NotTo(BeEmpty())
			})
		})
	})
})
