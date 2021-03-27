package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/martinv13/go-shiny/models"
)

type userData struct {
	Username      string
	DisplayedName string
	Groups        map[string]bool
}

// Get all users in a slice of struct with a boolean map to represent groups
func GetUsersData(db *gorm.DB) ([]userData, error) {
	var user models.User
	users, err := user.GetAll(db)
	if err != nil {
		return nil, err
	}
	usersData := make([]userData, len(users), len(users))
	for i, u := range users {
		usersData[i] = userData{Username: u.Username, DisplayedName: u.DisplayedName, Groups: u.GroupsMap(db)}
	}
	return usersData, nil
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		usersData, err := GetUsersData(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "users.html",
				gin.H{"loggedUserName": GetLoggedName(c), "selTab": "users", "errorMessage": "Unable to retrieve users."})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "users.html", gin.H{
			"loggedUserName": GetLoggedName(c),
			"selTab":         "users",
			"users":          usersData,
		})
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		username := c.Param("username")

		user := models.User{Username: username}

		if username == "new" {
			c.HTML(http.StatusOK, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
			})
			return
		}

		err := user.Get(db)

		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   fmt.Sprintf("User '%s' not found.", username),
			})
			c.Abort()
			return
		}

		c.HTML(http.StatusOK, "user.html", gin.H{
			"selTab":         "users",
			"loggedUserName": GetLoggedName(c),
			"username":       user.Username,
			"displayedName":  user.DisplayedName,
			"groups":         user.GroupsMap(db),
		})
	}
}

type userInfo struct {
	Username      string   `form:"username"`
	DisplayedName string   `form:"displayedname"`
	Groups        []string `form:"groups"`
	Password      string   `form:"password"`
}

func AdminUpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		username := c.Param("username")
		var info userInfo
		err := c.ShouldBind(&info)
		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   "User update failed. Please check the info provided.",
			})
			c.Abort()
			return
		}
		groups := make([]*models.Group, len(info.Groups), len(info.Groups))
		for i := range info.Groups {
			groups[i] = &models.Group{Name: info.Groups[i]}
		}
		user := models.User{
			Username:      info.Username,
			DisplayedName: info.DisplayedName,
			Groups:        groups,
			Password:      info.Password,
		}
		err = user.AdminUpdate(db, username)
		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   err.Error(),
			})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "user.html", gin.H{
			"selTab":         "users",
			"loggedUserName": GetLoggedName(c),
			"successMessage": "User has been updated",
			"username":       user.Username,
			"displayedName":  user.DisplayedName,
			"groups":         user.GroupsMap(db),
		})
	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		username := c.Param("username")
		user := models.User{Username: username}
		resData := gin.H{"loggedUserName": GetLoggedName(c), "selTab": "users"}
		err := user.Delete(db)
		if err != nil {
			resData["errorMessage"] = fmt.Sprintf("Could note delete user '%s'.", username)
			c.HTML(http.StatusBadRequest, "user.html", resData)
			c.Abort()
			return
		}
		usersData, err := GetUsersData(db)
		resData["successMessage"] = "User has been deleted"
		if err != nil {
			resData["errorMessage"] = "Unable to retrieve users"
			c.HTML(http.StatusInternalServerError, "users.html", resData)
			c.Abort()
			return
		}
		resData["users"] = usersData
		c.HTML(http.StatusOK, "users.html", resData)
	}
}
