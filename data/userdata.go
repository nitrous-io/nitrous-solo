package data

import (
	"os/user"

	"golang.org/x/crypto/bcrypt"
)

type UserData struct {
	Login             string // ide login username
	EncryptedPassword []byte // ide login password
	User              string // username on system
}

var HomePath = homePath

func (d *UserData) SetPassword(pw string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), 0)
	if err != nil {
		return err
	}
	d.EncryptedPassword = b
	return nil
}

func (d *UserData) VerifyPassword(pw string) bool {
	err := bcrypt.CompareHashAndPassword(d.EncryptedPassword, []byte(pw))
	return err == nil
}

func homePath(d *UserData) (homeDir string, err error) {
	u, err := user.Lookup(d.User)
	if err != nil {
		return "", err
	}

	return u.HomeDir, nil
}
