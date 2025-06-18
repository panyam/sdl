package runtime

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelOff
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelOff:
		return "OFF"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LogLevelDebug, nil
	case "INFO":
		return LogLevelInfo, nil
	case "WARN", "WARNING":
		return LogLevelWarn, nil
	case "ERROR":
		return LogLevelError, nil
	case "OFF", "NONE":
		return LogLevelOff, nil
	default:
		return LogLevelInfo, fmt.Errorf("unknown log level: %s", s)
	}
}

// Logger interface for structured logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

// DefaultLogger implements the Logger interface
type DefaultLogger struct {
	level  LogLevel
	output io.Writer
	logger *log.Logger
	mu     sync.RWMutex
}

// NewLogger creates a new logger instance
func NewLogger(output io.Writer, level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:  level,
		output: output,
		logger: log.New(output, "", log.LstdFlags),
	}
}

// SetLevel sets the minimum log level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *DefaultLogger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// log writes a log message if the level is enabled
func (l *DefaultLogger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	if level < l.level {
		return
	}
	
	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s", level.String(), msg)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// Global logger instance
var globalLogger = NewLogger(os.Stderr, LogLevelInfo)

// Package-level convenience functions

// SetLogLevel sets the global log level
func SetLogLevel(level LogLevel) {
	globalLogger.SetLevel(level)
}

// GetLogLevel returns the current global log level
func GetLogLevel() LogLevel {
	return globalLogger.GetLevel()
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

// Initialize logger from environment
func init() {
	// Check SDL_LOG_LEVEL environment variable
	if levelStr := os.Getenv("SDL_LOG_LEVEL"); levelStr != "" {
		if level, err := ParseLogLevel(levelStr); err == nil {
			SetLogLevel(level)
		}
	}
	
	// In test mode, default to ERROR level only
	if strings.HasSuffix(os.Args[0], ".test") {
		SetLogLevel(LogLevelError)
	}
}