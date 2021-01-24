package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/services/auth"
)

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
		user := &models.UserData{}
		err = user.LoginUser(credentials.Username, credentials.Password)
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

	}
}
