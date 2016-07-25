package token_test

import (
	"testing"
	"time"

	"github.com/nitrous-io/nitrous-solo/pkg/msgverifier"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "token")
}

var _ = Describe("token", func() {
	Describe("GenerateTokenString", func() {
		It("generates unique tokens", func() {
			t1, err := token.GenerateTokenString("foo")
			Expect(err).To(BeNil())
			t2, err := token.GenerateTokenString("foo")
			Expect(err).To(BeNil())
			t3, err := token.GenerateTokenString("foo")
			Expect(err).To(BeNil())
			Expect(t1).NotTo(Equal(t2))
			Expect(t1).NotTo(Equal(t3))
			Expect(t2).NotTo(Equal(t3))

			t4, err := token.GenerateTokenString("bar")
			Expect(err).To(BeNil())
			Expect(t1).NotTo(Equal(t4))
		})
	})

	Describe("GenerateTokenStringWithExpiry", func() {
		It("generates unique tokens with given expiry", func() {
			t1, err := token.GenerateTokenStringWithExpiry("foo", 1*time.Second)
			Expect(err).To(BeNil())
			t2, err := token.GenerateTokenStringWithExpiry("foo", 1*time.Second)
			Expect(err).To(BeNil())
			Expect(t1).NotTo(Equal(t2))
		})
	})

	Describe("ValidateTokenString", func() {
		Context("when a given token is valid", func() {
			It("returns login (username)", func() {
				t1, err := token.GenerateTokenString("foo")
				Expect(err).To(BeNil())
				t2, err := token.GenerateTokenString("bar")
				Expect(err).To(BeNil())

				Expect(token.ValidateTokenString(t1)).To(Equal("foo"))
				Expect(token.ValidateTokenString(t2)).To(Equal("bar"))
			})
		})

		Context("when a given token is valid, but expired", func() {
			It("returns an empty string", func() {
				t, err := token.GenerateTokenStringWithExpiry("foo", 1*time.Millisecond)
				Expect(err).To(BeNil())
				time.Sleep(2 * time.Millisecond)
				Expect(token.ValidateTokenString(t)).To(BeEmpty())
			})
		})

		Context("when a given token is valid, but signed with a different key", func() {
			var origMsgVerifier *msgverifier.MsgVerifier

			BeforeEach(func() {
				origMsgVerifier = msgverifier.DefaultMsgVerifier
				msgverifier.DefaultMsgVerifier = msgverifier.New([]byte("f00b4rb4zqUx"))
			})

			AfterEach(func() {
				msgverifier.DefaultMsgVerifier = origMsgVerifier
			})

			It("returns an empty string", func() {
				t, err := token.GenerateTokenString("foo")
				Expect(err).To(BeNil())
				msgverifier.DefaultMsgVerifier = origMsgVerifier
				Expect(token.ValidateTokenString(t)).To(BeEmpty())
			})
		})
	})
})
