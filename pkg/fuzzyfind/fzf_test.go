package fuzzyfind

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FzfFinder")
}

var _ = Describe("FzfFinder", func() {
	var (
		dir string
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("/tmp", "test-fzf-finder")
		Expect(err).To(BeNil())

		testFiles := []string{
			"README.md",
			"controllers/containers/containers.go",
			"controllers/containers/containers_test.go",
			"models/container.go",
			"models/workspace.go",
			"models/config.json",
			"dist/server",
			"pkg/dockerclient/doc.a",
			"pkg/dockerclient/docker/pkg.a",
			".git/HEAD",
		}

		for _, p := range testFiles {
			path := filepath.Join(dir, p)
			dir := filepath.Dir(path)
			err := os.MkdirAll(dir, 0755)
			Expect(err).To(BeNil())
			f, err := os.Create(path)
			Expect(err).To(BeNil())
			f.Close()
		}
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	Describe("FzfFinder#Find()", func() {

		DescribeTable("finding files in a directory",
			func(query string, expected []File) {
				finder := FzfFinder{BinaryPath: "fzf"}

				result, err := finder.Find(dir, query)

				Expect(err).To(BeNil())
				Expect(result).To(ConsistOf(expected)) // Order is NOT tested.
			},

			Entry("with matching folders", "models",
				[]File{
					{Path: "models/container.go"},
					{Path: "models/workspace.go"},
					{Path: "models/config.json"},
				},
			),

			Entry("with matching filename", "server",
				[]File{
					{Path: "dist/server"},
				},
			),

			Entry("with matching folder and filename", "pkg",
				[]File{
					{Path: "pkg/dockerclient/doc.a"},
					{Path: "pkg/dockerclient/docker/pkg.a"},
				},
			),

			Entry("with hidden file", "HEAD",
				[]File{
					{Path: ".git/HEAD"},
				},
			),

			Entry("with query that matches fuzzily", "mocon",
				[]File{
					{Path: "models/container.go"},
					{Path: "models/config.json"},
				},
			),
		)

		Context("when not given a directory", func() {

		})

		It("returns Files whose names match the query", func() {
			finder := FzfFinder{BinaryPath: "fzf"}

			files, err := finder.Find(dir, "models")

			Expect(err).To(BeNil())
			Expect(files).To(HaveLen(3))

			paths := []string{}
			for _, f := range files {
				paths = append(paths, f.Path)
			}
			Expect(paths).To(ConsistOf(
				"models/container.go",
				"models/workspace.go",
				"models/config.json"))
		})
	})
})
