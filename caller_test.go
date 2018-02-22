package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCaller(t *testing.T) {
	caller := getCaller(1)
	assert.Equal(t, "github.com/hsiafan/vlog", caller.packageName)
	assert.Equal(t, "caller_test.go", caller.fileName)
	assert.Equal(t, "TestCaller", caller.functionName)
	assert.Equal(t, 9, caller.line)
}
