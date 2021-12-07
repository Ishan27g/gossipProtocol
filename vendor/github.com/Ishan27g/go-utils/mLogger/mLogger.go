package mLogger

import (
	"sync"

	"github.com/hashicorp/go-hclog"
)

const defaultLevel = "info"
var once sync.Once

// loggers added as Named
var loggers map[string]hclog.Logger

// top level logger
var logger hclog.Logger

func init() {
	once.Do(func() {
		logger = nil // asserts New is called once
		loggers = make(map[string]hclog.Logger)
	})
}

// New create a new top level logger with hclog.LevelFromString
// Subsequent modules should call Get
func New(name, lvl string) hclog.Logger {
	m := sync.Mutex{}
	if lvl == ""{
		lvl = defaultLevel
	}
	opts := hclog.LoggerOptions{
		Name:        "[" + name + "]",
		Level:       hclog.LevelFromString(lvl),
		Mutex:       &m,
		DisableTime: true,
		Color:       hclog.AutoColor,
	}
	logger = hclog.New(&opts)
	loggers[name] = hclog.New(&opts)
	return loggers[name]
}

// Get returns a named logger by either creating a sub logger or
// returning existing one. If no top level logger exists, the first call to Get
// creates a top level logger
func Get(name string) hclog.Logger {
	if logger == nil{
		return New(name, defaultLevel)
	}
	if loggers[name] == nil {
		loggers[name] = logger.Named(name)
	}
	return loggers[name]
}
