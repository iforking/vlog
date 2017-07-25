package vlog

import (
	"errors"
	"log/syslog"
	"strconv"
	"time"
)

var _ Appender = (*SyslogAppender)(nil)

// SyslogAppender write log to syslogd, using go syslog package.
// This appender always send only raw log message, SyslogAppender will not take effect.
//
// SyslogAppender will map log levels from vlog to syslog by the following rules:
// TRACE		-- LOG_DEBUG
// DEBUG		-- LOG_DEBUG
// INFO			-- LOG_INFO
// WARN			-- LOG_WARNING
// ERROR		-- LOG_ERR
// CRITICAL		-- LOG_CRIT
type SyslogAppender struct {
	log         *syslog.Writer
	levelMap    map[Level]syslog.Priority
	transformer sysLogTransformer
}

// defaultLevelMap is the default level map from vlog to syslog
var defaultLevelMap = map[Level]syslog.Priority{
	Trace:    syslog.LOG_DEBUG,
	Debug:    syslog.LOG_DEBUG,
	Info:     syslog.LOG_INFO,
	Warn:     syslog.LOG_WARNING,
	Error:    syslog.LOG_ERR,
	Critical: syslog.LOG_CRIT,
}

type sysLogTransformer struct {
}

func (st sysLogTransformer) Transform(logger string, level Level, now time.Time, message string) []byte {
	return []byte(message)
}

// NewSyslogAppender create syslog appender, to system syslog daemon.
func NewSyslogAppender(tag string) (*SyslogAppender, error) {
	log, err := syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL0, tag)
	if err != nil {
		return nil, err
	}
	return &SyslogAppender{log: log, levelMap: defaultLevelMap}, nil
}

// NewSyslogAppenderToAddress create syslog appender, to a log daemon connected by network address.
func NewSyslogAppenderToAddress(network string, address string, tag string) (Appender, error) {
	log, err := syslog.Dial(network, address, syslog.LOG_INFO|syslog.LOG_LOCAL0, tag)
	if err != nil {
		return nil, err
	}
	return &SyslogAppender{log: log, levelMap: defaultLevelMap}, nil
}

// SetLevelMap set level map from vlog to syslog, replace the default log level map.
// This method should be called before appender start to work.
func (sa *SyslogAppender) SetLevelMap(levelMap map[Level]syslog.Priority) {
	sa.levelMap = levelMap
}

// Append write one log entry to syslog
func (sa *SyslogAppender) Append(name string, level Level, data []byte) error {
	if priority, ok := sa.levelMap[level]; ok {
		switch priority {
		case syslog.LOG_DEBUG:
			return sa.log.Debug(string(data))
		case syslog.LOG_INFO:
			return sa.log.Info(string(data))
		case syslog.LOG_NOTICE:
			return sa.log.Notice(string(data))
		case syslog.LOG_WARNING:
			return sa.log.Warning(string(data))
		case syslog.LOG_ERR:
			return sa.log.Err(string(data))
		case syslog.LOG_CRIT:
			return sa.log.Crit(string(data))
		case syslog.LOG_ALERT:
			return sa.log.Alert(string(data))
		case syslog.LOG_EMERG:
			return sa.log.Emerg(string(data))
		default:
			return errors.New("unknown syslog level: " + strconv.Itoa(int(priority)))
		}
	}

	_, err := sa.log.Write(data)
	return err
}

// Transformer always return the default, non-
func (sa *SyslogAppender) Transformer() Transformer {
	return sa.transformer
}

// SetTransformer not take effect for SyslogAppender, which always only send log message
func (sa *SyslogAppender) SetTransformer(transformer Transformer) {

}

// Close the syslog connection
func (sa *SyslogAppender) Close() error {
	return sa.log.Close()
}
