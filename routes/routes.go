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

	ws.RegisterHandler("sh.start", shell.Start)
	ws.RegisterHandler("sh.attach", shell.Attach)
	ws.RegisterHandler("sh.detach", shell.Detach)
	// short names for messages that get sent repeatedly
	ws.RegisterHandler("sh.i", shell.Input)
	ws.RegisterHandler("sh.r", shell.Resize)
	ws.RegisterHandler("sh.kill", shell.Kill)
	ws.RegisterHandler("sh.list", shell.List)
}
