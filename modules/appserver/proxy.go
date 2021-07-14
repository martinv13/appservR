package appserver

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Get the app for a specific request, based on request path and cookies
func (appServer *AppServer) GetApp(c *gin.Context) (*AppProxy, bool, error) {
	appServer.RLock()
	defer appServer.RUnlock()
	r := c.Request
	reqURI, _ := url.Parse(r.RequestURI)
	reqPath := strings.TrimSuffix(reqURI.Path, "/")
	for _, app := range appServer.byPath {
		appPath := strings.TrimSuffix(app.ShinyApp.Path, "/")
		if appPath == reqPath {
			if reqURI.Path != reqPath+"/" {
				c.Redirect(http.StatusMovedPermanently, reqPath+"/")
				c.Abort()
				return nil, false, nil
			}
			return app, true, nil
		}
	}
	appCookie, err := r.Cookie("go_shiny_appid")
	if err == nil {
		if app, ok := appServer.appsByName[appCookie.Value]; ok {
			return app, false, nil
		}
	}
	return nil, false, errors.New("No matching app found")
}

// Create a proxy handler
func (s *AppServer) CreateProxy() gin.HandlerFunc {

	director := func(req *http.Request) {}
	proxy := &httputil.ReverseProxy{Director: director}

	return func(c *gin.Context) {
		// Find matching app
		app, root, err := s.GetApp(c)
		if err != nil {
			c.HTML(http.StatusNotFound, "appnotfound.html", nil)
			c.Abort()
			return
		}
		// Find matching session or start new session
		var sess *Session
		if root {
			sess, err = app.GetSession("")
		} else {
			sessCookie, err := c.Request.Cookie("go_shiny_session")
			if err == nil {
				sess, err = app.GetSession(sessCookie.Value)
			}
		}
		if sess == nil {
			c.HTML(http.StatusNotFound, "appnotfound.html", nil)
			c.Abort()
			return
		}
		origin, _ := url.Parse("http://localhost:" + sess.Instance.Port())

		if username, ok := c.Get("username"); ok {
			c.Request.Header.Set("appservR-username", username.(string))
		}
		if displayedname, ok := c.Get("displayedname"); ok {
			c.Request.Header.Set("appservR-displayedname", displayedname.(string))
		}
		c.Request.Header.Set("appservR-appname", app.ShinyApp.AppName)

		c.Request.URL.Scheme = "http"
		c.Request.URL.Host = origin.Host
		if app.ShinyApp.Path != "/" {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, app.ShinyApp.Path, "", -1)
		}
		cookieApp := http.Cookie{
			Name:  "go_shiny_appid",
			Value: app.ShinyApp.AppName,
			Path:  "/",
		}
		http.SetCookie(c.Writer, &cookieApp)
		cookieSess := http.Cookie{
			Name:  "go_shiny_session",
			Value: sess.ID,
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
		// In case of websocket connection, close session when socket is disconnected
		if ws {
			app.CloseSession(sess.ID)
		}
	}
}
