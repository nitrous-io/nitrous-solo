package le

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
)

// VerifyExistingCert verifies existing cert with given hostname
func VerifyExistingCert(hostname string) (bool, error) {
	certPath := CertificatePath(hostname)
	keyPath := PrivateKeyPath(hostname)

	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return false, err
	} else {
		x509Cert, err := x509.ParseCertificate(certificate.Certificate[0])
		if err != nil {
			return false, err
		} else if x509Cert.NotAfter.Add(-2 * 24 * time.Hour).Before(time.Now()) {
			// If expiry date is less than 2 days left
			return false, nil
		}
	}

	return true, nil
}

// loadPrivateKey loads a PEM-encoded ECC/RSA private key from an array of bytes.
func loadPrivateKey(keyBytes []byte) (crypto.PrivateKey, error) {
	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private key type")
}

// SavePrivateKey saves a PEM-encoded ECC/RSA private key to an array of bytes.
func savePrivateKey(key crypto.PrivateKey) ([]byte, error) {
	var pemType string
	var keyBytes []byte
	switch key := key.(type) {
	case *ecdsa.PrivateKey:
		var err error
		pemType = "EC"
		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
	case *rsa.PrivateKey:
		pemType = "RSA"
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
	}

	pemKey := pem.Block{Type: pemType + " PRIVATE KEY", Bytes: keyBytes}
	return pem.EncodeToMemory(&pemKey), nil
}
