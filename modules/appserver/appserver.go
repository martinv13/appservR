package appserver

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/ssehandler"
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
	appServer := &AppServer{
		broker:     msgBroker,
		appsByName: make(map[string]*AppProxy),
		config:     config,
	}
	apps, err := appModel.All()
	if err != nil {
		return nil, err
	}
	appServer.byPath = make([]*AppProxy, len(apps), len(apps))
	for i := range apps {
		app, err := NewAppProxy(apps[i], msgBroker, config)
		if err != nil {
			return nil, err
		}
		appServer.appsByName[apps[i].AppName] = app
		appServer.byPath[i] = app
	}
	sort.SliceStable(appServer.byPath, appServer.prefixSort)
	return appServer, nil
}

// Apply app settings changes
func (s *AppServer) Update(appName string, app models.ShinyApp) error {
	s.Lock()
	defer s.Unlock()
	appProxy, ok := s.appsByName[appName]
	if !ok {
		// new app
		appProxy, err := NewAppProxy(app, s.broker, s.config)
		if err != nil {
			return err
		}
		s.appsByName[app.AppName] = appProxy
		s.byPath = append(s.byPath, appProxy)
		sort.SliceStable(s.byPath, s.prefixSort)
	} else {
		// updated app
		prevApp := appProxy.ShinyApp
		appProxy.Update(app)
		if app.AppName != prevApp.AppName {
			delete(s.appsByName, prevApp.AppName)
			s.appsByName[app.AppName] = appProxy
		}
		if app.Path != prevApp.Path {
			sort.SliceStable(s.byPath, s.prefixSort)
		}
	}
	return nil
}

// Delete an app proxy object from the running apps
func (s *AppServer) Delete(appName string) error {
	s.Lock()
	defer s.Unlock()
	appProxy, ok := s.appsByName[appName]
	i := findAppProxy(s.byPath, appName)
	if !ok || i >= len(s.byPath) {
		return errors.New("App does not exist")
	}
	appProxy.Cleanup()
	delete(s.appsByName, appName)
	s.byPath[i] = s.byPath[len(s.byPath)-1]
	s.byPath = s.byPath[:len(s.byPath)-1]
	sort.SliceStable(s.byPath, s.prefixSort)
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
