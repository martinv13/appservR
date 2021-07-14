package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/controllers"
	"github.com/martinv13/go-shiny/middlewares"
	"github.com/martinv13/go-shiny/modules/appserver"
	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/ssehandler"
	"github.com/martinv13/go-shiny/modules/vfsdata"
)

type AppRouter struct {
	router *gin.Engine
	config config.Config
}

// Create the router instance
func NewAppRouter(config config.Config, staticPaths *vfsdata.StaticPaths,
	appServer *appserver.AppServer, msgBroker *ssehandler.MessageBroker,
	appsCtl *controllers.AppController, usersCtl *controllers.UserController,
	groupsCtl *controllers.GroupController, authCtl *controllers.AuthController) (*AppRouter, error) {

	mode := config.GetString("mode")
	if mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	t := template.New("")
	t, err := loadTemplate(t, "/", staticPaths)
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(t)

	router.Use(gin.Recovery())
	if mode != "prod" {
		router.Use(gin.Logger())
	}

	router.StaticFS("/assets", staticPaths.Assets)

	router.Use(middlewares.Auth())

	auth := router.Group("/auth")
	auth = addAuthRoutes(auth, authCtl)

	admin := router.Group("/admin")
	admin.Use(middlewares.AdminAuth())
	admin = addAdminRoutes(admin, msgBroker, appsCtl, usersCtl, groupsCtl)

	router.Use(appServer.CreateProxy())

	server := &AppRouter{router: router, config: config}

	return server, nil
}

func (s *AppRouter) Start() error {
	s.config.Logger().Warning(fmt.Sprintf("Starting server on %s:%s", s.config.GetString("server.host"), s.config.GetString("server.port")))
	return s.router.Run(fmt.Sprintf("%s:%s", s.config.GetString("server.host"), s.config.GetString("server.port")))
}

// Load templates recursively using the embeded files if no equivalent file exist in the local directory
func loadTemplate(t *template.Template, path string, staticPaths *vfsdata.StaticPaths) (*template.Template, error) {
	bd, err := staticPaths.Templates.BundledFS.Open(path)
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
			t, err = loadTemplate(t, path+bfi.Name(), staticPaths)
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(bfi.Name(), ".html") {
			file, err := staticPaths.Templates.Open(path + "/" + bfi.Name())
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
