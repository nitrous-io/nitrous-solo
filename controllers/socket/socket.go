package socket

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
	"github.com/nitrous-io/nitrous-solo/pkg/ws"
)

func Socket(c *gin.Context) {
	tok := c.Query("token")
	if tok == "" {
		c.JSON(422, gin.H{
			"error": "invalid_params",
			"errors": map[string]string{
				"token": "is required",
			},
		})
		return
	}

	ud := controllers.GetUserDataByToken(tok, token.TempType)
	if ud == nil {
		c.JSON(422, gin.H{
			"error": "invalid_credentials",
			"errors": map[string]string{
				"token": "is invalid",
			},
		})
		return
	}

	if err := ws.Upgrade(c.Writer, c.Request, ws.M{
		"login": ud.Login,
		"user":  ud.User,
	}); err != nil {
		controllers.InternalServerError(c, err)
		return
	}
}
