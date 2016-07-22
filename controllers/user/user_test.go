package user_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	"github.com/nitrous-io/nitrous-solo/server"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user")
}

var _ = Describe("user", func() {
	var (
		s               *httptest.Server
		res             *http.Response
		err             error
		tok             string
		origDataDirPath string
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New(""))

		data.D = data.Reset()

		origDataDirPath = data.DataDirPath
		currPath, _ := filepath.Abs(".")
		fixturesPath := filepath.Join(currPath, "..", "..", "testhelper", "fixtures")
		data.DataDirPath = filepath.Join(fixturesPath, "nitrous-solo")

		ud := &data.UserData{
			Login: "foobar",
			User:  "nitrous",
		}

		ud.SetDefaultSettings()
		err = ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.Users[ud.Login] = ud

		tok, err = token.GenerateTokenString("foobar")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if res != nil {
			res.Body.Close()
		}
		s.Close()
		os.RemoveAll(data.DataDirPath)
		data.DataDirPath = origDataDirPath
	})

	Describe("GET /user", func() {
		Context("with a valid access token", func() {
			It("returns 200 OK and a json message containing user info", func() {
				res, err = MakeRequest("GET", s.URL+"/user", url.Values{}, http.Header{
					"Authorization": {"Token " + tok},
				}, nil)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusOK))

				Expect(ReadBody(res)).To(MatchJSON(`
					{
						"user": {
							"login": "foobar",
							"user": "nitrous",
							"auto_close_brackets": true,
							"font_size": 14,
							"highlight_line": true,
							"keybindings": "default",
							"match_brackets": true,
							"remove_trailing_whitespace_on_save": true,
							"ruler_column": 0,
							"show_line_numbers": true,
							"show_trailing_space": false,
							"soft_tabs": true,
							"tab_size": 2,
							"theme": "solarized dark",
							"wrap_lines": true,
							"root_path": "/home/nitrous",
							"home_path": "/home/nitrous",
							"readme_url": "https://nitrous-readmes.s3.amazonaws.com/ruby.2015-06-29.html",
							"readme_opened": false,
							"open_tabs": []
						}
					}
				`))
			})
		})

		Context("with an invalid access token", func() {
			It("returns 401 Unauthorized", func() {
				tok, err = token.GenerateTokenString("barqux")
				Expect(err).To(BeNil())

				res, err = MakeRequest("GET", s.URL+"/user", url.Values{}, http.Header{
					"Authorization": {"Token " + tok},
				}, nil)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))

				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_token"))
				Expect(j["error_description"]).To(Equal("access token is invalid"))
			})
		})

		Context("with no access token", func() {
			It("returns 401 Unauthorized", func() {
				res, err = MakeRequest("GET", s.URL+"/user", url.Values{}, http.Header{}, nil)
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))

				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_token"))
				Expect(j["error_description"]).To(Equal("access token is required"))
			})
		})
	})

	Describe("POST /user", func() {
		doRequest := func(json string) (*http.Response, error) {
			var jsonStr = []byte(json)
			req, err := http.NewRequest("POST", s.URL+"/user", bytes.NewBuffer(jsonStr))
			Expect(err).To(BeNil())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", fmt.Sprintf("Token %s", tok))
			return http.DefaultClient.Do(req)
		}

		It("updates nitrous-solo-data file and returns user data json", func() {
			ud, ok := data.D.Users["foobar"]
			Expect(ok).To(BeTrue())
			Expect(ud.AutoCloseBracket).To(Equal(true))
			Expect(ud.Keybindings).To(Equal("default"))

			json := `{
				"login": "foobar",
				"user": "nitrous",
				"auto_close_brackets": false,
				"font_size": 14,
				"highlight_line": true,
				"keybindings": "vim",
				"match_brackets": true,
				"remove_trailing_whitespace_on_save": true,
				"ruler_column": 0,
				"show_line_numbers": true,
				"show_trailing_space": false,
				"soft_tabs": true,
				"tab_size": 2,
				"theme": "solarized dark",
				"wrap_lines": true,
				"root_path": "/home/nitrous",
				"home_path": "/home/nitrous",
				"readme_url": "https://nitrous-readmes.s3.amazonaws.com/ruby.2015-06-29.html",
				"readme_opened": false,
				"open_tabs": []
			}`
			res, err := doRequest(json)
			Expect(err).To(BeNil())
			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(ReadBody(res)).To(MatchJSON(`
				{
					"user": {
						"login": "foobar",
						"user": "nitrous",
						"auto_close_brackets": false,
						"font_size": 14,
						"highlight_line": true,
						"keybindings": "vim",
						"match_brackets": true,
						"remove_trailing_whitespace_on_save": true,
						"ruler_column": 0,
						"show_line_numbers": true,
						"show_trailing_space": false,
						"soft_tabs": true,
						"tab_size": 2,
						"theme": "solarized dark",
						"wrap_lines": true,
						"root_path": "/home/nitrous",
						"home_path": "/home/nitrous",
						"readme_url": "https://nitrous-readmes.s3.amazonaws.com/ruby.2015-06-29.html",
						"readme_opened": false,
						"open_tabs": []
					}
				}
			`))

			data.Load()

			ud, ok = data.D.Users["foobar"]
			Expect(ok).To(BeTrue())
			Expect(ud.AutoCloseBracket).To(Equal(false))
			Expect(ud.Keybindings).To(Equal("vim"))
		})
	})
})
