package appserver

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID         string
	LastActive int64
	Instance   *Instance
	app        *AppProxy
}

func NewSession(app *AppProxy) *Session {
	sess := &Session{
		ID:         uuid.NewV4().String(),
		LastActive: time.Now().Unix(),
		app:        app,
	}
	return sess
}
