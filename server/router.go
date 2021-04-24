package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/controllers"
	"github.com/martinv13/go-shiny/middlewares"
	"github.com/martinv13/go-shiny/modules/appproxy"
	"github.com/martinv13/go-shiny/modules/ssehandler"
	"github.com/martinv13/go-shiny/vfsdata/assets"
	"github.com/martinv13/go-shiny/vfsdata/templates"
	"gorm.io/gorm"
)

func loadTemplate(t *template.Template, path string) (*template.Template, error) {
	bd, err := templates.BundledTemplates.Open(path)
	if err != nil {
		return nil, err
	}
	defer bd.Close()
	list, err := bd.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, bfi := range list {
		if bfi.IsDir() {
			t, err = loadTemplate(t, path+bfi.Name())
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(bfi.Name(), ".html") {
			file, err := templates.Templates.Open(path + "/" + bfi.Name())
			if err != nil {
				return nil, err
			}
			defer file.Close()
			h, err := ioutil.ReadAll(file)
			if err != nil {
				return nil, err
			}
			t, err = t.New(bfi.Name()).Parse(string(h))
			if err != nil {
				return nil, err
			}
		}
	}
	return t, nil
}

func NewRouter(db *gorm.DB, stream *ssehandler.Event) *gin.Engine {
	router := gin.New()

	t := template.New("")
	t, err := loadTemplate(t, "/")
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(t)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.StaticFS("/assets", &assets.Assets)

	router.Use(middlewares.SetDB(db))
	router.Use(middlewares.Auth())

	auth := router.Group("/auth")
	{
		auth.GET("/login", func(c *gin.Context) {
			ref := c.Request.Header["Referer"][0]
			fmt.Println(ref)
			if ref != "" {
				url, _ := url.Parse(ref)
				ref = url.Path
			} else {
				ref = "/"
			}
			c.HTML(http.StatusOK, "login.html", gin.H{"Referer": ref})
		})
		auth.POST("/login", controllers.DoLogin())
		auth.GET("/logout", controllers.DoLogout())
		auth.GET("/signup", func(c *gin.Context) {
			c.HTML(http.StatusOK, "signup.html", nil)
		})
		auth.POST("/signup", controllers.DoSignup())
	}

	admin := router.Group("/admin")
	admin.Use(middlewares.AdminAuth())
	{
		admin.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/admin/apps")
		})

		admin.GET("/apps", controllers.GetShinyApps())
		admin.GET("/apps/:appname", controllers.GetShinyApp())
		admin.POST("/apps/:appname", controllers.UpdateShinyApp())
		admin.GET("/apps/:appname/delete", controllers.DeleteShinyApp())

		admin.GET("/apps.json", stream.Controller())

		admin.GET("/users", controllers.GetUsers())
		admin.GET("/users/:username", controllers.GetUser())
		admin.POST("/users/:username", controllers.AdminUpdateUser())
		admin.GET("/users/:username/delete", controllers.DeleteUser())

		admin.GET("/groups", controllers.GetGroups())
		admin.GET("/groups/:groupname", controllers.GetGroup())
		admin.POST("/groups/:groupname", controllers.UpdateGroup())
		admin.GET("/groups/:groupname/delete", controllers.DeleteGroup())
		admin.GET("/groups/:groupname/add/:username", controllers.AddGroupMember())
		admin.GET("/groups/:groupname/remove/:username", controllers.RemoveGroupMember())

	}

	router.Use(appproxy.CreateProxy())

	return router

}
