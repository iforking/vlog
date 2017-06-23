package vlog

import (
	"fmt"
	"strings"
	"errors"
	"strconv"
	"time"
)

const (
	STRING   = 0
	PACKAGE  = 1
	FILE     = 2
	FUNCTION = 3
	LINE     = 4
	LEVEL    = 10
	LOGGER   = 11
	TIME     = 20
	MESSAGE  = 21
)

// Transformer convert one log record to byte array data.
// Transformer should can be share across goroutines, and user Should always reuse transformers.
type Transformer interface {
	Transform(logger string, level Level, message string, args []interface{}) []byte
}

var defaultTransformer = NewDefaultPatternTransformer()
// The default transformer used if not set
func DefaultTransformer() Transformer {
	return defaultTransformer
}

// Transform one log record using pattern, to string
type PatternTransformer struct {
	pattern string
	types   []uint32
	helpers []string
}

// below vars can be used in format string:
// {file} filename
// {package} package name
// {line} line number
// {function} function name
// {time} time
// {logger} the logger name
// {message} the log message
// use {{ to escape  {, use }} to escape }
func NewPatternFormatter(pattern string) (Transformer, error) {
	type State int
	const (
		NORMAL      = 0
		LEFT_FIRST  = 1
		V_NAME      = 2
		RIGHT_FIRST = 3
		FILTER      = 4
	)

	state := NORMAL
	buffer := []rune{}
	types := []uint32{}
	helpers := []string{}
	for idx, r := range []rune(pattern) {
		switch state {
		case NORMAL:
			if r == '{' {
				state = LEFT_FIRST
			} else if r == '}' {
				state = RIGHT_FIRST
			} else {
				buffer = append(buffer, r)
			}
		case LEFT_FIRST:
			if r == '{' {
				buffer = append(buffer, '{')
				state = NORMAL
			} else {
				idx := len(helpers)
				helpers = append(helpers, string(buffer))
				types = append(types, uint32((STRING<<16)|idx))
				buffer = buffer[:0]
				buffer = append(buffer, r)
				state = V_NAME
			}
		case V_NAME:
			if r == '}' || r == '|' {
				name := string(buffer)
				buffer = buffer[:0]
				if name == "file" {
					types = append(types, FILE<<16)
				} else if name == "package" {
					types = append(types, PACKAGE<<16)
				} else if name == "function" {
					types = append(types, FUNCTION<<16)
				} else if name == "line" {
					types = append(types, LINE<<16)
				} else if name == "time" {
					types = append(types, TIME<<16)
				} else if name == "logger" {
					types = append(types, LOGGER<<16)
				} else if name == "message" {
					types = append(types, MESSAGE<<16)
				} else if name == "level" {
					types = append(types, LEVEL<<16)
				} else {
					return nil, errors.New("unknown name: " + name)
				}
				if r == '}' {
					state = NORMAL
				} else if r == '|' {
					state = FILTER
				}
			} else {
				buffer = append(buffer, r)
			}
		case RIGHT_FIRST:
			if r == '}' {
				buffer = append(buffer, '}')
				state = NORMAL
			} else {
				return nil, errors.New("unexpected } at " + strconv.Itoa(idx-1))
			}
		case FILTER:
			if r == '}' {
				idx := len(helpers)
				helpers = append(helpers, string(buffer))
				buffer = buffer[:0]
				types[len(types)-1] |= uint32(idx)
				state = NORMAL
			} else {
				buffer = append(buffer, r)
			}
		default:
			return nil, errors.New("unhandled state: " + strconv.Itoa(state))
		}
	}
	if state != NORMAL {
		return nil, errors.New("format str does not finish rightly")
	}
	idx := len(helpers)
	helpers = append(helpers, string(buffer))
	types = append(types, uint32((STRING<<16)|idx))

	return &PatternTransformer{pattern: pattern, types: types, helpers: helpers}, nil
}

// return formatter with default format
func NewDefaultPatternTransformer() Transformer {
	formatter, err := NewPatternFormatter("{time} [{level}] {logger} - {message}\n")
	if err != nil {
		panic(err)
	}
	return formatter
}

// format log data to byte array data
func (f *PatternTransformer) Transform(logger string, level Level, message string, args []interface{}) []byte {

	logItems := []string{}
	var caller *caller
	depth := 4
	for _, t := range f.types {
		t1 := t >> 16
		idx := t & 0xffff
		switch t1 {
		case STRING:
			logItems = append(logItems, f.helpers[idx])
		case TIME:
			var timeFormat string
			if idx == 0 {
				timeFormat = "2006-01-02 15:04:05.999"
			} else {
				timeFormat = f.helpers[idx]
			}
			logItems = append(logItems, time.Now().Format(timeFormat))
		case LOGGER:
			logItems = append(logItems, logger)
		case LEVEL:
			logItems = append(logItems, level.Name())
		case MESSAGE:
			messageItems := f.formatMessage(message, args...)
			logItems = append(logItems, messageItems...)
		case PACKAGE:
			if caller == nil {
				caller = getCaller(depth)
			}
			logItems = append(logItems, caller.packageName)
		case FILE:
			if caller == nil {
				caller = getCaller(depth)
			}
			logItems = append(logItems, caller.fileName)
		case FUNCTION:
			if caller == nil {
				caller = getCaller(depth)
			}
			logItems = append(logItems, caller.functionName)
		case LINE:
			if caller == nil {
				caller = getCaller(depth)
			}
			logItems = append(logItems, strconv.Itoa(caller.line))
		default:
			panic("unsupported type: " + strconv.Itoa(int(t1)))
		}
	}

	return []byte(strings.Join(logItems, ""))
}

func (f *PatternTransformer) formatMessage(message string, args ...interface{}) []string {
	argNum := len(args)
	items := strings.SplitN(message, "{}", argNum+1)

	results := []string{}
	for idx, item := range items {
		results = append(results, item)
		if idx >= 0 && idx < len(items)-1 && idx < argNum {
			results = append(results, f.formatArg(args[idx]))
		}
	}

	for idx := len(items) - 1; idx < argNum; idx += 1 {
		results = append(results, " ")
		results = append(results, f.formatArg(args[idx]))
	}

	return results
}

func (f *PatternTransformer) formatArg(arg interface{}) string {
	return fmt.Sprintf("%v", arg)
}
