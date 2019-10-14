package logger

import "errors"

// A global variable so that log functions can be directly accessed
var log Logger

//Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default log level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The sytem shutsdown after logging the message.
	Fatal = "fatal"
)

//InstanceZapLogger int
const (
	InstanceZapLogger int = iota
)

var (
	errInvalidLoggerInstance = errors.New("Invalid logger instance")
)

//Logger is our contract for the logger
type Logger interface {
	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	WithFields(keyValues Fields) Logger
}

// Configuration stores the config for the logger
// For some loggers there can only be one level across writers, for such the level of Console is picked by default
type Configuration struct {
	EnableConsole     bool
	ConsoleJSONFormat bool
	ConsoleLevel      string
	EnableFile        bool
	FileJSONFormat    bool
	FileLevel         string
	FileLocation      string
}

//NewLogger returns an instance of logger
func NewLogger(config Configuration, loggerInstance int) error {
	switch loggerInstance {
	case InstanceZapLogger:
		logger, err := newZapLogger(config)
		if err != nil {
			return err
		}
		log = logger
		return nil

	default:
		return errInvalidLoggerInstance
	}
}

//Debugf log debug message
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

//Infof log info message
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

//Warnf log warn message
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

//Errorf log error message
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

//Fatalf log fatal message
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

//Panicf log panic message
func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

//WithFields return logger with object parameters
func WithFields(keyValues Fields) Logger {
	return log.WithFields(keyValues)
}
