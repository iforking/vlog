package vlog

import (
	"os"
	"bytes"
	"sync/atomic"
)

// Appender write the log to one destination, and can provider a transformer to convert the log message to desired data.
// Appender Should be reused across loggers.
type Appender interface {
	// append new data
	Append(data []byte) (written int, err error)
	// get the transformer of this appender
	Transformer() Transformer
	// set transformer to this appender
	SetTransformer(transformer Transformer)
}

// Used for impl Appender Transformer/Name... methods
type CanFormattedMixin struct {
	transformer atomic.Value //*Transformer
}

func NewAppenderMixin() *CanFormattedMixin {
	return &CanFormattedMixin{}
}

func (am *CanFormattedMixin) Transformer() Transformer {
	iface := am.transformer.Load()
	if iface == nil {
		return DefaultTransformer()
	}
	return *iface.(*Transformer)
}

func (am *CanFormattedMixin) SetTransformer(transformer Transformer) {
	am.transformer.Store(&transformer)
}

// appender write log to stdout
type ConsoleAppender struct {
	*CanFormattedMixin
	file *os.File
}

func (ca *ConsoleAppender) Append(data []byte) (written int, err error) {
	return ca.file.Write(data)
}

var defaultAppender = NewConsoleAppender()
// The default appender all logger use
func DefaultAppender() Appender {
	return defaultAppender
}

// create console appender, which write log to stdout
func NewConsoleAppender() Appender {
	return &ConsoleAppender{file: os.Stdout, CanFormattedMixin: NewAppenderMixin()}
}

// create console appender, which write log to stderr
func NewConsole2Appender() Appender {
	return &ConsoleAppender{file: os.Stderr, CanFormattedMixin: NewAppenderMixin()}
}

// appender discard all logs
type NopAppender struct {
	*CanFormattedMixin
}

// create nop appender
func NewNopAppender() Appender {
	return &NopAppender{CanFormattedMixin: NewAppenderMixin()}
}

func (NopAppender) Append(data []byte) (written int, err error) {
	return len(data), nil
}

// appender write log into memory
type BytesAppender struct {
	*CanFormattedMixin
	buffer bytes.Buffer
}

func NewBytesAppender() Appender {
	return &BytesAppender{CanFormattedMixin: NewAppenderMixin()}
}

func (b *BytesAppender) Append(data []byte) (written int, err error) {
	return b.buffer.Write(data)
}
