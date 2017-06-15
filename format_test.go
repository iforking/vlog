package verylog

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestFormatter_FormatMessage(t *testing.T) {
	formatter := NewDefaultFormatter()
	assert.Equal(t, "This is a test", strings.Join(formatter.formatMessage("This is a test"), ""))
	assert.Equal(t, "This is a test", strings.Join(formatter.formatMessage("This is a", "test"), ""))
	assert.Equal(t, "This is 1", strings.Join(formatter.formatMessage("This is", 1), ""))
	assert.Equal(t, "This is 1", strings.Join(formatter.formatMessage("This is {}", 1), ""))
	assert.Equal(t, "This is 1 2", strings.Join(formatter.formatMessage("This is {}", 1, 2), ""))
	assert.Equal(t, "1, 2", strings.Join(formatter.formatMessage("{}, {}", 1, 2), ""))
}
