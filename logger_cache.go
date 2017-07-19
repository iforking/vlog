package vlog

import (
	"sync"
	"os"
	"errors"
	"strings"
	"unsafe"
)

var loggerCache = initLogCache()

func initLogCache() *LoggerCache {
	configPath := os.Getenv("VLOG_CONFIG_FILE")
	if configPath == "" {
		configPath = "vlog.xml"
	}
	info, err := os.Stat(configPath)
	if err != nil || info.IsDir() {
		// no log config file
		return newDefaultLogCache()
	}
	root, err := LoadXmlConfig(configPath)
	if err != nil {
		panic(wrapError("load vlog config file error", err))
	}

	cache, err := createFromConfig(root)
	if err != nil {
		panic(wrapError("parse vlog config file error", err))
	}
	FreezeLoggerSetting()
	return cache
}

func createFromConfig(root *RootElement) (*LoggerCache, error) {
	var transformerMap = map[string]Transformer{}
	for _, e := range root.TransformerElements.TransformerElements {
		if e.Name == "" {
			return nil, errors.New("transformer name not set")
		}
		if _, ok := transformerMap[e.Name]; ok {
			return nil, errors.New("transformer " + e.Name + " already defined")
		}
		if e.Type == "" {
			return nil, errors.New("transformer type not set, name:" + e.Name)
		}
		builder, ok := builderRegistry.GetTransformerBuilder(e.Type)
		if !ok {
			return nil, errors.New("unknown transformer type:" + e.Type)
		}
		transformer, err := builder(concatBytes([]byte("<root>"), e.InnerXML, []byte("</root>")))
		if err != nil {
			return nil, wrapError("build transformer from config error", err)
		}
		transformerMap[e.Name] = transformer
	}

	var appenderMap = map[string]Appender{}
	for _, e := range root.AppenderElements.AppenderElements {
		if e.Name == "" {
			return nil, errors.New("appender name not set")
		}
		if _, ok := appenderMap[e.Name]; ok {
			return nil, errors.New("appender " + e.Name + " already defined")
		}
		if e.Type == "" {
			return nil, errors.New("appender " + e.Name + " type not set")
		}
		builder, ok := builderRegistry.GetAppenderBuilder(e.Type)
		if !ok {
			return nil, errors.New("unknown appender type: " + e.Type)
		}
		appender, err := builder(concatBytes([]byte("<root>"), e.InnerXML, []byte("</root>")))
		if err != nil {
			return nil, wrapError("build appender from config error", err)
		}

		var transformer Transformer
		if e.TransformerName == "" {
			transformer = DefaultTransformer()
		} else {
			if t, ok := transformerMap[e.TransformerName]; !ok {
				return nil, errors.New("transformer " + e.TransformerName + " not exists")
			} else {
				transformer = t
			}
		}
		appender.SetTransformer(transformer)
		appenderMap[e.Name] = appender
	}

	var loggers = map[string]bool{}
	var logConfigs []*LoggerConfig
	var rootConfig = &LoggerConfig{prefix: "", level: DefaultLevel, appenders: []Appender{DefaultAppender()}}
	for _, e := range root.LoggerElements {
		if loggers[e.Name] {
			return nil, errors.New("logger " + e.Name + " already defined")
		}
		if e.Level == "" {
			return nil, errors.New("logger " + e.Name + " level not set")
		}
		level, ok := levelNamesReverse[strings.ToUpper(e.Level)]
		if !ok {
			return nil, errors.New("unknown level " + e.Level + " for log " + e.Name)
		}
		if len(e.AppenderRefs) == 0 {
			return nil, errors.New("logger " + e.Name + " appenders not set")
		}
		var appenders []Appender
		for _, ar := range e.AppenderRefs {
			appender, ok := appenderMap[ar.Name]
			if !ok {
				return nil, errors.New("appender " + ar.Name + " not found for logger " + e.Name)
			}
			appenders = append(appenders, appender)
		}
		if e.Name == "" {
			rootConfig = &LoggerConfig{prefix: e.Name, level: level, appenders: appenders}
		} else {
			logConfigs = append(logConfigs, &LoggerConfig{prefix: e.Name, level: level, appenders: appenders})
		}
	}
	logConfigs = append(logConfigs, rootConfig)

	return &LoggerCache{
		loggerMap:  make(map[string]*Logger),
		logConfigs: logConfigs,
	}, nil
}

func newDefaultLogCache() *LoggerCache {
	return &LoggerCache{
		loggerMap:  make(map[string]*Logger),
		logConfigs: []*LoggerConfig{{prefix: "", level: DefaultLevel, appenders: []Appender{DefaultAppender()}}},
	}
}

type LoggerCache struct {
	logConfigs []*LoggerConfig
	loggerMap  map[string]*Logger
	lock       sync.Mutex
}

func (lc *LoggerCache) Load(name string) *Logger {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	logger, ok := lc.loggerMap[name]
	if ok {
		return logger
	}

	logger, ok = lc.loggerMap[name]
	if ok {
		return logger
	}

	logConfig := lc.matchConfig(name)

	appenders := &logConfig.appenders
	logger = &Logger{
		name:      name,
		level:     int32(logConfig.level),
		appenders: unsafe.Pointer(appenders),
	}
	lc.loggerMap[name] = logger
	return logger
}

func (lc *LoggerCache) matchConfig(name string) *LoggerConfig {
	var maxMatchLen = 0
	var config *LoggerConfig
	for _, logConfig := range lc.logConfigs {
		if matchPrefix(name, logConfig.prefix) {
			matchLen := len(logConfig.prefix)
			if matchLen >= maxMatchLen {
				config = logConfig
				maxMatchLen = matchLen
			}
		}
	}
	return config
}

func (lc *LoggerCache) filter(prefix string) []*Logger {
	loggers := []*Logger{}
	for _, logger := range lc.loggerMap {
		if matchPrefix(logger.Name(), prefix) {
			loggers = append(loggers, logger)
		}
	}
	return loggers
}

func matchPrefix(name string, prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(prefix) > len(name) {
		return false
	}
	i := 0
	for ; i < len(prefix); i++ {
		if prefix[i] != name[i] {
			return false
		}
	}
	if prefix[i-1] == '/' || len(prefix) == len(name) || name[i] == '/' {
		return true
	}
	return false
}

type LoggerConfig struct {
	prefix    string
	level     Level
	appenders []Appender
}
