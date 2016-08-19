package le

import (
	"bytes"
	"errors"
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
	CertDataPath = ""
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
		return "", "", nil
	}

	return certPath, keyPath, nil
}

func GenerateCert(email, hostname string) error {
	logrus.Infof(`Obtaining certificates from lets encrypt with email "%s" for host "%s"`, email, hostname)
	if err := os.MkdirAll(certDataDirPath(), 0700); err != nil {
		return err
	}

	certBytes, pkeyBytes, err := obtainCert(email, hostname)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(certBytes)
	if err := atomicwrite.AtomicWrite(buf, CertificatePath(hostname), "", 0644, true); err != nil {
		return err
	}

	buf = bytes.NewBuffer(pkeyBytes)
	if err := atomicwrite.AtomicWrite(buf, PrivateKeyPath(hostname), "", 0600, true); err != nil {
		return err
	}

	return nil
}

func obtainCert(email, hostname string) (certBytes, pkeyBytes []byte, err error) {
	// Look up or create the LE user account
	u, err := GetUser()
	if err != nil {
		if err == ErrUserNotFound {
			// Create new account when current account does not exist
			u, err = NewUser(email, hostname)
			if err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}

	// The client facilitates our communication with the CA server.
	client, err := NewACMEClient(config.AcmeURL, u, KeyType)
	if err != nil {
		return nil, nil, err
	}

	// If not registered, the user must register an account with the CA
	// and agree to terms
	if u.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			return nil, nil, err
		}
		u.Registration = reg

		if err := SaveUser(u); err != nil {
			return nil, nil, err
		}

		logrus.Info("\nYour sites will be served over HTTPS automatically using Let's Encrypt.")
		logrus.Info("By continuing, you agree to the Let's Encrypt Subscriber Agreement at:")
		logrus.Info("  https://acme-v01.api.letsencrypt.org/terms")

		// Agree to TOS
		if err := client.AgreeToTOS(); err != nil {
			return nil, nil, err
		}
	}

	// The acme library takes care of completing the challenges to obtain the certificate(s).
	// The domains must resolve to this machine or you have to use the DNS challenge.
	certificate, failures := client.ObtainCertificate([]string{hostname}, true, nil)
	if len(failures) > 0 {
		errorMessage := ""
		for r, err := range failures {
			errorMessage += r + " : " + err.Error() + "\n"
		}
		return nil, nil, errors.New(errorMessage)
	}

	return certificate.Certificate, certificate.PrivateKey, nil
}

func certDataDirPath() string {
	return CertDataPath
}

func certDataPath() string {
	return filepath.Join(certDataDirPath(), "cert_data.json")
}

func CertificatePath(hostname string) string {
	return filepath.Join(certDataDirPath(), hostname+".crt")
}

func PrivateKeyPath(hostname string) string {
	return filepath.Join(certDataDirPath(), hostname+".key")
}
