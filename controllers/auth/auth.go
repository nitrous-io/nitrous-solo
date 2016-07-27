package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
)

var DefaultTemporaryTokenExpiry = 5 * time.Second

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

	t, err := token.GenerateTokenString(login, token.LoginType)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": t,
	})
}

func CreateTemporaryToken(c *gin.Context) {
	ud := CurrentUserData(c)

	t, err := token.GenerateTokenStringWithExpiry(ud.Login, DefaultTemporaryTokenExpiry, token.TempType)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": t,
	})
}
