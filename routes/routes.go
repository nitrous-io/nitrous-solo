package routes

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers/ping"
	"github.com/nitrous-io/nitrous-solo/controllers/root"
	"github.com/nitrous-io/nitrous-solo/controllers/socket"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	ping_ "github.com/nitrous-io/nitrous-solo/handlers/ping"
	"github.com/nitrous-io/nitrous-solo/handlers/shell"
)

func Draw(r *gin.Engine, appDir string) {
	if gin.Mode() != gin.TestMode {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	r.GET("/", root.Index)
	r.Static("/assets", filepath.Join(appDir, "assets"))
	r.GET("/ping", ping.Ping)

	r.GET("/socket", socket.Socket)

	ws.RegisterHandler("ping", ping_.Ping)
	ws.RegisterHandler("fs.ls", fs.List)
	ws.RegisterHandler("fs.read", fs.Read)

	ws.RegisterHandler("sh.s", shell.Start) // short name because many replies will happen
	ws.RegisterHandler("sh.a", shell.Attach)
	ws.RegisterHandler("sh.detach", shell.Detach)
	ws.RegisterHandler("sh.i", shell.Input)
}
