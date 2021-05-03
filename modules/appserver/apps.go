package appserver

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/config"
	"github.com/martinv13/go-shiny/modules/ssehandler"
)

type AppServer struct {
	sync.RWMutex
	broker     *ssehandler.MessageBroker
	config     config.Config
	appsByName map[string]*AppProxy
	byPath     []*AppProxy
}

// Create a new struct to hold running app proxies
func NewAppServer(appModel models.AppModel, msgBroker *ssehandler.MessageBroker, config config.Config) (*AppServer, error) {
	appServer := AppServer{
		broker:     msgBroker,
		appsByName: make(map[string]*AppProxy),
		byPath:     []*AppProxy{},
		config:     config,
	}
	apps, err := appModel.All()
	if err != nil {
		return nil, err
	}
	for i := range apps {
		app, err := NewAppProxy(apps[i], msgBroker, config)
		appServer.appsByName[apps[i].AppName] = app
		if err != nil {
			return nil, err
		}
		appServer.byPath = append(appServer.byPath, app)
	}
	sort.SliceStable(appServer.byPath, appServer.prefixSort)
	return &appServer, nil
}

// Get session info for a specific request, based on request path and cookies
func (appServer *AppServer) GetSession(c *gin.Context) (*Session, error) {
	appServer.RLock()
	defer appServer.RUnlock()

	r := c.Request

	reqURI, _ := url.Parse(r.RequestURI)
	reqPath := strings.TrimSuffix(reqURI.Path, "/")
	for i := range appServer.byPath {
		appPath := strings.TrimSuffix(appServer.byPath[i].ShinyApp.Path, "/")
		if appPath == reqPath {
			if reqURI.Path != reqPath+"/" {
				c.Redirect(http.StatusMovedPermanently, reqPath+"/")
				c.Abort()
				return nil, nil
			}
			session, err := appServer.byPath[i].GetSession("")
			if err != nil {
				return nil, err
			}
			return session, nil
		}
	}
	appCookie, err := r.Cookie("go_shiny_appid")
	if err == nil {
		if app, ok := appServer.appsByName[appCookie.Value]; ok {
			sessCookie, _ := r.Cookie("go_shiny_session")
			session, err := app.GetSession(sessCookie.Value)
			if err != nil {
				return nil, err
			}
			return session, nil
		}
	}
	return nil, errors.New("No matching app found")
}

// Apply app settings changes
func (s *AppServer) Update(appName string, app models.ShinyApp) error {
	s.Lock()
	defer s.Unlock()

	appProxy, ok := s.appsByName[appName]
	// new app
	if !ok {
		appProxy, err := NewAppProxy(app, s.broker, s.config)
		s.appsByName[app.AppName] = appProxy
		if err != nil {
			return err
		}
		s.byPath = append(s.byPath, appProxy)
		sort.SliceStable(s.byPath, s.prefixSort)
	} else {
		prevApp := appProxy.ShinyApp
		appProxy.Update(app)
		if app.AppName != prevApp.AppName {
			delete(s.appsByName, prevApp.AppName)
			if app.AppName != "" {
				// case of a deleted app
				s.appsByName[app.AppName] = appProxy
			} else {
				i := findAppProxy(s.byPath, appName)
				if i < len(s.byPath) {
					s.byPath[i] = s.byPath[len(s.byPath)-1]
					s.byPath = s.byPath[:len(s.byPath)-1]
				}
			}
		}
		if app.Path != prevApp.Path || app.AppName == "" {
			sort.SliceStable(s.byPath, s.prefixSort)
		}
	}
	return nil
}

// Returns the status of all apps as a map indexed with app names
func (s *AppServer) GetAllStatus() map[string]interface{} {
	status := map[string]interface{}{}
	for n, app := range s.appsByName {
		status[n] = app.GetStatus(false)
	}
	return status
}

// Returns the status of a given app
func (s *AppServer) GetStatus(appName string) (map[string]interface{}, error) {
	app, ok := s.appsByName[appName]
	if !ok {
		return nil, errors.New("App not found")
	}
	return app.GetStatus(true), nil
}

func (s *AppServer) prefixSort(i, j int) bool {
	return !strings.HasPrefix(s.byPath[i].ShinyApp.Path,
		s.byPath[j].ShinyApp.Path)
}

func findAppProxy(p []*AppProxy, appName string) int {
	for i, a := range p {
		if a.ShinyApp.AppName == appName {
			return i
		}
	}
	return len(p)
}
