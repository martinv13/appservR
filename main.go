package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/martinv13/go-shiny/internal/shinyapp"

	"github.com/gin-gonic/gin"
)

func main() {

	shinyapp.RunShinyApp()

	origin, _ := url.Parse("http://localhost:3053/")

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	proxy := &httputil.ReverseProxy{Director: director}

	r := gin.Default()

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")

	r.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.tpl", gin.H{
			"title": "Main website",
		})
	})

	r.Use(func(c *gin.Context) {
		cookie := http.Cookie{
			Name:  "SHINYPROXY_APP",
			Value: "test",
		}
		http.SetCookie(c.Writer, &cookie)
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	r.Run()

}
