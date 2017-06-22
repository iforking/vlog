package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	logCache := &LoggerCache{loggerMap: make(map[string]*Logger)}
	logger1 := logCache.load("logger1")
	logger2 := logCache.load("logger2")
	logger3 := logCache.load("logger1")

	assert.Equal(t, logger1, logger3)
	assert.NotEqual(t, logger1, logger2)
}

func TestLoggerCache_Filter(t *testing.T) {
	logCache := &LoggerCache{loggerMap: make(map[string]*Logger)}
	_ = logCache.load("package0")
	_ = logCache.load("package1")
	_ = logCache.load("gopkg.in/package1")
	_ = logCache.load("github.com/user1/package1")
	_ = logCache.load("github.com/user2/package2")
	_ = logCache.load("github.com/user3/package3")

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
