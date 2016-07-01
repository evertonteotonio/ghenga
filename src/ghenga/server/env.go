package server

import "ghenga/db"

// Env is an environment for a handler function.
type Env struct {
	DB  db.DB
	Cfg Config

	Logger struct {
		Debug Logger
		Error Logger
	}
}

// Logger logs messages.
type Logger interface {
	Printf(format string, args ...interface{})
}

// Debugf logs a debug message.
func (e Env) Debugf(format string, args ...interface{}) {
	if e.Logger.Debug == nil {
		return
	}

	e.Logger.Debug.Printf("[debug] "+format, args...)
}

// Logf an infomational message.
func (e Env) Logf(format string, args ...interface{}) {
	if e.Logger.Error == nil {
		return
	}

	e.Logger.Error.Printf(format, args...)
}
