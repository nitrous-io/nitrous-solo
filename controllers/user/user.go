package user

import (
	"net/http"
	"os/user"

	"github.com/gin-gonic/gin"
	. "github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/data"
)

func Show(c *gin.Context) {
	ud := CurrentUserData(c)

	systemUser, err := user.Lookup(ud.User)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	u := map[string]interface{}{
		"login": ud.Login,
		"user":  ud.User,
		"home":  systemUser.HomeDir,
	}

	if ide := CurrentUserIDESettings(c); ide != nil {
		u["ide_settings"] = ide
	}

	c.JSON(http.StatusOK, gin.H{
		"user": u,
	})
}

func SaveIDESettings(c *gin.Context) {
	ud := CurrentUserData(c)

	ide := CurrentUserIDESettings(c)
	if ide == nil {
		ide = data.IDESettings{}
	}

	j := data.IDESettings{}
	if err := c.Bind(&j); err != nil {
		c.JSON(422, gin.H{"error": "invalid_params"})
		return
	}

	for k, v := range j {
		ide[k] = v
	}

	data.D.IDESettings[ud.Login] = ide

	if err := data.SaveIDESettings(); err != nil {
		InternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ide_settings": data.D.IDESettings[ud.Login],
	})
}
