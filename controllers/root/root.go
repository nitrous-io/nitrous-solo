package root

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "root.index", gin.H{
		"config": template.JS(`{ "foo": "bar" }`),
	})
}
