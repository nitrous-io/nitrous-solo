package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/routes"
	"github.com/nitrous-io/nitrous-solo/templates"
)

func New(appDir string) *gin.Engine {
	r := gin.New()
	r.SetHTMLTemplate(templates.T)
	routes.Draw(r, appDir)
	return r
}
