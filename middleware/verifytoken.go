package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
)

var tokenAuthHeaderRe = regexp.MustCompile(`\A\s*Token\s+([\S]+)\s*\z`)

func VerifyToken(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	match := tokenAuthHeaderRe.FindStringSubmatch(authHeader)
	if match == nil || len(match) < 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "access token is required",
		})
		c.Abort()
		return
	}

	ud := controllers.GetUserDataByToken(match[1])
	if ud == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "access token is invalid",
		})
		c.Abort()
		return
	}

	c.Set(controllers.CurrentUserDataKey, ud)
	c.Next()
}
