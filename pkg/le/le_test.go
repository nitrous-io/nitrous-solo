package le_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509/pkix"
	"io/ioutil"
	"os"
	"time"

	jose "gopkg.in/square/go-jose.v1"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/pkg/certgen"
	"github.com/nitrous-io/nitrous-solo/pkg/le"
	"github.com/nitrous-io/nitrous-solo/testhelper/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/xenolf/lego/acme"

	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/le")
}

var _ = Describe("le", func() {
	var (
		origLogLevel logrus.Level

		email    string
		hostname string

		dir string
		err error

		origNewACMEClient func(url string, u acme.User, keyType acme.KeyType) (le.ACMEClient, error)

		testCert []byte
		testPKey []byte
	)

	BeforeEach(func() {
		// suppress warning log message
		origLogLevel = logrus.GetLevel()
		logrus.SetLevel(logrus.FatalLevel)

		origNewACMEClient = le.NewACMEClient
		le.NewACMEClient = fake.NewACMEClient

		dir, err = ioutil.TempDir("", "cert_test_path")
		le.CertDataPath = dir

		testPKey, testCert, err = certgen.GenerateWithOptions(pkix.Name{}, hostname, time.Now(), time.Now().Add(10*24*time.Hour))
		Expect(err).To(BeNil())

		fake.ACMECertificate = testCert
		fake.ACMEPrivateKey = testPKey

		hostname = "www.foobar.com"
		email = "foo@example.com"
	})

	AfterEach(func() {
		logrus.SetLevel(origLogLevel)
		le.NewACMEClient = origNewACMEClient
		fake.FakeACMEReset()
		os.RemoveAll(dir)
	})

	createAndRegisterUser := func(email string, hostname string) {
		u, err := le.NewUser(email, hostname)
		Expect(err).To(BeNil())

		pkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(err).To(BeNil())

		reg := &acme.RegistrationResource{
			Body: acme.Registration{
				ID: 1,
				Key: jose.JsonWebKey{
					Key: &pkey.PublicKey,
				},
			},
		}
		u.Registration = reg

		err = le.SaveUser(u)
		Expect(err).To(BeNil())
	}

	Describe("GenerateCert", func() {
		It("generates certificate and private key from acme provider with email and hostname", func() {
			hostname := "www.foobar.com"
			err = le.GenerateCert("foo@example.com", hostname)
			Expect(err).To(BeNil())

			certBytes, err := ioutil.ReadFile(le.CertificatePath(hostname))
			Expect(err).To(BeNil())
			Expect(certBytes).To(Equal(testCert))

			pkeyBytes, err := ioutil.ReadFile(le.PrivateKeyPath(hostname))
			Expect(err).To(BeNil())
			Expect(pkeyBytes).To(Equal(testPKey))
		})

		Context("when a user does not exist", func() {
			It("creates a user", func() {
				u, err := le.GetUser()
				Expect(err).To(Equal(le.ErrUserNotFound))

				err = le.GenerateCert(email, hostname)
				Expect(err).To(BeNil())

				u, err = le.GetUser()
				Expect(err).To(BeNil())
				Expect(u.Email).To(Equal(email))
				Expect(u.Hostname).To(Equal(hostname))
			})

			It("registers a user to acme", func() {
				err = le.GenerateCert(email, hostname)
				Expect(err).To(BeNil())

				Expect(fake.RegisterCalls.Count()).To(Equal(1))
			})

			It("agrees on TOS for a user", func() {
				err = le.GenerateCert(email, hostname)
				Expect(err).To(BeNil())

				Expect(fake.AgreeToTOSCalls.Count()).To(Equal(1))
			})
		})

		Context("when a user already exists and has registered", func() {
			BeforeEach(func() {
				createAndRegisterUser(email, hostname)
			})

			It("does not register a user to acme", func() {
				err = le.GenerateCert(email, hostname)
				Expect(err).To(BeNil())

				Expect(fake.RegisterCalls.Count()).To(Equal(0))
			})

			It("does not agree on TOS for a user", func() {
				err = le.GenerateCert(email, hostname)
				Expect(err).To(BeNil())

				Expect(fake.AgreeToTOSCalls.Count()).To(Equal(0))
			})
		})
	})

	Describe("GetCertPathes", func() {
		Context("when a user does not exist", func() {
			It("returns an error", func() {
				certPath, keyPath, err := le.GetCertPathes()
				Expect(err).NotTo(BeNil())
				Expect(certPath).To(BeZero())
				Expect(keyPath).To(BeZero())
			})
		})

		Context("when a user exist", func() {
			BeforeEach(func() {
				createAndRegisterUser(email, hostname)
			})

			Context("when certificate and private key don't exist", func() {
				It("returns an error", func() {
					certPath, keyPath, err := le.GetCertPathes()
					Expect(err).NotTo(BeNil())
					Expect(certPath).To(BeZero())
					Expect(keyPath).To(BeZero())
				})
			})

			Context("when certificate and private key exist", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(le.CertificatePath(hostname), testCert, 0644)
					Expect(err).To(BeNil())

					err = ioutil.WriteFile(le.PrivateKeyPath(hostname), testPKey, 0644)
					Expect(err).To(BeNil())
				})

				It("returns certificate and private key path", func() {
					certPath, keyPath, err := le.GetCertPathes()
					Expect(err).To(BeNil())
					Expect(certPath).To(Equal(le.CertificatePath(hostname)))
					Expect(keyPath).To(Equal(le.PrivateKeyPath(hostname)))
				})

				Context("when cert is expires within 2 days", func() {
					BeforeEach(func() {
						var err error
						testPKey, testCert, err = certgen.GenerateWithOptions(pkix.Name{}, hostname, time.Now(), time.Now().Add(10*time.Hour))
						Expect(err).To(BeNil())

						err = ioutil.WriteFile(le.CertificatePath(hostname), testCert, 0644)
						Expect(err).To(BeNil())

						err = ioutil.WriteFile(le.PrivateKeyPath(hostname), testPKey, 0644)
						Expect(err).To(BeNil())
					})

					It("does not return certificate and private key path", func() {
						certPath, keyPath, err := le.GetCertPathes()
						Expect(err).To(BeNil())
						Expect(certPath).To(BeZero())
						Expect(keyPath).To(BeZero())
					})
				})
			})
		})
	})
})
