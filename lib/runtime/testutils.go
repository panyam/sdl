package runtime

import (
	"bytes"
	"strings"
	"testing"
)

// QuietTest disables logging for the duration of a test
// Usage: defer QuietTest(t)()
func QuietTest(t *testing.T) func() {
	oldLevel := GetLogLevel()
	SetLogLevel(LogLevelOff)
	return func() {
		SetLogLevel(oldLevel)
	}
}

// CaptureLog captures log output during test execution
// Returns the captured output and a cleanup function
func CaptureLog(t *testing.T, level LogLevel) (*bytes.Buffer, func()) {
	oldLogger := globalLogger
	buffer := &bytes.Buffer{}
	
	// Create new logger that writes to buffer
	globalLogger = NewLogger(buffer, level)
	
	return buffer, func() {
		globalLogger = oldLogger
	}
}

// VerboseTest enables debug logging if test is run with -v flag
func VerboseTest(t *testing.T) func() {
	oldLevel := GetLogLevel()
	if testing.Verbose() {
		SetLogLevel(LogLevelDebug)
	}
	return func() {
		SetLogLevel(oldLevel)
	}
}

// AssertNoLogErrors checks that no ERROR level logs were produced
func AssertNoLogErrors(t *testing.T, logs string) {
	if strings.Contains(logs, "[ERROR]") {
		t.Errorf("Unexpected error logs found:\n%s", logs)
	}
}

// AssertLogContains checks that logs contain expected message
func AssertLogContains(t *testing.T, logs string, expected string) {
	if !strings.Contains(logs, expected) {
		t.Errorf("Expected log message not found.\nExpected: %s\nActual logs:\n%s", expected, logs)
	}
}

// TestLogLevel sets log level for a specific test
// Useful for debugging individual tests
func TestLogLevel(t *testing.T, level LogLevel) func() {
	oldLevel := GetLogLevel()
	SetLogLevel(level)
	t.Logf("Set log level to %s for test %s", level.String(), t.Name())
	return func() {
		SetLogLevel(oldLevel)
	}
}