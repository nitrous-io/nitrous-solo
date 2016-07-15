package user_test

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
	RunSpecs(t, "user")
}

var _ = Describe("user", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error
		tok string
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

		tok, err = token.GenerateTokenString("foobar")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		if res != nil {
			res.Body.Close()
		}
		s.Close()
	})

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
						"user": "nitrous"
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
