package data_test

import (
	"os/user"

	"github.com/nitrous-io/nitrous-solo/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var currentSystemUser *user.User

func init() {
	currentSystemUser, _ = user.Current()
}

var _ = Describe("userdata", func() {
	Describe("password", func() {
		It("sets and verifies password", func() {
			ud := &data.UserData{
				Login: "foo",
				User:  "bar",
			}

			err := ud.SetPassword("bazqux")
			Expect(err).To(BeNil())

			Expect(ud.EncryptedPassword).NotTo(BeNil())

			Expect(ud.VerifyPassword("bazqux")).To(BeTrue())
			Expect(ud.VerifyPassword("wrong password")).To(BeFalse())
		})
	})

	Describe("home path", func() {
		It("returns current user home path", func() {
			ud := &data.UserData{
				Login: "foo",
				User:  currentSystemUser.Username,
			}

			homePath, err := data.HomePath(ud)
			Expect(err).To(BeNil())
			Expect(homePath).To(Equal(currentSystemUser.HomeDir))
		})
	})
})
