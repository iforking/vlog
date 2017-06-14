package verylog

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
)

func TestFileAppender_Write(t *testing.T) {
	appender, err := NewFileAppender("test_file.log")
	assert.NoError(t, err)
	defer os.Remove("test_file.log")

	appender.Write("This is a test\n")
}

func TestFileAppender_Write2(t *testing.T) {
	appender, err := NewFileAppender("multi/path/test_file.log")
	defer os.RemoveAll("multi/")
	assert.NoError(t, err)

	appender.Write("This is a test\n")
}
