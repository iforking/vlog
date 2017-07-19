package vlog

import (
	"os"
	"bytes"
	"sync/atomic"
)

// Appender write the log to one destination, and can provider a transformer to convert the log message to desired data.
// Appender Should be reused across loggers.
type Appender interface {
	// Append append new data to destination. name is the name of logger, level is the level of logger
	Append(name string, level Level, data []byte) error
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

func (ca *ConsoleAppender) Append(name string, level Level, data []byte) error {
	_, err := ca.file.Write(data)
	return err
}

var defaultAppender Appender = NewConsoleAppender()
// The default appender all logger use
func DefaultAppender() Appender {
	return defaultAppender
}

// create console appender, which write log to stdout
func NewConsoleAppender() *ConsoleAppender {
	return &ConsoleAppender{file: os.Stdout, CanFormattedMixin: NewAppenderMixin()}
}

// create console appender, which write log to stderr
func NewConsole2Appender() *ConsoleAppender {
	return &ConsoleAppender{file: os.Stderr, CanFormattedMixin: NewAppenderMixin()}
}

var _ Appender = (*NopAppender)(nil)
// NopAppender discard all logs
type NopAppender struct {
	*CanFormattedMixin
}

// NewNopAppender create nop appender
func NewNopAppender() *NopAppender {
	return &NopAppender{CanFormattedMixin: NewAppenderMixin()}
}

func (NopAppender) Append(name string, level Level, data []byte) error {
	return nil
}

var _ Appender = (*BytesAppender)(nil)
// BytesAppender write log into memory
type BytesAppender struct {
	*CanFormattedMixin
	buffer bytes.Buffer
}

func NewBytesAppender() *BytesAppender {
	return &BytesAppender{CanFormattedMixin: NewAppenderMixin()}
}

func (b *BytesAppender) Append(name string, level Level, data []byte) error {
	_, err := b.buffer.Write(data)
	return err
}
