package vlog

import (
	"sync/atomic"
	"unsafe"
	"strings"
	"time"
)

var loggerLocked int32 = 0

// unlock logger, so later modifications to loggers will take effect
func UnfreezeLoggerSetting() {
	atomic.StoreInt32(&loggerLocked, 0)
}

// lock logger, so all modifications to loggers will not take effect
func FreezeLoggerSetting() {
	atomic.StoreInt32(&loggerLocked, 1)
}

// if return true, all modifications to loggers will not take effect
func LoggerSettingFroze() bool {
	return atomic.LoadInt32(&loggerLocked) == 1
}

type Level int32

var levelNames = map[Level]string{
	Trace:    "Trace",
	Debug:    "Debug",
	Info:     "Info",
	Warn:     "Warn",
	Error:    "Error",
	Critical: "Critical",
}

var levelNamesReverse = reverseLevelNames(levelNames)

func reverseLevelNames(levelNames map[Level]string) map[string]Level {
	var m = map[string]Level{}
	for level, str := range levelNames {
		m[strings.ToUpper(str)] = level
	}
	return m
}

func (l Level) Name() string {
	return levelNames[l]
}

// log levels
const (
	Trace        Level = 10
	Debug        Level = 20
	Info         Level = 30
	Warn         Level = 40
	Error        Level = 50
	Critical     Level = 60
	Off          Level = 70
	DefaultLevel Level = Info
)

type Logger struct {
	name      string
	level     int32          //Level
	appenders unsafe.Pointer //*[]Appender
}

// the name of this logger
func (l *Logger) Name() string {
	return l.name
}

// set new Level to this logger. the default log level is Debug
func (l *Logger) SetLevel(level Level) {
	if LoggerSettingFroze() {
		return
	}
	atomic.StoreInt32(&l.level, int32(level))
}

// current level of this logger
func (l *Logger) Level() Level {
	return Level(atomic.LoadInt32(&l.level))
}

// Set appender for this logger
func (l *Logger) SetAppenders(appender []Appender) {
	if LoggerSettingFroze() {
		return
	}
	atomic.StorePointer(&l.appenders, unsafe.Pointer(&appender))
}

// The appenders this logger have
func (l *Logger) Appenders() []Appender {
	return *(*[]Appender)(atomic.LoadPointer(&l.appenders))
}

// Add one new appender to logger
func (l *Logger) AddAppender(appender Appender) {
	if LoggerSettingFroze() {
		return
	}
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
	l.log(Trace, message, args...)
}

// log message with debug level
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(Debug, message, args...)
}

// log message with info level
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(Info, message, args...)
}

// log message with warn level
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(Warn, message, args...)
}

// log message with error level
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(Error, message, args...)
}

// log message with critical level
func (l *Logger) Critical(message string, args ...interface{}) {
	l.log(Critical, message, args...)
}

// if this logger log trace message
func (l *Logger) IsTraceEnable() bool {
	return l.Level() <= Trace
}

// if this logger log debug message
func (l *Logger) IsDebugEnable() bool {
	return l.Level() <= Debug
}

// if this logger log info message
func (l *Logger) IsInfoEnable() bool {
	return l.Level() <= Info
}

// if this logger log warn level message
func (l *Logger) IsWarnEnable() bool {
	return l.Level() <= Warn
}

// if this logger log error message
func (l *Logger) IsErrorEnable() bool {
	return l.Level() <= Error
}

// if this logger log critical message
func (l *Logger) IsCriticalEnable() bool {
	return l.Level() <= Critical
}

func (l *Logger) log(level Level, message string, args ...interface{}) {
	if l.Level() <= level {
		now := time.Now()
		for _, appender := range l.Appenders() {
			transformer := appender.Transformer()
			data := transformer.Transform(l.Name(), level, now, message, args)
			_, err := appender.Append(data)
			if err != nil {
				//what we can do?
			}
		}
	}
}

// create new logger, with name and
func GetLogger(name string) *Logger {
	return loggerCache.Load(name)
}

// return the log of current package, use package name as logger name
func CurrentPackageLogger() *Logger {
	caller := getCaller(2)
	return GetLogger(caller.packageName)
}
