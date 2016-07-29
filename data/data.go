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
	"github.com/go-fsnotify/fsnotify"
	"github.com/nitrous-io/nitrous-solo/config"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
)

var (
	D           *Data
	DataDirPath string
)

const (
	defaultUserDataFileName    = "nitrous-solo-users.dat"
	defaultIDESettingsFileName = "nitrous-solo-ide.json"
)

func init() {
	D = Reset()
}

type IDESettings map[string]interface{}

type Data struct {
	UserData    map[string]*UserData   // <Login, *UserData>
	IDESettings map[string]IDESettings // <Login, IDESettings (JSON)>
}

func Reset() *Data {
	return &Data{
		UserData:    map[string]*UserData{},
		IDESettings: map[string]IDESettings{},
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

func LoadIDESettings() error {
	if DataDirPath == "" {
		logrus.Fatalf("path to data directory must be defined")
	}

	// load ide data
	p := ideSettingsFileName()

	b, err := ioutil.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf(`could not load IDE data file "%s" - %v`, p, err))
	}

	d := map[string]IDESettings{}

	if err == nil {
		if err := json.Unmarshal(b, &d); err != nil {
			return errors.New(fmt.Sprintf(`could not parse IDE data file "%s" - %v`, p, err))
		}
	}

	D.IDESettings = d
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

func SaveIDESettings() error {
	if DataDirPath == "" {
		return errors.New("path to data directory must be defined")
	}

	if err := os.MkdirAll(DataDirPath, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, DataDirPath))
	}

	// save ide data
	p := ideSettingsFileName()

	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(D.IDESettings); err != nil {
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

func ideSettingsFileName() string {
	return filepath.Join(DataDirPath, addEnvSuffix(defaultIDESettingsFileName))
}

func ReloadUserDataUponUpdate() error {
	watcher, err := fsnotify.NewWatcher()
	if err = watcher.Add(userDataFileName()); err != nil {
		return err
	}
	defer watcher.Close()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Remove > 0 || event.Op&fsnotify.Rename > 0 {
				return errors.New("user data file is renamed or removed")
			}

			if event.Op&fsnotify.Write > 0 || event.Op&fsnotify.Create > 0 {
				if err := LoadIDESettings(); err != nil {
					return err
				}
			}
		case err := <-watcher.Errors:
			return err
		}
	}

	return nil
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
