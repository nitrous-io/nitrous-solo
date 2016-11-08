package config

import "os"

var (
	Env     = "development"
	AcmeURL = "https://acme-staging.api.letsencrypt.org/directory"
)

func init() {
	if e := os.Getenv("NITROUS_ENV"); e != "" {
		Env = e
	}

	if acmeURL := os.Getenv("ACME_URL"); acmeURL != "" {
		AcmeURL = acmeURL
	}
}
