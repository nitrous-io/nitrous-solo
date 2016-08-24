package main

import (
	"crypto/x509/pkix"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/osext"
	"github.com/nitrous-io/nitrous-solo/config"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/certgen"
	"github.com/nitrous-io/nitrous-solo/pkg/le"
	"github.com/nitrous-io/nitrous-solo/server"
)

func main() {
	if config.Env == "production" {
		gin.SetMode("release")
	}

	appDir, err := osext.ExecutableFolder()
	if err != nil {
		logrus.Fatal(err)
	}

	dataDirPathFlag := flag.String("d", "", "path to data directory")
	bindAddrFlag := flag.String("b", "0.0.0.0:6877", "bind address and port")
	sslKeyPathFlag := flag.String("key", "", "path to SSL private key")
	sslCertPathFlag := flag.String("cert", "", "path to SSL certificate")
	noSSLFlag := flag.Bool("nossl", false, "disable HTTPS")
	helpFlag := flag.Bool("h", false, "show help")

	flag.Parse()

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	dataDirPath := *dataDirPathFlag
	bindAddr := *bindAddrFlag
	noSSL := *noSSLFlag
	sslKeyPath := *sslKeyPathFlag
	sslCertPath := *sslCertPathFlag

	if dataDirPath == "" {
		u, err := user.Current()
		if err != nil {
			logrus.Fatal("could not determine current user")
		}
		dataDirPath = filepath.Join(u.HomeDir, ".nitrous-solo")
	}

	data.DataDirPath = dataDirPath
	le.CertDataPath = filepath.Join(dataDirPath, "global")

	if err := data.LoadUserData(); err != nil {
		logrus.Fatal(err)
	}
	if err := data.LoadIDESettings(); err != nil {
		logrus.Fatal(err)
	}

	r := server.New(appDir)
	if noSSL {
		log.Fatal(r.Run(bindAddr))
	} else {
		if sslKeyPath != "" && sslCertPath != "" {
			log.Fatal(r.RunTLS(bindAddr, sslCertPath, sslKeyPath))
		}

		if sslKeyPath == "" && sslCertPath != "" {
			logrus.Fatal("path to SSL private key must also be specified (--key)")
		}

		if sslKeyPath != "" && sslCertPath == "" {
			logrus.Fatal("path to SSL certificate must also be specified (--cert)")
		}

		// Use Let's Encrypt cert if it exists
		sslCertPath, sslKeyPath, err := le.GetCertPathes()
		if err != nil {
			if err == le.ErrCertExpired {
				if err := le.RenewCert(); err == nil {
					var secondErr error
					sslCertPath, sslKeyPath, secondErr = le.GetCertPathes()
					if secondErr != nil {
						logrus.Warnf("Failed to read letsencrpyt cert after renew: %v", err)
					}
				} else {
					logrus.Warnf("Failed to renew letsencrypt cert: %v", err)
				}
			} else {
				logrus.Warnf("Failed to read letsencrpyt cert file: %v", err)
			}
		}

		// When a user hasn't generated cert with lets-encrypt,
		// it generates self-signed cert
		if sslKeyPath == "" && sslCertPath == "" {
			logrus.Warn("SSL certificate and private key aren't specified; falling back to a self-signed certificate.")

			sslKeyPath = filepath.Join(dataDirPath, "selfsigned.key")
			sslCertPath = filepath.Join(dataDirPath, "selfsigned.crt")

			_, err := os.Stat(sslKeyPath)
			shouldGen := os.IsNotExist(err)
			_, err = os.Stat(sslCertPath)
			shouldGen = shouldGen || os.IsNotExist(err)

			if shouldGen {
				logrus.Info("Generating a self-signed certificate...")
				keyData, certData, err := certgen.Generate(pkix.Name{
					Country:      []string{"US"},
					Locality:     []string{"San Francisco"},
					Province:     []string{"CA"},
					Organization: []string{"Nitrous Solo Self-signed Certificate", "This certificate is provided without warranty!"},
				})
				if err != nil {
					logrus.Fatal(err)
				}

				if err := ioutil.WriteFile(sslKeyPath, keyData, 0600); err != nil {
					logrus.Fatalf("could not write private key file: %v", err)
				}

				if err := ioutil.WriteFile(sslCertPath, certData, 0600); err != nil {
					logrus.Fatalf("could not write certificate file: %v", err)
				}
			}
		}

		log.Fatal(r.RunTLS(bindAddr, sslCertPath, sslKeyPath))
	}
}
