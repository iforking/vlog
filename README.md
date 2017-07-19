The Very Log lib for golang

## Get logger

Each logger has a name, there is only one logger for same name. You can pass a name, or just using current package name as logger name:

```go
var logger = vlog.GetLogger(loggerName) // specify a logger name
var logger = vlog.CurrentPackageLogger() // using full package name as logger name
```

## Log message

Logger has six levels: Trace/Debug/Info/Warn/Error/Critical.
Log methods can use format string to format params, if has more params than placeholders, the remain params will be output after formatted string.

```go
logger.Info("start the server")
logger.Info("start the server at {}:{}", host, port)
logger.Info("start the server at", host+":"+strconv.itoa(port))
logger.Error("start server error:", err)
logger.Error("start server {}:{} error:", host, port, err)
```

Loggers also have IsXxxxEnable methods, to avoid unnecessary converting cost:

```go
if logger.IsDebugEnable() {
	logger.Debug("server accept connection:", expensiveConvert(conn))
}
```

## Setting by code

By default, logger only output message with info level or above, using default message format, to standard output.
To change this, set custom Appender/Level/Transformer to the logger.

```go
var logger = vlog.CurrentPackageLogger()

func init() {
	appender := vlog.NewConsoleAppender()
	// custom log format
	transformer, _ := vlog.NewPatternTransformer("{time} [{Level}] {file}:{line} - {message}\n")
	appender.SetTransformer(transformer)
	// using custom appender
	logger.SetAppenders([]vlog.Appender{appender})
	// set level to debug, will output all message with level equal or higher than Debug
	logger.SetLevel(vlog.Debug)
}
```

## FileAppender log rotate

If using FileAppender to write log into file, a log rotater can be set to rotate log file, by log file size or time.

```go
// appender without rotater
appender := vlog.NewFileAppender("path/to/logfile", nil)
// appender with rotater rotate log file every 800m
rotater := vlog.NewSizeRotater(800 * 1024*1024, 6)
appender := vlog.NewFileAppender("path/to/logfile", rotater)
// appender with rotater rotate log file every day
rotater := vlog.NewDailyRotater("20060102")
appender := vlog.NewFileAppender("path/to/logfile", rotater)
```

## Setting by config file

Loggers can also be set by a xml format config file.
If a config file is used, all settings by code will not take effect, vlog only obey the config file.
So you can set logger by code in your final or in your lib in development.
When final routine is deployed, you can use a config file to meet your real need.

Vlog will load config file from path "vlog.xml" by default.
To use a different path, set env VLOG_CONFIG_FILE to you path, before start the routine.

Note that the logger config select logger by name segment prefix, separated by "/".
If having multi logger config matched, A Logger will use logger config that have the longest prefix.
This means a logger setting named with "github.com/user1" will apply to
loggers with name "github.com/user1", "github.com/user1/project1",
but not affect loggers with name "github.com/user2", "github.com/user123", or "github.com/user".
The logger setting with name "" will apply to all loggers, except logger be config by other logger setting.

Click to see a [sample config file](https://raw.githubusercontent.com/clearthesky/vlog/master/vlog_sample.xml).

## Appenders

Appenders supportted now:

| Appender Type | Create by Code | Name in Config File |
| :------: | :------: | :------: |
| ConsoleAppender | NewConsoleAppender | ConsoleAppender |
| ConsoleAppender | NewConsole2Appender | Console2Appender |
| FileAppender | NewFileAppender | FileAppender |
| SyslogAppender | SyslogAppender | SyslogAppender |
| NopAppender | NewNopAppender | NopAppender |

## Rotaters

| Rotater Type | Create by Code | Name in Config File |
| :------: | :------: | :------: |
| TimeRotater | NewDailyRotater | DailyRotater |
| TimeRotater | NewHourlyRotater | HourlyRotater |
| TimeRotater | NewTimeRotater | - |
| SizeRotater | NewSizeRotater | SizeRotater |

## Transformers

| Transformer Type | Create by Code | Name in Config File |
| :------: | :------: | :------: |
| PatternTransformer | NewPatternTransformer | PatternTransformer |

Below variables can be used in PatternTransformer format string:

* {file} filename
* {package} package name
* {line} line number
* {function} function name
* {time} time
* {logger} the logger name
* {Level}/{level}/{LEVEL} the logger level, with different character case
* {message} the log message

Use {{ to escape  {, use }} to escape }

{time} can set custom format via filter, by {time|2006-01-02 15:04:05.000}

