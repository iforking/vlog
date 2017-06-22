package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
	"time"
)

func TestLogger_Critical(t *testing.T) {
	logger := CurrentPackageLogger()
	assert.Equal(t, "github.com/clearthesky/vlog", logger.Name())
	assert.Equal(t, DEFAULT_LEVEL, logger.level.Load().(Level))

	appender := NewBytesAppender()
	logger.SetAppender(appender)
	logger.Info("this is a test")
	assert.True(t, strings.HasSuffix(appender.(*BytesAppender).buffer.String(),
		" [INFO] github.com/clearthesky/vlog - this is a test\n"))

	appender = NewBytesAppender()
	logger.SetAppender(appender)
	formatter, _ := NewFormatter("{time|2006-01-02} {package}/{file} - {message}\n")
	logger.SetFormatter(formatter)
	logger.Info("this is a test")
	date := time.Now().Format("2006-01-02")
	assert.Equal(t, date+" github.com/clearthesky/vlog/logger_test.go - this is a test\n",
		appender.(*BytesAppender).buffer.String())

	logger2 := CurrentPackageLogger()
	assert.Equal(t, logger, logger2)
}
