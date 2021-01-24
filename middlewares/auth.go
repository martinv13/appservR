package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/services/auth"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Request.Cookie("token")
		if err == nil {
			fmt.Println("token cookie found")
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
				fmt.Println(claims)
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
			c.Redirect(http.StatusFound, "/login")
		}
		c.Next()
	}
}
