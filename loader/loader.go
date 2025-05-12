package loader

import (
	"fmt"
	"sync" // To handle potential concurrent loads if needed later, though starting sequential.

	"github.com/panyam/sdl/decl"
)

// LoadResult holds the outcome of a loading operation.
type LoadResult struct {
	RootFile    *decl.FileDecl            // The AST for the initially requested root file.
	LoadedFiles map[string]*decl.FileDecl // All files loaded, keyed by canonical path.
	Errors      []error                   // Parsing, resolution, cycle, depth, and type inference errors.
}

// Loader handles parsing and recursively loading imported SDL files.
type Loader struct {
	parser   Parser
	resolver FileResolver
	maxDepth int

	// Internal state during a load operation
	mutex       sync.Mutex // Protects shared state if concurrency is added
	loadedFiles map[string]*decl.FileDecl
	pending     map[string]bool // Tracks files currently being loaded in the recursion stack for cycle detection
}

// NewLoader creates a new SDL loader.
// maxDepth specifies the maximum import recursion depth (0 means no limit, 1 means root only, etc.).
func NewLoader(parser Parser, resolver FileResolver, maxDepth int) *Loader {
	return &Loader{
		parser:      parser,
		resolver:    resolver,
		maxDepth:    maxDepth,
		loadedFiles: make(map[string]*decl.FileDecl),
		pending:     make(map[string]bool),
	}
}

// LoadRootFile parses the specified root file and recursively loads its imports.
// It performs type inference on the entire loaded set of files afterwards.
func (l *Loader) LoadRootFile(rootPath string) (*LoadResult, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Reset state for this load operation
	l.loadedFiles = make(map[string]*decl.FileDecl)
	l.pending = make(map[string]bool)

	var errors []error

	// Start recursive loading
	rootFile, err := l.loadFileRecursive(rootPath, rootPath, 0) // Initial importer path is the root path itself
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to load root file '%s': %w", rootPath, err))
		// Return immediately if the root fails to load/parse
		return &LoadResult{Errors: errors}, errors[0]
	}

	// --- Type Inference Pass ---
	// Infer types only after all files are successfully parsed and loaded.
	// Iterate through the map of all loaded files.
	// We need to ensure correct resolution order if inference depends on it,
	// but decl.InferTypesForFile should handle resolving symbols within the loaded set.
	// log.Printf("Starting type inference for %d loaded files...", len(l.loadedFiles))
	for path, fileDecl := range l.loadedFiles {
		// log.Printf("Inferring types for: %s", path)
		inferenceErrors := decl.InferTypesForFile(fileDecl) // Call the function from decl package
		if len(inferenceErrors) > 0 {
			for _, inferErr := range inferenceErrors {
				errors = append(errors, fmt.Errorf("type inference error in '%s': %w", path, inferErr))
			}
		}
		// log.Printf("Finished inferring types for: %s", path)
	}

	result := &LoadResult{
		RootFile:    rootFile,
		LoadedFiles: l.loadedFiles, // Return a copy? Or the internal map? Returning internal for now.
		Errors:      errors,
	}

	// Return the first error encountered, or nil if only inference errors occurred.
	// The caller can inspect result.Errors for all issues.
	if err != nil {
		return result, err // The original loading error takes precedence
	}
	if len(errors) > 0 {
		return result, errors[0] // Return the first type inference error as the primary error
	}

	return result, nil
}

// loadFileRecursive handles the actual loading and parsing logic.
func (l *Loader) loadFileRecursive(importerPath, filePath string, depth int) (*decl.FileDecl, error) {
	// 1. Check Max Depth
	// Note: depth 0 is the root, depth 1 is its direct imports, etc.
	// maxDepth 1 means only root file. maxDepth 2 means root + direct imports.
	if l.maxDepth > 0 && depth >= l.maxDepth {
		return nil, fmt.Errorf("max import depth (%d) exceeded near '%s'", l.maxDepth, filePath)
	}

	// 2. Resolve the path using the resolver to get the canonical path
	contentReader, canonicalPath, err := l.resolver.Resolve(importerPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve import '%s' from '%s': %w", filePath, importerPath, err)
	}
	defer contentReader.Close() // Ensure the reader is closed

	// Use canonicalPath for all checks and storage from now on
	// 3. Check if already loaded
	if fileDecl, found := l.loadedFiles[canonicalPath]; found {
		return fileDecl, nil
	}

	// 4. Check for circular dependency
	if l.pending[canonicalPath] {
		return nil, fmt.Errorf("circular import detected: '%s' is already being loaded", canonicalPath)
	}

	// 5. Mark as pending
	l.pending[canonicalPath] = true
	defer delete(l.pending, canonicalPath) // Ensure cleanup on return

	// 6. Parse the file content
	// log.Printf("Parsing: %s (Importer: %s, Depth: %d)", canonicalPath, importerPath, depth) // VDebug
	fileDecl, err := l.parser.Parse(contentReader, canonicalPath)
	if err != nil {
		return nil, fmt.Errorf("parsing error in '%s': %w", canonicalPath, err)
	}

	// 7. Store the successfully parsed file
	l.loadedFiles[canonicalPath] = fileDecl

	// 8. Recursively load imports
	// First, call Resolve on the FileDecl itself to populate internal maps
	if err := fileDecl.Resolve(); err != nil {
		return nil, fmt.Errorf("error resolving definitions in '%s': %w", canonicalPath, err)
	}

	// Now process imports
	imports, err := fileDecl.Imports()
	if err != nil {
		return nil, err
	}
	for _, importDecl := range imports { // Assuming FileDecl has an Imports() method
		if importDecl.Path == nil || importDecl.Path.Value == nil {
			return nil, fmt.Errorf("invalid import statement (missing path) in '%s' at pos %d", canonicalPath, importDecl.Pos())
		}
		importPathStr, ok := importDecl.Path.Value.Value.(string)
		if !ok {
			return nil, fmt.Errorf("import path is not a string literal in '%s' at pos %d", canonicalPath, importDecl.Pos())
		}

		_, err := l.loadFileRecursive(canonicalPath, importPathStr, depth+1)
		if err != nil {
			// Wrap the error to show the import chain
			return nil, fmt.Errorf("failed to load import '%s' from '%s': %w", importPathStr, canonicalPath, err)
		}
		// Optional: Link the loaded FileDecl back to importDecl.ResolvedFile = loadedDecl
		// However, the central l.loadedFiles map is the source of truth.
	}

	return fileDecl, nil
}

// Add helper to FileDecl to get Imports if it doesn't exist
// (This should ideally be in decl/ast.go, but adding concept here)
// We might need to modify decl.FileDecl.Resolve or add an Imports() method there.
// Let's assume decl.FileDecl will have an Imports() method after Resolve() is called.
// If not, we'll need to iterate f.Declarations here.
