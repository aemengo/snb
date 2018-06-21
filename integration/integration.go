package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gbytes"
	"os"
	"path/filepath"
)

var _ = Describe("Integration", func() {

	var workingDir string

	AfterEach(func() {
		os.RemoveAll(filepath.Join(workingDir, ".snb"))
	})

	Describe("happy path", func() {

		BeforeEach(func() {
			workingDir = fixturePath("happy-path")
		})

		It("completes the build and renders the output", func() {
			By("running with initial state")

			command := exec.Command(binaryPath, workingDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say(`Step 1/2`))
			Expect(session).To(gbytes.Say(`---> Running`))
			Expect(session).To(gbytes.Say(`executing operation 1`))
			Expect(session).To(gbytes.Say(`operation 1 stderr message`))

			Expect(session).To(gbytes.Say(`Step 2/2`))
			Expect(session).To(gbytes.Say(`---> Running`))
			Expect(session).To(gbytes.Say(`executing operation 2`))
			Expect(session).To(gbytes.Say(`operation 2 stderr message`))

			Expect(session).To(gbytes.Say(`Build completed`))

			By("running with the state cached")

			command = exec.Command(binaryPath, workingDir)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say(`Step 1/2`))
			Expect(session).To(gbytes.Say(`---> Using cache`))
			Expect(session).To(gbytes.Say(`Step 2/2`))
			Expect(session).To(gbytes.Say(`---> Using cache`))
			Expect(session.Out.Contents()).NotTo(SatisfyAny(
				ContainSubstring("executing operation 1"),
				ContainSubstring("executing operation 2"),
			))

			By("running in the working directory and passing no args")

			command = exec.Command(binaryPath)
			command.Dir = workingDir
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session).To(gbytes.Say(`Build completed`))
		})
	})

	Describe("changing state", func() {
		BeforeEach(func() {
			workingDir = fixturePath("state-change")
		})

		It("correctly caches operation steps", func() {
			defer os.RemoveAll(filepath.Join(fixturePath("state-change"), "build-artifacts"))

			command := exec.Command(binaryPath, workingDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say(`Step 1/1`))
			Expect(session).To(gbytes.Say(`---> Running`))

			command = exec.Command(binaryPath, workingDir)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say(`Step 1/1`))
			Expect(session).To(gbytes.Say(`---> Using cache`))
		})
	})


	Describe("missing ShakeAndBakeFile", func() {
		BeforeEach(func() {
			workingDir = fixturePath("no-snb-file")
		})

		It("returns an error message", func() {
			command := exec.Command(binaryPath, workingDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).NotTo(Equal(0))

			Expect(session).To(gbytes.Say(`ShakeAndBakeFile not found.`))
		})
	})

	Describe("failed run", func() {
		BeforeEach(func() {
			workingDir = fixturePath("failed-run")
		})

		It("returns an error", func() {
			command := exec.Command(binaryPath, workingDir)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).NotTo(Equal(0))
		})
	})

	Describe("help", func() {
		It("returns an error", func() {
			command := exec.Command(binaryPath, "--help")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-session.Exited
			Expect(session.ExitCode()).NotTo(Equal(0))
			Expect(session).To(gbytes.Say(`USAGE`))
		})
	})
})
