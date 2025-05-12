package loader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DefaultFileResolver implements FileResolver for the local filesystem.
type DefaultFileResolver struct{}

// NewDefaultFileResolver creates a standard filesystem resolver.
func NewDefaultFileResolver() *DefaultFileResolver {
	return &DefaultFileResolver{}
}

// Resolve handles filesystem paths.
func (r *DefaultFileResolver) Resolve(importerPath, importPath string) (io.ReadCloser, string, error) {
	var resolvedPath string

	if filepath.IsAbs(importPath) {
		resolvedPath = importPath
	} else {
		// Assume importerPath is the canonical path of the importing file
		importerDir := filepath.Dir(importerPath)
		resolvedPath = filepath.Join(importerDir, importPath)
	}

	// Get the absolute path to use as the canonical path
	canonicalPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return nil, "", fmt.Errorf("could not get absolute path for '%s': %w", resolvedPath, err)
	}

	// Check if file exists and open
	file, err := os.Open(canonicalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("file not found: %s (resolved from '%s')", canonicalPath, importPath)
		}
		return nil, "", fmt.Errorf("could not open file '%s': %w", canonicalPath, err)
	}

	return file, canonicalPath, nil
}

// END_OF_FILE ./loader/resolver.go
