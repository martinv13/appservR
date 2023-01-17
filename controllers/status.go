package controllers

import (
	"fmt"

	"github.com/appservR/appservR/modules/appserver"
	"github.com/appservR/appservR/modules/config"
	"github.com/gin-gonic/gin"
)

type StatusController struct {
	appServer *appserver.AppServer
	config    config.Config
}

// Create a new controller object
func NewStatusController(appServer *appserver.AppServer, config config.Config) *StatusController {
	return &StatusController{
		appServer: appServer,
		config:    config,
	}
}

// Render apps page
func (ctl *StatusController) Apps() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

// Render an app details page
func (ctl *StatusController) App() gin.HandlerFunc {
	return func(c *gin.Context) {

		appName := c.Param("appname")

		fmt.Println(appName)

	}
}
