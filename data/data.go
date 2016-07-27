package data

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/config"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
)

var (
	D           *Data
	DataDirPath string
)

const (
	defaultUserDataFileName = "nitrous-solo-users.dat"
	defaultIDEDataFileName  = "nitrous-solo-ide.json"
)

func init() {
	D = Reset()
}

type Data struct {
	UserData map[string]*UserData              // <Login, *UserData>
	IDEData  map[string]map[string]interface{} // <Login, JSON-data>
}

func Reset() *Data {
	return &Data{
		UserData: map[string]*UserData{},
		IDEData:  map[string]map[string]interface{}{},
	}
}

func LoadUserData() error {
	if DataDirPath == "" {
		logrus.Fatalf("path to data directory must be defined")
	}

	// load user data
	p := userDataFileName()

	f, err := os.Open(p)
	if err != nil && !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`could not load user data file "%s" - %v`, p, err))
	}

	d := map[string]*UserData{}

	if err == nil {
		defer f.Close()

		if err := gob.NewDecoder(f).Decode(&d); err != nil {
			return errors.New(fmt.Sprintf(`could not parse user data file "%s" - %v`, p, err))
		}
	}

	D.UserData = d

	return nil
}

func LoadIDEData() error {
	if DataDirPath == "" {
		logrus.Fatalf("path to data directory must be defined")
	}

	// load ide data
	p := ideDataFileName()

	b, err := ioutil.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`could not load IDE data file "%s" - %v`, p, err))
	}

	d := map[string]map[string]interface{}{}

	if err == nil {
		if err := json.Unmarshal(b, &d); err != nil {
			return errors.New(fmt.Sprintf(`could not parse IDE data file "%s" - %v`, p, err))
		}
	}

	D.IDEData = d
	return nil
}

func SaveUserData() error {
	if DataDirPath == "" {
		return errors.New("path to data directory must be defined")
	}

	if err := os.MkdirAll(DataDirPath, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, DataDirPath))
	}

	// save user data
	p := userDataFileName()

	buf := &bytes.Buffer{}

	if err := gob.NewEncoder(buf).Encode(D.UserData); err != nil {
		return errors.New("could not serialize user data!")
	}

	if err := atomicwrite.AtomicWrite(buf, p, "", 0600, true); err != nil {
		return errors.New(fmt.Sprintf(`could not save user data file "%s" - %v`, p, err))
	}

	return nil
}

func SaveIDEData() error {
	if DataDirPath == "" {
		return errors.New("path to data directory must be defined")
	}

	if err := os.MkdirAll(DataDirPath, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, DataDirPath))
	}

	// save ide data
	p := ideDataFileName()

	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(D.IDEData); err != nil {
		return errors.New("could not serialize ide data!")
	}

	if err := atomicwrite.AtomicWrite(buf, p, "", 0600, true); err != nil {
		return errors.New(fmt.Sprintf(`could not save ide data file "%s" - %v`, p, err))
	}

	return nil
}

func userDataFileName() string {
	return filepath.Join(DataDirPath, addEnvSuffix(defaultUserDataFileName))
}

func ideDataFileName() string {
	return filepath.Join(DataDirPath, addEnvSuffix(defaultIDEDataFileName))
}

func addEnvSuffix(p string) string {
	suffix := ""
	if e := config.Env; e != "production" {
		suffix = "-" + e
	}

	if suffix == "" {
		return p
	}

	ext := filepath.Ext(p)
	return p[:len(p)-len(ext)] + suffix + ext
}
