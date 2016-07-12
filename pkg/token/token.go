package token

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/nitrous-io/nitrous-solo/pkg/msgverifier"
)

var Expiry = 72 * time.Hour

type token struct {
	Nonce     string    `json:"n"` // to ensure uniqueness
	Login     string    `json:"l"`
	ExpiresAt time.Time `json:"e"`
}

func GenerateTokenString(login string) (string, error) {
	nonce, err := generateNonce(48)
	if err != nil {
		return "", err
	}
	return msgverifier.DefaultMsgVerifier.Generate(token{
		Nonce:     nonce,
		Login:     login,
		ExpiresAt: time.Now().Add(Expiry),
	})
}

// returns login (username) if token is valid
func ValidateTokenString(str string) (login string) {
	var v token
	if err := msgverifier.DefaultMsgVerifier.Verify(str, &v); err != nil {
		return ""
	}
	if v.ExpiresAt.Before(time.Now()) {
		return ""
	}
	return v.Login
}

func generateNonce(n uint) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
