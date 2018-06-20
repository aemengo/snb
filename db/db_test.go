package db_test

import (
	. "github.com/aemengo/snb/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"github.com/aemengo/snb/fs"
)

var _ = Describe("Store", func() {
	var (
		dir string
		dbClient *DB
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "snb-store-")
		Expect(err).NotTo(HaveOccurred())

		dbClient, err = New(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		dbClient.Close()
		os.RemoveAll(dir)
	})

	Describe("When nothing is saved to the store", func() {
		It("returns false for cached steps", func() {
			cached, err := dbClient.IsCached("some-step", 1, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cached).To(BeFalse())
		})
	})

	Describe("when steps are saved to the store", func() {
		BeforeEach(func() {
			err := dbClient.Save("some-step", 1, []fs.Object{
				{Path: "./some-path-1", Sha: "abc-some-sha"},
				{Path: "./some-path-2", Sha: "def-some-sha"},
			})

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when queried for step and matching objects", func() {
			It("returns true for cached steps", func() {
				cached, err := dbClient.IsCached("some-step", 1, []fs.Object{
					{Path: "./some-path-1", Sha: "abc-some-sha"},
					{Path: "./some-path-2", Sha: "def-some-sha"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).To(BeTrue())
			})
		})

		Context("when queried for step and some of matching objects", func() {
			It("returns true for cached steps", func() {
				cached, err := dbClient.IsCached("some-step", 1, []fs.Object{
					{Path: "./some-path-1", Sha: "abc-some-sha"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).To(BeTrue())
			})
		})

		Context("when queried for mismatched step instructions", func() {
			It("returns false for cached steps", func() {
				cached, err := dbClient.IsCached("some-non-existent-step", 1, []fs.Object{
					{Path: "./some-path-1", Sha: "abc-some-sha"},
					{Path: "./some-path-2", Sha: "def-some-sha"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).To(BeFalse())
			})
		})

		Context("when queried for mismatched step index", func() {
			It("returns false for cached steps", func() {
				cached, err := dbClient.IsCached("some-step", -1, []fs.Object{
					{Path: "./some-path-1", Sha: "abc-some-sha"},
					{Path: "./some-path-2", Sha: "def-some-sha"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).To(BeFalse())
			})
		})

		Context("when queried for step with no objects", func() {
			It("returns false for cached steps", func() {
				cached, err := dbClient.IsCached("some-step", 1, []fs.Object{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).To(BeFalse())
			})
		})
	})
})
