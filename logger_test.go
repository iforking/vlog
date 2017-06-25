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
	assert.Equal(t, DEFAULT_LEVEL, logger.Level())

	appender := NewBytesAppender()
	logger.SetAppenders([]Appender{appender})
	logger.Info("this is a test")
	assert.True(t, strings.HasSuffix(appender.(*BytesAppender).buffer.String(),
		" [INFO] github.com/clearthesky/vlog - this is a test\n"))

	appender = NewBytesAppender()
	transformer, _ := NewPatternFormatter("{time|2006-01-02} {package}/{file} - {message}\n")
	appender.SetTransformer(transformer)
	logger.SetAppenders([]Appender{appender})
	logger.Info("this is a test")
	date := time.Now().Format("2006-01-02")
	assert.Equal(t, date+" github.com/clearthesky/vlog/logger_test.go - this is a test\n",
		appender.(*BytesAppender).buffer.String())

	logger2 := CurrentPackageLogger()
	assert.Equal(t, logger, logger2)
}

func TestLoggerJudge(t *testing.T) {
	logger := CurrentPackageLogger()
	logger.SetLevel(OFF)
	assert.False(t, logger.IsTraceEnable())
	assert.False(t, logger.IsErrorEnable())

	logger.SetLevel(CRITICAL)
	assert.True(t, logger.IsCriticalEnable())
	assert.False(t, logger.IsErrorEnable())

	logger.SetLevel(ERROR)
	assert.True(t, logger.IsErrorEnable())
	assert.False(t, logger.IsInfoEnable())

	logger.SetLevel(TRACE)
	assert.True(t, logger.IsTraceEnable())
	assert.True(t, logger.IsInfoEnable())
}

func TestLogger_AddAppender(t *testing.T) {
	logger := CurrentPackageLogger()
	assert.Equal(t, 1, len(logger.Appenders()))
	appender := NewConsole2Appender()
	logger.AddAppender(appender)
	assert.Equal(t, 2, len(logger.Appenders()))
	assert.Equal(t, appender, logger.Appenders()[1])
}
