package data_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"testing"

	"github.com/nitrous-io/nitrous-solo/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data")
}

var currentSystemUser *user.User

func init() {
	currentSystemUser, _ = user.Current()
}

var _ = Describe("data", func() {
	var (
		origDataDirPath string
	)

	BeforeEach(func() {
		var err error
		origDataDirPath = data.DataDirPath
		data.DataDirPath, err = ioutil.TempDir("", "nitrous-test")
		Expect(err).To(BeNil())

		data.D = data.Reset()
		data.CurrentUser = currentSystemUser.Username
	})

	AfterEach(func() {
		data.DataDirPath = origDataDirPath
		data.CurrentUser = ""
		os.Remove(data.DataDirPath)
	})

	Describe("user data", func() {
		It("saves user data to a file and load from disk", func() {
			data.D.UserData = map[string]*data.UserData{
				"foobar": &data.UserData{
					Login:             "foobar",
					User:              "foobar-user",
					EncryptedPassword: []byte("wh@73v3r"),
				},
			}

			err := data.SaveUserData()
			Expect(err).To(BeNil())

			data.D = data.Reset()
			Expect(data.D.UserData).To(BeEmpty())

			err = data.LoadUserData()
			Expect(err).To(BeNil())

			Expect(data.D.UserData).NotTo(BeEmpty())
			Expect(data.D.UserData).To(HaveKey("foobar"))
			Expect(data.D.UserData["foobar"].Login).To(Equal("foobar"))
			Expect(data.D.UserData["foobar"].User).To(Equal("foobar-user"))
			Expect(data.D.UserData["foobar"].EncryptedPassword).To(Equal([]byte("wh@73v3r")))
		})

		It("loads empty user data if the file does not exist", func() {
			err := data.LoadUserData()
			Expect(err).To(BeNil())

			Expect(data.D.UserData).To(BeEmpty())
		})
	})

	Describe("ide settings", func() {
		It("saves ide settings to a file and load from disk", func() {
			data.D.IDESettings = map[string]data.IDESettings{
				"foobar": data.IDESettings{
					"font-size":   float64(10),
					"key-binding": "vim",
				},
			}

			err := data.SaveIDESettings()
			Expect(err).To(BeNil())

			data.D = data.Reset()
			Expect(data.D.IDESettings).To(BeEmpty())

			err = data.LoadIDESettings()
			Expect(err).To(BeNil())

			Expect(data.D.IDESettings).NotTo(BeEmpty())
			Expect(data.D.IDESettings).To(HaveKey("foobar"))
			Expect(data.D.IDESettings["foobar"]["font-size"]).To(Equal(float64(10)))
			Expect(data.D.IDESettings["foobar"]["key-binding"]).To(Equal("vim"))
		})

		It("loads empty ide settings if the file does not exist", func() {
			err := data.LoadIDESettings()
			Expect(err).To(BeNil())

			Expect(data.D.IDESettings).To(BeEmpty())
		})
	})
})
