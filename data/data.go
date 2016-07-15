package data

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/nitrous-io/nitrous-solo/config"
)

var D *Data
var DataDirPath string

const defaultDataFileName = "nitrous-solo-data"

type Data struct {
	Users map[string]*UserData // <Login, *UserData>
}

func Reset() *Data {
	return &Data{
		Users: map[string]*UserData{},
	}
}

func Load() {
	if DataDirPath == "" {
		logrus.Fatalf("path to data directory must be defined")
	}

	p := dataFileName()

	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			// write default file
			D = Reset()
			if err := Save(); err != nil {
				logrus.Fatal(err)
			}
			return
		}
		logrus.Fatalf(`could not load data file "%s"`, p)
	}
	defer f.Close()

	var d Data
	if err := gob.NewDecoder(f).Decode(&d); err != nil {
		logrus.Fatalf(`could not parse data file "%s"`, p)
	}
	D = &d
}

func Save() error {
	if DataDirPath == "" {
		return errors.New("path to data directory must be defined")
	}

	p := dataFileName()

	f, err := ioutil.TempFile("", "nitrous-data")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Chmod(0600); err != nil {
		return err
	}

	if err := gob.NewEncoder(f).Encode(D); err != nil {
		return errors.New(fmt.Sprintf(`could not write to file "%s"`, f.Name()))
	}

	f.Close()

	if err := os.MkdirAll(DataDirPath, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, DataDirPath))
	}

	if err := os.Rename(f.Name(), p); err != nil {
		return errors.New(fmt.Sprintf(`could not write to file "%s"`, p))
	}

	return nil
}

func dataFileName() string {
	suffix := ""
	if e := config.Env; e != "production" {
		suffix = "-" + e
	}

	return filepath.Join(DataDirPath, defaultDataFileName+suffix)
}
