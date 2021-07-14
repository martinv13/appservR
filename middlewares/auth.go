package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/appservR/appservR/modules/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")
		if err == nil {
			token, err := auth.ValidateToken(token.Value)
			if err == nil && token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				c.Set("username", fmt.Sprintf("%s", claims["username"]))
				c.Set("displayedname", fmt.Sprintf("%s", claims["name"]))
				gs := strings.Split(fmt.Sprintf("%s", claims["groups"]), ",")
				groups := map[string]bool{}
				for i := range gs {
					groups[gs[i]] = true
				}
				c.Set("groups", groups)
			}
		}
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorized := false
		groups, ok := c.Get("groups")
		if ok {
			groupsMap, ok := groups.(map[string]bool)
			if ok {
				val, ok := groupsMap["admins"]
				authorized = ok && val
			}
		}
		if !authorized {
			c.HTML(http.StatusNotFound, "appnotfound.html", gin.H{})
			c.Abort()
		} else {
			c.Next()
		}
	}
}
