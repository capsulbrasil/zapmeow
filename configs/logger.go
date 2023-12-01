package configs

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	// Calls os.Exit(1) after logging
	Fatal(args ...interface{})
	// Calls panic() after logging
	Panic(args ...interface{})
}

type logger struct {
	log *logrus.Logger
}

func NewLogger() *logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	return &logger{
		log: log,
	}
}

func (l *logger) Trace(args ...interface{}) {
	l.log.Trace(args...)
}
func (l *logger) Debug(args ...interface{}) {
	l.log.Debug(args...)
}
func (l *logger) Info(args ...interface{}) {
	l.log.Info(args...)
}
func (l *logger) Warn(args ...interface{}) {
	l.log.Warn(args...)
}
func (l *logger) Error(args ...interface{}) {
	l.log.Error(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.log.Fatal(args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.log.Panic(args...)
}
