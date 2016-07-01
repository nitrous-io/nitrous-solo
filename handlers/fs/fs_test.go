package fs_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fs")
}

var _ = Describe("fs", func() {
	currPath, _ := filepath.Abs(".")
	fixturesPath := filepath.Join(currPath, "..", "..", "testhelper", "fixtures")

	var res []byte

	cb := func(r []byte) {
		res = r
	}

	BeforeEach(func() {
		res = nil
	})

	Describe("List", func() {
		Context("when the given path exists", func() {
			It("returns a list of files in a given path", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j struct {
					Files []string `json:"files"`
				}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())

				Expect(j.Files).To(ConsistOf(
					"bar/",
					"baz",
					"brokensymlink",
					"dirsymlink/",
					"filesymlink",
					"foo",
					".hiddendir/",
					".hiddenfile",
					"qux/",
					"recursymlink1",
					"recursymlink2",
				))
			})
		})

		Context("when the given path is a symlink to a directory", func() {
			It("follows symlink and returns a list of files in the path, to which symlink is pointing", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs", "dirsymlink")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j struct {
					Files []string `json:"files"`
				}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())

				Expect(j.Files).To(ConsistOf(
					"quxdir/",
					"quxfoo",
					".quxhidden",
				))
			})
		})

		Context("when the given path is empty", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": ""}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("invalid_path"))
			})
		})

		Context("when the given path does not exist", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "ffs")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("invalid_path"))
			})
		})

		Context("when the given path is not a directory", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs", "foo")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("path_not_directory"))
			})
		})

		Context("when the given path is a symlink to a file", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs", "filesymlink")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("path_not_directory"))
			})
		})

		Context("when the given path is a broken symlink", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs", "brokensymlink")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("invalid_path"))
			})
		})

		Context("when the given path is a recursive symlink", func() {
			It("returns an error", func() {
				fs.List("fs", []byte(`{"path": "`+filepath.Join(fixturesPath, "fs", "recursymlink1")+`"}`), cb)
				Expect(res).NotTo(BeNil())

				var j map[string]interface{}
				err := json.Unmarshal(res, &j)
				Expect(err).To(BeNil())
				Expect(j["error"]).To(Equal("symlink_too_deep"))
			})
		})
	})
})
