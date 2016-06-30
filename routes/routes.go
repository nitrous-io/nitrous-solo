package routes

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers/ping"
	"github.com/nitrous-io/nitrous-solo/controllers/root"
	"github.com/nitrous-io/nitrous-solo/controllers/socket"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"

	pinghandler "github.com/nitrous-io/nitrous-solo/handlers/ping"
)

func Draw(r *gin.Engine, appDir string) {
	if gin.Mode() != gin.TestMode {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	r.GET("/", root.Index)
	fmt.Println(filepath.Join(appDir, "assets"))
	r.Static("/assets", filepath.Join(appDir, "assets"))
	r.GET("/ping", ping.Ping)

	r.GET("/socket", socket.Socket)

	ws.RegisterHandler("ping", pinghandler.Ping)
}
