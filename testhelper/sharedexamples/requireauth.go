package sharedexamples

import (
	"net/http"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	. "github.com/nitrous-io/nitrous-solo/testhelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ItRequiresAuthentication(varFn func() (ud *data.UserData, headers *http.Header), reqFn func() *http.Response) {
	var (
		ud      *data.UserData
		headers *http.Header
	)

	BeforeEach(func() {
		ud, headers = varFn()
	})

	Context("with an invalid access token", func() {
		It("returns 401 Unauthorized", func() {
			tok, err := token.GenerateTokenString(ud.Login+"barqux", token.LoginType)
			Expect(err).To(BeNil())
			headers.Set("Authorization", "Token "+tok)

			res := reqFn()
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))

			j := ParseJSONBody(res.Body)
			Expect(j["error"]).To(Equal("invalid_token"))
			Expect(j["error_description"]).To(Equal("access token is invalid"))
		})
	})

	Context("with wrong type of token", func() {
		It("returns 401 Unauthorized", func() {
			tok, err := token.GenerateTokenString(ud.Login, token.TempType)
			Expect(err).To(BeNil())
			headers.Set("Authorization", "Token "+tok)

			res := reqFn()
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))

			j := ParseJSONBody(res.Body)
			Expect(j["error"]).To(Equal("invalid_token"))
			Expect(j["error_description"]).To(Equal("access token is invalid"))
		})
	})

	Context("with no access token", func() {
		It("returns 401 Unauthorized", func() {
			headers.Del("Authorization")
			res := reqFn()
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))

			j := ParseJSONBody(res.Body)
			Expect(j["error"]).To(Equal("invalid_token"))
			Expect(j["error_description"]).To(Equal("access token is required"))
		})
	})
}
