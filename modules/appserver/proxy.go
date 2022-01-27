package appserver

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Get the app for a specific request, based on request path and cookies, and check access right
func (appServer *AppServer) GetApp(c *gin.Context) (*AppProxy, bool, error) {
	appServer.RLock()
	defer appServer.RUnlock()
	r := c.Request
	reqURI, _ := url.Parse(r.RequestURI)
	reqPath := strings.TrimSuffix(reqURI.Path, "/")
	for _, app := range appServer.byPath {
		appPath := strings.TrimSuffix(app.App.Path, "/")
		if appPath == reqPath {
			// check user auth
			if !app.Authorized(c) {
				return nil, false, errors.New("unauthorized")
			}
			if reqURI.Path != reqPath+"/" {
				c.Redirect(http.StatusMovedPermanently, reqPath+"/")
				c.Abort()
				return nil, false, nil
			}
			return app, true, nil
		}
	}
	appCookie, err := r.Cookie("appservr_appid")
	if err == nil {
		if app, ok := appServer.appsByName[appCookie.Value]; ok {
			return app, false, nil
		}
	}
	return nil, false, errors.New("no matching app found")
}

// Create a proxy handler
func (s *AppServer) CreateProxy() gin.HandlerFunc {

	director := func(req *http.Request) {}
	proxy := &httputil.ReverseProxy{Director: director}
	logger := s.config.Logger()

	abortWithError := func(c *gin.Context, err error) {
		if err != nil {
			logger.Debug(err.Error())
		}
		c.HTML(http.StatusNotFound, "appnotfound.html", nil)
		c.Abort()
	}

	return func(c *gin.Context) {
		// Find matching app and check auth
		app, root, err := s.GetApp(c)
		if err != nil {
			abortWithError(c, err)
			return
		}
		// Case for redirects
		if app == nil {
			return
		}
		// Is current reqest a websocket upgrade?
		var ws = c.Request.Header.Get("Upgrade") == "websocket"
		// Find matching session or start new session
		var sess *Session
		sessCookie, err := c.Request.Cookie("appservr_session")
		var sessNotFound error
		if err == nil {
			sess, sessNotFound = app.GetSession(sessCookie.Value, ws)
		}
		if root || sessNotFound != nil {
			sess, _ = app.GetSession("", ws)
		}
		if sess == nil {
			abortWithError(c, err)
			return
		}
		sessID := sess.ID
		inst := sess.Instance
		origin, _ := url.Parse("http://localhost:" + sess.Instance.Port())

		if username, ok := c.Get("username"); ok {
			c.Request.Header.Set("appservR-username", username.(string))
		}
		if displayedname, ok := c.Get("displayedname"); ok {
			c.Request.Header.Set("appservR-displayedname", displayedname.(string))
		}
		c.Request.Header.Set("appservR-appname", app.App.Name)

		c.Request.URL.Scheme = "http"
		c.Request.URL.Host = origin.Host
		if app.App.Path != "/" {
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, app.App.Path, "", -1)
		}
		cookieApp := http.Cookie{
			Name:  "appservr_appid",
			Value: app.App.Name,
			Path:  "/",
		}
		http.SetCookie(c.Writer, &cookieApp)
		cookieSess := http.Cookie{
			Name:  "appservr_session",
			Value: sessID,
			Path:  "/",
		}
		http.SetCookie(c.Writer, &cookieSess)
		modifyResponse := func(res *http.Response) error {
			if res.StatusCode == 404 || res.StatusCode == 500 {
				return errors.New("error from server")
			}
			return nil
		}
		errorHandler := func(res http.ResponseWriter, req *http.Request, err error) {
			c.HTML(404, "appnotfound.html", gin.H{"err": err.Error()})
		}
		proxy.ModifyResponse = modifyResponse
		proxy.ErrorHandler = errorHandler

		proxy.ServeHTTP(c.Writer, c.Request)
		// In case of websocket connection, close session when socket is disconnected
		if ws {
			app.CloseSession(sessID)
			inst.SetUserCount(-1, true)
		}
	}
}
