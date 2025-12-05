package services

import (
	"bytes"
	"testing"

	"github.com/panyam/sdl/lib/runtime"
)

// QuietTest disables console logging for the duration of a test
func QuietTest(t *testing.T) func() {
	oldLevel := consoleLogger.GetLevel()
	oldEmojis := consoleLogger.useEmojis

	consoleLogger.SetLevel(runtime.LogLevelOff)
	consoleLogger.SetUseEmojis(false)

	return func() {
		consoleLogger.SetLevel(oldLevel)
		consoleLogger.SetUseEmojis(oldEmojis)
	}
}

// CaptureConsoleLog captures console log output during test execution
func CaptureConsoleLog(t *testing.T, level runtime.LogLevel) (*bytes.Buffer, func()) {
	oldLogger := consoleLogger
	buffer := &bytes.Buffer{}

	// Create new logger that writes to buffer
	consoleLogger = NewConsoleLogger(buffer, level)
	consoleLogger.SetUseEmojis(false) // No emojis in tests

	return buffer, func() {
		consoleLogger = oldLogger
	}
}
