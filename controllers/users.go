package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/martinv13/go-shiny/models"
)

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)

		username := c.Param("username")

		var user models.UserData
		err := user.Get(db, username)

		if err == nil {

			c.HTML(http.StatusOK, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"username":       user.Username,
				"displayedName":  user.DisplayedName,
				"groups":         user.Groups,
			})
		}
	}
}
