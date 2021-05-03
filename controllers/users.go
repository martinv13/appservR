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

// Get all users
func (userCtl *UserController) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userCtl.userModel.All()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "users.html",
				gin.H{"loggedUserName": GetLoggedName(c), "selTab": "users", "errorMessage": "Unable to retrieve users."})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "users.html", userCtl.buildUsersTemplateData(users, c))
	}
}

func (userCtl *UserController) GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")

		if username == "new" {
			c.HTML(http.StatusOK, "user.html", userCtl.buildUserTemplateData(models.User{}, c))
			return
		}

		user, err := userCtl.userModel.Find(username)

		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   fmt.Sprintf("User '%s' not found.", username),
			})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "user.html", userCtl.buildUserTemplateData(user, c))
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
		groups := make([]models.Group, len(info.Groups), len(info.Groups))
		isAdmin := false
		for i := range info.Groups {
			groups[i] = models.Group{Name: info.Groups[i]}
			isAdmin = isAdmin || groups[i].Name == "admins"
		}
		loggedUser, ok := c.Get("username")
		if ok {
			if username == loggedUser && !isAdmin {
				groups = append(groups, models.Group{Name: "admins"})
			}
		}
		user := models.User{
			Username:      info.Username,
			DisplayedName: info.DisplayedName,
			Groups:        groups,
			Password:      info.Password,
		}
		err = userCtl.userModel.AdminSave(user, username)
		if err != nil {
			c.HTML(http.StatusBadRequest, "user.html", gin.H{
				"selTab":         "users",
				"loggedUserName": GetLoggedName(c),
				"errorMessage":   err.Error(),
			})
			c.Abort()
			return
		}
		res := userCtl.buildUserTemplateData(user, c)
		res["successMessage"] = "User has been updated"
		c.HTML(http.StatusOK, "user.html", res)
	}
}

func (userCtl *UserController) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		resData := gin.H{"loggedUserName": GetLoggedName(c), "selTab": "users"}
		loggedUsername, ok := c.Get("username")
		var err error
		if ok && username != loggedUsername {
			err = userCtl.userModel.Delete(username)
		}
		if !ok || username == loggedUsername || err != nil {
			resData["errorMessage"] = fmt.Sprintf("Could note delete user '%s'.", username)
			c.HTML(http.StatusBadRequest, "user.html", resData)
			c.Abort()
			return
		}
		users, err := userCtl.userModel.All()
		if err != nil {
			resData["errorMessage"] = "Unable to retrieve users"
			c.HTML(http.StatusInternalServerError, "users.html", resData)
			c.Abort()
			return
		}
		resData = userCtl.buildUsersTemplateData(users, c)
		resData["successMessage"] = "User has been deleted"
		c.HTML(http.StatusOK, "users.html", resData)
	}
}

// Get user data
func (ctl *UserController) buildUserTemplateData(user models.User, c *gin.Context) map[string]interface{} {
	res, _ := ctl.userModel.AsMap(user)
	res["selTab"] = "users"
	res["loggedUserName"] = GetLoggedName(c)
	return res
}

// Get all users data
func (ctl *UserController) buildUsersTemplateData(users []models.User, c *gin.Context) map[string]interface{} {
	usersData, _ := ctl.userModel.AsMapSlice(users)
	return map[string]interface{}{
		"selTab":         "users",
		"loggedUserName": GetLoggedName(c),
		"users":          usersData,
	}
}
