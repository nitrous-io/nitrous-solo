package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
)

func VerifyTempToken(c *gin.Context) {
	tok := c.Query("token")
	if tok == "" {
		c.JSON(422, gin.H{
			"error": "invalid_params",
			"errors": map[string]string{
				"token": "is required",
			},
		})
		c.Abort()
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
		c.Abort()
		return
	}

	c.Set(controllers.CurrentUserDataKey, ud)
	c.Next()
}
