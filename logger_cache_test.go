package verylog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	logger1 := logCache.load("logger1")
	logger2 := logCache.load("logger2")
	logger3 := logCache.load("logger1")

	assert.Equal(t, logger1, logger3)
	assert.NotEqual(t, logger1, logger2)
}
