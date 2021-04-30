package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/martinv13/go-shiny/models"
)

type UserController struct {
	userModel models.UserModel
}

func NewUserController(userModel models.UserModel) *UserController {
	return &UserController{
		userModel: userModel,
	}
}

type userData struct {
	Username      string
	DisplayedName string
	Groups        map[string]bool
}

// Get all users in a slice of struct with a boolean map to represent groups
func (userCtl *UserController) GetUsersData() ([]userData, error) {
	users, err := userCtl.userModel.All()
	if err != nil {
		return nil, err
	}
	usersData := make([]userData, len(users), len(users))
	for i, u := range users {
		usersData[i] = userData{Username: u.Username, DisplayedName: u.DisplayedName, Groups: userCtl.userModel.GroupsMap(&u)}
	}
	return usersData, nil
}

func (userCtl *UserController) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		usersData, err := userCtl.GetUsersData()
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

func (userCtl *UserController) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		if username == "new" {
			c.HTML(http.StatusOK, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
			})
			return
		}

		user, err := userCtl.userModel.FindByUsername(username)

		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   fmt.Sprintf("User '%s' not found.", username),
			})
			c.Abort()
			return
		}

		allowedGroups := userCtl.userModel.GroupsMap(user)
		c.HTML(http.StatusOK, "user.html", gin.H{
			"selTab":         "users",
			"loggedUserName": GetLoggedName(c),
			"username":       user.Username,
			"displayedName":  user.DisplayedName,
			"groups":         allowedGroups,
		})
	}
}

type userInfo struct {
	Username      string   `form:"username"`
	DisplayedName string   `form:"displayedname"`
	Groups        []string `form:"groups"`
	Password      string   `form:"password"`
}

func (userCtl *UserController) AdminUpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		err = userCtl.userModel.AdminSave(&user, username)
		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   err.Error(),
			})
			c.Abort()
			return
		}
		allowedGroups := userCtl.userModel.GroupsMap(&user)
		c.HTML(http.StatusOK, "user.html", gin.H{
			"selTab":         "users",
			"loggedUserName": GetLoggedName(c),
			"successMessage": "User has been updated",
			"username":       user.Username,
			"displayedName":  user.DisplayedName,
			"groups":         allowedGroups,
		})
	}
}

func (userCtl *UserController) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		resData := gin.H{"loggedUserName": GetLoggedName(c), "selTab": "users"}
		err := userCtl.userModel.DeleteByUsername(username)
		if err != nil {
			resData["errorMessage"] = fmt.Sprintf("Could note delete user '%s'.", username)
			c.HTML(http.StatusBadRequest, "user.html", resData)
			c.Abort()
			return
		}
		usersData, err := userCtl.GetUsersData()
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
