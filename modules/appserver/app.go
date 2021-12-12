package appserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/appservR/appservR/models"
	"github.com/appservR/appservR/modules/appsource"
	"github.com/appservR/appservR/modules/config"
	"github.com/appservR/appservR/modules/ssehandler"
)

type AppProxy struct {
	sync.RWMutex
	RApp            models.RApp
	AppSource       appsource.AppSource
	StatusStream    *ssehandler.MessageBroker
	Instances       map[string]*Instance
	Sessions        map[string]*Session
	SessionsGCTimer *time.Timer
	config          config.Config
}

// Create a new app proxy
func NewAppProxy(app models.RApp, msgBroker *ssehandler.MessageBroker, config config.Config) (*AppProxy, error) {
	p := &AppProxy{
		RApp:            app,
		AppSource:       appsource.NewAppSource(app, config),
		StatusStream:    msgBroker,
		Instances:       map[string]*Instance{},
		Sessions:        map[string]*Session{},
		SessionsGCTimer: time.NewTimer(time.Second * time.Duration(30)),
		config:          config,
	}
	go p.Rescale()
	// Delete unused sessions every 30s
	go func() {
		<-p.SessionsGCTimer.C
		p.Lock()
		defer p.Unlock()
		anyExpired := false
		for id, sess := range p.Sessions {
			expired := sess.LastActive < time.Now().Unix()-30*60
			if expired {
				p.doCloseSession(id)
			}
			anyExpired = anyExpired || expired
		}
		if anyExpired {
			go p.Rescale()
		}
	}()
	return p, nil
}

// Cleanup before deleting app
func (p *AppProxy) Cleanup() {
	p.Lock()
	defer p.Unlock()
	// Stop all instances
	for _, inst := range p.Instances {
		inst.Stop()
	}
	// Stop sessions cleanup timer
	p.SessionsGCTimer.Stop()
}

// Find an existing session or create a new session and selects
// the most appropriate running instance according to current load
func (p *AppProxy) GetSession(sessionID string) (*Session, error) {
	p.Lock()
	defer func() {
		p.Unlock()
		go p.ReportStatus()
	}()

	sess, ok := p.Sessions[sessionID]
	if !ok {
		sess = NewSession(p)
	}

	// if session already exist and is still valid
	if sess.Instance != nil {
		if sess.Instance.Status() == instStatus.RUNNING {
			sess.LastActive = time.Now().Unix()
			return sess, nil
		} else {
			sess.Instance.SetUserCount(-1, true)
		}
	}

	// else, simple choice strategy: lowest user count of all running instances
	bestInstID := ""
	bestUC := 0
	for id, inst := range p.Instances {
		uc := inst.UserCount()
		if inst.Status() == instStatus.RUNNING && (bestInstID == "" || uc < bestUC) {
			bestInstID = id
			bestUC = uc
		}
	}
	if bestInstID != "" {
		sess.Instance = p.Instances[bestInstID]
		p.Sessions[sess.ID] = sess
		sess.Instance.SetUserCount(1, true)
		return sess, nil
	}
	return nil, errors.New("no running instance available")
}

// Rescale to appropriate number of workers (for now, a fixed user-defined number of workers)
func (p *AppProxy) Rescale() {
	p.Lock()
	defer func() {
		p.Unlock()
		go p.ReportStatus()
	}()
	nbInst := 0
	for _, inst := range p.Instances {
		status := inst.Status()
		if status == instStatus.PHASING_OUT {
			if inst.UserCount() == 0 {
				err := inst.Stop()
				if err == nil {
					delete(p.Instances, inst.ID)
				}
			}
		} else {
			nbInst++
		}
	}
	targetWorkers := p.RApp.Workers
	if !p.RApp.Active {
		targetWorkers = 0
	}
	// if too few instances, start new ones
	for w := 0; w < targetWorkers-nbInst; w++ {
		inst := NewInstance(p.RApp.AppName, p.AppSource.Path(), p.config)
		inst.Start()
		p.Instances[inst.ID] = inst
	}
	// if too many instances, phase out the one with less users connected
	if nbInst > targetWorkers {
		insts := make([]*Instance, nbInst)
		i := 0
		for _, inst := range p.Instances {
			if inst.Status() != instStatus.PHASING_OUT {
				insts[i] = inst
				i++
			}
		}
		sort.Slice(insts, func(i int, j int) bool {
			return insts[i].UserCount() < insts[j].UserCount()
		})
		for i = 0; i < nbInst-targetWorkers; i++ {
			insts[i].PhaseOut()
		}
	}
}

// End a specific session without lock and without rescaling
func (p *AppProxy) doCloseSession(sessionID string) error {
	if sess, ok := p.Sessions[sessionID]; ok {
		sess.Instance.SetUserCount(-1, true)
		delete(p.Sessions, sessionID)
		return nil
	}
	return errors.New("cannot find session")
}

// End a specific session
func (p *AppProxy) CloseSession(sessionID string) error {
	p.Lock()
	defer p.Unlock()
	err := p.doCloseSession(sessionID)
	go p.Rescale()
	return err
}

// Stop or restart all instances, while keeping existing connections
func (p *AppProxy) phaseOut() {
	for _, i := range p.Instances {
		i.PhaseOut()
	}
	go p.Rescale()
}

// Apply changes to app settings
func (p *AppProxy) Update(app models.RApp) {
	p.Lock()
	defer p.Unlock()
	prevApp := p.RApp
	p.RApp = app
	if prevApp.AppDir != app.AppDir || prevApp.Active != app.Active {
		p.phaseOut()
	} else if prevApp.Workers != app.Workers {
		go p.Rescale()
	}
}

// Remove an instance which has been stopped
func (p *AppProxy) DeleteInstance(ID string) {
	p.Lock()
	defer p.Unlock()
	delete(p.Instances, ID)
}

// Return app status info as a map
func (app *AppProxy) GetStatus(detailed bool) map[string]interface{} {
	nbRunning := 0
	nbPhasingOut := 0
	stdErr := []string{}
	for _, i := range app.Instances {
		status := i.Status()
		if status == instStatus.RUNNING {
			nbRunning++
		} else if status == instStatus.PHASING_OUT {
			nbPhasingOut++
		}
		if detailed {
			stdErr = append(stdErr, i.StdErr())
		}
	}
	status := map[string]interface{}{
		"RunningInst":    nbRunning,
		"PhasingOutInst": nbPhasingOut,
		"ConnectedUsers": len(app.Sessions),
	}
	if detailed {
		status["StdErr"] = stdErr
	}
	return status
}

// Stream status update
func (p *AppProxy) ReportStatus() {
	p.RLock()
	defer p.RUnlock()
	users := len(p.Sessions)
	msg := ""
	if users == 0 {
		msg = "no connected user"
	} else if users == 1 {
		msg = "1 connected user"
	} else {
		msg = fmt.Sprintf("%d connected users", users)
	}
	msgData, _ := json.Marshal(map[string]string{
		"appName": p.RApp.AppName,
		"value":   msg,
	})
	p.StatusStream.Message <- string(msgData)
}
