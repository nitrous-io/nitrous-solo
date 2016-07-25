package controllers

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/data"
	"github.com/nitrous-io/nitrous-solo/pkg/token"

	"github.com/Sirupsen/logrus"
)

const (
	CurrentUserDataKey = "current_user_data"
)

func InternalServerError(c *gin.Context, err error) {
	var (
		errMsg  = "internal server error"
		errHash string
	)

	if err != nil {
		errMsg = err.Error()
		errHash = fmt.Sprintf("%x", sha1.Sum([]byte(errMsg)))
	}

	req := c.Request

	fields := logrus.Fields{
		"req": fmt.Sprintf("%s %s", req.Method, req.URL.String()),
		"ip":  c.ClientIP(),
	}

	j := gin.H{
		"error": "internal_server_error",
	}

	if errHash != "" {
		fields["hash"] = errHash
		j["error_hash"] = errHash
	}

	if (req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH") && strings.Contains(c.ContentType(), "application/x-www-form-urlencoded") {
		if err := req.ParseForm(); err == nil {
			fields["form"] = req.PostForm.Encode()
		}
	}

	logrus.WithFields(fields).Error(errMsg)
	c.JSON(http.StatusInternalServerError, j)
}

func GetUserDataByToken(tok string, tokenType token.Type) *data.UserData {
	login := token.ValidateTokenString(tok, tokenType)
	if login == "" {
		return nil
	}

	ud, ok := data.D.Users[login]
	if !ok {
		return nil
	}

	return ud
}

func CurrentUserData(c *gin.Context) *data.UserData {
	ui, exists := c.Get(CurrentUserDataKey)
	if ui == nil || !exists {
		return nil
	}

	ud, ok := ui.(*data.UserData)
	if !ok {
		return nil
	}
	return ud
}
