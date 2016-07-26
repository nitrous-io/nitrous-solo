package socket

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func Socket(c *gin.Context) {
	ud := controllers.CurrentUserData(c)

	if err := ws.Upgrade(c.Writer, c.Request, ws.M{
		"login": ud.Login,
		"user":  ud.User,
	}); err != nil {
		controllers.InternalServerError(c, err)
		return
	}
}
