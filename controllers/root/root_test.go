package root_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nitrous-io/nitrous-solo/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "root")
}

var _ = Describe("root", func() {
	var (
		s   *httptest.Server
		res *http.Response
		err error
	)

	BeforeEach(func() {
		s = httptest.NewServer(server.New())
	})

	AfterEach(func() {
		if res != nil {
			res.Body.Close()
		}
		s.Close()
	})

	It("returns 200 OK and the rendered index template", func() {
		res, err = http.Get(s.URL + "/")
		Expect(err).To(BeNil())
		Expect(res.StatusCode).To(Equal(http.StatusOK))
	})
})
