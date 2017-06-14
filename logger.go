package verylog

import (
	"sync"
)

type Level uint32

var levelNames = map[Level]string{
	TRACE:    "TRACE",
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARN:     "WARN",
	ERROR:    "ERROR",
	CRITICAL: "CRITICAL",
}

func (l Level) Name() string {
	return levelNames[l]
}

// log levels
const (
	TRACE         Level = 10
	DEBUG         Level = 20
	INFO          Level = 30
	WARN          Level = 40
	ERROR         Level = 50
	CRITICAL      Level = 60
	OFF           Level = 70
	DEFAULT_LEVEL Level = INFO
)

type Logger struct {
	name      string
	level     Level
	appender  Appender
	formatter *Formatter
	lock      sync.RWMutex
}

// the name of this logger
func (l *Logger) Name() string {
	return l.name
}

// set new Level to this logger. the default log level is DEBUG
func (l *Logger) SetLevel(level Level) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

// set appender for this logger
func (l *Logger) SetAppender(appender Appender) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.appender = appender
}

// set format for this logger
// below vars can be used in format string:
// {file} filename
// {package} package name
// {line} line number
// {function} function name
// {time} time
// {logger} the logger name
// {message} the log message
// use {{ to escape  {, use }} to escape }
func (l *Logger) SetFormatter(format string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	formatter, err := NewFormatter(format)
	if err != nil {
		panic(err)
	}
	l.formatter = formatter
}

// log message with trace level
func (l *Logger) Trace(message string, args ...interface{}) {
	l.log(TRACE, message, args...)
}

// log message with debug level
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// log message with info level
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// log message with warn level
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// log message with error level
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// log message with critical level
func (l *Logger) Critical(message string, args ...interface{}) {
	l.log(CRITICAL, message, args...)
}

func (l *Logger) log(level Level, message string, args ...interface{}) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.level <= level {
		str := l.formatter.Format(l.Name(), level, message, args...)
		l.appender.Write(str)
	}
}

func createLogger(name string) *Logger {
	logger := &Logger{
		name:      name,
		level:     DEFAULT_LEVEL,
		appender:  NewConsoleAppender(),
		formatter: DefaultFormatter(),
	}
	return logger
}

// create new logger, with name and
func NewLogger(name string) *Logger {
	return logCache.load(name)
}

// return the log of current package, use package name as logger name
func CurrentPackageLogger() *Logger {
	caller := getCaller(2)
	return NewLogger(caller.packageName)
}
