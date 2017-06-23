package vlog

import (
	"sync"
)

var logCache = newDefaultLogCache()

func newDefaultLogCache() *LoggerCache {
	return &LoggerCache{
		loggerMap:       make(map[string]*Logger),
		appenderConfigs: []*appenderConfig{{prefix: "", appenders: []Appender{DefaultAppender()}}},
		levelConfigs:    []*levelConfig{{prefix: "", level: DEFAULT_LEVEL}},
	}
}

type LoggerCache struct {
	levelConfigs    []*levelConfig
	appenderConfigs []*appenderConfig
	loggerMap       map[string]*Logger
	lock            sync.Mutex
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

	logger = &Logger{
		name: name,
	}
	logger.SetLevel(lc.matchLevel(name))
	logger.SetAppenders(lc.matchAppender(name))
	lc.loggerMap[name] = logger
	return logger
}

func (lc *LoggerCache) matchLevel(name string) Level {
	var maxMatchLen = 0
	var level Level
	for _, levelConfig := range lc.levelConfigs {
		if matchPrefix(name, levelConfig.prefix) {
			matchLen := len(levelConfig.prefix)
			if matchLen >= maxMatchLen {
				level = levelConfig.level
				maxMatchLen = matchLen
			}
		}
	}
	return level
}

func (lc *LoggerCache) matchAppender(name string) []Appender {
	var maxMatchLen = 0
	var appenders []Appender
	for _, appenderConfig := range lc.appenderConfigs {
		if matchPrefix(name, appenderConfig.prefix) {
			matchLen := len(appenderConfig.prefix)
			if matchLen >= maxMatchLen {
				appenders = appenderConfig.appenders
				maxMatchLen = matchLen
			}
		}
	}
	return appenders
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

func (lc *LoggerCache) SetPrefixLevel(prefix string, level Level) {
	for _, logger := range lc.filter(prefix) {
		logger.SetLevel(level)
	}
	lc.levelConfigs = append(lc.levelConfigs, &levelConfig{prefix: prefix, level: level})
}

func (lc *LoggerCache) SetPrefixAppenders(prefix string, appenders []Appender) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	for _, logger := range lc.filter(prefix) {
		logger.SetAppenders(appenders)
	}
	lc.appenderConfigs = append(lc.appenderConfigs, &appenderConfig{prefix: prefix, appenders: appenders})
}

func (lc *LoggerCache) AddPrefixAppender(prefix string, appender Appender) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	for _, logger := range lc.filter(prefix) {
		logger.AddAppender(appender)
	}
	var found = false
	for _, appenderConfig := range lc.appenderConfigs {
		if appenderConfig.prefix == prefix {
			appenderConfig.appenders = append(appenderConfig.appenders, appender)
			found = true
		}
	}
	if !found {
		appenders := lc.matchAppender(prefix)
		appenders = append(appenders, appender)
		lc.appenderConfigs = append(lc.appenderConfigs, &appenderConfig{prefix: prefix, appenders: appenders})
	}
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

type levelConfig struct {
	prefix string
	level  Level
}

type appenderConfig struct {
	prefix    string
	appenders []Appender
}
