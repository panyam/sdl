package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test the core filesystem filtering logic
func TestFileSystemFiltering(t *testing.T) {
	ws := &WebServer{}

	tests := []struct {
		name       string
		filename   string
		extensions []string
		expected   bool
	}{
		{
			name:       "SDL file allowed",
			filename:   "example.sdl",
			extensions: []string{".sdl", ".recipe"},
			expected:   true,
		},
		{
			name:       "Recipe file allowed",
			filename:   "demo.recipe",
			extensions: []string{".sdl", ".recipe"},
			expected:   true,
		},
		{
			name:       "Markdown file blocked",
			filename:   "README.md",
			extensions: []string{".sdl", ".recipe"},
			expected:   false,
		},
		{
			name:       "No extension blocked",
			filename:   "Makefile",
			extensions: []string{".sdl", ".recipe"},
			expected:   false,
		},
		{
			name:       "Empty filter allows all",
			filename:   "anything.txt",
			extensions: []string{},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := FileSystemConfig{
				Extensions: tt.extensions,
			}
			result := ws.isAllowedFile(tt.filename, config)
			if result != tt.expected {
				t.Errorf("isAllowedFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// Test filesystem security - path traversal prevention
func TestFileSystemSecurity(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sdl-security-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override default filesystems for testing
	oldDefaultFS := defaultFileSystems
	defaultFileSystems = map[string]FileSystemConfig{
		"test": {
			ID:         "test",
			BasePath:   tmpDir,
			ReadOnly:   false,
			Extensions: []string{".sdl", ".recipe"},
		},
	}
	defer func() {
		defaultFileSystems = oldDefaultFS
	}()

	ws := &WebServer{}

	// Test various path traversal attempts
	securityTests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "Basic path traversal with ..",
			url:            "/api/filesystems/test/../../../etc/passwd",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Path traversal in middle",
			url:            "/api/filesystems/test/subdir/../../etc/passwd",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "URL encoded path traversal",
			url:            "/api/filesystems/test/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Valid subdirectory access",
			url:            "/api/filesystems/test/subdir/file.sdl",
			expectedStatus: http.StatusNotFound, // File doesn't exist but path is valid
		},
	}

	for _, tt := range securityTests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			ws.handleFilesystemOperations(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for URL: %s",
					tt.expectedStatus, w.Code, tt.url)
			}
		})
	}
}

// Test file operations with a mock filesystem
func TestFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sdl-ops-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testContent := "system Test {}"
	testFile := filepath.Join(tmpDir, "test.sdl")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Override default filesystems for testing
	oldDefaultFS := defaultFileSystems
	defaultFileSystems = map[string]FileSystemConfig{
		"test": {
			ID:         "test",
			BasePath:   tmpDir,
			ReadOnly:   false,
			Extensions: []string{".sdl", ".recipe"},
		},
	}
	defer func() {
		defaultFileSystems = oldDefaultFS
	}()

	ws := &WebServer{}

	t.Run("Read file", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/filesystems/test/test.sdl", nil)
		w := httptest.NewRecorder()

		ws.handleFilesystemOperations(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body, _ := io.ReadAll(w.Result().Body)
		if string(body) != testContent {
			t.Errorf("Expected content %q, got %q", testContent, string(body))
		}
	})

	t.Run("Write file", func(t *testing.T) {
		newContent := "system Updated {}"
		req := httptest.NewRequest("PUT", "/api/filesystems/test/new.sdl",
			strings.NewReader(newContent))
		w := httptest.NewRecorder()

		ws.handleFilesystemOperations(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify file was written
		written, err := os.ReadFile(filepath.Join(tmpDir, "new.sdl"))
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}
		if string(written) != newContent {
			t.Errorf("Expected written content %q, got %q", newContent, string(written))
		}
	})

	t.Run("List directory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/filesystems/test/", nil)
		w := httptest.NewRecorder()

		ws.handleFilesystemOperations(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp ListFilesResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Should have test.sdl and new.sdl from previous tests
		if len(resp.Files) < 1 {
			t.Errorf("Expected at least 1 file, got %d", len(resp.Files))
		}
	})
}
