package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/nitrous-io/nitrous-solo/controllers"
)

func Show(c *gin.Context) {
	ud := CurrentUserData(c)

	c.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"login": ud.Login,
			"user":  ud.User,
		},
	})
}
