package mLogger

import (
	"sync"

	"github.com/hashicorp/go-hclog"
)

type LoggerI interface{}
type loggerI struct{}
type Option func(l *LoggerI)

func Apply(options ...Option) LoggerI {
	newLogger := func() LoggerI {
		return &loggerI{}
	}
	l := newLogger()
	for _, option := range options {
		option(&l)
	}
	return l
}
func Level(level hclog.Level) Option {
	return func(l *LoggerI) {
		logLevel = level
	}
}
func Color(color bool) Option {
	return func(l *LoggerI) {
		if !color {
			colorOn = hclog.ColorOff
		}
	}
}

var once sync.Once
var logLevel = hclog.Trace
var colorOn = hclog.AutoColor

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
func New(name string) hclog.Logger {
	m := sync.Mutex{}

	opts := hclog.LoggerOptions{
		Name:        "[" + name + "]",
		Level:       logLevel,
		Mutex:       &m,
		DisableTime: true,
		Color:       colorOn,
	}
	logger = hclog.New(&opts)
	loggers[name] = hclog.New(&opts)
	return loggers[name]
}

// Get returns a named logger by either creating a sub logger or
// returning existing one. If no top level logger exists, the first call to Get
// creates a top level logger
func Get(name string) hclog.Logger {
	if logger == nil {
		return New(name)
	}
	if loggers[name] == nil {
		loggers[name] = New(name)
	}
	return loggers[name]
}
