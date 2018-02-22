// +build linux darwin

package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewSyslogAppender(t *testing.T) {
	appender, err := NewSyslogAppender("vlog")
	assert.NoError(t, err)
	defer appender.Close()
}

func TestSyslogAppender_Append(t *testing.T) {
	appender, _ := NewSyslogAppender("vlog")
	defer appender.Close()
	appender.Append("vlog", Info, []byte("This is a test"))
}
