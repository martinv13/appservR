package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/services/auth"
	"gorm.io/gorm"
)

func GetLoggedName(c *gin.Context) string {
	name := "unknown"
	nameVal, ok := c.Get("displayedname")
	if ok {
		name = fmt.Sprintf("%s", nameVal)
	}
	return name
}

type loginCredentials struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

func DoLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials loginCredentials
		err := c.ShouldBind(&credentials)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.html",
				gin.H{"errorMessage": "Login failed. Please check your username and password."})
		}
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		user := models.User{Username: credentials.Username, Password: credentials.Password}
		err = user.Login(db)
		if err == nil {
			token := auth.GenerateToken(user)
			cookie := http.Cookie{
				Name:  "token",
				Value: token,
			}
			http.SetCookie(c.Writer, &cookie)
			c.Redirect(http.StatusFound, "/admin/settings")
		} else {
			c.HTML(http.StatusUnauthorized, "login.html",
				gin.H{"errorMessage": "Login failed. Please check your username and password."})
		}
	}
}

func DoLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie := http.Cookie{
			Name:  "token",
			Value: "",
		}
		http.SetCookie(c.Writer, &cookie)
		c.Redirect(http.StatusFound, "/")
	}
}

type signupInfo struct {
	Username      string `form:"username"`
	DisplayedName string `form:"displayedname"`
	Password      string `form:"password"`
	Password2     string `form:"password2"`
}

func DoSignup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var info signupInfo
		err := c.ShouldBind(&info)
		if err != nil {
			c.HTML(http.StatusBadRequest, "signup.html",
				gin.H{"errorMessage": "Signup failed. Please check the info provided."})
			c.Abort()
			return
		}
		if info.Password != info.Password2 {
			c.HTML(http.StatusBadRequest, "signup.html",
				gin.H{"errorMessage": "Signup failed. Passwords do not match."})
			c.Abort()
			return
		}
		dbi, _ := c.Get("DB")
		db := dbi.(*gorm.DB)
		user := models.User{
			Username:      info.Username,
			DisplayedName: info.DisplayedName,
			Password:      info.Password,
		}
		err = user.Create(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "signup.html",
				gin.H{"errorMessage": "Signup failed. Could not create user."})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "signupsuccess.html", gin.H{})
	}
}
