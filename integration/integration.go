package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"os"
	"path/filepath"
)

var _ = Describe("Integration", func() {

	var (
		fixtureDir string
		resultPath string
	)

	AfterEach(func() {
		os.RemoveAll(resultPath)
	})

	Describe("When a build has never been ran", func() {

		BeforeEach(func() {
			fixtureDir = fixturePath("fixture1")
			resultPath = filepath.Join(fixtureDir, "result.txt")
		})

		It("completes the build and renders the output", func() {
			command := exec.Command(binaryPath)
			command.Dir = fixtureDir

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(contentsAt(resultPath)).To(SatisfyAll(
				ContainSubstring("line 1"),
				ContainSubstring("line 2"),
			))
		})
	})
})
