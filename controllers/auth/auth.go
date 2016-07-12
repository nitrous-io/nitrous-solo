package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
)

func CreateToken(c *gin.Context) {
	login := strings.ToLower(c.PostForm("login"))
	password := c.PostForm("password")

	if login == "" || password == "" {
		c.JSON(422, gin.H{
			"error":             "invalid_params",
			"error_description": "login and password should not be blank",
		})
		return
	}

	ud, ok := data.D.Users[login]

	if !ok || !ud.VerifyPassword(password) {
		c.JSON(422, gin.H{
			"error":             "invalid_credentials",
			"error_description": "invalid login or password",
		})
		return
	}

	t, err := token.GenerateTokenString(login)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": t,
	})
}
