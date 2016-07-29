package fs_test

import (
	"encoding/json"
	"os/user"
	"path/filepath"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var currentSystemUser *user.User

func init() {
	currentSystemUser, _ = user.Current()
}

var _ = Describe("fs.List", func() {
	var (
		res          []byte
		origHomePath func(u *data.UserData) (string, error)
	)

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

		data.D = data.Reset()
		ud := &data.UserData{
			Login: "foobar",
			User:  currentSystemUser.Username,
		}
		data.D.UserData[ud.Login] = ud

		origHomePath = data.HomePath
		data.HomePath = func(u *data.UserData) (string, error) {
			return fixturesPath, nil
		}
	})

	AfterEach(func() {
		data.HomePath = origHomePath
	})

	doList := func(path string) {
		fs.List(&ws.Context{
			Data:  EncodeJSON(H{"path": path}),
			Reply: cb,
			Meta: ws.M{
				"user": "foobar",
			},
			Done: Noop,
		})
	}

	Context("when the given path exists", func() {
		It("returns a list of files in a given path", func() {
			doList(filepath.Join(fixturesPath, "fs"))
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
			doList(filepath.Join(fixturesPath, "fs", "dirsymlink"))
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
			doList("")
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path does not exist", func() {
		It("returns an error", func() {
			doList(filepath.Join(fixturesPath, "ffs"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path is not a directory", func() {
		It("returns an error", func() {
			doList(filepath.Join(fixturesPath, "fs", "foo"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("path_not_directory"))
		})
	})

	Context("when the given path is a symlink to a file", func() {
		It("returns an error", func() {
			doList(filepath.Join(fixturesPath, "fs", "filesymlink"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("path_not_directory"))
		})
	})

	Context("when the given path is a broken symlink", func() {
		It("returns an error", func() {
			doList(filepath.Join(fixturesPath, "fs", "brokensymlink"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})

	Context("when the given path is a recursive symlink", func() {
		It("returns an error", func() {
			doList(filepath.Join(fixturesPath, "fs", "recursymlink1"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("symlink_too_deep"))
		})
	})

	Context("when the given path is not in home directory", func() {
		It("returns an error", func() {
			data.HomePath = func(u *data.UserData) (string, error) {
				return "/root", nil
			}

			doList(filepath.Join(fixturesPath, "fs"))
			Expect(res).NotTo(BeNil())
			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_path"))
		})
	})
})
