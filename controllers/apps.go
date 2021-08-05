package controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appserver"
	"github.com/gin-gonic/gin"
)

type AppController struct {
	appModel  models.AppModel
	appServer *appserver.AppServer
}

func NewAppController(appModel models.AppModel, appServer *appserver.AppServer) *AppController {
	return &AppController{
		appModel:  appModel,
		appServer: appServer,
	}
}

// Render apps page
func (ctl *AppController) GetRApps() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "apps.html", ctl.buildAppsTemplateData(c))
	}
}

// Render an app details page
func (ctl *AppController) GetRApp() gin.HandlerFunc {
	return func(c *gin.Context) {

		appName := c.Param("appname")

		var app models.RApp
		var err error
		var data gin.H

		if appName != "new" {
			app, err = ctl.appModel.Find(appName)
			if err == nil {
				data, err = ctl.buildAppTemplateData(app, c)
			}
		}

		if appName == "new" || err != nil {
			data, err = ctl.buildAppTemplateData(models.RApp{}, c)
			data["Title"] = "New App"
			c.HTML(http.StatusOK, "app.html", data)
			return
		}
		c.HTML(http.StatusOK, "app.html", data)
	}
}

type RAppSettings struct {
	AppName       string   `form:"appname" binding:"required"`
	Path          string   `form:"path" binding:"required"`
	Properties    []string `form:"properties[]"`
	AllowedGroups []string `form:"allowedgroups"`
	AppDir        string   `form:"appdir"`
	Workers       int      `form:"workers"`
}

// Update or create an app
func (ctl *AppController) UpdateRApp() gin.HandlerFunc {
	return func(c *gin.Context) {

		var appInfo RAppSettings
		var res map[string]interface{}
		appname := c.Param("appname")
		err := c.ShouldBind(&appInfo)
		if err == nil && appname != "" {
			isActive := false
			isRestricted := false
			for _, val := range appInfo.Properties {
				if val == "active" {
					isActive = true
				}
				if val == "restrictaccess" {
					isRestricted = true
				}
			}
			app := models.RApp{
				AppName:        appInfo.AppName,
				Path:           appInfo.Path,
				AppDir:         appInfo.AppDir,
				Workers:        appInfo.Workers,
				Active:         isActive,
				RestrictAccess: isRestricted,
			}
			err = ctl.appModel.Save(app, appname)
			if err == nil {
				ctl.appServer.Update(appname, app)
				res, err = ctl.buildAppTemplateData(app, c)
				res["successMessage"] = "App updated successfuly."
				c.HTML(http.StatusOK, "app.html", res)
				return
			}
		}
		res = make(gin.H)
		v := reflect.ValueOf(appInfo)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			res[t.Field(i).Name] = v.Field(i).Interface()
		}
		res = gin.H{"AppSettings": res}
		res["selTab"] = "apps"
		res["loggedUserName"] = GetLoggedName(c)
		res["errorMessage"] = "App update failed. Please check the info provided."
		c.HTML(http.StatusBadRequest, "app.html", res)
		c.Abort()
		return
	}
}

// Controller function to delete a R app
func (ctl *AppController) DeleteRApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appName := c.Param("appname")
		err := ctl.appModel.Delete(appName)
		res := ctl.buildAppsTemplateData(c)
		ctl.appServer.Update(appName, models.RApp{})
		if err != nil {
			res["errorMessage"] = fmt.Sprintf("An error occured while deleting app %s.", appName)
			c.HTML(http.StatusOK, "apps.html", res)
			c.Abort()
			return
		}
		res["successMessage"] = fmt.Sprintf("App %s has been deleted.", appName)
		c.HTML(http.StatusOK, "apps.html", res)
	}
}

// Build map for use in template
func (ctl *AppController) buildAppTemplateData(app models.RApp, c *gin.Context) (gin.H, error) {
	appMap, err := ctl.appModel.AsMap(app)
	if err != nil {
		return nil, err
	}
	data := gin.H{
		"loggedUserName": GetLoggedName(c),
		"selTab":         "apps",
		"Title":          strings.Title(app.AppName),
		"AppSettings":    appMap,
	}
	if app.AppName != "" {
		status, err := ctl.appServer.GetStatus(app.AppName)
		if err != nil {
			return nil, err
		}
		data["Status"] = status
	}
	return data, nil
}

// Build map with all apps for use in a template
func (ctl *AppController) buildAppsTemplateData(c *gin.Context) gin.H {
	apps, _ := ctl.appModel.All()
	status := ctl.appServer.GetAllStatus()
	res := make(map[string]interface{})
	for _, a := range apps {
		res[a.AppName] = map[string]interface{}{
			"AppName": a.AppName,
			"Path":    a.Path,
			"Title":   strings.Title(a.AppName),
			"Status":  status[a.AppName],
		}
	}
	return gin.H{
		"loggedUserName": GetLoggedName(c),
		"selTab":         "apps",
		"apps":           res,
	}
}
