package sharedexamples

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ItRequiresTempToken(varFn func() (ud *data.UserData, params *url.Values), reqFn func() *http.Response) {
	var (
		ud     *data.UserData
		params *url.Values
	)

	BeforeEach(func() {
		ud, params = varFn()
	})

	Context("when token is missing", func() {
		It("returns 422 with invalid_params", func() {
			params.Del("token")
			res := reqFn()

			b := &bytes.Buffer{}
			_, err := b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(res.StatusCode).To(Equal(422))
			Expect(b.String()).To(MatchJSON(`{
				"error": "invalid_params",
				"errors": {
					"token": "is required"
				}
			}`))
		})
	})

	Context("when token is wrong", func() {
		It("returns 422 with invalid_token", func() {
			tok, err := token.GenerateTokenString(ud.Login+"1", token.TempType)
			Expect(err).To(BeNil())

			params.Set("token", tok)
			res := reqFn()

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(res.StatusCode).To(Equal(422))
			Expect(b.String()).To(MatchJSON(`{
				"error": "invalid_credentials",
				"errors": {
					"token": "is invalid"
				}
			}`))
		})
	})

	Context("when token is not temp type", func() {
		It("returns 422 with invalid_token", func() {
			tok, err := token.GenerateTokenString(ud.Login, token.LoginType)
			Expect(err).To(BeNil())

			params.Set("token", tok)
			res := reqFn()

			b := &bytes.Buffer{}
			_, err = b.ReadFrom(res.Body)
			Expect(err).To(BeNil())

			Expect(res.StatusCode).To(Equal(422))
			Expect(b.String()).To(MatchJSON(`{
				"error": "invalid_credentials",
				"errors": {
					"token": "is invalid"
				}
			}`))
		})
	})
}
