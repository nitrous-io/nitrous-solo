package msgverifier

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Sirupsen/logrus"
)

var DefaultMsgVerifier *MsgVerifier

func init() {
	key := make([]byte, 256)
	if _, err := rand.Read(key); err != nil {
		logrus.Fatal("could not generate secret key")
	}

	DefaultMsgVerifier = New(key)
}

var ErrInvalidMessage = errors.New("msgverifier: Invalid Message")

type MsgVerifier struct {
	secret []byte
}

func New(secret []byte) *MsgVerifier {
	return &MsgVerifier{secret: secret}
}

func (m *MsgVerifier) Verify(signed string, v interface{}) error {
	if signed == "" || len(signed) < 4 {
		return ErrInvalidMessage
	}
	spl := strings.SplitN(signed, "--", 2)
	if len(spl) < 2 {
		return ErrInvalidMessage
	}
	encodedJSON := spl[0]
	signature := spl[1]
	validSig := m.makeSignature(encodedJSON)
	if validSig != signature {
		return ErrInvalidMessage
	}
	j, err := base64.StdEncoding.DecodeString(encodedJSON)
	if err != nil {
		return err
	}
	err = json.Unmarshal(j, v)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgVerifier) Generate(obj interface{}) (string, error) {
	json, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	encodedJSON := base64.StdEncoding.EncodeToString(json)
	signature := m.makeSignature(encodedJSON)
	return fmt.Sprintf("%s--%s", encodedJSON, signature), nil
}

func (m *MsgVerifier) makeSignature(input string) string {
	mac := hmac.New(sha1.New, m.secret)
	mac.Write([]byte(input))
	return hex.EncodeToString(mac.Sum(nil))
}
