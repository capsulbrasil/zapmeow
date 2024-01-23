package logger

import (
	"os"
	"zapmeow/config"

	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields

var log *logrus.Logger

func Init() {
	cfg := config.Load()
	log = logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	if cfg.Environment == config.Production {
		log.SetLevel(logrus.ErrorLevel)
	}

	log.SetOutput(os.Stdout)
}

func InfoWithFields(message string, fields Fields) {
	log.WithFields(fields).Info(message)
}

func DebugWithFields(message string, fields Fields) {
	log.WithFields(fields).Debug(message)
}

func ErrorWithFields(message string, fields Fields) {
	log.WithFields(fields).Error(message)
}

func FatalWithFields(message string, fields Fields) {
	log.WithFields(fields).Fatal(message)
}

func PanicWithFields(message string, fields Fields) {
	log.WithFields(fields).Panic(message)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}
