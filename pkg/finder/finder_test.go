package finder_test

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/pkg/finder"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "finder")
}

var _ = Describe("FindContent", func() {
	var rootPath string

	BeforeEach(func() {
		logrus.SetLevel(logrus.FatalLevel)

		currPath, _ := filepath.Abs(".")
		fixturesPath := filepath.Join(currPath, "..", "..", "testhelper", "fixtures")
		rootPath = filepath.Join(fixturesPath, "search")
	})

	It("searches specified term in root folder", func() {
		matches, err := finder.FindContent(rootPath, "The", "", false, false)
		Expect(err).To(BeNil())
		Expect(matches).NotTo(BeNil())
		Expect(matches).To(HaveLen(2))

		sort.Sort(finder.ByFileName(matches))

		Expect(matches[0].File).To(Equal("bar.rb"))
		Expect(matches[0].Matches).To(HaveLen(1))
		Expect(matches[0].Matches[0].Row).To(Equal(8))
		Expect(matches[0].Matches[0].Content).To(Equal("# The [n,l] substring gives the starting position and length.  The [n..m]"))
		Expect(matches[0].Matches[0].HighlightedContent).To(Equal("# \x1b[30;43mThe\x1b[0m\x1b[K [n,l] substring gives the starting position and length.  \x1b[30;43mThe\x1b[0m\x1b[K [n..m]"))
		Expect(matches[0].Matches[0].Count).To(Equal(2))

		Expect(matches[1].File).To(Equal("foo.md"))
		Expect(matches[1].Matches).To(HaveLen(1))
		Expect(matches[1].Matches[0].Row).To(Equal(11))
		Expect(matches[1].Matches[0].Content).To(Equal("# The [n,l] substring gives the starting position and length.  The [n..m]"))
		Expect(matches[1].Matches[0].HighlightedContent).To(Equal("# \x1b[30;43mThe\x1b[0m\x1b[K [n,l] substring gives the starting position and length.  \x1b[30;43mThe\x1b[0m\x1b[K [n..m]"))
		Expect(matches[1].Matches[0].Count).To(Equal(2))
	})

	Context("number of search results are more than ContentSearchLimit", func() {
		var origContentSearchLimit int

		BeforeEach(func() {
			origContentSearchLimit = finder.ContentSearchLimit
			finder.ContentSearchLimit = 1
		})

		AfterEach(func() {
			finder.ContentSearchLimit = origContentSearchLimit
		})

		It("limits number of search results based on ContentSearchLimit", func() {
			rootPath := "../../testhelper/fixtures/search"
			matches, err := finder.FindContent(rootPath, "The", "", false, false)
			Expect(err).To(BeNil())
			Expect(matches).NotTo(BeNil())
			Expect(matches).To(HaveLen(1))
		})
	})

	Describe("File pattern", func() {
		It("searches specified term in specified files match pattern", func() {
			matches, err := finder.FindContent(rootPath, "character", "*.md", false, false)
			Expect(err).To(BeNil())
			Expect(matches).NotTo(BeNil())
			Expect(matches).To(HaveLen(1))

			Expect(matches[0].File).To(Equal("foo.md"))
			Expect(matches[0].Matches).To(HaveLen(1))
			Expect(matches[0].Matches[0].Row).To(Equal(7))
			Expect(matches[0].Matches[0].Content).To(Equal("# Selecting a character in a string gives an integer ascii code."))
			Expect(matches[0].Matches[0].HighlightedContent).To(Equal("# Selecting a \x1b[30;43mcharacter\x1b[0m\x1b[K in a string gives an integer ascii code."))
			Expect(matches[0].Matches[0].Count).To(Equal(1))
		})
	})

	Describe("Case sensitivity", func() {
		It("does case sensitive search if it ignoreCase is false", func() {
			matches, err := finder.FindContent(rootPath, "CharacTer", "", false, false)
			Expect(err).To(BeNil())
			Expect(matches).To(BeNil())
		})

		It("ignores case sensitivity if it ignoreCase is true", func() {
			matches, err := finder.FindContent(rootPath, "CharacTer", "", true, false)
			Expect(err).To(BeNil())
			Expect(matches).NotTo(BeNil())
			Expect(matches).To(HaveLen(2))

			sort.Sort(finder.ByFileName(matches))

			Expect(matches[0].File).To(Equal("bar.rb"))
			Expect(matches[0].Matches).To(HaveLen(1))
			Expect(matches[0].Matches[0].Row).To(Equal(4))
			Expect(matches[0].Matches[0].Content).To(Equal("# Selecting a character in a string gives an integer ascii code."))
			Expect(matches[0].Matches[0].HighlightedContent).To(Equal("# Selecting a \x1b[30;43mcharacter\x1b[0m\x1b[K in a string gives an integer ascii code."))
			Expect(matches[0].Matches[0].Count).To(Equal(1))

			Expect(matches[1].File).To(Equal("foo.md"))
			Expect(matches[1].Matches).To(HaveLen(1))
			Expect(matches[1].Matches[0].Row).To(Equal(7))
			Expect(matches[1].Matches[0].Content).To(Equal("# Selecting a character in a string gives an integer ascii code."))
			Expect(matches[1].Matches[0].HighlightedContent).To(Equal("# Selecting a \x1b[30;43mcharacter\x1b[0m\x1b[K in a string gives an integer ascii code."))
			Expect(matches[1].Matches[0].Count).To(Equal(1))
		})
	})

	Describe("Regexp", func() {
		It("searches specified term as regular expression", func() {
			matches, err := finder.FindContent(rootPath, "char.*", "", false, true)
			Expect(err).To(BeNil())
			Expect(matches).NotTo(BeNil())
			Expect(matches).To(HaveLen(2))

			sort.Sort(finder.ByFileName(matches))

			Expect(matches[0].File).To(Equal("bar.rb"))
			Expect(matches[0].Matches).To(HaveLen(1))
			Expect(matches[0].Matches[0].Row).To(Equal(4))
			Expect(matches[0].Matches[0].Content).To(Equal("# Selecting a character in a string gives an integer ascii code."))
			Expect(matches[0].Matches[0].HighlightedContent).To(Equal("# Selecting a \x1b[30;43mcharacter in a string gives an integer ascii code.\x1b[0m\x1b[K"))
			Expect(matches[0].Matches[0].Count).To(Equal(1))

			Expect(matches[1].File).To(Equal("foo.md"))
			Expect(matches[1].Matches).To(HaveLen(1))
			Expect(matches[1].Matches[0].Row).To(Equal(7))
			Expect(matches[1].Matches[0].Content).To(Equal("# Selecting a character in a string gives an integer ascii code."))
			Expect(matches[1].Matches[0].HighlightedContent).To(Equal("# Selecting a \x1b[30;43mcharacter in a string gives an integer ascii code.\x1b[0m\x1b[K"))
			Expect(matches[1].Matches[0].Count).To(Equal(1))
		})
	})

	Describe("Symlink", func() {
		It("should follow symlink for search", func() {
			matches, err := finder.FindContent(rootPath, "hello", "", false, false)
			Expect(err).To(BeNil())
			Expect(matches).NotTo(BeNil())
			Expect(matches).To(HaveLen(2))

			sort.Sort(finder.ByFileName(matches))

			Expect(matches[0].File).To(Equal("foo/hello.go"))
			Expect(matches[0].Matches).To(HaveLen(1))
			Expect(matches[0].Matches[0].Row).To(Equal(6))
			Expect(matches[0].Matches[0].Content).To(Equal("	fmt.Println(\"hello\")"))
			Expect(matches[0].Matches[0].HighlightedContent).To(Equal("	fmt.Println(\"\x1b[30;43mhello\x1b[0m\x1b[K\")"))
			Expect(matches[0].Matches[0].Count).To(Equal(1))

			Expect(matches[0].File).To(Equal("foo/hello.go"))
			Expect(matches[0].Matches).To(HaveLen(1))
			Expect(matches[0].Matches[0].Row).To(Equal(6))
			Expect(matches[0].Matches[0].Content).To(Equal("	fmt.Println(\"hello\")"))
			Expect(matches[0].Matches[0].HighlightedContent).To(Equal("	fmt.Println(\"\x1b[30;43mhello\x1b[0m\x1b[K\")"))
			Expect(matches[0].Matches[0].Count).To(Equal(1))
		})
	})
})
