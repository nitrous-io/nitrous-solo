package files_test

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	"github.com/nitrous-io/nitrous-solo/server"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	"github.com/nitrous-io/nitrous-solo/testhelper/sharedexamples"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "files")
}

var _ = Describe("files", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error

		dir string
		tok string

		ud         *data.UserData
		runAs      string
		runUserUID uint32
		runUserGID uint32
	)

	BeforeEach(func() {
		dir, err = ioutil.TempDir("", "nitrous-solo-upload-test")
		s = httptest.NewServer(server.New(dir))

		data.D = data.Reset()
		runAs = "nitrous"
		ud = &data.UserData{
			Login: "foobar",
			User:  runAs,
		}
		err := ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.Users[ud.Login] = ud

		u, err := user.Lookup(runAs)

		uid, err := strconv.Atoi(u.Uid)
		Expect(err).To(BeNil())
		runUserUID = uint32(uid)

		gid, err := strconv.Atoi(u.Gid)
		Expect(err).To(BeNil())
		runUserGID = uint32(gid)

		tok, err = token.GenerateTokenString("foobar", token.LoginType)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		os.RemoveAll(dir)

		if res != nil {
			res.Body.Close()
		}
		s.Close()
	})

	Describe("Upload", func() {
		var (
			location     string
			filename     string
			uploadedPath string
			headers      http.Header
		)

		doRequest := func(location, partName, content string) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormField("location")
			Expect(err).To(BeNil())

			_, err = part.Write([]byte(location))
			Expect(err).To(BeNil())

			part, err = writer.CreateFormFile(partName, filename)
			Expect(err).To(BeNil())

			_, err = part.Write([]byte(content))
			Expect(err).To(BeNil())

			Expect(writer.Close()).To(BeNil())

			req, err := http.NewRequest("POST", s.URL+"/files", body)
			Expect(err).To(BeNil())

			req.Header.Set("Content-Type", writer.FormDataContentType())
			if headers != nil {
				for k, v := range headers {
					for _, h := range v {
						req.Header.Add(k, h)
					}
				}
			}

			res, err = http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
		}

		doRequestWithoutMultipart := func() {
			res, err = MakeRequest("POST", s.URL+"/files", nil, headers, nil)
			Expect(err).To(BeNil())
		}

		BeforeEach(func() {
			location = dir
			filename = "foo-bar.txt"
			uploadedPath = filepath.Join(location, filename)

			headers = http.Header{
				"Authorization": {"Token " + tok},
			}
		})

		It("returns 200 OK", func() {
			doRequest(location, "files[]", "hello\nworld")

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(b.String()).To(MatchJSON(`{"status":"ok"}`))
		})

		It("creates the file in root folder as current user", func() {
			doRequest(location, "files[]", "hello\nworld")

			content, err := ioutil.ReadFile(uploadedPath)
			Expect(err).To(BeNil())

			Expect(string(content)).To(Equal("hello\nworld"))

			fi, err := os.Stat(uploadedPath)
			Expect(err).To(BeNil())
			st := fi.Sys().(*syscall.Stat_t)
			Expect(st.Uid).To(Equal(runUserUID))
			Expect(st.Gid).To(Equal(runUserGID))
		})

		Context("when the file exists at the location", func() {
			BeforeEach(func() {
				ioutil.WriteFile(location, []byte("existing content"), 0644)
			})

			It("overwrites the file", func() {
				doRequest(location, "files[]", "hello\nworld")

				content, err := ioutil.ReadFile(uploadedPath)
				Expect(err).To(BeNil())

				Expect(string(content)).To(Equal("hello\nworld"))
			})
		})

		Context("when the request is not multipart", func() {
			It("returns 400 with bad_request", func() {
				doRequestWithoutMultipart()

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_request",
					"error_description": "the request should be encoded in multipart/form-data format"
				}`))
			})
		})

		Context("when location param is missing", func() {
			It("returns 422 with invalid_params", func() {
				doRequest("", "files[]", "hello\nworld")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_params",
					"errors": {
						"location": "is required"
					}
				}`))
			})
		})

		Context("when files param is missing", func() {
			It("returns 422 with invalid_params", func() {
				doRequest(location, "", "")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_params",
					"errors": {
						"files": "is required"
					}
				}`))
			})
		})

		Context("when the user does not exist", func() {
			BeforeEach(func() {
				runAs = "someone-else"
				ud.User = runAs
			})

			It("returns 403 with invalid_request", func() {
				doRequest(location, "files[]", "hello\nworld")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusForbidden))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_request",
					"error_description": "the user could not be found"
				}`))
			})
		})

		Context("when the user does not have permission to write to the path", func() {
			BeforeEach(func() {
				Expect(os.Chmod(dir, 0400)).To(BeNil())
			})

			It("returns 403 with invalid_request", func() {
				doRequest(location, "files[]", "hello\nworld")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusForbidden))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_request",
					"error_description": "the user does not have permission to write to the path"
				}`))
			})
		})

		sharedexamples.ItRequiresAuthentication(func() (*data.UserData, *http.Header) {
			return ud, &headers
		}, func() *http.Response {
			doRequest(location, "files[]", "hello\nworld")
			return res
		})
	})

	Describe("Download", func() {
		var (
			filePath string
			dir      string
			err      error

			params url.Values
		)

		doRequest := func() {
			res, err = MakeRequest("GET", s.URL+"/files", params, nil, nil)
			Expect(err).To(BeNil())
		}

		BeforeEach(func() {
			dir, err = ioutil.TempDir("", "")
			Expect(err).To(BeNil())

			filePath = filepath.Join(dir, "hello.txt")
			Expect(ioutil.WriteFile(filePath, []byte("hello"), 0644)).To(BeNil())

			tok, err = token.GenerateTokenString("foobar", token.TempType)
			Expect(err).To(BeNil())

			params = url.Values{
				"path":  {filePath},
				"token": {tok},
			}
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("returns 200 OK with the file", func() {
			doRequest()

			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(res.Header.Get("Content-Disposition")).To(Equal("attachment; filename=hello.txt"))
			Expect(res.Header.Get("Content-Type")).To(Equal("text/plain; charset=utf-8"))

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())
			Expect(b.Bytes()).To(Equal([]byte(`hello`)))
		})

		Context("the path is directory", func() {
			BeforeEach(func() {
				otherFile := filepath.Join(dir, "world.sh")
				Expect(ioutil.WriteFile(otherFile, []byte("#!/bin/sh\necho 'hello world'\n"), 0755)).To(BeNil())

				otherDir := filepath.Join(dir, "foo")
				Expect(os.Mkdir(otherDir, 0777)).To(BeNil())

				fileInDir := filepath.Join(otherDir, "bar.txt")
				Expect(ioutil.WriteFile(fileInDir, []byte("qux"), 0644)).To(BeNil())

				params.Set("path", dir)
			})

			It("returns 200 OK with zip file", func() {
				doRequest()

				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(res.Header.Get("Content-Disposition")).To(Equal("attachment; filename=" + filepath.Base(dir) + ".zip"))
				Expect(res.Header.Get("Content-Type")).To(Equal("application/zip"))
				Expect(res.Header.Get("Content-Transfer-Encoding")).To(Equal("binary"))

				b, err := ioutil.ReadAll(res.Body)
				r := bytes.NewReader(b)

				zr, err := zip.NewReader(r, int64(len(b)))
				Expect(err).To(BeNil())

				var expectedFiles = map[string]struct {
					body []byte
					perm os.FileMode
				}{
					"foo/bar.txt": {[]byte(`qux`), 0644},
					"hello.txt":   {[]byte(`hello`), 0644},
					"world.sh":    {[]byte("#!/bin/sh\necho 'hello world'\n"), 0755},
				}

				for _, zf := range zr.File {
					f, err := zf.Open()
					defer f.Close()
					Expect(err).To(BeNil())

					expectedFile, ok := expectedFiles[zf.Name]
					Expect(ok).To(BeTrue())

					Expect(zf.Mode().Perm()).To(Equal(expectedFile.perm))

					fb, err := ioutil.ReadAll(f)
					Expect(err).To(BeNil())
					Expect(fb).To(Equal(expectedFile.body))
					delete(expectedFiles, zf.Name)
					f.Close()
				}

				Expect(expectedFiles).To(HaveLen(0))
			})
		})

		Context("when the file does not exist", func() {
			BeforeEach(func() {
				params.Set("path", "nowhere")
			})

			It("returns 422 with invalid_params", func() {
				doRequest()

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_params",
					"errors": {
						"path": "could not be found"
					}
				}`))
			})
		})

		Context("when path is missing", func() {
			BeforeEach(func() {
				params.Del("path")
			})

			It("returns 422 with invalid_params", func() {
				doRequest()

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_params",
					"errors": {
						"path": "is required"
					}
				}`))
			})
		})

		sharedexamples.ItRequiresTempToken(func() (*data.UserData, *url.Values) {
			return ud, &params
		}, func() *http.Response {
			doRequest()
			return res
		})
	})
})
