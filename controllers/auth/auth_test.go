package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/nitrous-io/nitrous-solo/controllers/auth"
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
	RunSpecs(t, "auth")
}

var _ = Describe("auth", func() {
	var (
		ud *data.UserData

		s   *httptest.Server
		res *http.Response
		err error
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New(""))

		data.D = data.Reset()
		ud = &data.UserData{
			Login: "foobar",
			User:  "nitrous",
		}
		err := ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.Users[ud.Login] = ud
	})

	AfterEach(func() {
		if res != nil {
			res.Body.Close()
		}
		s.Close()
	})

	Describe("POST /token", func() {
		doRequest := func(login, password string) {
			res, err = http.PostForm(s.URL+"/token", url.Values{
				"login":    {login},
				"password": {password},
			})
		}

		Context("when login is blank", func() {
			It("returns an error", func() {
				doRequest("", "foo123")
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_params"))
				Expect(j["error_description"]).To(Equal("login and password should not be blank"))
			})
		})

		Context("when password is blank", func() {
			It("returns an error", func() {
				doRequest("foobar", "")
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_params"))
				Expect(j["error_description"]).To(Equal("login and password should not be blank"))
			})
		})

		Context("when login is wrong", func() {
			It("returns an error", func() {
				doRequest("foobro", "foo123")
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_credentials"))
				Expect(j["error_description"]).To(Equal("invalid login or password"))
			})
		})

		Context("when password is wrong", func() {
			It("returns an error", func() {
				doRequest("foobar", "foo456")
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_credentials"))
				Expect(j["error_description"]).To(Equal("invalid login or password"))
			})
		})

		Context("when credentials are valid", func() {
			It("returns a valid token", func() {
				doRequest("foobar", "foo123")
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusOK))
				j := ParseJSONBody(res.Body)
				t, ok := j["token"].(string)
				Expect(ok).To(BeTrue())
				Expect(t).NotTo(BeEmpty())
				Expect(token.ValidateTokenString(t, token.LoginType)).To(Equal("foobar"))
			})
		})
	})

	Describe("POST /token/temp", func() {
		var (
			tok                 string
			headers             http.Header
			params              url.Values
			origTempTokenExpiry time.Duration
		)

		BeforeEach(func() {
			tok, err = token.GenerateTokenString("foobar", token.LoginType)
			Expect(err).To(BeNil())

			headers = http.Header{
				"Authorization": {"Token " + tok},
			}

			origTempTokenExpiry = auth.DefaultTemporaryTokenExpiry
			auth.DefaultTemporaryTokenExpiry = 10 * time.Millisecond
		})

		AfterEach(func() {
			auth.DefaultTemporaryTokenExpiry = origTempTokenExpiry
		})

		doRequest := func() {
			res, err = MakeRequest("POST", s.URL+"/token/temp", params, headers, nil)
			Expect(err).To(BeNil())
		}

		sharedexamples.ItRequiresAuthentication(func() (*data.UserData, *http.Header) {
			return ud, &headers
		}, func() *http.Response {
			doRequest()
			return res
		})

		Context("when vaild token is given", func() {
			It("returns a valid temporary token", func() {
				doRequest()
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusOK))
				j := ParseJSONBody(res.Body)
				t, ok := j["token"].(string)
				Expect(ok).To(BeTrue())
				Expect(t).NotTo(BeEmpty())
				Expect(token.ValidateTokenString(t, token.TempType)).To(Equal("foobar"))
			})

			It("the token expires after default expiry", func() {
				doRequest()
				Expect(err).To(BeNil())

				Expect(res.StatusCode).To(Equal(http.StatusOK))
				j := ParseJSONBody(res.Body)
				t, ok := j["token"].(string)
				Expect(ok).To(BeTrue())
				Expect(t).NotTo(BeEmpty())
				Expect(token.ValidateTokenString(t, token.TempType)).To(Equal("foobar"))

				time.Sleep(10 * time.Millisecond)

				Expect(token.ValidateTokenString(t, token.TempType)).To(BeEmpty())
			})
		})
	})
})
