package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/controllers"
	"github.com/martinv13/go-shiny/data/assets"
	"github.com/martinv13/go-shiny/data/templates"
	"github.com/martinv13/go-shiny/middlewares"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/services/appproxy"
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

func NewRouter() *gin.Engine {
	router := gin.Default()

	t := template.New("")
	t, err := loadTemplate(t, "/")
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(t)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.StaticFS("/assets", &assets.Assets)

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
			user := models.UserData{}
			c.HTML(http.StatusOK, "users.html", gin.H{"displayedName": getName(c), "selTab": "users", "users": user.GetAll()})
		})
		admin.GET("/groups", func(c *gin.Context) {
			c.HTML(http.StatusOK, "groups.html", gin.H{"displayedName": getName(c), "selTab": "groups"})
		})
	}

	router.Use(appproxy.CreateProxy())

	return router

}
