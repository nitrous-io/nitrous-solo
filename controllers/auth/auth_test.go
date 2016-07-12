package auth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
	RunSpecs(t, "auth")
}

var _ = Describe("auth", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New(""))

		data.D = data.Reset()
		ud := &data.UserData{
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
				Expect(token.ValidateTokenString(t)).To(Equal("foobar"))
			})
		})
	})
})
