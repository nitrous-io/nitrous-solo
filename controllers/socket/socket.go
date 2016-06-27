package socket

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func Socket(c *gin.Context) {
	err := ws.Upgrade(c.Writer, c.Request)
	if err != nil {
		controllers.InternalServerError(c, err)
		return
	}
}
