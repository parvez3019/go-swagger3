package logger

import log "github.com/sirupsen/logrus"

type Logger struct {
	debugMode bool
}

var loggerInstance *Logger

func init() {
	loggerInstance = &Logger{}
}

func SetDebugMode(debugMode bool) *Logger {
	loggerInstance.debugMode = debugMode
	return loggerInstance
}

func (l *Logger) Debug(v ...interface{}) {
	if l.debugMode {
		log.Debugln(v...)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.debugMode {
		log.Debugf(format, args...)
	}
}
