package loader

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// FileSystemResolver implements FileResolver using a FileSystem backend
type FileSystemResolver struct {
	fs FileSystem
}

// NewFileSystemResolver creates a resolver that uses the provided FileSystem
func NewFileSystemResolver(fs FileSystem) *FileSystemResolver {
	return &FileSystemResolver{fs: fs}
}

// Resolve implements the FileResolver interface
func (r *FileSystemResolver) Resolve(importerPath, importPath string, open bool) (content io.ReadCloser, canonicalPath string, err error) {
	// Resolve the import path
	resolvedPath := r.resolveImportPath(importPath, importerPath)
	
	// Check if file exists
	if !r.fs.Exists(resolvedPath) {
		return nil, "", fmt.Errorf("file not found: %s", resolvedPath)
	}
	
	// Use the resolved path as canonical path
	canonicalPath = resolvedPath
	
	if !open {
		// Just return the canonical path without opening
		return nil, canonicalPath, nil
	}
	
	// Read the file content
	data, err := r.fs.ReadFile(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file %s: %w", resolvedPath, err)
	}
	
	// Wrap in a ReadCloser
	content = io.NopCloser(bytes.NewReader(data))
	
	return content, canonicalPath, nil
}

// resolveImportPath resolves an import path relative to the importing file
func (r *FileSystemResolver) resolveImportPath(importPath, importerPath string) string {
	// Handle different import types
	
	// 1. Named imports (@stdlib/, @examples/, etc.)
	if strings.HasPrefix(importPath, "@") {
		// These are handled as-is by the filesystem mount
		return importPath
	}
	
	// 2. URL imports (https://, http://)
	if strings.Contains(importPath, "://") {
		return importPath
	}
	
	// 3. GitHub imports (github.com/...)
	if strings.HasPrefix(importPath, "github.com/") {
		// Let the GitHub filesystem handle the transformation
		return importPath
	}
	
	// 4. Absolute paths
	if strings.HasPrefix(importPath, "/") {
		return importPath
	}
	
	// 5. Relative paths - resolve relative to the importer's directory
	if importerPath == "" {
		// If no importer path, treat as relative to root
		return importPath
	}
	
	dir := filepath.Dir(importerPath)
	return filepath.Join(dir, importPath)
}

// CreateDefaultFileSystem creates a composite filesystem suitable for server usage
func CreateDefaultFileSystem() FileSystem {
	cfs := NewCompositeFS()
	
	// Local filesystem as fallback
	cfs.SetFallback(NewLocalFS("."))
	
	// GitHub support
	cfs.Mount("github.com/", NewGitHubFS())
	
	// HTTP/HTTPS support
	cfs.Mount("https://", NewHTTPFileSystem(""))
	cfs.Mount("http://", NewHTTPFileSystem(""))
	
	return cfs
}

// CreateDevelopmentFileSystem creates a filesystem for development with hot-reload support
func CreateDevelopmentFileSystem(devServerURL string) FileSystem {
	cfs := NewCompositeFS()
	
	// Memory filesystem for temporary files
	cfs.Mount("/tmp/", NewMemoryFS())
	
	// Development server mounts
	if devServerURL != "" {
		cfs.Mount("/examples/", NewHTTPFileSystem(devServerURL + "/examples"))
		cfs.Mount("/lib/", NewHTTPFileSystem(devServerURL + "/lib"))
		cfs.Mount("/demos/", NewHTTPFileSystem(devServerURL + "/demos"))
	}
	
	// GitHub support
	cfs.Mount("github.com/", NewGitHubFS())
	
	// HTTP/HTTPS support for external imports
	cfs.Mount("https://", NewHTTPFileSystem(""))
	cfs.Mount("http://", NewHTTPFileSystem(""))
	
	// Local filesystem as fallback
	cfs.SetFallback(NewLocalFS("."))
	
	return cfs
}

// CreateProductionFileSystem creates a filesystem with bundled files and caching
func CreateProductionFileSystem(bundledFiles map[string][]byte) FileSystem {
	cfs := NewCompositeFS()
	
	// Memory filesystem with preloaded bundles
	memFS := NewMemoryFS()
	memFS.PreloadFiles(bundledFiles)
	cfs.Mount("/examples/", memFS)
	cfs.Mount("/lib/", memFS)
	
	// GitHub support (with caching)
	cfs.Mount("github.com/", NewGitHubFS())
	
	// HTTP/HTTPS support for external imports
	cfs.Mount("https://", NewHTTPFileSystem(""))
	cfs.Mount("http://", NewHTTPFileSystem(""))
	
	return cfs
}