package appproxy

import (
	"encoding/json"
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
	sync.RWMutex
	ShinyApp     *models.ShinyApp
	StatusStream *ssehandler.Event
	Instances    map[int]*appInstance
	nextID       int
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
	appProxy.Lock()
	defer appProxy.Unlock()
	for w := 0; w < appProxy.ShinyApp.Workers; w++ {
		inst := &appInstance{App: appProxy.ShinyApp}
		err := inst.Start()
		if err != nil {
			return err
		}
		appProxy.Instances[appProxy.nextID] = inst
		appProxy.nextID++
	}
	return nil
}

// GetSession returns an existing session or a new session and selects
// the most appropriate running instance according to current load
func (appProxy *AppProxy) GetSession(sessionID string) (*Session, error) {
	appProxy.Lock()
	defer func() {
		appProxy.Unlock()
		appProxy.reportStatus()
	}()

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
	app.Lock()
	delete(app.Sessions, sess.ID)
	app.Unlock()
	app.reportStatus()
}

func (appProxy *AppProxy) reportStatus() {
	appProxy.RLock()
	defer appProxy.RUnlock()
	users := len(appProxy.Sessions)
	msg := ""
	if users == 0 {
		msg = "no connected user"
	} else if users == 1 {
		msg = "1 connected user"
	} else {
		msg = fmt.Sprintf("%d connected users", users)
	}
	msgData, _ := json.Marshal(map[string]string{
		"appName": appProxy.ShinyApp.AppName,
		"value":   msg,
	})
	appProxy.StatusStream.Message <- string(msgData)
}
