package files_test

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	"github.com/nitrous-io/nitrous-solo/server"
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
	)

	BeforeEach(func() {
		dir, err = ioutil.TempDir("", "nitrous-solo-upload-test")
		s = httptest.NewServer(server.New(dir))

		data.D = data.Reset()
		ud := &data.UserData{
			Login: "foobar",
			User:  "nitrous",
		}
		err := ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.Users[ud.Login] = ud

		tok, err = token.GenerateTokenString("foobar")
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
			location string
			headers  http.Header
		)

		doRequest := func(location, partName, content string) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormField("location")
			Expect(err).To(BeNil())

			_, err = part.Write([]byte(location))
			Expect(err).To(BeNil())

			part, err = writer.CreateFormFile(partName, location)
			Expect(err).To(BeNil())

			_, err = part.Write([]byte(content))
			Expect(err).To(BeNil())

			Expect(writer.Close()).To(BeNil())

			req, err := http.NewRequest("POST", s.URL+"/files/upload", body)
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
			req, err := http.NewRequest("POST", s.URL+"/files/upload", nil)
			Expect(err).To(BeNil())

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

		BeforeEach(func() {
			location = filepath.Join(dir, "foo-bar.txt")
			headers = http.Header{
				"Authorization": {"Token " + tok},
			}
		})

		It("returns 200 OK", func() {
			doRequest(location, "files", "hello\nworld")

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(b.String()).To(MatchJSON(`{"status":"ok"}`))
		})

		It("creates the file in root folder", func() {
			doRequest(location, "files", "hello\nworld")

			content, err := ioutil.ReadFile(location)
			Expect(err).To(BeNil())

			Expect(string(content)).To(Equal("hello\nworld"))
		})

		Context("when the file exists at the location", func() {
			BeforeEach(func() {
				ioutil.WriteFile(location, []byte("existing content"), 0644)
			})

			It("overwrites the file", func() {
				doRequest(location, "files", "hello\nworld")

				content, err := ioutil.ReadFile(location)
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
				doRequest("", "files", "hello\nworld")

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

		Context("when token is missing", func() {
			BeforeEach(func() {
				headers = http.Header{}
			})

			It("returns 401 with invalid_token", func() {
				doRequest(location, "files", "hello\nworld")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_token",
					"error_description": "access token is required"
				}`))
			})
		})

		Context("when token is wrong", func() {
			BeforeEach(func() {
				tok, err = token.GenerateTokenString("foobarqux")
				Expect(err).To(BeNil())

				headers = http.Header{
					"Authorization": {"Token " + tok},
				}
			})

			It("returns 401 with invalid_token", func() {
				doRequest(location, "files", "hello\nworld")

				b := &bytes.Buffer{}
				_, err = b.ReadFrom(res.Body)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(b.String()).To(MatchJSON(`{
					"error": "invalid_token",
					"error_description": "access token is invalid"
				}`))
			})
		})
	})
})
