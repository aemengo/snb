package parser_test

import (
	. "github.com/aemengo/snb/parser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	Describe("Parse", func() {
		var specContents []byte

		Context("when there are multiple steps indicated", func() {
			BeforeEach(func() {
				specContents = []byte(`
				RUN ./some-command 1

				RUN ./some-command 2

				RUN ./some-command 3`)
			})

			It("returns multiple steps", func() {
				spec, err := Parse(specContents)
				Expect(err).NotTo(HaveOccurred())

				Expect(spec).To(Equal(Spec{
					Steps: []string{
						"./some-command 1",
						"./some-command 2",
						"./some-command 3",
					},
				}))

			})
		})

		Context("when there are multiple multi-line steps indicated", func() {
			BeforeEach(func() {
				specContents = []byte(`
				RUN ./some-command 1 && \
				    ./some-added-command 1

				RUN ./some-command 2 && \
				    ./some-added-command 2

				RUN ./some-command 3 && \
				    ./some-added-command 3`)
			})

			It("returns multiple steps", func() {
				spec, err := Parse(specContents)
				Expect(err).NotTo(HaveOccurred())

				Expect(spec.Steps[0]).To(SatisfyAll(
					ContainSubstring(`./some-command 1 && \`),
					ContainSubstring(`./some-added-command 1`),
				))
				Expect(spec.Steps[1]).To(SatisfyAll(
					ContainSubstring(`./some-command 2 && \`),
					ContainSubstring(`./some-added-command 2`),
				))
				Expect(spec.Steps[2]).To(SatisfyAll(
					ContainSubstring(`./some-command 3 && \`),
					ContainSubstring(`./some-added-command 3`),
				))
			})
		})
	})
})
