package verylog

import (
	"os"
	"bytes"
)

// the log dest
type Appender interface {
	Write(data string)
}

// appender write log to stdout
type ConsoleAppender struct {
	file *os.File
}

func (ca *ConsoleAppender) Write(data string) {
	ca.file.WriteString(data)
}

// create console appender, which write log to stdout
func NewConsoleAppender() Appender {
	return &ConsoleAppender{file: os.Stdout}
}

// create console appender, which write log to stderr
func NewConsole2Appender() Appender {
	return &ConsoleAppender{file: os.Stderr}
}

// appender discard all logs
type NopAppender struct {
}

// create nop appender
func NewNopAppender() Appender {
	return &NopAppender{}
}

func (NopAppender) Write(data string) {
}

// appender write log into memory
type BytesAppender struct {
	buffer bytes.Buffer
}

func NewBytesAppender() Appender {
	return &BytesAppender{}
}

func (b *BytesAppender) Write(data string) {
	b.buffer.WriteString(data)
}
