package vlog

import (
	"sync"
	"os"
	"fmt"
	"errors"
	"strings"
	"sync/atomic"
	"unsafe"
)

var logCache = initLogCache()

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
		fmt.Fprintln(os.Stderr, "load vlog file error, use default setting:", err)
		return newDefaultLogCache()
	}

	cache, err := createFromConfig(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse vlog file error, use default setting:", err)
		return newDefaultLogCache()
	}
	LockLogger()
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
		transformer, err := builder.Build([]byte("<root>" + string(e.InnerXML) + "</root>"))
		if err != nil {
			return nil, err
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
		appender, err := builder.Build([]byte("<root>" + string(e.InnerXML) + "</root>"))
		if err != nil {
			return nil, err
		}
		if e.TransformerName == "" {
			return nil, errors.New("appender " + e.Name + " transformer not set")
		}
		transformer, ok := transformerMap[e.TransformerName]
		if !ok {
			return nil, errors.New("transformer " + e.TransformerName + " not exists")
		}
		appender.SetTransformer(transformer)
		appenderMap[e.Name] = appender
	}

	var loggers = map[string]bool{}
	var logConfigs []*LogConfig
	var rootConfig = &LogConfig{prefix: "", level: DEFAULT_LEVEL, appenders: []Appender{DefaultAppender()}}
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
			rootConfig = &LogConfig{prefix: e.Name, level: level, appenders: appenders}
		} else {
			logConfigs = append(logConfigs, &LogConfig{prefix: e.Name, level: level, appenders: appenders})
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
		logConfigs: []*LogConfig{{prefix: "", level: DEFAULT_LEVEL, appenders: []Appender{DefaultAppender()}}},
	}
}

type LoggerCache struct {
	logConfigs []*LogConfig
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
	levelValue := &atomic.Value{}
	levelValue.Store(logConfig.level)

	appenders := &logConfig.appenders
	logger = &Logger{
		name:      name,
		level:     levelValue,
		appenders: unsafe.Pointer(appenders),
	}
	lc.loggerMap[name] = logger
	return logger
}

func (lc *LoggerCache) matchConfig(name string) *LogConfig {
	var maxMatchLen = 0
	var config *LogConfig
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

type LogConfig struct {
	prefix    string
	level     Level
	appenders []Appender
}
