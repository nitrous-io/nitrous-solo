package fs_test

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fs.Write", func() {
	var (
		res         []byte
		pathToWrite string
		dir         string
		err         error

		runAs      string
		runUserUID uint32
		runUserGID uint32
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

		dir, err = ioutil.TempDir("", "")
		Expect(err).To(BeNil())

		runAs = "nitrous"
		u, err := user.Lookup(runAs)

		uid, err := strconv.Atoi(u.Uid)
		Expect(err).To(BeNil())
		runUserUID = uint32(uid)

		gid, err := strconv.Atoi(u.Gid)
		Expect(err).To(BeNil())
		runUserGID = uint32(gid)

		pathToWrite = filepath.Join(dir, "hello.txt")
	})

	AfterEach(func() {
		os.RemoveAll(dir)
	})

	It("writes the content to a file", func() {
		fs.Write(&ws.Context{
			Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
			Reply: cb,
			Meta: map[string]interface{}{
				"login": "foo",
				"user":  runAs,
			},
			Done: Noop,
		})

		Expect(res).NotTo(BeNil())

		j := parseResponseJSON()
		Expect(j["written"]).To(BeTrue())

		fi, err := os.Stat(pathToWrite)
		Expect(err).To(BeNil())

		st := fi.Sys().(*syscall.Stat_t)
		Expect(st.Uid).To(Equal(runUserUID))
		Expect(st.Gid).To(Equal(runUserGID))

		content, err := ioutil.ReadFile(pathToWrite)
		Expect(err).To(BeNil())

		Expect(content).To(Equal([]byte(`hello world`)))
	})

	Context("when user does not exist", func() {
		BeforeEach(func() {
			runAs = "somethingelse"
		})

		It("follows the symlink to write to the actual file", func() {
			fs.Write(&ws.Context{
				Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
				Reply: cb,
				Meta: map[string]interface{}{
					"login": "foo",
					"user":  runAs,
				},
				Done: Noop,
			})
			Expect(res).NotTo(BeNil())

			j := parseResponseJSON()
			Expect(j["error"]).To(Equal("invalid_user"))
		})
	})

	Context("when the file exists", func() {
		BeforeEach(func() {
			Expect(ioutil.WriteFile(pathToWrite, []byte(`old content`), 0644)).To(BeNil())
		})

		It("overwrites the file", func() {
			fs.Write(&ws.Context{
				Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
				Reply: cb,
				Meta: map[string]interface{}{
					"login": "foo",
					"user":  runAs,
				},
				Done: Noop,
			})

			j := parseResponseJSON()
			Expect(j["written"]).To(BeTrue())

			content, err := ioutil.ReadFile(pathToWrite)
			Expect(err).To(BeNil())

			Expect(content).To(Equal([]byte(`hello world`)))
		})

		Context("when the user does not have the permission to write to the given path", func() {
			BeforeEach(func() {
				Expect(os.Chmod(pathToWrite, 0400)).To(BeNil())
			})

			It("returns an error", func() {
				fs.Write(&ws.Context{
					Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  runAs,
					},
					Done: Noop,
				})

				Expect(res).NotTo(BeNil())

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("permission_denied"))

				content, err := ioutil.ReadFile(pathToWrite)
				Expect(err).To(BeNil())

				Expect(content).To(Equal([]byte(`old content`)))
			})
		})

		Context("when the file is a directory", func() {
			BeforeEach(func() {
				pathToWrite = filepath.Join(dir, "foo")
				Expect(os.Mkdir(pathToWrite, 0644)).To(BeNil())
			})

			It("returns an error", func() {
				fs.Write(&ws.Context{
					Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  runAs,
					},
					Done: Noop,
				})

				Expect(res).NotTo(BeNil())

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("path_is_directory"))
			})
		})

		Context("when the file is a symlink", func() {
			var symlinkPath string

			BeforeEach(func() {
				symlinkPath = filepath.Join(dir, "symlink.txt")
				Expect(os.Symlink(pathToWrite, symlinkPath)).To(BeNil())
			})

			It("writes to the actual file to which the symlink is pointing", func() {
				fs.Write(&ws.Context{
					Data:  EncodeJSON(H{"path": symlinkPath, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  runAs,
					},
					Done: Noop,
				})
				Expect(res).NotTo(BeNil())

				j := parseResponseJSON()
				Expect(j["written"]).To(BeTrue())

				content, err := ioutil.ReadFile(pathToWrite)
				Expect(err).To(BeNil())

				Expect(content).To(Equal([]byte(`hello world`)))
			})

			Context("when the symlink is broken", func() {
				BeforeEach(func() {
					os.Remove(symlinkPath)
					brokenSymlinkPath := filepath.Join(dir, "broken.txt")
					Expect(os.Symlink(brokenSymlinkPath, symlinkPath)).To(BeNil())
					Expect(os.Symlink(symlinkPath, brokenSymlinkPath)).To(BeNil())
				})

				It("returns an error", func() {
					fs.Write(&ws.Context{
						Data:  EncodeJSON(H{"path": symlinkPath, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
						Reply: cb,
						Meta: map[string]interface{}{
							"login": "foo",
							"user":  runAs,
						},
						Done: Noop,
					})
					Expect(res).NotTo(BeNil())

					j := parseResponseJSON()
					Expect(j["error"]).To(Equal("symlink_too_deep"))
				})
			})
		})

		Context("when the file path is not absolute", func() {
			BeforeEach(func() {
				pathToWrite = "bar.txt"
			})

			It("returns an error", func() {
				fs.Write(&ws.Context{
					Data:  EncodeJSON(H{"path": pathToWrite, "content": base64.StdEncoding.EncodeToString([]byte("hello world"))}),
					Reply: cb,
					Meta: map[string]interface{}{
						"login": "foo",
						"user":  runAs,
					},
					Done: Noop,
				})

				Expect(res).NotTo(BeNil())

				j := parseResponseJSON()
				Expect(j["error"]).To(Equal("path_is_not_absolute"))
			})
		})

	})
})
