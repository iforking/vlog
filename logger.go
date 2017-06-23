package vlog

import (
	"sync/atomic"
	"unsafe"
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
	level     atomic.Value   //Level
	appenders unsafe.Pointer //*[]Appender
}

// the name of this logger
func (l *Logger) Name() string {
	return l.name
}

// set new Level to this logger. the default log level is DEBUG
func (l *Logger) SetLevel(level Level) {
	l.level.Store(level)
}

// current level of this logger
func (l *Logger) Level() Level {
	iface := l.level.Load()
	return iface.(Level)
}

// Set appender for this logger
func (l *Logger) SetAppenders(appender []Appender) {
	atomic.StorePointer(&l.appenders, unsafe.Pointer(&appender))
}

// The appenders this logger have
func (l *Logger) Appenders() []Appender {
	return *(*[]Appender)(atomic.LoadPointer(&l.appenders))
}

// Add one new appender to logger
func (l *Logger) AddAppender(appender Appender) {
	for {
		p := atomic.LoadPointer(&l.appenders)
		appenders := *(*[]Appender)(p)
		newAppenders := make([]Appender, len(appenders)+1)
		copy(newAppenders, appenders)
		newAppenders[len(appenders)] = appender
		if atomic.CompareAndSwapPointer(&l.appenders, p, unsafe.Pointer(&newAppenders)) {
			break
		}
	}
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
	return l.Level() <= TRACE
}

// if this logger log debug message
func (l *Logger) IsDebugEnable() bool {
	return l.Level() <= DEBUG
}

// if this logger log info message
func (l *Logger) IsInfoEnable() bool {
	return l.Level() <= INFO
}

// if this logger log warn level message
func (l *Logger) IsWarnEnable() bool {
	return l.Level() <= WARN
}

// if this logger log error message
func (l *Logger) IsErrorEnable() bool {
	return l.Level() <= ERROR
}

// if this logger log critical message
func (l *Logger) IsCriticalEnable() bool {
	return l.Level() <= CRITICAL
}

func (l *Logger) log(level Level, message string, args ...interface{}) {
	if l.Level() <= level {
		for _, appender := range l.Appenders() {
			transformer := appender.Transformer()
			data := transformer.Transform(l.Name(), level, message, args)
			_, err := appender.Append(data)
			if err != nil {
				//what we can do?
			}
		}
	}
}

// create new logger, with name and
func GetLogger(name string) *Logger {
	return logCache.Load(name)
}

// return the log of current package, use package name as logger name
func CurrentPackageLogger() *Logger {
	caller := getCaller(2)
	return GetLogger(caller.packageName)
}
