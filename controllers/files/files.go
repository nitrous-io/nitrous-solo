package files

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
)

func Upload(c *gin.Context) {
	ud := controllers.CurrentUserData(c)

	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "the request should be encoded in multipart/form-data format",
		})
		return
	}

	var location string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			c.JSON(422, gin.H{
				"error": "invalid_params",
				"errors": map[string]interface{}{
					"files": "is required",
				},
			})
			return
		}

		if part.FormName() == "location" {
			out, err := ioutil.ReadAll(part)
			if err != nil {
				controllers.InternalServerError(c, err)
				return
			}

			location = string(out)
		} else if part.FormName() == "files" {
			if location == "" {
				c.JSON(422, gin.H{
					"error": "invalid_params",
					"errors": map[string]interface{}{
						"location": "is required",
					},
				})
				return
			}

			if err := atomicwrite.AtomicWrite(part, location, ud.User, true); err != nil {
				if err == atomicwrite.ErrPermissionDenied {
					c.JSON(http.StatusForbidden, gin.H{
						"error":             "invalid_request",
						"error_description": "the user does not have permission to write to the path",
					})
					return
				}

				if err == atomicwrite.ErrInvalidUser {
					c.JSON(http.StatusForbidden, gin.H{
						"error":             "invalid_request",
						"error_description": "the user could not be found",
					})
					return
				}

				controllers.InternalServerError(c, err)
				return
			}

			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
