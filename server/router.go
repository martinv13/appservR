package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/services/appproxy"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	origin, _ := url.Parse("http://localhost:3053/")

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
		app, _ := appproxy.MatchApp(req)
		port, err := app.GetPort()
		if err == nil {
			req.URL.Host = "localhost:" + port
		}
		fmt.Println(app)
	}

	proxy := &httputil.ReverseProxy{Director: director}

	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	router.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.tpl", gin.H{
			"title": "Main website",
		})
	})

	router.Use(func(c *gin.Context) {
		cookie := http.Cookie{
			Name:  "SHINYPROXY_APP",
			Value: "test",
		}
		http.SetCookie(c.Writer, &cookie)
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	return router

}
