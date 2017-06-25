package vlog

import (
	"io/ioutil"
	"encoding/xml"
	"sync"
	"errors"
	"strconv"
	"strings"
)

// the root of config elements
type RootElement struct {
	XMLName             struct{}    `xml:"vlog"`
	LoggerElements      []*LoggerElement `xml:"logger"`
	AppenderElements    *AppenderElements `xml:"appenders"`
	TransformerElements *TransformerElements `xml:"transformers"`
}

type AppenderRef struct {
	Name string `xml:"name,attr"`
}

type LoggerElement struct {
	Name         string `xml:"name,attr"`
	Level        string `xml:"level,attr"`
	AppenderRefs []AppenderRef `xml:"appender-ref"`
}

type AppenderElements struct {
	AppenderElements []*AppenderElement `xml:"appender"`
}

type AppenderElement struct {
	Name            string `xml:"name,attr"`
	Type            string `xml:"type,attr"`
	TransformerName string `xml:"transformer-ref,attr"`
	InnerXML        []byte `xml:",innerxml"`
}

type TransformerElements struct {
	TransformerElements []*TransformerElement `xml:"transformer"`
}

type TransformerElement struct {
	Name     string `xml:"name,attr"`
	Type     string `xml:"type,attr"`
	InnerXML []byte `xml:",innerxml"`
}

// load logger config from xml config file
func LoadXmlConfig(path string) (*RootElement, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, wrapError("load log config file failed", err)
	}
	var root RootElement
	err = xml.Unmarshal(data, &root)
	if err != nil {
		return nil, wrapError("unmarshal log config file failed", err)
	}
	return &root, nil
}

// Accept transformer xml config, create an transformer
type TransformerBuilder func(xmlData []byte) (Transformer, error)

func buildPatternTransformer(xmlData []byte) (Transformer, error) {
	var p struct {
		Pattern string `xml:"pattern"`
	}
	err := xml.Unmarshal(xmlData, &p)
	if err != nil {
		return nil, wrapError("unmarshal PatternTransformer config failed", err)
	}
	pattern, err := strconv.Unquote("\"" + strings.Replace(p.Pattern, "\"", "__double_quote__", -1) + "\"")
	if err != nil {
		return nil, wrapError("unquote PatternTransformer pattern string failed", err)
	}
	pattern = strings.Replace(pattern, "__double_quote__", "\"", -1)
	return NewPatternFormatter(pattern)
}

// Accept appender xml config, create an appender
type AppenderBuilder func(xmlData []byte) (Appender, error)

// for build FileAppender
func buildFileAppender(xmlData []byte) (Appender, error) {
	var setting = &struct {
		Path string `xml:"path"`
		Rotater *struct {
			Type        string `xml:"type,attr"`
			Pattern     string `xml:"pattern,attr"`    // for time rotater
			Size        int64 `xml:"rotate-size,attr"` // for size rotater
			SuffixWidth int `xml:"suffix-width,attr"`  // for size rotater
		} `xml:"rotater"`
	}{}
	err := xml.Unmarshal(xmlData, setting)
	if err != nil {
		return nil, wrapError("unmarshal FileAppender config failed", err)
	}

	if setting.Path == "" {
		return nil, errors.New("FileAppender path not set")
	}
	var rotater Rotater = nil
	rotaterSetting := setting.Rotater
	if rotaterSetting != nil {
		rType := rotaterSetting.Type
		if rType == "SizeRotater" {
			if rotaterSetting.Size == 0 || rotaterSetting.SuffixWidth == 0 {
				return nil, errors.New("should set rotate-size and suffix-width")
			}
			rotater = NewSizeRotater(rotaterSetting.Size, rotaterSetting.SuffixWidth)
		} else if rType == "DailyRotater" {
			rotater = NewDayRotater(rotaterSetting.Pattern)
		} else if rType == "HourlyRotater" {
			rotater = NewHourRotater(rotaterSetting.Pattern)
		} else {
			return nil, errors.New("unknown rotater: " + rType)
		}
	}
	return NewFileAppender(setting.Path, rotater)
}

var builderRegistry *BuilderRegistry = &BuilderRegistry{
	transformerBuilderMap: map[string]TransformerBuilder{},
	appenderBuilderMap:    map[string]AppenderBuilder{},
}

type BuilderRegistry struct {
	transformerBuilderMap map[string]TransformerBuilder
	appenderBuilderMap    map[string]AppenderBuilder
	lock                  sync.Mutex
}

// register one transformer builder, so can be used in config file
func (tr *BuilderRegistry) RegisterTransformerBuilder(_type string, builder TransformerBuilder) {
	tr.lock.Lock()
	tr.transformerBuilderMap[_type] = builder
	tr.lock.Unlock()
}

func (tr *BuilderRegistry) GetTransformerBuilder(_type string) (builder TransformerBuilder, ok bool) {
	tr.lock.Lock()
	builder, ok = tr.transformerBuilderMap[_type]
	tr.lock.Unlock()
	return
}

// register one appender builder, so can be used in config file
func (tr *BuilderRegistry) RegisterAppenderBuilder(_type string, builder AppenderBuilder) {
	tr.lock.Lock()
	tr.appenderBuilderMap[_type] = builder
	tr.lock.Unlock()
}

func (tr *BuilderRegistry) GetAppenderBuilder(_type string) (builder AppenderBuilder, ok bool) {
	tr.lock.Lock()
	builder, ok = tr.appenderBuilderMap[_type]
	tr.lock.Unlock()
	return
}

func init() {
	builderRegistry.RegisterTransformerBuilder("PatternTransformer", buildPatternTransformer)
	builderRegistry.RegisterAppenderBuilder("ConsoleAppender", func(xmlData []byte) (Appender, error) {
		return NewConsoleAppender(), nil
	})
	builderRegistry.RegisterAppenderBuilder("Console2Appender", func(xmlData []byte) (Appender, error) {
		return NewConsole2Appender(), nil
	})
	builderRegistry.RegisterAppenderBuilder("FileAppender", buildFileAppender)
	builderRegistry.RegisterAppenderBuilder("NopAppender", func(xmlData []byte) (Appender, error) {
		return NewNopAppender(), nil
	})
}
