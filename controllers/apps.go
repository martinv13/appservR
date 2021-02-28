package controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/appproxy"
	"gorm.io/gorm"
)

// Get all app infos
func GetShinyApps() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "apps.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "apps",
			"apps":           appproxy.GetAllStatus(),
		})
	}
}

// Build map for use in template
func BuildShinyAppMap(app models.ShinyApp, c *gin.Context) (gin.H, error) {
	dbi, _ := c.Get("DB")
	db := dbi.(*gorm.DB)
	data := gin.H{
		"loggedUserName": GetLoggedName(c),
		"selTab":         "apps",
		"Title":          strings.Title(app.AppName),
		"AppName":        app.AppName,
		"Path":           app.Path,
		"AppDir":         app.AppDir,
		"Workers":        app.Workers,
		"Active":         app.Active,
		"RestrictAccess": app.RestrictAccess,
		"AllowedGroups":  app.GroupsMap(db),
	}
	status, err := appproxy.GetStatus(app.AppName)
	if err != nil {
		return nil, err
	}
	for k, v := range status {
		data[k] = v
	}
	return data, nil
}

func GetShinyApp() gin.HandlerFunc {
	return func(c *gin.Context) {

		appName := c.Param("appname")

		var app models.ShinyApp
		var err error
		var data gin.H

		if appName != "new" {
			app = models.ShinyApp{AppName: appName}
			err = app.Get()
			if err == nil {
				data, err = BuildShinyAppMap(app, c)
			}
		}

		if appName == "new" || err != nil {
			c.HTML(http.StatusOK, "app.html", gin.H{
				"selTab":         "apps",
				"loggedUserName": GetLoggedName(c),
				"Title":          "New App",
			})
			return
		}
		c.HTML(http.StatusOK, "app.html", data)
	}
}

type ShinyAppSettings struct {
	AppName       string   `form:"appname" binding:"required"`
	Path          string   `form:"path" binding:"required"`
	Properties    []string `form:"properties[]"`
	AllowedGroups []string `form:"allowedgroups"`
	AppDir        string   `form:"appdir"`
	Workers       int      `form:"workers"`
}

// Update or create app
func UpdateShinyApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)

		var appInfo ShinyAppSettings

		if err := c.ShouldBind(&appInfo); err != nil {
			fmt.Println(err)
			res := make(gin.H)
			v := reflect.ValueOf(appInfo)
			t := v.Type()
			for i := 0; i < v.NumField(); i++ {
				res[t.Field(i).Name] = v.Field(i).Interface()
			}
			res["selTab"] = "apps"
			res["loggedUserName"] = GetLoggedName(c)
			res["errorMessage"] = "App update failed. Please check the info provided."
			c.HTML(http.StatusBadRequest, "app.html", res)
			c.Abort()
			return
		}
		appname := c.Param("appname")
		isActive := false
		isRestricted := false
		fmt.Println(appInfo.Properties)
		for _, val := range appInfo.Properties {
			if val == "active" {
				isActive = true
			}
			if val == "restrictaccess" {
				isRestricted = true
			}
		}
		app := models.ShinyApp{
			AppName:        appInfo.AppName,
			Path:           appInfo.Path,
			AppDir:         appInfo.AppDir,
			Workers:        appInfo.Workers,
			Active:         isActive,
			RestrictAccess: isRestricted,
		}

		if appname != "" {
			err := app.Update(db, appname)
			if err != nil {
				c.HTML(http.StatusBadRequest, "app.html", gin.H{
					"selTab":         "apps",
					"loggedUserName": GetLoggedName(c),
					"errorMessage":   "Update failed. Please check the info provided.",
				})
				c.Abort()
				return
			}
			res, err := BuildShinyAppMap(app, c)
			if err != nil {
				fmt.Println(err)
				c.HTML(http.StatusBadRequest, "app.html", gin.H{
					"selTab":         "apps",
					"loggedUserName": GetLoggedName(c),
					"errorMessage":   "Update failed. Please check the info provided.",
				})
				c.Abort()
				return
			}
			res["successMessage"] = "App updated successfuly."
			c.HTML(http.StatusOK, "app.html", res)
			return
		}
		c.HTML(http.StatusBadRequest, "app.html", gin.H{
			"selTab":         "apps",
			"loggedUserName": GetLoggedName(c),
			"errorMessage":   "App update failed. Please check the info provided.",
		})
		c.Abort()
	}
}

// Controller function to delete a Shiny app
func DeleteShinyApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		app := models.ShinyApp{AppName: c.Param("appname")}
		err := app.Delete(db)
		if err == nil {
			c.HTML(http.StatusOK, "apps.html", gin.H{
				"loggedUserName": GetLoggedName(c),
				"selTab":         "apps",
				"AppName":        app.AppName,
				"Path":           app.Path,
				"AppDir":         app.AppDir,
				"Workers":        app.Workers,
				"Active":         app.Active,
				"RestrictAccess": app.RestrictAccess,
				"AllowedGroups":  app.GroupsMap(db),
			})
		}
	}
}
