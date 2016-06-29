package fs_test

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fs.findInFiles", func() {
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

	btoa := base64.StdEncoding.EncodeToString

	Context("when path and term are specified", func() {
		It("returns search results", func() {
			fs.FindInFiles(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "search"), "term": "README"}),
				Reply: cb,
				Done:  Noop,
			})

			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j.Results).To(HaveLen(1))
			Expect(j.Results[0].File).To(Equal("foo.md"))
			Expect(j.Results[0].Matches).To(HaveLen(1))
			Expect(j.Results[0].Matches[0].Row).To(Equal(1))
			Expect(j.Results[0].Matches[0].Content).To(Equal(btoa([]byte("== README"))))
			Expect(j.Results[0].Matches[0].HighlightedContent).To(Equal(btoa([]byte("== \x1b[30;43mREADME\x1b[0m\x1b[K"))))
			Expect(j.Results[0].Matches[0].Count).To(Equal(1))
		})

		Context("when file pattern is specified", func() {
			It("searches in files that match the pattern", func() {
				input := EncodeJSON(H{
					"path":         filepath.Join(fixturesPath, "search"),
					"term":         "reverse",
					"file_pattern": "*.rb",
				})
				fs.FindInFiles(&ws.Context{
					Data:  input,
					Reply: cb,
					Done:  Noop,
				})

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
				fs.FindInFiles(&ws.Context{
					Data:  input,
					Reply: cb,
					Done:  Noop,
				})

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
				fs.FindInFiles(&ws.Context{
					Data:  input,
					Reply: cb,
					Done:  Noop,
				})

				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(1))
				Expect(j.Results[0].File).To(Equal("foo.md"))
				Expect(j.Results[0].Matches).To(HaveLen(1))
			})
		})

		Context("when there is no match", func() {
			var origLogLevel logrus.Level

			BeforeEach(func() {
				// suppress warning log message
				origLogLevel = logrus.GetLevel()
				logrus.SetLevel(logrus.FatalLevel)
			})

			AfterEach(func() {
				logrus.SetLevel(origLogLevel)
			})

			It("returns empty array", func() {
				input := EncodeJSON(H{
					"path": filepath.Join(fixturesPath, "search"),
					"term": "Readme",
				})
				fs.FindInFiles(&ws.Context{
					Data:  input,
					Reply: cb,
					Done:  Noop,
				})

				Expect(res).NotTo(BeNil())
				j := parseResponseJSON()
				Expect(j.Results).To(HaveLen(0))
			})
		})
	})

	Context("when the given path is empty string", func() {
		It("returns search results", func() {
			input := EncodeJSON(H{"path": "", "term": "README"})
			fs.FindInFiles(&ws.Context{
				Data:  input,
				Reply: cb,
				Done:  Noop,
			})

			Expect(res).NotTo(BeNil())
			j := parseErrorResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given term is empty string", func() {
		It("returns search results", func() {
			input := EncodeJSON(H{"path": filepath.Join(fixturesPath, "search"), "term": ""})
			fs.FindInFiles(&ws.Context{
				Data:  input,
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())
			j := parseErrorResponseJSON()
			Expect(j["error"]).To(Equal("invalid_term"))
		})
	})
})
