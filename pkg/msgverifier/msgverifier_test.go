package msgverifier_test

import (
	"github.com/nitrous-io/nitrous-solo/pkg/msgverifier"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "msgverifier")
}

var _ = Describe("msgverifier", func() {
	var verifier *msgverifier.MsgVerifier
	var message string

	BeforeEach(func() {
		verifier = msgverifier.New([]byte("foobar"))
		message = "WyJmb28iLCJiYXIiXQ==--0898cb188ac378e6ba570f7fe2a946d2a51a4b13" // '["foo","bar"]' in base64
	})

	Describe("Verify", func() {
		Context("valid signed message", func() {
			It("returns the original object", func() {
				var o []string
				err := verifier.Verify(message, &o)
				Expect(err).To(BeNil())
				Expect(o).To(Equal([]string{"foo", "bar"}))
			})
		})

		Context("invalid signed message", func() {
			BeforeEach(func() {
				message = "WyJmb28iLCJiYXIiXQ==--236i789834569723950781239048710293471024"
			})

			It("returns null", func() {
				var o []string
				err := verifier.Verify(message, &o)
				Expect(err).To(Equal(msgverifier.ErrInvalidMessage))
				Expect(o).To(BeNil())
			})
		})

		Context("empty message", func() {
			It("returns null", func() {
				var o []string
				err := verifier.Verify("", &o)
				Expect(err).To(Equal(msgverifier.ErrInvalidMessage))
				Expect(o).To(BeNil())
			})
		})
	})

	Describe("Generate", func() {
		It("generates signed message", func() {
			msg, _ := verifier.Generate([]string{"foo", "bar"})
			Expect(msg).To(Equal(message))
		})
	})
})
