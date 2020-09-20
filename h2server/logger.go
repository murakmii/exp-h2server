package h2server

import "log"

type (
	LogLevel uint8

	Logger interface {
		Write(level LogLevel, format string, args ...interface{})
	}

	loggerImpl struct {
		logger *log.Logger
		level  LogLevel
	}

	nullLogger struct{}
)

const (
	DebugLog LogLevel = 1
	InfoLog  LogLevel = 2
	ErrorLog LogLevel = 3
)

var (
	_ Logger = (*loggerImpl)(nil)
	_ Logger = (*nullLogger)(nil)
)

func NewLogger(logger *log.Logger, level LogLevel) Logger {
	return &loggerImpl{logger: logger, level: level}
}

func (impl *loggerImpl) Write(level LogLevel, format string, args ...interface{}) {
	if level < impl.level {
		return
	}

	impl.logger.Printf(format, args...)
}

func NullLogger() Logger {
	return &nullLogger{}
}

func (null *nullLogger) Write(_ LogLevel, _ string, _ ...interface{}) {}
