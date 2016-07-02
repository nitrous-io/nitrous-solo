package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/osext"
	"github.com/nitrous-io/nitrous-solo/server"
)

func main() {
	if os.Getenv("NITROUS_ENV") == "production" {
		gin.SetMode("release")
	}
	appDir, err := osext.ExecutableFolder()
	if err != nil {
		logrus.Fatal(err)
	}
	r := server.New(appDir)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":3000"
	}
	r.Run(addr)
}
