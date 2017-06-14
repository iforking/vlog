package verylog

import (
	"sync"
)

var logCache = &LoggerCache{logs: make(map[string]*Logger)}

type LoggerCache struct {
	logs map[string]*Logger
	lock sync.Mutex
}

func (lc *LoggerCache) load(name string) *Logger {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	logger, ok := lc.logs[name]
	if ok {
		return logger
	}

	logger, ok = lc.logs[name]
	if ok {
		return logger
	}
	logger = createLogger(name)
	lc.logs[name] = logger
	return logger
}
