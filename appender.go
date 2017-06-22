package vlog

import (
	"os"
	"bytes"
	"sync/atomic"
)

// Appender write the log to one destination.
// Appender Should  be reused across loggers.
type Appender interface {
	// the name of this appender
	Name() string
	// append new data
	Append(data []byte) (written int, err error)
	// get the transformer of this appender
	Transformer() Transformer
	// set transformer to this appender
	SetTransformer(transformer Transformer)
}

// Used for impl Appender Transformer/Name... methods
type AppenderMixin struct {
	name        string
	transformer atomic.Value //*Transformer
}

func NewAppenderMixin(name string) *AppenderMixin {
	return &AppenderMixin{name: name}
}

func (am *AppenderMixin) Name() string {
	return am.name
}

func (am *AppenderMixin) Transformer() Transformer {
	iface := am.transformer.Load()
	if iface == nil {
		return DefaultTransformer()
	}
	return *iface.(*Transformer)
}

func (am *AppenderMixin) SetTransformer(transformer Transformer) {
	am.transformer.Store(&transformer)
}

// appender write log to stdout
type ConsoleAppender struct {
	*AppenderMixin
	file *os.File
}

func (ca *ConsoleAppender) Append(data []byte) (written int, err error) {
	return ca.file.Write(data)
}

var defaultAppender = NewConsoleAppender("default")
// The default appender all logger use
func DefaultAppender() Appender {
	return defaultAppender
}

// create console appender, which write log to stdout
func NewConsoleAppender(name string) Appender {
	return &ConsoleAppender{file: os.Stdout, AppenderMixin: NewAppenderMixin(name)}
}

// create console appender, which write log to stderr
func NewConsole2Appender() Appender {
	return &ConsoleAppender{file: os.Stderr}
}

// appender discard all logs
type NopAppender struct {
	*AppenderMixin
}

// create nop appender
func NewNopAppender(name string) Appender {
	return &NopAppender{AppenderMixin: NewAppenderMixin(name)}
}

func (NopAppender) Append(data []byte) (written int, err error) {
	return len(data), nil
}

// appender write log into memory
type BytesAppender struct {
	*AppenderMixin
	buffer bytes.Buffer
}

func NewBytesAppender(name string) Appender {
	return &BytesAppender{AppenderMixin: NewAppenderMixin(name)}
}

func (b *BytesAppender) Append(data []byte) (written int, err error) {
	return b.buffer.Write(data)
}
