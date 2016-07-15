package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/nitrous-io/nitrous-solo/pkg/ws"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type resCollection struct {
	responses [][]byte
	index     int
}

func (r *resCollection) next() []byte {
	if len(r.responses) < r.index+1 {
		return []byte{}
	}

	res := r.responses[r.index]
	r.index = r.index + 1
	return res
}

var _ = Describe("fs", func() {
	var rc resCollection

	cb := func(r []byte) {
		rc.responses = append(rc.responses, r)
	}

	parseResponseJSON := func(res []byte) (j map[string]interface{}) {
		err := json.Unmarshal(res, &j)
		Expect(err).To(BeNil())
		return j
	}

	BeforeEach(func() {
		rc = resCollection{}
		SetDefaultEventuallyTimeout(10 * time.Second)
	})

	AfterEach(func() {
		SetDefaultEventuallyTimeout(1 * time.Second)
	})

	Describe("fs.Watch", func() {
		var onClose chan struct{}

		BeforeEach(func() {
			onClose = make(chan struct{}, 1)
		})

		AfterEach(func() {
			// close connection
			onClose <- struct{}{}
		})

		Context("when invalid JSON is given", func() {
			It("returns an error", func() {
				Watch(&ws.Context{
					Data:    []byte(`{invalid JSON`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["error"]).To(Equal("invalid_json"))
			})
		})

		Context("when the given path does not exist", func() {
			It("returns an error", func() {
				Watch(&ws.Context{
					Data:    []byte(`{"path": "/does-not-exist"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})
				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["error"]).To(Equal("invalid_path"))
			})
		})

		Context("when the given path is not a directory", func() {
			It("returns an error", func() {
				tmpfile, err := ioutil.TempFile("", "fs-watch-test")
				Expect(err).To(BeNil())
				defer os.Remove(tmpfile.Name())

				Watch(&ws.Context{
					Data:    []byte(`{"path": "` + tmpfile.Name() + `"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["error"]).To(Equal("path_not_directory"))
			})
		})

		Context("when the given path is a directory", func() {
			var testDir string

			BeforeEach(func() {
				var err error
				testDir, err = ioutil.TempDir("/tmp", "subdirectories-test")
				Expect(err).To(BeNil())

				// Create some files in the directory.
				f, err := os.Create(filepath.Join(testDir, "README.md"))
				Expect(err).To(BeNil())
				f.Close()

				err = os.MkdirAll(filepath.Join(testDir, "script"), 0755)
				Expect(err).To(BeNil())
				f, err = os.Create(filepath.Join(testDir, "script", "server"))
				Expect(err).To(BeNil())
				f.Close()
			})

			AfterEach(func() {
				os.RemoveAll(testDir)
			})

			It("responds with events when there are filesystem changes in the directory", func() {
				Watch(&ws.Context{
					Data:    []byte(`{"path": "` + testDir + `"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["evt"]).To(Equal("started"))
				watchID, ok := j["id"].(string)
				Expect(ok).To(BeTrue())
				Expect(watchID).NotTo(Equal(""))

				// There is no guarantee that ioutil.WriteFile writes to disk all at once,
				// so it was sometimes producing multiple "modified" events due to multiple
				// writes happening, causing random test failures. Write only one byte to ensure Write().
				err := ioutil.WriteFile(filepath.Join(testDir, "README.md"), []byte("h"), 0644)
				Expect(err).To(BeNil())

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["path"]).To(Equal(filepath.Join(testDir, "README.md")))
				Expect(j["evt"]).To(Equal("modified"))

				// Create a file.
				f, err := os.Create(filepath.Join(testDir, "CHANGELOG.md"))
				Expect(err).To(BeNil())
				f.Close()

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["path"]).To(Equal(filepath.Join(testDir, "CHANGELOG.md")))
				Expect(j["evt"]).To(Equal("created"))

				// Delete a file.
				err = os.Remove(filepath.Join(testDir, "README.md"))
				Expect(err).To(BeNil())

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["path"]).To(Equal(filepath.Join(testDir, "README.md")))
				Expect(j["evt"]).To(Equal("removed"))

				// Rename a file - this will trigger a rename and a create event, in
				// any order.
				err = os.Rename(
					filepath.Join(testDir, "script", "server"),
					filepath.Join(testDir, "script", "server-old"),
				)
				Expect(err).To(BeNil())

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j1 := parseResponseJSON(res)

				res = rc.next()
				Expect(res).NotTo(BeEmpty())
				j2 := parseResponseJSON(res)

				Expect([]map[string]interface{}{j1, j2}).To(ConsistOf(
					map[string]interface{}{
						"id":   watchID,
						"evt":  "renamed",
						"path": filepath.Join(testDir, "script", "server"),
					},
					map[string]interface{}{
						"id":   watchID,
						"evt":  "created",
						"path": filepath.Join(testDir, "script", "server-old"),
					},
				))
			})

			It("watches any newly created subdirectories for changes", func() {
				Watch(&ws.Context{
					Data:    []byte(`{"path": "` + testDir + `"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["evt"]).To(Equal("started"))
				watchID, ok := j["id"].(string)
				Expect(ok).To(BeTrue())
				Expect(watchID).NotTo(Equal(""))

				// Create subdirectory 2 levels deep.
				err := os.MkdirAll(filepath.Join(testDir, "spec", "helpers"), 0755)
				Expect(err).To(BeNil())

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["path"]).To(Equal(filepath.Join(testDir, "spec")))
				Expect(j["evt"]).To(Equal("created"))

				// The spec/helpers directory does not trigger an event.

				// Make changes in the spec/helpers directory.
				f, err := os.Create(filepath.Join(testDir, "spec", "helpers", "app_helper.rb"))
				Expect(err).To(BeNil())
				f.Close()

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["path"]).To(Equal(filepath.Join(testDir, "spec", "helpers", "app_helper.rb")))
				Expect(j["evt"]).To(Equal("created"))
			})
		})

		Context("when the given path is a symlink to a directory", func() {
			It("responds with events from the source directory the symlink resolves to", func() {
				testDir1, err := ioutil.TempDir("/tmp", "subdirectories-test-1")
				Expect(err).To(BeNil())
				defer os.RemoveAll(testDir1)
				testDir2, err := ioutil.TempDir("/tmp", "subdirectories-test-2")
				Expect(err).To(BeNil())
				defer os.RemoveAll(testDir2)

				symlinkDir := testDir2 + "/symlinkdir"
				err = os.Symlink(testDir1, symlinkDir)
				Expect(err).To(BeNil())

				Watch(&ws.Context{
					Data:    []byte(`{"path": "` + symlinkDir + `"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["evt"]).To(Equal("started"))
				watchID, ok := j["id"].(string)
				Expect(ok).To(BeTrue())
				Expect(watchID).NotTo(Equal(""))

				// Create a file.
				f, err := os.Create(filepath.Join(testDir1, "CHANGELOG.md"))
				Expect(err).To(BeNil())
				f.Close()

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				// We currently return the _source_ directory's path (instead of the
				// symlink directory's path), which is probably not good. Let's change
				// it when we know for sure after we integrate with the IDE.
				Expect(j["path"]).To(Equal(filepath.Join(testDir1, "CHANGELOG.md")))
				Expect(j["evt"]).To(Equal("created"))
			})
		})

		Context("when OnClose channel receives a message", func() {
			It("stops watching a directory for changes and calls ws.Context.Done", func() {
				testDir, err := ioutil.TempDir("/tmp", "subdirectories-test")
				Expect(err).To(BeNil())
				defer os.RemoveAll(testDir)

				// Create some files in the directory.
				f, err := os.Create(filepath.Join(testDir, "README.md"))
				Expect(err).To(BeNil())
				f.Close()
				err = os.MkdirAll(filepath.Join(testDir, "script"), 0755)
				Expect(err).To(BeNil())
				f, err = os.Create(filepath.Join(testDir, "script", "server"))
				Expect(err).To(BeNil())
				f.Close()

				doneCalled := false
				doneFn := func() {
					doneCalled = true
				}

				Watch(&ws.Context{
					Data:    []byte(`{"path": "` + testDir + `"}`),
					Reply:   cb,
					OnClose: onClose,
					Done:    doneFn,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["evt"]).To(Equal("started"))
				watchID, ok := j["id"].(string)
				Expect(ok).To(BeTrue())
				Expect(watchID).NotTo(Equal(""))

				onClose <- struct{}{}

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				Expect(res).NotTo(BeEmpty())
				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["evt"]).To(Equal("stopped"))

				// done function should have been called to clean up onClose hash
				Expect(doneCalled).To(BeTrue())

				// Modify a file and make sure we don't get any events.
				err = ioutil.WriteFile(filepath.Join(testDir, "README.md"), []byte("h"), 0644)
				Expect(err).To(BeNil())

				time.Sleep(10 * time.Millisecond) // wait a little to give fsnotify a chance to detect changes
				Expect(rc.next()).To(BeEmpty())
			})
		})
	})

	Describe("fs.Unwatch", func() {
		It("stops watching a directory for changes", func() {
			testDir, err := ioutil.TempDir("/tmp", "subdirectories-test")
			Expect(err).To(BeNil())
			defer os.RemoveAll(testDir)

			// Create some files in the directory.
			f, err := os.Create(filepath.Join(testDir, "README.md"))
			Expect(err).To(BeNil())
			f.Close()
			err = os.MkdirAll(filepath.Join(testDir, "script"), 0755)
			Expect(err).To(BeNil())
			f, err = os.Create(filepath.Join(testDir, "script", "server"))
			Expect(err).To(BeNil())
			f.Close()

			Watch(&ws.Context{
				Data:  []byte(`{"path": "` + testDir + `"}`),
				Reply: cb,
				Done:  Noop,
			})

			res := rc.next()
			Expect(res).NotTo(BeEmpty())

			j := parseResponseJSON(res)
			Expect(j["evt"]).To(Equal("started"))
			watchID, ok := j["id"].(string)
			Expect(ok).To(BeTrue())
			Expect(watchID).NotTo(Equal(""))

			Unwatch(&ws.Context{
				Data:  []byte(`{"id": "` + watchID + `"}`),
				Reply: cb,
				Done:  Noop,
			})

			Eventually(func() Any {
				res = rc.next()
				return res
			}).ShouldNot(BeEmpty())

			Expect(res).NotTo(BeEmpty())
			j = parseResponseJSON(res)
			Expect(j["id"]).To(Equal(watchID))
			Expect(j["evt"]).To(Equal("stopped"))

			// Modify a file and make sure we don't get any events.
			err = ioutil.WriteFile(filepath.Join(testDir, "README.md"), []byte("h"), 0644)
			Expect(err).To(BeNil())

			time.Sleep(10 * time.Millisecond) // wait a little to give fsnotify a chance to detect changes
			Expect(rc.next()).To(BeEmpty())
		})

		Context("when unwatch is performed twice", func() {
			It("doesn't deadlock and the 2nd unwatch is basically a no-op", func() {
				// This test is to ensure we don't deadlock (e.g. sending on a blocked
				// channel) because of the synchronization code.

				testDir, err := ioutil.TempDir("/tmp", "subdirectories-test")
				Expect(err).To(BeNil())
				defer os.RemoveAll(testDir)

				Watch(&ws.Context{
					Data:  []byte(`{"path": "` + testDir + `"}`),
					Reply: cb,
					Done:  Noop,
				})

				res := rc.next()
				Expect(res).NotTo(BeEmpty())

				j := parseResponseJSON(res)
				Expect(j["evt"]).To(Equal("started"))
				watchID, ok := j["id"].(string)
				Expect(ok).To(BeTrue())
				Expect(watchID).NotTo(Equal(""))

				Unwatch(&ws.Context{
					Data:  []byte(`{"id": "` + watchID + `"}`),
					Reply: cb,
					Done:  Noop,
				})

				Eventually(func() Any {
					res = rc.next()
					return res
				}).ShouldNot(BeEmpty())

				j = parseResponseJSON(res)
				Expect(j["id"]).To(Equal(watchID))
				Expect(j["evt"]).To(Equal("stopped"))

				Unwatch(&ws.Context{
					Data:  []byte(`{"id": "` + watchID + `"}`),
					Reply: cb,
					Done:  Noop,
				})

				time.Sleep(10 * time.Millisecond) // wait a little to see if another message is received
				// There should only be a single "unwatched" event, and no "stopped"
				// event (since the previous unwatch already stopped it).
				Expect(rc.next()).To(BeEmpty())
			})
		})
	})

	Describe("dirtree()", func() {
		var testDir string

		setupTestFiles := func(dir string, files []string) {
			for _, p := range files {
				path := filepath.Join(testDir, p)
				dir := filepath.Dir(path)
				err := os.MkdirAll(dir, 0755)
				Expect(err).To(BeNil())
				f, err := os.Create(path)
				Expect(err).To(BeNil())
				f.Close()
			}
		}

		BeforeEach(func() {
			var err error
			testDir, err = ioutil.TempDir("/tmp", "subdirectories-test")
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			os.RemoveAll(testDir)
		})

		It("returns a tree of all directories in a given directory, including the directory itself", func() {
			files := []string{
				"README.md",
				"controllers/ping/ping.go",
				"controllers/ping/ping_test.go",
				"testhelper/fixtures/fs/file.txt",
				".git/HEAD",
				".git/refs/remotes/origin/HEAD",
				".git/refs/remotes/origin/release",
			}
			setupTestFiles(testDir, files)

			dirs, err := dirtree(testDir)
			Expect(err).To(BeNil())

			Expect(dirs).To(ConsistOf(
				testDir,
				testDir+"/controllers",
				testDir+"/controllers/ping",
				testDir+"/testhelper",
				testDir+"/testhelper/fixtures",
				testDir+"/testhelper/fixtures/fs",
				testDir+"/.git",
				testDir+"/.git/refs",
				testDir+"/.git/refs/remotes",
				testDir+"/.git/refs/remotes/origin",
			))
		})

		It("does not resolve subdirectories that are actually symlinks", func() {
			files := []string{
				"README.md",
				"controllers/ping/ping.go",
			}
			setupTestFiles(testDir, files)

			testDir2, err := ioutil.TempDir("/tmp", "subdirectories-test-2")
			Expect(err).To(BeNil())
			defer os.RemoveAll(testDir2)

			err = os.Symlink(testDir2, testDir+"/symlinkdir")
			Expect(err).To(BeNil())

			dirs, err := dirtree(testDir)
			Expect(err).To(BeNil())

			Expect(dirs).To(ConsistOf(
				testDir,
				testDir+"/controllers",
				testDir+"/controllers/ping",
			))
		})

		Context("when given directory is a symlink", func() {
			It("it resolves symlink to the source directory", func() {
				files := []string{
					"README.md",
					"controllers/ping/ping.go",
				}
				setupTestFiles(testDir, files)

				testDir2, err := ioutil.TempDir("/tmp", "subdirectories-test-2")
				Expect(err).To(BeNil())
				defer os.RemoveAll(testDir2)

				symlinkDir := testDir2 + "/symlinkdir"
				err = os.Symlink(testDir, symlinkDir)
				Expect(err).To(BeNil())

				dirs, err := dirtree(symlinkDir)
				Expect(err).To(BeNil())

				Expect(dirs).To(ConsistOf(
					testDir,
					testDir+"/controllers",
					testDir+"/controllers/ping",
				))
			})
		})
	})
})
