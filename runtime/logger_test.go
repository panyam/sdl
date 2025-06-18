package runtime

import (
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Test that logs are quiet by default in tests
	buffer, cleanup := CaptureLog(t, LogLevelInfo)
	defer cleanup()
	
	Debug("This should not appear")
	Info("This should appear")
	Warn("This warning should appear")
	Error("This error should appear")
	
	logs := buffer.String()
	
	if strings.Contains(logs, "This should not appear") {
		t.Error("Debug log appeared when log level was Info")
	}
	
	if !strings.Contains(logs, "This should appear") {
		t.Error("Info log did not appear")
	}
	
	if !strings.Contains(logs, "This warning should appear") {
		t.Error("Warning log did not appear")
	}
	
	if !strings.Contains(logs, "This error should appear") {
		t.Error("Error log did not appear")
	}
}

func TestQuietTest(t *testing.T) {
	// Test that QuietTest helper works
	defer QuietTest(t)()
	
	// These should not produce any output
	Debug("Debug message")
	Info("Info message")
	Warn("Warning message")
	Error("Error message")
	
	// No way to assert no output without capturing, but at least verify it runs
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		hasError bool
	}{
		{"DEBUG", LogLevelDebug, false},
		{"debug", LogLevelDebug, false},
		{"INFO", LogLevelInfo, false},
		{"WARN", LogLevelWarn, false},
		{"WARNING", LogLevelWarn, false},
		{"ERROR", LogLevelError, false},
		{"OFF", LogLevelOff, false},
		{"NONE", LogLevelOff, false},
		{"INVALID", LogLevelInfo, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLogLevel(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("Expected error for input %s", tt.input)
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error for input %s: %v", tt.input, err)
			}
			if level != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestVerboseTest(t *testing.T) {
	// Test that VerboseTest enables debug logging when -v is used
	defer VerboseTest(t)()
	
	buffer, cleanup := CaptureLog(t, GetLogLevel())
	defer cleanup()
	
	Debug("Debug message in verbose mode")
	
	logs := buffer.String()
	
	// If running with -v, debug should appear
	if testing.Verbose() && !strings.Contains(logs, "Debug message in verbose mode") {
		t.Error("Debug log did not appear in verbose mode")
	}
}