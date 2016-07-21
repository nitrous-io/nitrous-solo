package atomicwrite_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/atomicwrite")
}

var _ = Describe("pkg/atomicwrite", func() {
	Describe("AtomicWrite", func() {
		var (
			content     *bytes.Buffer
			pathToWrite string
			dir         string
			err         error

			runAs      string
			runUserUID uint32
			runUserGID uint32
		)
		BeforeEach(func() {
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
			content = bytes.NewBufferString("hello world")
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("writes the content to a file", func() {
			Expect(atomicwrite.AtomicWrite(content, pathToWrite, runAs, false)).To(BeNil())

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
				Expect(
					atomicwrite.AtomicWrite(content, pathToWrite, runAs, false),
				).To(Equal(atomicwrite.ErrInvalidUser))
			})
		})

		Context("when the file exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(pathToWrite, []byte(`old content`), 0644)).To(BeNil())
			})

			Context("when overwrite param is true", func() {
				It("overwrites the file", func() {
					Expect(atomicwrite.AtomicWrite(content, pathToWrite, runAs, true)).To(BeNil())

					content, err := ioutil.ReadFile(pathToWrite)
					Expect(err).To(BeNil())

					Expect(content).To(Equal([]byte(`hello world`)))
				})

				Context("when the user does not have the permission to write to the given path", func() {
					BeforeEach(func() {
						Expect(os.Chmod(pathToWrite, 0400)).To(BeNil())
					})

					It("returns an error", func() {
						Expect(
							atomicwrite.AtomicWrite(content, pathToWrite, runAs, true),
						).To(Equal(atomicwrite.ErrPermissionDenied))

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
						Expect(
							atomicwrite.AtomicWrite(content, pathToWrite, runAs, true),
						).To(Equal(atomicwrite.ErrPathIsDirectory))
					})
				})

				Context("when the file is a symlink", func() {
					var symlinkPath string

					BeforeEach(func() {
						symlinkPath = filepath.Join(dir, "symlink.txt")
						Expect(os.Symlink(pathToWrite, symlinkPath)).To(BeNil())
					})

					It("writes to the actual file to which the symlink is pointing", func() {
						Expect(atomicwrite.AtomicWrite(content, symlinkPath, runAs, true)).To(BeNil())

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
							Expect(
								atomicwrite.AtomicWrite(content, symlinkPath, runAs, true),
							).To(Equal(atomicwrite.ErrSymlinkTooDeep))
						})
					})
				})
			})

			Context("when the overwrite param is false", func() {
				It("returns an error", func() {
					Expect(
						atomicwrite.AtomicWrite(content, pathToWrite, runAs, false),
					).To(Equal(atomicwrite.ErrFileExist))

					content, err := ioutil.ReadFile(pathToWrite)
					Expect(err).To(BeNil())

					Expect(content).NotTo(Equal([]byte(`hello world`)))
				})

				Context("when the user does not have the permission to write to the given path", func() {
					BeforeEach(func() {
						Expect(os.Chmod(pathToWrite, 0400)).To(BeNil())
					})

					It("returns an error", func() {
						Expect(
							atomicwrite.AtomicWrite(content, pathToWrite, runAs, false),
						).To(Equal(atomicwrite.ErrPermissionDenied))

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
						Expect(
							atomicwrite.AtomicWrite(content, pathToWrite, runAs, false),
						).To(Equal(atomicwrite.ErrPathIsDirectory))
					})
				})

				Context("when the file is a symlink", func() {
					var symlinkPath string

					BeforeEach(func() {
						symlinkPath = filepath.Join(dir, "symlink.txt")
						Expect(os.Symlink(pathToWrite, symlinkPath)).To(BeNil())
					})

					It("returns an error if symlink points to a file", func() {
						Expect(
							atomicwrite.AtomicWrite(content, symlinkPath, runAs, false),
						).To(Equal(atomicwrite.ErrFileExist))

						content, err := ioutil.ReadFile(pathToWrite)
						Expect(err).To(BeNil())

						Expect(content).NotTo(Equal([]byte(`hello world`)))
					})

					Context("when the symlink is broken", func() {
						BeforeEach(func() {
							os.Remove(symlinkPath)
							brokenSymlinkPath := filepath.Join(dir, "broken.txt")
							Expect(os.Symlink(brokenSymlinkPath, symlinkPath)).To(BeNil())
							Expect(os.Symlink(symlinkPath, brokenSymlinkPath)).To(BeNil())
						})

						It("returns an error", func() {
							Expect(
								atomicwrite.AtomicWrite(content, symlinkPath, runAs, false),
							).To(Equal(atomicwrite.ErrSymlinkTooDeep))
						})
					})
				})
			})
		})
	})
})
