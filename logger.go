package vlog

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// Level the logger level
type Level int32

var levelNames = map[Level]string{
	Trace:    "Trace",
	Debug:    "Debug",
	Info:     "Info",
	Warn:     "Warn",
	Error:    "Error",
	Critical: "Critical",
}

// Name return the name of level, using captical form
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

// Logger the logger
type Logger struct {
	name      string
	level     int32          //Level
	appenders unsafe.Pointer //*[]Appender
	frozen    bool           // frozen level. the level is set by env, following level set in code will not take effect
}

// Name the name of this logger
func (l *Logger) Name() string {
	return l.name
}

// SetLevel set new Level to this logger. the default log level is Debug
func (l *Logger) SetLevel(level Level) {
	if l.frozen {
		return
	}
	atomic.StoreInt32(&l.level, int32(level))
}

// Level current level of this logger
func (l *Logger) Level() Level {
	return Level(atomic.LoadInt32(&l.level))
}

// SetAppenders set one or multi appenders for this logger
func (l *Logger) SetAppenders(appenders ...Appender) {
	atomic.StorePointer(&l.appenders, unsafe.Pointer(&appenders))
}

// Appenders return the appenders this logger have
func (l *Logger) Appenders() []Appender {
	return *(*[]Appender)(atomic.LoadPointer(&l.appenders))
}

// AddAppender add one new appender to logger
func (l *Logger) AddAppenders(appenders ...Appender) {
	if len(appenders) == 0 {
		return
	}

	for {
		p := atomic.LoadPointer(&l.appenders)
		originAppenders := *(*[]Appender)(p)
		newAppenders := make([]Appender, len(originAppenders)+len(appenders))
		copy(newAppenders, originAppenders)
		copy(newAppenders[len(originAppenders):], appenders)
		if atomic.CompareAndSwapPointer(&l.appenders, p, unsafe.Pointer(&newAppenders)) {
			break
		}
	}
}

// SetTransformerForAppenders set transformer, apply to all appenders the logger current have
func (l *Logger) SetTransformerForAppenders(transformer Transformer) {
	for _, appender := range l.Appenders() {
		appender.SetTransformer(transformer)
	}
}

// Trace log message with trace level
func (l *Logger) Trace(message string, args ...interface{}) error {
	return l.log(Trace, message, args...)
}

// Debug log message with debug level
func (l *Logger) Debug(message string, args ...interface{}) error {
	return l.log(Debug, message, args...)
}

// Info log message with info level
func (l *Logger) Info(message string, args ...interface{}) error {
	return l.log(Info, message, args...)
}

// Warn log message with warn level
func (l *Logger) Warn(message string, args ...interface{}) error {
	return l.log(Warn, message, args...)
}

// log message with error level
func (l *Logger) Error(message string, args ...interface{}) error {
	return l.log(Error, message, args...)
}

// Critical log message with critical level
func (l *Logger) Critical(message string, args ...interface{}) error {
	return l.log(Critical, message, args...)
}

// TraceEnabled if this logger log trace message
func (l *Logger) TraceEnabled() bool {
	return l.Level() <= Trace
}

// DebugEnabled if this logger log debug message
func (l *Logger) DebugEnabled() bool {
	return l.Level() <= Debug
}

// InfoEnabled if this logger log info message
func (l *Logger) InfoEnabled() bool {
	return l.Level() <= Info
}

// WarnEnabled if this logger log warn level message
func (l *Logger) WarnEnabled() bool {
	return l.Level() <= Warn
}

// ErrorEnabled if this logger log error message
func (l *Logger) ErrorEnabled() bool {
	return l.Level() <= Error
}

// CriticalEnabled if this logger log critical message
func (l *Logger) CriticalEnabled() bool {
	return l.Level() <= Critical
}

// Trace log message with trace level
func (l *Logger) LazyTrace(f func() string) error {
	if !l.TraceEnabled() {
		return nil
	}
	return l.Trace(f())
}

// Debug log message with debug level
func (l *Logger) LazyDebug(f func() string) error {
	if !l.DebugEnabled() {
		return nil
	}
	return l.Debug(f())
}

// Info log message with info level
func (l *Logger) LazyInfo(f func() string) error {
	if !l.InfoEnabled() {
		return nil
	}
	return l.Info(f())
}

// Warn log message with warn level
func (l *Logger) LazyWarn(f func() string) error {
	if !l.WarnEnabled() {
		return nil
	}
	return l.Warn(f())
}

// log message with error level
func (l *Logger) LazyError(f func() string) error {
	if !l.ErrorEnabled() {
		return nil
	}
	return l.Error(f())
}

// Critical log message with critical level
func (l *Logger) LazyCritical(f func() string) error {
	if !l.CriticalEnabled() {
		return nil
	}
	return l.Critical(f())
}

func (l *Logger) log(level Level, message string, args ...interface{}) error {
	appenders := l.Appenders()
	if l.Level() <= level && len(appenders) > 0 {
		now := time.Now()
		fMessage := formatMessage(message, args...)
		name := l.Name()
		//TODO: async, parallel write
		for _, appender := range appenders {
			transformer := appender.Transformer()
			data := transformer.Transform(name, level, now, fMessage)
			err := appender.Append(name, level, data)
			if err != nil {
				//TODO: collection errors
				return err
			}
		}
	}
	return nil
}

func formatMessage(message string, args ...interface{}) string {
	argNum := len(args)
	items := strings.SplitN(message, "{}", argNum+1)

	var results []string
	for idx, item := range items {
		results = append(results, item)
		if idx >= 0 && idx < len(items)-1 && idx < argNum {
			results = append(results, formatArg(args[idx]))
		}
	}

	for idx := len(items) - 1; idx < argNum; idx++ {
		results = append(results, " ")
		results = append(results, formatArg(args[idx]))
	}

	return strings.Join(results, "")
}

func formatArg(arg interface{}) string {
	return fmt.Sprintf("%v", arg)
}

// GetLogger return the logger with name
func GetLogger(name string) *Logger {
	return loggerCache.Load(name)
}

// CurrentPackageLogger return the log of current package, use package name as logger name
func CurrentPackageLogger() *Logger {
	caller := getCaller(2)
	return GetLogger(caller.packageName)
}
