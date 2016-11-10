package ping_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/user"
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
	RunSpecs(t, "ping")
}

var currentSystemUser *user.User

func init() {
	currentSystemUser, _ = user.Current()
}

var _ = Describe("ping", func() {
	var (
		s   *httptest.Server
		res *http.Response

		headers http.Header
		err     error
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New(""))

		data.D = data.Reset()
		data.CurrentUser = currentSystemUser.Username

		ud := &data.UserData{
			Login: "foobar",
			User:  currentSystemUser.Username,
		}

		err := ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.UserData[ud.Login] = ud

		tok, err := token.GenerateTokenString("foobar", token.LoginType)
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

	It("returns 200 OK and a json message containing pong", func() {
		res, err = MakeRequest("GET", s.URL+"/ping", nil, headers, nil)
		Expect(err).To(BeNil())

		var j map[string]string
		err = json.NewDecoder(res.Body).Decode(&j)
		Expect(err).To(BeNil())

		Expect(res.StatusCode).To(Equal(http.StatusOK))
		Expect(j["message"]).To(Equal("pong"))
	})
})
