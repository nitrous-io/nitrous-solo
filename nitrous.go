package main

import (
	"flag"
	"os/user"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/osext"
	"github.com/nitrous-io/nitrous-solo/config"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/server"
)

func main() {
	if config.Env == "production" {
		gin.SetMode("release")
	}

	appDir, err := osext.ExecutableFolder()
	if err != nil {
		logrus.Fatal(err)
	}

	dataDirPathFlag := flag.String("d", "", "path to data directory")
	bindAddrFlag := flag.String("b", "0.0.0.0:6877", "bind address and port")
	flag.Parse()

	dataDirPath := *dataDirPathFlag
	bindAddr := *bindAddrFlag

	if dataDirPath == "" {
		u, err := user.Current()
		if err != nil {
			logrus.Fatal("could not determine current user")
		}
		dataDirPath = filepath.Join(u.HomeDir, ".nitrous-solo")
	}

	data.DataDirPath = dataDirPath
	if err := data.LoadUserData(); err != nil {
		logrus.Fatal(err)
	}
	if err := data.LoadIDESettings(); err != nil {
		logrus.Fatal(err)
	}

	r := server.New(appDir)
	r.Run(bindAddr)
}
