package vlog

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCaller(t *testing.T) {
	caller := getCaller(1)
	assert.Equal(t, "github.com/clearthesky/vlog", caller.packageName)
	assert.Equal(t, "reflect_test.go", caller.fileName)
	assert.Equal(t, "TestCaller", caller.functionName)
	assert.Equal(t, 9, caller.line)
}
