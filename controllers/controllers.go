package controllers

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Sirupsen/logrus"
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
