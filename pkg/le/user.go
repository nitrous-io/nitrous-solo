package le

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
	"github.com/xenolf/lego/acme"
)

// User represents a Let's Encrypt user account.
type User struct {
	Hostname     string
	Email        string
	Registration *acme.RegistrationResource
	key          crypto.PrivateKey
}

// UserData contains persisted items pertaining to a user.
type UserData struct {
	Reg []byte `json: reg`
	Key []byte `json: key`
}

var (
	ErrUserNotFound = errors.New("user not found")
)

// GetEmail gets u's email.
func (u User) GetEmail() string {
	return u.Email
}

// GetPrivateKey gets u's private key.
func (u User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// GetRegistration gets u's registration resource.
func (u User) GetRegistration() *acme.RegistrationResource {
	return u.Registration
}

// newUser creates a new User for the given email address with a new private key.
func NewUser(email string, hostname string) (*User, error) {
	user := User{Email: email, Hostname: hostname}
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return &user, err
	}
	user.key = privateKey
	return &user, nil
}

// getUser loads the user with the given email from disk.
func GetUser() (*User, error) {
	var user User
	b, err := ioutil.ReadFile(userDataPath())
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return nil, ErrUserNotFound
	}

	var userData UserData
	if err := json.Unmarshal(b, &userData); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(userData.Reg, &user); err != nil {
		return nil, err
	}

	user.key, err = loadPrivateKey(userData.Key)
	return &user, err
}

// saveUser persists a user's key and account registration to the file system.
func SaveUser(user *User) error {
	if err := os.MkdirAll(userDataDirPath(), 0700); err != nil {
		return err
	}

	// Save the private key and registration
	userData := new(UserData)
	var err error
	userData.Key, err = savePrivateKey(user.key)
	if err != nil {
		return err
	}

	userData.Reg, err = json.MarshalIndent(&user, "", "\t")
	if err != nil {
		return err
	}

	b, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if err := atomicwrite.AtomicWrite(buf, userDataPath(), "", 0600, true); err != nil {
		return err
	}

	return nil
}

func userDataDirPath() string {
	return CertDataPath
}

func userDataPath() string {
	return filepath.Join(userDataDirPath(), "user_data.json")
}
