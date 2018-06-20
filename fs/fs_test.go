package fs_test

import (
	. "github.com/aemengo/snb/fs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
	"os"
)

var _ = Describe("Fs", func() {
	var (
		fsClient *FS
		workingDir string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "snb-fs-")
		Expect(err).NotTo(HaveOccurred())

		fsClient, err = New(workingDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(workingDir)
	})

	Describe("Get", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(
				filepath.Join(workingDir, "sample-file"),
				[]byte("some-sample-file-content"),
				0600,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a relative path is provided", func() {
			It("returns the contents of the file", func() {
				contents, err := fsClient.Get("sample-file")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-sample-file-content"))
			})
		})

		Context("when an absolute path is provided", func() {
			It("returns the contents of the file", func() {
				contents, err := fsClient.Get(filepath.Join(workingDir, "sample-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-sample-file-content"))
			})
		})

		Context("when the file does not exist", func() {
			It("returns the error", func() {
				_, err := fsClient.Get("sample-non-existent-file")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Exists", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(
				filepath.Join(workingDir, "sample-file"),
				[]byte("some-sample-file-content"),
				0600,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a relative path is provided", func() {
			It("returns true", func() {
				exists := fsClient.Exists("sample-file")
				Expect(exists).To(BeTrue())
			})
		})

		Context("when an absolute path is provided", func() {
			It("returns true", func() {
				exists := fsClient.Exists(filepath.Join(workingDir, "sample-file"))
				Expect(exists).To(BeTrue())
			})
		})

		Context("when the file does not exist", func() {
			It("returns false", func() {
				exists := fsClient.Exists("sample-non-existent-file")
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("GetSrcFiles", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(
				filepath.Join(workingDir, "sample-file-1"),
				[]byte("some-sample-file-content-1"),
				0600,
			)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(
				filepath.Join(workingDir, "sample-file-2"),
				[]byte("some-sample-file-content-2"),
				0600,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a step with matching filenames is provided", func() {
			It("returns objects with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles("./sample-file-1 && ./sample-file-2")
				Expect(err).NotTo(HaveOccurred())
				Expect(len(objs)).To(Equal(2))
				Expect(objs[0].Path).To(Equal("./sample-file-1"))
				Expect(objs[0].Sha).NotTo(BeEmpty())

				Expect(objs[1].Path).To(Equal("./sample-file-2"))
				Expect(objs[1].Sha).NotTo(BeEmpty())
			})
		})

		Context("when a step with matching filenames and plenty of whitespace is provided", func() {
			It("returns objects with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles("./sample-file-1 && \n \n     ./sample-file-2")
				Expect(err).NotTo(HaveOccurred())
				Expect(len(objs)).To(Equal(2))
				Expect(objs[0].Path).To(Equal("./sample-file-1"))
				Expect(objs[0].Sha).NotTo(BeEmpty())

				Expect(objs[1].Path).To(Equal("./sample-file-2"))
				Expect(objs[1].Sha).NotTo(BeEmpty())
			})
		})

		Context("when a step with matching and nonmatching filenames is provided", func() {
			It("returns matching objects with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles("./sample-file-1 && ./some-non-existent-file")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(objs)).To(Equal(1))
				Expect(objs[0].Path).To(Equal("./sample-file-1"))
				Expect(objs[0].Sha).NotTo(BeEmpty())
			})
		})

		Context("when a step with an absolute path", func() {
			It("returns matching objects with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles(filepath.Join(workingDir, "sample-file-1"))
				Expect(err).NotTo(HaveOccurred())

				Expect(len(objs)).To(Equal(1))
				Expect(objs[0].Path).To(Equal(filepath.Join(workingDir, "sample-file-1")))
				Expect(objs[0].Sha).NotTo(BeEmpty())
			})
		})

		Context("when a step matches a directory", func() {
			BeforeEach(func() {
				err := os.MkdirAll(filepath.Join(workingDir, "some-directory"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(workingDir, "some-directory", "sample-file"),
					[]byte("some-sample-file-content"),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns matching objects with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles("cp some-directory")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(objs)).To(Equal(1))
				Expect(objs[0].Path).To(Equal("some-directory"))
				Expect(objs[0].Sha).NotTo(BeEmpty())
			})
		})

		Context("when a step matches a golang repository", func() {
			BeforeEach(func() {
				err := os.MkdirAll(filepath.Join(workingDir, "src", "golang-repo"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(
					filepath.Join(workingDir, "src", "golang-repo", "sample-file"),
					[]byte("some-sample-file-content"),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns matching golang dir with the sha representations", func() {
				objs, err := fsClient.GetSrcFiles("go build golang-repo")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(objs)).To(Equal(1))
				Expect(objs[0].Path).To(Equal(filepath.Join("src", "golang-repo")))
				Expect(objs[0].Sha).NotTo(BeEmpty())
			})
		})
	})
})
