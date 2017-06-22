package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestFormatter_FormatMessage(t *testing.T) {
	transformer := NewDefaultPatternTransformer().(*PatternTransformer)
	assert.Equal(t, "This is a test", strings.Join(transformer.formatMessage("This is a test"), ""))
	assert.Equal(t, "This is a test", strings.Join(transformer.formatMessage("This is a", "test"), ""))
	assert.Equal(t, "This is 1", strings.Join(transformer.formatMessage("This is", 1), ""))
	assert.Equal(t, "This is 1", strings.Join(transformer.formatMessage("This is {}", 1), ""))
	assert.Equal(t, "This is 1 2", strings.Join(transformer.formatMessage("This is {}", 1, 2), ""))
	assert.Equal(t, "1, 2", strings.Join(transformer.formatMessage("{}, {}", 1, 2), ""))
}
