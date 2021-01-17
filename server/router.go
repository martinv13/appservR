package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/**/*")

	admin := router.Group("/admin")
	{
		admin.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/admin/settings")
		})
		admin.GET("/settings", func(c *gin.Context) {
			c.HTML(http.StatusOK, "settings.tpl", gin.H{})
		})
		admin.GET("/apps", func(c *gin.Context) {
			c.HTML(http.StatusOK, "apps.tpl", gin.H{})
		})
		admin.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "users.tpl", gin.H{})
		})
		admin.GET("/groups", func(c *gin.Context) {
			c.HTML(http.StatusOK, "groups.tpl", gin.H{})
		})
	}

	router.Use(appproxy.CreateProxy())

	return router

}
