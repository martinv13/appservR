package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"gorm.io/gorm"
)

type ShinyAppPayload struct {
	AppName        string   `form:"appname" binding:"required"`
	Path           string   `form:"path" binding:"required"`
	AppDir         string   `form:"appdir" binding:"required"`
	Workers        int      `form:"workers" binding:"required"`
	Active         bool     `form:"active"`
	RestrictAccess bool     `form:"restrictaccess"`
	AllowedGroups  []string `form:"allowedgroups"`
}

func GetShinyApps() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		app := models.ShinyApp{}
		c.HTML(http.StatusOK, "apps.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "apps",
			"apps":           app.GetAllMapSlice(db),
		})
	}
}

func GetShinyApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		app := models.ShinyApp{AppName: c.Param("appname")}
		err := app.Get()
		if err == nil {
			c.HTML(http.StatusOK, "app.html", gin.H{
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

func UpdateShinyApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)

		var appInfo ShinyAppPayload

		if err := c.ShouldBind(&appInfo); err != nil {
			c.HTML(http.StatusBadRequest, "app.html", gin.H{
				"selTab":         "apps",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   "App update failed. Please check the info provided.",
			})
			c.Abort()
			return
		}
		appname := c.Param("appname")
		app := models.ShinyApp{
			AppName:        appInfo.AppName,
			Path:           appInfo.Path,
			AppDir:         appInfo.AppDir,
			Workers:        appInfo.Workers,
			Active:         appInfo.Active,
			RestrictAccess: appInfo.RestrictAccess,
		}

		if appname != "" {
			err := app.Update(db, appname)
			if err != nil {
				c.HTML(http.StatusBadRequest, "app.html", gin.H{
					"selTab":         "apps",
					"loggedUserName": GetLoggedName(c),
					"errorMessage":   "App update failed. Please check the info provided.",
				})
				c.Abort()
				return
			}
			c.HTML(http.StatusOK, "app.html", gin.H{
				"selTab":         "apps",
				"loggedUserName": GetLoggedName(c),
				"successMessage": "App updated successfuly.",
				"AppName":        app.AppName,
				"Path":           app.Path,
				"AppDir":         app.AppDir,
				"Workers":        app.Workers,
				"Active":         app.Active,
				"RestrictAccess": app.RestrictAccess,
				"AllowedGroups":  app.GroupsMap(db),
			})
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
