package user

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	. "github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/data"
)

func Show(c *gin.Context) {
	ud := CurrentUserData(c)

	c.JSON(http.StatusOK, gin.H{
		"user": ud.AsJSON(),
	})
}

func Update(c *gin.Context) {
	var updatedUserData data.UserData
	if err := c.BindJSON(&updatedUserData); err != nil {
		InternalServerError(c, err)
	}

	data.D.Users[updatedUserData.Login] = &updatedUserData
	if err := data.Save(); err != nil {
		logrus.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"user": updatedUserData.AsJSON(),
	})
}
