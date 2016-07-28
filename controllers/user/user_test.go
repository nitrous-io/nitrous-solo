package user_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/user"
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

var currentSystemUser *user.User

func init() {
	currentSystemUser, _ = user.Current()
}

var _ = Describe("user", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error
		tok string

		ud      *data.UserData
		headers http.Header

		origDataDirPath string
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New(""))

		origDataDirPath = data.DataDirPath
		data.DataDirPath, err = ioutil.TempDir("", "nitrous-test")
		Expect(err).To(BeNil())

		data.D = data.Reset()
		ud = &data.UserData{
			Login: "foobar",
			User:  currentSystemUser.Username,
		}
		err := ud.SetPassword("foo123")
		Expect(err).To(BeNil())
		data.D.UserData[ud.Login] = ud
		data.D.IDESettings[ud.Login] = data.IDESettings{
			"font-size": float64(10),
			"tab-size":  float64(4),
		}

		tok, err = token.GenerateTokenString("foobar", token.LoginType)
		Expect(err).To(BeNil())

		headers = http.Header{
			"Authorization": {"Token " + tok},
		}
	})

	AfterEach(func() {
		data.DataDirPath = origDataDirPath

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

				Expect(ReadBody(res)).To(MatchJSON(`{
					"user": {
						"login": "foobar",
						"user": "` + currentSystemUser.Username + `",
						"home": "` + currentSystemUser.HomeDir + `",
						"ide_settings": {
							"font-size": 10,
							"tab-size": 4
						}
					}
				}`))
			})
		})

		sharedexamples.ItRequiresAuthentication(func() (*data.UserData, *http.Header) {
			return ud, &headers
		}, func() *http.Response {
			doRequest()
			return res
		})
	})

	Describe("PUT /user/ide_settings", func() {
		var ide map[string]interface{}

		doRequest := func() {
			body, err := json.Marshal(ide)
			Expect(err).To(BeNil())

			req, err := http.NewRequest("PUT", s.URL+"/user/ide_settings", bytes.NewBuffer(body))
			Expect(err).To(BeNil())

			req.Header.Set("Content-Type", "application/json")
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

		It("updates ide data", func() {
			ide = map[string]interface{}{
				"font-size":   float64(12),
				"key-binding": "vim",
			}
			doRequest()

			Expect(res.StatusCode).To(Equal(http.StatusOK))

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(b.String()).To(MatchJSON(`{
				"ide_settings": {
					"font-size": 12,
					"tab-size": 4,
					"key-binding": "vim"
				}
			}`))

			// check that data in memory is updated
			d := data.D.IDESettings[ud.Login]
			Expect(d).NotTo(BeNil())
			Expect(d["font-size"]).To(Equal(float64(12)))
			Expect(d["tab-size"]).To(Equal(float64(4)))
			Expect(d["key-binding"]).To(Equal("vim"))

			// load ide data from file and check that it is updated
			err := data.LoadIDESettings()
			Expect(err).To(BeNil())
			d = data.D.IDESettings[ud.Login]
			Expect(d).NotTo(BeNil())
			Expect(d["font-size"]).To(Equal(float64(12)))
			Expect(d["tab-size"]).To(Equal(float64(4)))
			Expect(d["key-binding"]).To(Equal("vim"))
		})

		sharedexamples.ItRequiresAuthentication(func() (*data.UserData, *http.Header) {
			return ud, &headers
		}, func() *http.Response {
			doRequest()
			return res
		})
	})
})
