package fake

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	jose "gopkg.in/square/go-jose.v1"

	"github.com/nitrous-io/nitrous-solo/pkg/le"
	. "github.com/onsi/gomega"
	"github.com/xenolf/lego/acme"
)

var (
	ACMECertificate []byte
	ACMEPrivateKey  []byte

	RegisterCalls   Calls
	AgreeToTOSCalls Calls
)

func FakeACMEReset() {
	RegisterCalls = Calls{}
	AgreeToTOSCalls = Calls{}
}

type ACMEClient struct {
	URL     string
	User    acme.User
	keyType acme.KeyType
}

func NewACMEClient(url string, u acme.User, keyType acme.KeyType) (le.ACMEClient, error) {
	return &ACMEClient{url, u, keyType}, nil
}

func (c *ACMEClient) ObtainCertificate(domains []string, bundle bool, privKey crypto.PrivateKey) (acme.CertificateResource, map[string]error) {
	return acme.CertificateResource{
		Domain:      domains[0],
		Certificate: ACMECertificate,
		PrivateKey:  ACMEPrivateKey,
	}, nil
}

func (c *ACMEClient) Register() (reg *acme.RegistrationResource, err error) {
	pkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).To(BeNil())

	reg = &acme.RegistrationResource{
		Body: acme.Registration{
			ID: 1,
			Key: jose.JsonWebKey{
				Key: &pkey.PublicKey,
			},
		},
	}

	RegisterCalls.Add(
		List{},
		List{reg, err},
		nil,
	)

	return reg, nil
}

func (c *ACMEClient) AgreeToTOS() (err error) {
	AgreeToTOSCalls.Add(
		List{},
		List{err},
		nil,
	)

	return nil
}
