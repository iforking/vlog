package vlog

import (
	"sync/atomic"
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
	level     *atomic.Value //Level
	appender  *atomic.Value //*Appender
	formatter *atomic.Value //*Transformer
}

// the name of this logger
func (l *Logger) Name() string {
	return l.name
}

// set new Level to this logger. the default log level is DEBUG
func (l *Logger) SetLevel(level Level) {
	l.level.Store(level)
}

func (l *Logger) loadLevel() Level {
	return l.level.Load().(Level)
}

// set appender for this logger
func (l *Logger) SetAppender(appender Appender) {
	l.appender.Store(&appender)
}

func (l *Logger) loadAppender() Appender {
	return *l.appender.Load().(*Appender)
}

// set format for this logger
func (l *Logger) SetFormatter(formatter Transformer) {
	l.formatter.Store(&formatter)
}

func (l *Logger) loadFormatter() Transformer {
	return *l.formatter.Load().(*Transformer)
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

// if this logger log trace message
func (l *Logger) IsTraceEnable() bool {
	return l.loadLevel() <= TRACE
}

// if this logger log debug message
func (l *Logger) IsDebugEnable() bool {
	return l.loadLevel() <= DEBUG
}

// if this logger log info message
func (l *Logger) IsInfoEnable() bool {
	return l.loadLevel() <= INFO
}

// if this logger log warn level message
func (l *Logger) IsWarnEnable() bool {
	return l.loadLevel() <= WARN
}

// if this logger log error message
func (l *Logger) IsErrorEnable() bool {
	return l.loadLevel() <= ERROR
}

// if this logger log critical message
func (l *Logger) IsCriticalEnable() bool {
	return l.loadLevel() <= CRITICAL
}

func (l *Logger) log(level Level, message string, args ...interface{}) {
	if l.loadLevel() <= level {
		str := l.loadFormatter().Transform(l.Name(), level, message, args...)
		_, err := l.loadAppender().Append(str)
		if err != nil {
			//what we can do?
		}
	}
}

func createLogger(name string) *Logger {
	logger := &Logger{
		name:      name,
		level:     &atomic.Value{},
		appender:  &atomic.Value{},
		formatter: &atomic.Value{},
	}
	logger.SetLevel(DEFAULT_LEVEL)
	logger.SetAppender(NewConsoleAppender())
	logger.SetFormatter(NewDefaultPatternTransformer())
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
