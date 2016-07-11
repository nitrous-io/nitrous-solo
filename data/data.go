package data

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
)

var D *Data
var DataDirPath string

const dataFileName = "nitrous-solo-data"

type Data struct {
	Users map[string]*UserData // <Login, *UserData>
}

func new() *Data {
	return &Data{
		Users: map[string]*UserData{},
	}
}

func Load() {
	if DataDirPath == "" {
		logrus.Fatalf("path to data directory must be defined")
	}
	path := filepath.Join(DataDirPath, dataFileName)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// write default file
			D = new()
			if err := Save(); err != nil {
				logrus.Fatal(err)
			}
			return
		}
		logrus.Fatalf(`could not load data file "%s"`, path)
	}
	defer f.Close()

	var d Data
	if err := gob.NewDecoder(f).Decode(&d); err != nil {
		logrus.Fatalf(`could not parse data file "%s"`, path)
	}
	D = &d
}

func Save() error {
	if DataDirPath == "" {
		return errors.New("path to data directory must be defined")
	}
	tmpPath := filepath.Join(DataDirPath, dataFileName+".save")
	origPath := filepath.Join(DataDirPath, dataFileName)

	if err := os.MkdirAll(DataDirPath, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, DataDirPath))
	}

	f, err := os.OpenFile(tmpPath, (os.O_CREATE | os.O_TRUNC | os.O_RDWR), 0600)
	if err != nil {
		return errors.New(fmt.Sprintf(`could not write to file "%s"`, tmpPath))
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(D); err != nil {
		return errors.New(fmt.Sprintf(`could not write to file "%s"`, tmpPath))
	}

	f.Close()

	if err := os.Rename(tmpPath, origPath); err != nil {
		return errors.New(fmt.Sprintf(`could not write to file "%s"`, origPath))
	}

	return nil
}
