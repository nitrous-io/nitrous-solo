package le

import (
	"crypto"

	"github.com/xenolf/lego/acme"
)

var NewACMEClient func(url string, u acme.User, keyType acme.KeyType) (ACMEClient, error)

type ACMEClient interface {
	ObtainCertificate(domains []string, bundle bool, privKey crypto.PrivateKey) (acme.CertificateResource, map[string]error)
	Register() (*acme.RegistrationResource, error)
	AgreeToTOS() error
}
