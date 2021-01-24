package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/controllers"
	"github.com/martinv13/go-shiny/middlewares"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*/*.html")

	router.Use(middlewares.Auth())

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	router.POST("/login", controllers.DoLogin())

	router.GET("/logout", controllers.DoLogout())

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	router.POST("/signup", controllers.DoSignup())

	admin := router.Group("/admin", middlewares.AdminAuth())
	{
		getName := func(c *gin.Context) string {
			name := "unknown"
			nameVal, ok := c.Get("displayedname")
			if ok {
				name = fmt.Sprintf("%s", nameVal)
			}
			return name
		}
		admin.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/admin/settings")
		})
		admin.GET("/settings", func(c *gin.Context) {

			c.HTML(http.StatusOK, "settings.html", gin.H{"displayedName": getName(c), "selTab": "settings"})
		})
		admin.GET("/apps", func(c *gin.Context) {
			app := models.ShinyApp{}
			c.HTML(http.StatusOK, "apps.html", gin.H{"displayedName": getName(c), "selTab": "apps", "apps": app.GetAll()})
		})
		admin.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "users.html", gin.H{"displayedName": getName(c), "selTab": "users"})
		})
		admin.GET("/groups", func(c *gin.Context) {
			c.HTML(http.StatusOK, "groups.html", gin.H{"displayedName": getName(c), "selTab": "groups"})
		})
	}

	router.Use(appproxy.CreateProxy())

	return router

}
