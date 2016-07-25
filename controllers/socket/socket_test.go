package socket_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	"github.com/nitrous-io/nitrous-solo/server"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "socket")
}

var _ = Describe("socket", func() {
	var (
		s    *httptest.Server
		res  *http.Response
		conn *websocket.Conn
		err  error
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
		if conn != nil {
			conn.WriteMessage(websocket.CloseMessage, []byte("bye"))
			conn.Close()
		}

		if res != nil {
			res.Body.Close()
		}

		s.Close()
	})

	Describe("GET /socket", func() {
		var params url.Values

		BeforeEach(func() {
			params = url.Values{}
		})

		doRequest := func() {
			var wsURL *url.URL
			wsURL, err = url.Parse(fmt.Sprintf("ws://%s/socket", s.Listener.Addr().String()))
			Expect(err).To(BeNil())
			wsURL.RawQuery = params.Encode()

			dialer := websocket.Dialer{HandshakeTimeout: 100 * time.Millisecond}
			conn, res, err = dialer.Dial(wsURL.String(), http.Header{})
		}

		Context("when valid token is given", func() {
			BeforeEach(func() {
				tok, err := token.GenerateTokenString("foobar", token.TempType)
				Expect(err).To(BeNil())

				params = url.Values{
					"token": {tok},
				}
			})

			It("opens websocket connection", func() {
				doRequest()
				Expect(err).To(BeNil())
				Expect(res.StatusCode).To(Equal(http.StatusSwitchingProtocols))
				Expect(res.Header.Get("Sec-WebSocket-Accept")).NotTo(BeEmpty())
			})
		})

		Context("when token is empty", func() {
			It("returns 422 invalid params", func() {
				doRequest()
				Expect(err).NotTo(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_params"))
				Expect(j["errors"]).To(Equal(map[string]interface{}{
					"token": "is required",
				}))
			})
		})

		Context("when token is not valid", func() {
			BeforeEach(func() {
				tok, err := token.GenerateTokenString("foo", token.TempType)
				Expect(err).To(BeNil())

				params = url.Values{
					"token": {tok},
				}
			})

			It("returns 422 invalid params", func() {
				doRequest()
				Expect(err).NotTo(BeNil())

				Expect(res.StatusCode).To(Equal(422))
				j := ParseJSONBody(res.Body)
				Expect(j["error"]).To(Equal("invalid_credentials"))
				Expect(j["errors"]).To(Equal(map[string]interface{}{
					"token": "is invalid",
				}))
			})
		})
	})
})
