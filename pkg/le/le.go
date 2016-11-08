package le

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/config"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
	"github.com/xenolf/lego/acme"
)

const (
	KeyType = acme.RSA2048
)

var (
	CertDataPath   = ""
	ErrCertExpired = errors.New("cert is expired or expirying soon")
)

func init() {
	NewACMEClient = NewLeACMEClient
}

func GetCertPathes() (certPath string, keyPath string, err error) {
	u, err := GetUser()
	if err != nil {
		return "", "", err
	}

	certPath = CertificatePath(u.Hostname)
	keyPath = PrivateKeyPath(u.Hostname)

	valid, err := VerifyExistingCert(u.Hostname)
	if err != nil {
		return "", "", err
	}

	if !valid {
		return "", "", ErrCertExpired
	}

	return certPath, keyPath, nil
}

func GenerateCert(email, hostname string) error {
	logrus.Infof(`Obtaining certificates from lets encrypt with email "%s" for host "%s"`, email, hostname)
	if err := os.MkdirAll(certDataDirPath(), 0700); err != nil {
		return err
	}

	cr, err := obtainCert(email, hostname)
	if err != nil {
		return err
	}

	if err := writeCertificateToDisk(hostname, cr); err != nil {
		return nil
	}

	return nil
}

func createOrFindClient(email, hostname string) (client ACMEClient, err error) {
	// Look up or create the LE user account
	u, err := GetUser()
	if err != nil {
		if err == ErrUserNotFound {
			// Create new account when current account does not exist
			u, err = NewUser(email, hostname)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// The client facilitates our communication with the CA server.
	client, err = NewACMEClient(config.AcmeURL, u, KeyType)
	if err != nil {
		return nil, err
	}

	// If not registered, the user must register an account with the CA
	// and agree to terms
	if u.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			return nil, err
		}
		u.Registration = reg

		if err := SaveUser(u); err != nil {
			return nil, err
		}

		logrus.Info("\nYour sites will be served over HTTPS automatically using Let's Encrypt.")
		logrus.Info("By continuing, you agree to the Let's Encrypt Subscriber Agreement at:")
		logrus.Info("  https://acme-v01.api.letsencrypt.org/terms")

		// Agree to TOS
		if err := client.AgreeToTOS(); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func obtainCert(email, hostname string) (*acme.CertificateResource, error) {
	client, err := createOrFindClient(email, hostname)
	if err != nil {
		return nil, err
	}

	// The acme library takes care of completing the challenges to obtain the certificate(s).
	// The domains must resolve to this machine or you have to use the DNS challenge.
	cr, failures := client.ObtainCertificate([]string{hostname}, true, nil)
	if len(failures) > 0 {
		errorMessage := ""
		for r, err := range failures {
			errorMessage += r + " : " + err.Error() + "\n"
		}
		return nil, errors.New(errorMessage)
	}

	return &cr, nil
}

func writeCertificateToDisk(hostname string, cr *acme.CertificateResource) error {
	buf := bytes.NewBuffer(cr.Certificate)
	if err := atomicwrite.AtomicWrite(buf, CertificatePath(hostname), "", 0644, true); err != nil {
		return err
	}

	buf = bytes.NewBuffer(cr.PrivateKey)
	if err := atomicwrite.AtomicWrite(buf, PrivateKeyPath(hostname), "", 0600, true); err != nil {
		return err
	}

	b, err := json.Marshal(cr)
	if err != nil {
		return err
	}

	buf = bytes.NewBuffer(b)
	if err := atomicwrite.AtomicWrite(buf, CertificateMetaPath(hostname), "", 0644, true); err != nil {
		return err
	}

	return nil
}

func readCertificateMetaFromDisk(hostname string) (*acme.CertificateResource, error) {
	b, err := ioutil.ReadFile(CertificateMetaPath(hostname))
	if err != nil {
		return nil, err
	}

	var cr acme.CertificateResource
	if err := json.Unmarshal(b, &cr); err != nil {
		return nil, err
	}

	b, err = ioutil.ReadFile(CertificatePath(hostname))
	if err != nil {
		return nil, err
	}
	cr.Certificate = b

	b, err = ioutil.ReadFile(PrivateKeyPath(hostname))
	if err != nil {
		return nil, err
	}
	cr.PrivateKey = b

	return &cr, nil
}

func RenewCert() error {
	u, err := GetUser()
	if err != nil {
		return err
	}

	client, err := createOrFindClient(u.Email, u.Hostname)
	if err != nil {
		return err
	}

	currentCR, err := readCertificateMetaFromDisk(u.Hostname)
	if err != nil {
		return err
	}

	cr, err := client.RenewCertificate(*currentCR, true)
	if err != nil {
		if _, ok := err.(acme.TOSError); ok {
			err := client.AgreeToTOS()
			if err != nil {
				return err
			}
			cr, err = client.RenewCertificate(*currentCR, true)
			if err != nil {
				return err
			}
		}
		return err
	}

	if err := writeCertificateToDisk(u.Hostname, &cr); err != nil {
		return err
	}

	return nil
}

func certDataDirPath() string {
	return CertDataPath
}

func CertificatePath(hostname string) string {
	return filepath.Join(certDataDirPath(), hostname+".crt")
}

func PrivateKeyPath(hostname string) string {
	return filepath.Join(certDataDirPath(), hostname+".key")
}

func CertificateMetaPath(hostname string) string {
	return filepath.Join(certDataDirPath(), hostname+".json")
}
