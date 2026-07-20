package appserver

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         string
	LastActive int64
	Instance   *Instance
	app        *AppProxy
}

func NewSession(app *AppProxy) *Session {
	sess := &Session{
		ID:         uuid.New().String(),
		LastActive: time.Now().Unix(),
		app:        app,
	}
	return sess
}
