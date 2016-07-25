package token

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/nitrous-io/nitrous-solo/pkg/msgverifier"
)

var DefaultExpiry = 72 * time.Hour

type Type int

const (
	LoginType Type = iota
	SocketType
	TempType
)

type token struct {
	Nonce     string    `json:"n"` // to ensure uniqueness
	Login     string    `json:"l"`
	TokenType Type      `json:"t"`
	ExpiresAt time.Time `json:"e"`
}

func GenerateTokenString(login string, tokenType Type) (string, error) {
	return GenerateTokenStringWithExpiry(login, DefaultExpiry, tokenType)
}

func GenerateTokenStringWithExpiry(login string, expiry time.Duration, tokenType Type) (string, error) {
	nonce, err := generateNonce(48)
	if err != nil {
		return "", err
	}
	return msgverifier.DefaultMsgVerifier.Generate(token{
		Nonce:     nonce,
		Login:     login,
		TokenType: tokenType,
		ExpiresAt: time.Now().Add(expiry),
	})
}

// returns login (username) if token is valid
func ValidateTokenString(str string, tokenType Type) (login string) {
	var v token
	if err := msgverifier.DefaultMsgVerifier.Verify(str, &v); err != nil {
		return ""
	}

	if v.TokenType != tokenType {
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
