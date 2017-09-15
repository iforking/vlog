package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
	"time"
)

func TestLogger(t *testing.T) {
	logger := CurrentPackageLogger()
	assert.Equal(t, "github.com/clearthesky/vlog", logger.Name())
	assert.Equal(t, DefaultLevel, logger.Level())

	appender := NewBytesAppender()
	logger.SetAppenders(appender)
	logger.Info("this is a test")
	assert.True(t, strings.HasSuffix(appender.buffer.String(),
		" [Info] github.com/clearthesky/vlog - this is a test\n"))

	appender = NewBytesAppender()
	transformer, _ := NewPatternTransformer("{time|2006-01-02} {package}/{file} - {message}\n")
	appender.SetTransformer(transformer)
	logger.SetAppenders(appender)
	logger.Info("this is a test")
	date := time.Now().Format("2006-01-02")
	assert.Equal(t, date+" github.com/clearthesky/vlog/logger_test.go - this is a test\n", appender.buffer.String())

	logger2 := CurrentPackageLogger()
	assert.Equal(t, logger, logger2)
}

func TestLoggerJudge(t *testing.T) {
	logger := CurrentPackageLogger()
	logger.SetLevel(Off)
	assert.False(t, logger.IsTraceEnable())
	assert.False(t, logger.IsErrorEnable())

	logger.SetLevel(Critical)
	assert.True(t, logger.IsCriticalEnable())
	assert.False(t, logger.IsErrorEnable())

	logger.SetLevel(Error)
	assert.True(t, logger.IsErrorEnable())
	assert.False(t, logger.IsInfoEnable())

	logger.SetLevel(Trace)
	assert.True(t, logger.IsTraceEnable())
	assert.True(t, logger.IsInfoEnable())
}

func TestLogger_AddAppender(t *testing.T) {
	logger := CurrentPackageLogger()
	assert.Equal(t, 1, len(logger.Appenders()))
	appender := NewConsole2Appender()
	logger.AddAppenders(appender)
	assert.Equal(t, 2, len(logger.Appenders()))
	assert.Equal(t, appender, logger.Appenders()[1])
}


func TestFormatMessage(t *testing.T) {
	assert.Equal(t, "This is a test", formatMessage("This is a test"), "")
	assert.Equal(t, "This is a test", formatMessage("This is a", "test"), "")
	assert.Equal(t, "This is 1", formatMessage("This is", 1), "")
	assert.Equal(t, "This is 1", formatMessage("This is {}", 1), "")
	assert.Equal(t, "This is 1 2", formatMessage("This is {}", 1, 2), "")
	assert.Equal(t, "1, 2", formatMessage("{}, {}", 1, 2), "")
}
