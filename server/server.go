package server

import (
	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/routes"
)

func New() *gin.Engine {
	r := gin.New()
	r.LoadHTMLGlob("./templates/*.tmpl")
	routes.Draw(r)
	return r
}
