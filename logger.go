package configwatchd

import "log"

var (
	ldebug *log.Logger
	lerror *log.Logger
)

func SetLoggers(loggers struct {
	Debug *log.Logger
	Error *log.Logger
}) {
	ldebug = loggers.Debug
	lerror = loggers.Error
}
