package console

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesystemOperations(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sdl-fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	testFiles := []struct {
		path    string
		content string
		isDir   bool
	}{
		{path: "hello.sdl", content: "system Hello {}", isDir: false},
		{path: "demo.recipe", content: "# Demo recipe", isDir: false},
		{path: "README.md", content: "# Readme", isDir: false}, // Should be filtered out
		{path: "subdir", content: "", isDir: true},
		{path: "subdir/nested.sdl", content: "system Nested {}", isDir: false},
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(tmpDir, tf.path)
		if tf.isDir {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("Failed to create dir %s: %v", tf.path, err)
			}
		} else {
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("Failed to create parent dir: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte(tf.content), 0644); err != nil {
				t.Fatalf("Failed to write file %s: %v", tf.path, err)
			}
		}
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
		"readonly": {
			ID:         "readonly",
			BasePath:   tmpDir,
			ReadOnly:   true,
			Extensions: []string{".sdl", ".recipe"},
		},
	}
	defer func() {
		defaultFileSystems = oldDefaultFS
	}()

	// Create web server for testing
	ws := &WebServer{}

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, resp *http.Response, body []byte)
	}{
		{
			name:           "List root directory",
			method:         "GET",
			url:            "/api/filesystems/test/",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response, body []byte) {
				var listResp ListFilesResponse
				if err := json.Unmarshal(body, &listResp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				
				// Should have 2 files (hello.sdl, demo.recipe) and 1 directory (subdir)
				// README.md should be filtered out
				if len(listResp.Files) != 3 {
					t.Errorf("Expected 3 items, got %d", len(listResp.Files))
				}
				
				// Check that README.md is filtered out
				for _, f := range listResp.Files {
					if f.Name == "README.md" {
						t.Error("README.md should be filtered out")
					}
				}
			},
		},
		{
			name:           "Read file content",
			method:         "GET",
			url:            "/api/filesystems/test/hello.sdl",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response, body []byte) {
				if string(body) != "system Hello {}" {
					t.Errorf("Unexpected file content: %s", string(body))
				}
			},
		},
		{
			name:           "Read filtered file should fail",
			method:         "GET",
			url:            "/api/filesystems/test/README.md",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Write new file",
			method:         "PUT",
			url:            "/api/filesystems/test/new.sdl",
			body:           "system New {}",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response, body []byte) {
				// Verify file was created
				content, err := os.ReadFile(filepath.Join(tmpDir, "new.sdl"))
				if err != nil {
					t.Errorf("Failed to read created file: %v", err)
				}
				if string(content) != "system New {}" {
					t.Errorf("Unexpected file content: %s", string(content))
				}
			},
		},
		{
			name:           "Write filtered extension should fail",
			method:         "PUT",
			url:            "/api/filesystems/test/bad.txt",
			body:           "some content",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Write to readonly filesystem should fail",
			method:         "PUT",
			url:            "/api/filesystems/readonly/new.sdl",
			body:           "system New {}",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Delete file",
			method:         "DELETE",
			url:            "/api/filesystems/test/demo.recipe",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response, body []byte) {
				// Verify file was deleted
				if _, err := os.Stat(filepath.Join(tmpDir, "demo.recipe")); !os.IsNotExist(err) {
					t.Error("File should have been deleted")
				}
			},
		},
		{
			name:           "Delete directory should fail",
			method:         "DELETE",
			url:            "/api/filesystems/test/subdir",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Create directory",
			method:         "POST",
			url:            "/api/filesystems/test/newdir",
			body:           `{"type": "directory"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *http.Response, body []byte) {
				// Verify directory was created
				info, err := os.Stat(filepath.Join(tmpDir, "newdir"))
				if err != nil {
					t.Errorf("Failed to stat created directory: %v", err)
				}
				if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			},
		},
		{
			name:           "Path traversal attack should fail",
			method:         "GET",
			url:            "/api/filesystems/test/../../../etc/passwd",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unknown filesystem should fail",
			method:         "GET",
			url:            "/api/filesystems/unknown/",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			
			req := httptest.NewRequest(tt.method, tt.url, body)
			if tt.method == "PUT" || tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}
			
			w := httptest.NewRecorder()
			ws.handleFilesystemOperations(w, req)
			
			resp := w.Result()
			respBody, _ := io.ReadAll(resp.Body)
			
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(respBody))
			}
			
			if tt.checkResponse != nil {
				tt.checkResponse(t, resp, respBody)
			}
		})
	}
}

func TestIsAllowedFile(t *testing.T) {
	ws := &WebServer{}
	
	tests := []struct {
		filename   string
		extensions []string
		expected   bool
	}{
		{"hello.sdl", []string{".sdl", ".recipe"}, true},
		{"demo.recipe", []string{".sdl", ".recipe"}, true},
		{"README.md", []string{".sdl", ".recipe"}, false},
		{"test.txt", []string{".sdl", ".recipe"}, false},
		{"no-extension", []string{".sdl", ".recipe"}, false},
		{"any-file.txt", []string{}, true}, // Empty extensions means allow all
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			fsConfig := FileSystemConfig{
				Extensions: tt.extensions,
			}
			result := ws.isAllowedFile(tt.filename, fsConfig)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.filename, result)
			}
		})
	}
}

func TestListDirectory(t *testing.T) {
	ws := &WebServer{}
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sdl-list-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []struct {
		name  string
		isDir bool
	}{
		{"hello.sdl", false},
		{"demo.recipe", false},
		{"README.md", false},
		{"subdir", true},
	}

	for _, f := range files {
		fullPath := filepath.Join(tmpDir, f.name)
		if f.isDir {
			os.Mkdir(fullPath, 0755)
		} else {
			os.WriteFile(fullPath, []byte("test"), 0644)
		}
	}

	fsConfig := FileSystemConfig{
		Extensions: []string{".sdl", ".recipe"},
	}

	fileInfos, err := ws.listDirectory(tmpDir, "/test", fsConfig)
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	// Should have 3 items: hello.sdl, demo.recipe, and subdir
	// README.md should be filtered out
	if len(fileInfos) != 3 {
		t.Errorf("Expected 3 files, got %d", len(fileInfos))
	}

	// Check file properties
	for _, f := range fileInfos {
		switch f.Name {
		case "hello.sdl", "demo.recipe":
			if f.IsDirectory {
				t.Errorf("%s should not be a directory", f.Name)
			}
			if !strings.HasPrefix(f.Path, "/test/") {
				t.Errorf("Path should start with /test/, got %s", f.Path)
			}
		case "subdir":
			if !f.IsDirectory {
				t.Error("subdir should be a directory")
			}
			if !strings.HasSuffix(f.Path, "/") {
				t.Error("Directory path should end with /")
			}
		case "README.md":
			t.Error("README.md should be filtered out")
		}
	}
}