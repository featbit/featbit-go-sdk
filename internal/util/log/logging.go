package log

import (
	"fmt"
	"strings"
	"time"
)

const (
	INFO = iota
	WARN
	ERROR

	TRACE = -2
	DEBUG = -1
)

type Logger interface {
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

type SimpleLogger struct {
	Level int
}

func (s *SimpleLogger) Errorf(format string, args ...interface{}) {
	format = strings.Join([]string{"[ERROR]", time.Now().Format(time.RFC3339), format, "\n"}, " ")
	fmt.Printf(format, args...)
}

func (s *SimpleLogger) Warnf(format string, args ...interface{}) {
	if s.Level <= WARN {
		format = strings.Join([]string{"[WARNING]", time.Now().Format(time.RFC3339), format, "\n"}, " ")
		fmt.Printf(format, args...)
	}
}

func (s *SimpleLogger) Infof(format string, args ...interface{}) {
	if s.Level <= INFO {
		format = strings.Join([]string{"[INFO]", time.Now().Format(time.RFC3339), format, "\n"}, " ")
		fmt.Printf(format, args...)
	}
}

func (s *SimpleLogger) Debugf(format string, args ...interface{}) {
	if s.Level <= DEBUG {
		format = strings.Join([]string{"[DEBUG]", time.Now().Format(time.RFC3339), format, "\n"}, " ")
		fmt.Printf(format, args...)
	}

}

func (s *SimpleLogger) Tracef(format string, args ...interface{}) {
	if s.Level == TRACE {
		format = strings.Join([]string{"[TRACE]", time.Now().Format(time.RFC3339), format, "\n"}, " ")
		fmt.Printf(format, args...)
	}
}

func SetLogger(l Logger) {
	logger = l
}

var logger Logger

func LogError(format string, args ...interface{}) {
	if logger != nil {
		logger.Errorf(format, args...)
	}
}

func LogWarn(format string, args ...interface{}) {
	if logger != nil {
		logger.Warnf(format, args...)
	}
}

func LogInfo(format string, args ...interface{}) {
	if logger != nil {
		logger.Infof(format, args...)
	}
}

func LogDebug(format string, args ...interface{}) {
	if logger != nil {
		logger.Debugf(format, args...)
	}
}

func LogTrace(format string, args ...interface{}) {
	if logger != nil {
		logger.Tracef(format, args...)
	}
}
