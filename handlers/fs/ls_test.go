package fs_test

import (
	"encoding/json"
	"path/filepath"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fs.List", func() {
	var res []byte

	cb := func(r []byte) {
		res = r
	}

	parseResponseJSON := func() (j map[string]interface{}) {
		err := json.Unmarshal(res, &j)
		Expect(err).To(BeNil())
		return j
	}

	BeforeEach(func() {
		res = nil
	})

	Context("when the given path exists", func() {
		It("returns a list of files in a given path", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["files"]).To(ConsistOf(
				"bar/",
				"baz",
				"binaryfile",
				"brokensymlink",
				"code.rb",
				"dirsymlink/",
				"filesymlink",
				"foo",
				"Gemfile",
				".hiddendir/",
				".hiddenfile",
				"qux/",
				"recursymlink1",
				"recursymlink2",
				"tiny.gif",
			))
		})
	})

	Context("when the given path is a symlink to a directory", func() {
		It("follows symlink and returns a list of files in the path, to which symlink is pointing", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "dirsymlink")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["files"]).To(ConsistOf(
				"quxdir/",
				"quxfoo",
				".quxhidden",
			))
		})
	})

	Context("when the given path is an empty string", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": ""}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path does not exist", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "ffs")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path is not a directory", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "foo")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("path_not_directory"))
		})
	})

	Context("when the given path is a symlink to a file", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "filesymlink")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("path_not_directory"))
		})
	})

	Context("when the given path is a broken symlink", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "brokensymlink")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path is a recursive symlink", func() {
		It("returns an error", func() {
			fs.List(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "recursymlink1")}),
				Reply: cb,
			})
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("symlink_too_deep"))
		})
	})
})
