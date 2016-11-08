package le

import (
	"crypto"

	"github.com/xenolf/lego/acme"
)

type LeACMEClient struct {
	client *acme.Client
}

func NewLeACMEClient(url string, u acme.User, keyType acme.KeyType) (ACMEClient, error) {
	c, err := acme.NewClient(url, u, keyType)
	if err != nil {
		return nil, err
	}

	return &LeACMEClient{c}, nil
}

func (c *LeACMEClient) ObtainCertificate(domains []string, bundle bool, privKey crypto.PrivateKey) (acme.CertificateResource, map[string]error) {
	return c.client.ObtainCertificate(domains, bundle, privKey)
}

func (c *LeACMEClient) RenewCertificate(cr acme.CertificateResource, bundle bool) (acme.CertificateResource, error) {
	return c.client.RenewCertificate(cr, bundle)
}

func (c *LeACMEClient) Register() (*acme.RegistrationResource, error) {
	return c.client.Register()
}

func (c *LeACMEClient) AgreeToTOS() error {
	return c.client.AgreeToTOS()
}
