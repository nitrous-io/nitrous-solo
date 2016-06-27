package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers/ping"
	"github.com/nitrous-io/nitrous-solo/controllers/root"
	"github.com/nitrous-io/nitrous-solo/controllers/socket"
)

func Draw(r *gin.Engine) {
	if gin.Mode() != gin.TestMode {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	r.GET("/", root.Index)
	r.Static("/assets", "./assets")
	r.GET("/ping", ping.Ping)

	r.GET("/socket", socket.Socket)
}
