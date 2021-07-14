package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/auth"
	"github.com/gin-gonic/gin"
)

func GetLoggedName(c *gin.Context) string {
	name := "unknown"
	nameVal, ok := c.Get("displayedname")
	if ok {
		name = fmt.Sprintf("%s", nameVal)
	}
	return name
}

type AuthController struct {
	userModel models.UserModel
}

func NewAuthController(userModel models.UserModel) *AuthController {
	return &AuthController{
		userModel: userModel,
	}
}

type loginCredentials struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Referer  string `form:"refurl"`
}

func (ctl *AuthController) DoLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials loginCredentials
		err := c.ShouldBind(&credentials)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.html",
				gin.H{
					"errorMessage": "Login failed. Please check your username and password.",
					"Referer":      credentials.Referer,
				})
		}
		user := models.User{Username: credentials.Username, Password: credentials.Password}
		user, err = ctl.userModel.Login(user)
		if err == nil {
			token := auth.GenerateToken(user)
			cookie := http.Cookie{
				Name:     "token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(c.Writer, &cookie)
			ref := credentials.Referer
			if strings.HasSuffix(ref, "/auth/signup") {
				ref = "/"
			}
			c.Redirect(http.StatusFound, ref)
		} else {
			c.HTML(http.StatusUnauthorized, "login.html",
				gin.H{
					"errorMessage": "Login failed. Please check your username and password.",
					"Referer":      credentials.Referer,
				})
		}
	}
}

func (ctl *AuthController) DoLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie := http.Cookie{
			Name:  "token",
			Value: "",
			Path:  "/",
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

func (ctl *AuthController) DoSignup() gin.HandlerFunc {
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
		user := models.User{
			Username:      info.Username,
			DisplayedName: info.DisplayedName,
			Password:      info.Password,
		}
		err = ctl.userModel.Save(user, "new")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "signup.html",
				gin.H{"errorMessage": "Signup failed. Username already taken."})
			c.Abort()
			return
		}
		c.HTML(http.StatusOK, "signupsuccess.html", gin.H{})
	}
}
