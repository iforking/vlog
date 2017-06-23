package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	logCache := newDefaultLogCache()
	logger1 := logCache.Load("logger1")
	logger2 := logCache.Load("logger2")
	logger3 := logCache.Load("logger1")

	assert.Equal(t, logger1, logger3)
	assert.NotEqual(t, logger1, logger2)
}

func TestLoggerCache_Filter(t *testing.T) {
	logCache := newDefaultLogCache()
	_ = logCache.Load("package0")
	_ = logCache.Load("package1")
	_ = logCache.Load("gopkg.in/package1")
	_ = logCache.Load("github.com/user1/package1")
	_ = logCache.Load("github.com/user2/package2")
	_ = logCache.Load("github.com/user3/package3")

	loggers := logCache.filter("")
	assert.Equal(t, 6, len(loggers))

	loggers = logCache.filter("github.com")
	assert.Equal(t, 3, len(loggers))

	loggers = logCache.filter("github.com/user1/package1")
	assert.Equal(t, 1, len(loggers))

	loggers = logCache.filter("gopkg.in/")
	assert.Equal(t, 1, len(loggers))
	assert.Equal(t, "gopkg.in/package1", loggers[0].Name())
}

func TestCache_SetPrefix(t *testing.T) {
	logCache := newDefaultLogCache()
	logger1 := logCache.Load("package1")
	logger2 := logCache.Load("gopkg.in/package1")
	logger3 := logCache.Load("github.com/user1/package1")

	logCache.SetPrefixLevel("", DEBUG)
	assert.Equal(t, DEBUG, logger1.Level())
	assert.Equal(t, DEBUG, logger2.Level())
	assert.Equal(t, DEBUG, logger3.Level())

	logCache.SetPrefixLevel("github.com", WARN)
	appender := NewConsole2Appender("")
	logCache.SetPrefixAppenders("github.com", []Appender{appender})
	assert.Equal(t, DEBUG, logger1.Level())
	assert.Equal(t, WARN, logger3.Level())
	assert.NotEqual(t, []Appender{appender}, logger1.Appenders())
	assert.Equal(t, []Appender{appender}, logger3.Appenders())
	logger4 := logCache.Load("github.com/user1/package2")
	assert.Equal(t, WARN, logger4.Level())
	assert.Equal(t, []Appender{appender}, logger4.Appenders())

	logCache.SetPrefixAppenders("", []Appender{appender})
	appender2 := NewConsoleAppender("")
	logCache.AddPrefixAppender("gopkg.in/", appender2)
	assert.Equal(t, []Appender{appender}, logger1.Appenders())
	assert.Equal(t, []Appender{appender, appender2}, logger2.Appenders())
	assert.Equal(t, []Appender{appender}, logger3.Appenders())
	logger5 := logCache.Load("gopkg.in/package2")
	logger6 := logCache.Load("github.com/user2/package2")
	assert.Equal(t, []Appender{appender, appender2}, logger5.Appenders())
	assert.Equal(t, []Appender{appender}, logger6.Appenders())
}
