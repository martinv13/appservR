package server

import (
	"net/http"
	"net/url"

	"github.com/appservR/appservR/controllers"
	"github.com/gin-gonic/gin"
)

func addAuthRoutes(auth *gin.RouterGroup, authCtl *controllers.AuthController) *gin.RouterGroup {
	auth.GET("/login", func(c *gin.Context) {
		refs := c.Request.Header["Referer"]
		ref := "/"
		if len(refs) > 0 {
			ref = refs[0]
		}
		if ref != "/" {
			url, _ := url.Parse(ref)
			ref = url.Path
		}
		c.HTML(http.StatusOK, "login.html", gin.H{"Referer": ref})
	})
	auth.POST("/login", authCtl.DoLogin())
	auth.GET("/logout", authCtl.DoLogout())
	auth.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})
	auth.POST("/signup", authCtl.DoSignup())
	return auth
}
