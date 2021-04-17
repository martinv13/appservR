package appproxy

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/martinv13/go-shiny/models"
	"github.com/martinv13/go-shiny/modules/ssehandler"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID         string
	startedAt  int64
	lastActive int64
	App        *AppProxy
	Instance   *appInstance
}

type AppProxy struct {
	ShinyApp     *models.ShinyApp
	StatusStream *ssehandler.Event
	Instances    map[int]*appInstance
	nextID       int
	mu           sync.Mutex
	Sessions     map[string]*Session
}

func NewAppProxy(app *models.ShinyApp, stream *ssehandler.Event) *AppProxy {
	return &AppProxy{
		ShinyApp:     app,
		StatusStream: stream,
		nextID:       0,
		Sessions:     map[string]*Session{},
		Instances:    map[int]*appInstance{},
	}
}

// Start initializes and starts app subprocesses instances
func (appProxy *AppProxy) Start() error {
	appProxy.mu.Lock()
	for w := 0; w < appProxy.ShinyApp.Workers; w++ {
		inst := &appInstance{App: appProxy.ShinyApp}
		err := inst.Start()
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

	defer appProxy.reportStatus()

	session, ok := appProxy.Sessions[sessionID]

	if ok {
		if session.Instance.Status == "RUNNING" {
			session.lastActive = time.Now().Unix()
			return session, nil
		}
		if len(appProxy.Instances) > 0 {
			for _, inst := range appProxy.Instances {
				if inst.Status == "RUNNING" {
					session.Instance = inst
					session.lastActive = time.Now().Unix()
					return session, nil
				}
			}
		}
	}

	if len(appProxy.Instances) > 0 {
		for _, inst := range appProxy.Instances {
			if inst.Status == "RUNNING" {
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
	app.reportStatus()
}

func (appProxy *AppProxy) reportStatus() {

	users := len(appProxy.Sessions)
	msg := ""
	if users == 0 {
		msg = "no connected user"
	} else if users == 1 {
		msg = "1 connected user"
	} else {
		msg = fmt.Sprintf("%d connected users", users)
	}
	appProxy.StatusStream.Message <- fmt.Sprintf("{\"appName\":\"%s\", \"value\": \"%s\"}",
		appProxy.ShinyApp.AppName, msg)
}
