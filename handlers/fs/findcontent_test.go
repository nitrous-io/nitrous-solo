package fs_test

import (
	"encoding/json"
	"path/filepath"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fs.find-content", func() {
	var res []byte

	cb := func(r []byte) {
		res = r
	}

	parseResponseJSON := func() (j fs.SearchResponse) {
		err := json.Unmarshal(res, &j)
		Expect(err).To(BeNil())
		return j
	}

	parseErrorResponseJSON := func() (j map[string]interface{}) {
		err := json.Unmarshal(res, &j)
		Expect(err).To(BeNil())
		return j
	}

	BeforeEach(func() {
		res = nil
	})

	Context("when path and term are specified", func() {
		It("returns search results", func() {
			fs.FindContent("fs.find-content", EncodeJSON(H{"path": filepath.Join(fixturesPath, "search"), "term": "README"}), cb)
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j.Results).To(HaveLen(1))
			Expect(j.Results[0].File).To(Equal("foo.md"))
			Expect(j.Results[0].Matches).To(HaveLen(1))
			Expect(j.Results[0].Matches[0].Row).To(Equal(1))
			Expect(j.Results[0].Matches[0].Content).To(Equal("== README"))
			Expect(j.Results[0].Matches[0].HighlightedContent).To(Equal("== \x1b[30;43mREADME\x1b[0m\x1b[K"))
			Expect(j.Results[0].Matches[0].Count).To(Equal(1))
		})

		Context("when file pattern is specified", func() {
			It("searches in files that match the pattern", func() {
				input := EncodeJSON(H{
					"path":         filepath.Join(fixturesPath, "search"),
					"term":         "reverse",
					"file_pattern": "*.rb",
				})

				fs.FindContent("fs.find-content", input, cb)
				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(1))
				Expect(j.Results[0].File).To(Equal("bar.rb"))
				Expect(j.Results[0].Matches).To(HaveLen(1))
			})
		})

		Context("when ignorecase is true", func() {
			It("does case-insensitive search", func() {
				input := EncodeJSON(H{
					"path":        filepath.Join(fixturesPath, "search"),
					"term":        "ReadMe",
					"ignore_case": true,
				})

				fs.FindContent("fs.find-content", input, cb)
				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(1))
				Expect(j.Results[0].File).To(Equal("foo.md"))
				Expect(j.Results[0].Matches).To(HaveLen(1))
			})
		})

		Context("when useregexp is true", func() {
			It("search term as regular expression", func() {
				input := EncodeJSON(H{
					"path":       filepath.Join(fixturesPath, "search"),
					"term":       "rea.*me",
					"use_regexp": true,
				})

				fs.FindContent("fs.find-content", input, cb)
				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(1))
				Expect(j.Results[0].File).To(Equal("foo.md"))
				Expect(j.Results[0].Matches).To(HaveLen(1))
			})
		})

		Context("when there is no matches", func() {
			It("returns empty array", func() {
				input := EncodeJSON(H{
					"path": filepath.Join(fixturesPath, "search"),
					"term": "Readme",
				})

				fs.FindContent("fs.find-content", input, cb)
				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(0))
			})
		})
	})

	Context("when the given path is empty string", func() {
		It("returns search results", func() {
			fs.FindContent("fs.find-content", EncodeJSON(H{"path": "", "term": "README"}), cb)
			Expect(res).NotTo(BeNil())
			j := parseErrorResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given term is empty string", func() {
		It("returns search results", func() {
			fs.FindContent("fs.find-content", EncodeJSON(H{"path": filepath.Join(fixturesPath, "search"), "term": ""}), cb)
			Expect(res).NotTo(BeNil())
			j := parseErrorResponseJSON()
			Expect(j["error"]).To(Equal("invalid_term"))
		})
	})
})
