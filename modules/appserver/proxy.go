package appserver

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *AppServer) CreateProxy() gin.HandlerFunc {

	director := func(req *http.Request) {}
	proxy := &httputil.ReverseProxy{Director: director}

	return func(c *gin.Context) {

		session, err := s.GetSession(c.Request)
		if err != nil {
			c.HTML(404, "appnotfound.html", nil)
			return
		}
		port := session.Instance.Port
		origin, _ := url.Parse("http://localhost:" + port)

		if username, ok := c.Get("username"); ok {
			c.Request.Header.Set("go-shiny-username", username.(string))
		}
		if displayedname, ok := c.Get("displayedname"); ok {
			c.Request.Header.Set("go-shiny-displayedname", displayedname.(string))
		}

		c.Request.URL.Scheme = "http"
		c.Request.URL.Host = origin.Host
		if session.App.ShinyApp.Path != "/" {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, session.App.ShinyApp.Path, "", -1)
		}
		cookieApp := http.Cookie{
			Name:  "go_shiny_appid",
			Value: session.App.ShinyApp.AppName,
			Path:  "/",
		}
		http.SetCookie(c.Writer, &cookieApp)
		cookieSess := http.Cookie{
			Name:  "go_shiny_session",
			Value: session.ID,
			Path:  "/",
		}
		http.SetCookie(c.Writer, &cookieSess)
		modifyResponse := func(res *http.Response) error {
			if res.StatusCode == 404 || res.StatusCode == 500 {
				return errors.New("Error from server")
			}
			return nil
		}
		errorHandler := func(res http.ResponseWriter, req *http.Request, err error) {
			c.HTML(404, "appnotfound.html", gin.H{"err": err.Error()})
		}
		proxy.ModifyResponse = modifyResponse
		proxy.ErrorHandler = errorHandler

		ws := c.Request.Header.Get("Upgrade") == "websocket"
		proxy.ServeHTTP(c.Writer, c.Request)
		if ws {
			session.Close()
		}
	}
}
