package appproxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateProxy() gin.HandlerFunc {

	director := func(req *http.Request) {}
	proxy := &httputil.ReverseProxy{Director: director}

	errorHandler := func(res http.ResponseWriter, req *http.Request, err error) {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte("App not found"))
	}

	return func(c *gin.Context) {
		app, err := MatchApp(c.Request)
		if err != nil {
			c.HTML(404, "appnotfound.tpl", nil)
			c.Error(errors.New("App not found"))
			c.Abort()
			return
		}
		port, err := app.GetPort()
		if err != nil {
			return
		}
		origin, _ := url.Parse("http://localhost:" + port)
		c.Request.Header.Add("X-Forwarded-Host", c.Request.Host)
		c.Request.Header.Add("X-Origin-Host", origin.Host)
		c.Request.URL.Scheme = "http"
		c.Request.URL.Host = origin.Host
		if app.ShinyApp.Path != "/" {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, app.ShinyApp.Path, "", -1)
		}
		cookie := http.Cookie{
			Name:  "GO_SHINY_APP_ID",
			Value: app.ShinyApp.ID,
		}
		http.SetCookie(c.Writer, &cookie)
		proxy.ErrorHandler = errorHandler
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
