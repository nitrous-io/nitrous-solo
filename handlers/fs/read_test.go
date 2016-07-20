package fs_test

import (
	"encoding/base64"
	"path/filepath"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("fs.Read", func() {
	var res []byte

	cb := func(r []byte) {
		res = r
	}

	BeforeEach(func() {
		res = nil
	})

	Context("when the file at the given path has a recognized text file extension", func() {
		It("returns file content", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "code.rb")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "text",
				"content": base64.StdEncoding.EncodeToString([]byte("puts 'hello world'\n")),
			})))
		})
	})

	Context("when the file at the given path has a recognized text file filename", func() {
		It("returns file content", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "Gemfile")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "text",
				"content": base64.StdEncoding.EncodeToString([]byte("source 'https://rubygems.org'\n")),
			})))
		})
	})

	Context("when the file at the given path is empty", func() {
		It("returns file content", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "foo")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "text",
				"content": "",
			})))
		})
	})

	Context("when the file at the given path has a recognized image file extension", func() {
		It("returns file content, encoded in base64", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "tiny.gif")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "image",
				"content": "R0lGODlhAQABAIABAP///wAAACwAAAAAAQABAAACAkQBADs=",
			})))
		})
	})

	Context("when the file at the given path is recognized as a text file", func() {
		It("returns file content", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "baz")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "text",
				"content": base64.StdEncoding.EncodeToString([]byte("hello world i am a text file with no extension\n")),
			})))
		})
	})

	Context("when the file at the given path is an unrecognized file", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "binaryfile")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "unrecognized_file_type",
			})))
		})
	})

	Context("when the given path is a symlink to a readable file", func() {
		It("returns file content", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "filesymlink")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"type":    "text",
				"content": base64.StdEncoding.EncodeToString([]byte("puts 'hello world'\n")),
			})))
		})
	})

	Context("when the given path is an empty string", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": ""}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "invalid_path",
			})))
		})
	})

	Context("when the given path does not exist", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "code.java")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "invalid_path",
			})))
		})
	})

	Context("when the given path is a directory", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "bar")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "path_not_file",
			})))
		})
	})

	Context("when the given path is a symlink to a directory", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "dirsymlink")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "path_not_file",
			})))
		})
	})

	Context("when the given path is a broken symlink", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "brokensymlink")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "invalid_path",
			})))
		})
	})

	Context("when the given path is a recursive symlink", func() {
		It("returns an error", func() {
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "recursymlink1")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "symlink_too_deep",
			})))
		})
	})

	Context("when the file at the given path has a size that is larger than the limit", func() {
		It("returns an error", func() {
			origLimit := fs.FileSizeLimit
			fs.FileSizeLimit = int64(10)
			defer func() {
				fs.FileSizeLimit = origLimit
			}()
			fs.Read(&ws.Context{
				Data:  EncodeJSON(H{"path": filepath.Join(fixturesPath, "fs", "code.rb")}),
				Reply: cb,
				Done:  Noop,
			})
			Expect(res).NotTo(BeNil())

			Expect(res).To(MatchJSON(EncodeJSON(H{
				"error": "file_too_big",
			})))
		})
	})
})
