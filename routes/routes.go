package routes

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers/auth"
	"github.com/nitrous-io/nitrous-solo/controllers/ping"
	"github.com/nitrous-io/nitrous-solo/controllers/root"
	"github.com/nitrous-io/nitrous-solo/controllers/socket"
	"github.com/nitrous-io/nitrous-solo/controllers/user"
	"github.com/nitrous-io/nitrous-solo/middleware"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"

	"github.com/nitrous-io/nitrous-solo/handlers/fs"
	ping_ "github.com/nitrous-io/nitrous-solo/handlers/ping"
	"github.com/nitrous-io/nitrous-solo/handlers/sh"
)

func Draw(r *gin.Engine, appDir string) {
	if gin.Mode() != gin.TestMode {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	r.GET("/", root.Index)
	r.Static("/assets", filepath.Join(appDir, "assets"))
	r.GET("/ping", ping.Ping)

	r.POST("/token", auth.CreateToken)
	r.GET("/socket", socket.Socket)

	{ // routes requiring authentication
		ar := r.Group("", middleware.VerifyToken)
		ar.GET("/user", user.Show)
	}

	ws.RegisterHandler("ping", ping_.Ping)
	ws.RegisterHandler("fs.ls", fs.List)
	ws.RegisterHandler("fs.read", fs.Read)
	ws.RegisterHandler("fs.write", fs.Write)
	ws.RegisterHandler("fs.watch", fs.Watch)
	ws.RegisterHandler("fs.unwatch", fs.Unwatch)

	ws.RegisterHandler("sh.start", sh.Start)
	ws.RegisterHandler("sh.attach", sh.Attach)
	ws.RegisterHandler("sh.detach", sh.Detach)
	// short names for messages that get sent repeatedly
	ws.RegisterHandler("sh.i", sh.Input)
	ws.RegisterHandler("sh.r", sh.Resize)
	ws.RegisterHandler("sh.kill", sh.Kill)
	ws.RegisterHandler("sh.list", sh.List)
	ws.RegisterHandler("sh.run", sh.Run)
}
