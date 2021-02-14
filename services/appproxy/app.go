package appproxy

import (
	"errors"
	"sync"
	"time"

	"github.com/martinv13/go-shiny/models"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID         string
	startedAt  int64
	lastActive int64
	App        *AppProxy
	Instance   *shinyAppInstance
}

type AppProxy struct {
	ShinyApp  *models.ShinyApp
	Instances map[int]*shinyAppInstance
	nextID    int
	mu        sync.Mutex
	Sessions  map[string]*Session
}

func NewAppProxy(app *models.ShinyApp) *AppProxy {
	return &AppProxy{
		ShinyApp:  app,
		nextID:    0,
		Sessions:  map[string]*Session{},
		Instances: map[int]*shinyAppInstance{},
	}
}

// Start initializes and starts app subprocesses instances
func (appProxy *AppProxy) Start() error {
	appProxy.mu.Lock()
	for w := 0; w < appProxy.ShinyApp.Workers; w++ {
		inst, err := SpawnApp(appProxy.ShinyApp.AppName, appProxy.ShinyApp.AppDir)
		if err != nil {
			return err
		}
		appProxy.Instances[appProxy.nextID] = inst
		appProxy.nextID++
	}
	appProxy.mu.Unlock()
	return nil
}

// GetSession returns an existing session or a new session and selects
// the most appropriate running instance according to current load
func (appProxy *AppProxy) GetSession(sessionID string) (*Session, error) {

	session, ok := appProxy.Sessions[sessionID]

	if ok {
		if session.Instance.State == "RUNNING" {
			session.lastActive = time.Now().Unix()
			return session, nil
		}
		if len(appProxy.Instances) > 0 {
			for _, inst := range appProxy.Instances {
				if inst.State == "RUNNING" {
					session.Instance = inst
					session.lastActive = time.Now().Unix()
					return session, nil
				}
			}
		}
	}

	if len(appProxy.Instances) > 0 {
		for _, inst := range appProxy.Instances {
			if inst.State == "RUNNING" {
				now := time.Now().Unix()
				session = &Session{
					ID:         uuid.NewV4().String(),
					startedAt:  now,
					lastActive: now,
					App:        appProxy,
					Instance:   inst,
				}
				appProxy.Sessions[session.ID] = session
				return session, nil
			}
		}
	}
	return nil, errors.New("No running instance available")
}

func (sess *Session) Close() {
	app := sess.App
	delete(app.Sessions, sess.ID)
}
