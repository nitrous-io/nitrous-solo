package domain

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

func Sanitize(name string) (string, error) {
	name = strings.TrimSpace(name)
	apexDomain, err := publicsuffix.EffectiveTLDPlusOne(name)
	if err != nil {
		return "", err
	}

	if name == apexDomain {
		name = "www." + name
	}

	return name, nil
}
