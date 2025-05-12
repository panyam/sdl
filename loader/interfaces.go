package loader

import (
	"io"

	"github.com/panyam/leetcoach/sdl/decl" // Loader needs to know about the AST structure
)

// Parser defines the interface for parsing SDL content.
type Parser interface {
	// Parse reads from the input reader and returns the root AST node.
	// sourceName is used for context in error messages (e.g., file path).
	Parse(input io.Reader, sourceName string) (*decl.FileDecl, error)
}

// FileResolver defines the interface for resolving import paths and reading file content.
type FileResolver interface {
	// Resolve takes the path of the importing file and the path string from the import statement.
	// It should return:
	// 1. An io.ReadCloser for the content of the resolved file.
	// 2. The canonical path (e.g., absolute path) of the resolved file, used for caching and cycle detection.
	// 3. An error if resolution or reading fails.
	Resolve(importerPath, importPath string) (content io.ReadCloser, canonicalPath string, err error)
}
