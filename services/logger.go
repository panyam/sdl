package services

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/panyam/sdl/lib/runtime"
)

// ConsoleLogger wraps runtime.Logger with console-specific features
type ConsoleLogger struct {
	runtime.Logger
	useEmojis bool
}

// NewConsoleLogger creates a logger suitable for console output
func NewConsoleLogger(output io.Writer, level runtime.LogLevel) *ConsoleLogger {
	return &ConsoleLogger{
		Logger:    runtime.NewLogger(output, level),
		useEmojis: true,
	}
}

// SetUseEmojis enables or disables emoji output
func (l *ConsoleLogger) SetUseEmojis(use bool) {
	l.useEmojis = use
}

// Event logs an event with optional emoji
func (l *ConsoleLogger) Event(emoji, format string, args ...any) {
	prefix := ""
	if l.useEmojis && emoji != "" {
		prefix = emoji + " "
	}
	msg := fmt.Sprintf(format, args...)
	l.Info("%s%s", prefix, msg)
}

// Success logs a success message
func (l *ConsoleLogger) Success(format string, args ...any) {
	l.Event("✅", format, args...)
}

// Failure logs a failure message
func (l *ConsoleLogger) Failure(format string, args ...any) {
	l.Event("❌", format, args...)
}

// Start logs a start event
func (l *ConsoleLogger) Start(format string, args ...any) {
	l.Event("🚀", format, args...)
}

// Stop logs a stop event
func (l *ConsoleLogger) Stop(format string, args ...any) {
	l.Event("🛑", format, args...)
}

// Global console logger
var consoleLogger = NewConsoleLogger(os.Stdout, runtime.LogLevelDebug)

// Package-level convenience functions

// SetLogLevel sets the console log level
func SetLogLevel(level runtime.LogLevel) {
	consoleLogger.SetLevel(level)
}

// SetUseEmojis enables or disables emoji output
func SetUseEmojis(use bool) {
	consoleLogger.SetUseEmojis(use)
}

// Event logs an event with optional emoji
func Event(emoji, format string, args ...any) {
	consoleLogger.Event(emoji, format, args...)
}

// Success logs a success message
func Success(format string, args ...any) {
	consoleLogger.Success(format, args...)
}

// Failure logs a failure message
func Failure(format string, args ...any) {
	consoleLogger.Failure(format, args...)
}

// Start logs a start event
func Start(format string, args ...any) {
	consoleLogger.Start(format, args...)
}

// Stop logs a stop event
func Stop(format string, args ...any) {
	consoleLogger.Stop(format, args...)
}

// Debug logs a debug message
func Debug(format string, args ...any) {
	consoleLogger.Debug(format, args...)
}

// Info logs an info message
func Info(format string, args ...any) {
	consoleLogger.Info(format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...any) {
	consoleLogger.Warn(format, args...)
}

// Error logs an error message
func Error(format string, args ...any) {
	consoleLogger.Error(format, args...)
}

// Initialize from environment
func init() {
	// Check SDL_LOG_LEVEL environment variable
	if levelStr := os.Getenv("SDL_LOG_LEVEL"); levelStr != "" {
		if level, err := runtime.ParseLogLevel(levelStr); err == nil {
			SetLogLevel(level)
		}
	}

	// Check SDL_NO_EMOJI environment variable
	if os.Getenv("SDL_NO_EMOJI") != "" {
		SetUseEmojis(false)
	}

	// In test mode, default to no emojis and ERROR level
	if strings.HasSuffix(os.Args[0], ".test") {
		SetUseEmojis(false)
		SetLogLevel(runtime.LogLevelError)
	}
}
