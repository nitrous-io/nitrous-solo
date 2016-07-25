package user_test

import (
	"net/http"
	"net/http/httptest"
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
	RunSpecs(t, "user")
}

var _ = Describe("user", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error
		tok string

		ud      *data.UserData
		headers http.Header
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

		tok, err = token.GenerateTokenString("foobar", token.LoginType)
		Expect(err).To(BeNil())

		headers = http.Header{
			"Authorization": {"Token " + tok},
		}
	})

	AfterEach(func() {
		if res != nil {
			res.Body.Close()
		}
		s.Close()
	})

	Describe("GET /user", func() {
		doRequest := func() {
			res, err = MakeRequest("GET", s.URL+"/user", nil, headers, nil)
			Expect(err).To(BeNil())
		}

		Context("with a valid access token", func() {
			It("returns 200 OK and a json message containing user info", func() {
				doRequest()
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

		sharedexamples.ItRequiresAuthentication(func() (*data.UserData, *http.Header) {
			return ud, &headers
		}, func() *http.Response {
			doRequest()
			return res
		})
	})
})
