package vlog

import (
	"sync"
)

var logCache = &LoggerCache{loggerMap: make(map[string]*Logger)}

type LoggerCache struct {
	loggerMap map[string]*Logger
	lock      sync.Mutex
}

func (lc *LoggerCache) load(name string) *Logger {
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
	logger = createLogger(name)
	lc.loggerMap[name] = logger
	return logger
}

func (lc *LoggerCache) filter(prefix string) []*Logger {
	loggers := []*Logger{}
	lc.lock.Lock()
	defer lc.lock.Unlock()
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

// set level of all loggers, which name has prefix segments, split by slash
func SetLevel(prefix string, level Level) {
	loggers := logCache.filter(prefix)
	for _, logger := range loggers {
		logger.SetLevel(level)
	}
}

// set appender to all loggers, which name has prefix segments, split by slash
func SetAppender(prefix string, appender Appender) {
	loggers := logCache.filter(prefix)
	for _, logger := range loggers {
		logger.SetAppender(appender)
	}
}

