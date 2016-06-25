package root

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"config": template.JS(`{ "foo": "bar" }`),
	})
}
