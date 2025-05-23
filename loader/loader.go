package loader

import (
	"fmt"
	"log"
	"os"
	"sync" // To handle potential concurrent loads if needed later, though starting sequential.
	"time"

	"github.com/panyam/sdl/decl"
)

type FileStatus struct {
	// The canonical path of a file
	FullPath string

	// The AST Node corresponding to this file
	FileDecl *decl.FileDecl

	// Errors for this file
	Errors []error

	// When the file was last parsed
	LastParsed time.Time

	// When the file was last validated
	LastValidated time.Time

	// Files imported from this file as an easy map
	ImportedFiles map[string]bool
}

func (f *FileStatus) AddImports(imported ...string) {
	if f.ImportedFiles == nil {
		f.ImportedFiles = map[string]bool{}
	}
	for _, path := range imported {
		f.ImportedFiles[path] = true
	}
}

func (f *FileStatus) HasErrors() bool {
	return len(f.Errors) > 0
}

func (f *FileStatus) PrintErrors() {
	for _, err := range f.Errors {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (f *FileStatus) AddErrors(errs ...error) {
	for _, err := range errs {
		f.Errors = append(f.Errors, err)
	}
}

// Loader handles parsing and recursively loading imported SDL files.
type Loader struct {
	parser   Parser
	resolver FileResolver
	maxDepth int

	// Internal state during a load operation
	mutex        sync.Mutex // Protects shared state if concurrency is added
	fileStatuses map[string]*FileStatus
	// loadedFiles  map[string]*decl.FileDecl
	pending map[string]bool // Tracks files currently being loaded in the recursion stack for cycle detection
}

// NewLoader creates a new SDL loader.
// maxDepth specifies the maximum import recursion depth (0 means no limit, 1 means root only, etc.).
func NewLoader(parser Parser, resolver FileResolver, maxDepth int) *Loader {
	return &Loader{
		parser:       parser,
		resolver:     resolver,
		maxDepth:     maxDepth,
		fileStatuses: make(map[string]*FileStatus),
		pending:      make(map[string]bool),
	}
}

func (l *Loader) LoadFiles(paths ...string) (allValid bool, results map[string]*FileStatus) {
	allValid = true
	results = make(map[string]*FileStatus)
	for _, filePath := range paths {
		result, err := l.LoadFile(filePath, "", 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, "  Error loading file: ")
			if result == nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				for _, err := range result.Errors {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			allValid = false
			continue
		}
		results[filePath] = result
	}
	return
}

// LoadFile starts loading file as the root and then recursives loads all its imports.
// Note that this stage the file is not validated.  It is only checked for parse errors.
//
// Params:
//
// filePath - Path of root to be imported.  This can be an absolute or relative path.  A relative path would be relative
// to the importerDir.  If the importerDir is empty then the current working dir is used.  When loading imported files
// the importerDir is set to the folder in which the "caller" file exists (filePath).
//
// importerDir - Directory relative to which the filePath is resolved if it is a relative path.
func (l *Loader) LoadFile(filePath string, importerPath string, depth int) (*FileStatus, error) {
	// 1. Check Max Depth
	// Note: depth 0 is the root, depth 1 is its direct imports, etc.
	// maxDepth 1 means only root file. maxDepth 2 means root + direct imports.
	if l.maxDepth > 0 && depth >= l.maxDepth {
		return nil, fmt.Errorf("max import depth (%d) exceeded near '%s'", l.maxDepth, filePath)
	}

	// 2. Resolve the path using the resolver to get the canonical path
	contentReader, canonicalPath, err := l.resolver.Resolve(importerPath, filePath, true)
	if err != nil {
		return nil, err
	}
	defer contentReader.Close() // Ensure the reader is closed

	// Use canonicalPath for all checks and storage from now on
	// 3. Check if already loaded
	fileStatus, found := l.fileStatuses[canonicalPath]
	if found {
		return fileStatus, nil
	}
	fileStatus = &FileStatus{FullPath: canonicalPath}
	l.fileStatuses[canonicalPath] = fileStatus

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
		fileStatus.Errors = append(fileStatus.Errors, err)
		return fileStatus, fmt.Errorf("parsing error in '%s': %w", canonicalPath, err)
	}

	// 7. Store the successfully parsed file
	fileStatus.FileDecl = fileDecl
	fileStatus.LastParsed = time.Now()

	// 8. Recursively load imports
	// First, call Resolve on the FileDecl itself to populate internal maps
	if err := fileDecl.Resolve(); err != nil {
		fileStatus.Errors = append(fileStatus.Errors, err)
		return fileStatus, fmt.Errorf("error resolving definitions in '%s': %w", canonicalPath, err)
	}

	// Now process imports
	imports, err := fileDecl.Imports()
	if err != nil {
		fileStatus.Errors = append(fileStatus.Errors, err)
		return fileStatus, err
	}

	for _, importDecl := range imports { // Assuming FileDecl has an Imports() method
		if importDecl.Path == nil || importDecl.Path.Value == nil {
			err := fmt.Errorf("invalid import statement (missing path) in '%s' at pos %d", canonicalPath, importDecl.Pos())
			fileStatus.Errors = append(fileStatus.Errors, err)
			return fileStatus, err
		}
		importPathStr, ok := importDecl.Path.Value.Value.(string)
		if !ok {
			err := fmt.Errorf("import path is not a string literal in '%s' at pos %d", canonicalPath, importDecl.Pos())
			fileStatus.Errors = append(fileStatus.Errors, err)
			return fileStatus, err
		}

		importedFS, err := l.LoadFile(importPathStr, canonicalPath, depth+1)
		if err != nil {
			// Wrap the error to show the import chain
			err := fmt.Errorf("failed to load import '%s' from '%s': %w", importPathStr, canonicalPath, err)
			fileStatus.Errors = append(fileStatus.Errors, err)
			return fileStatus, err
		}
		fileStatus.AddImports(importedFS.FullPath)
	}

	// Once imports are loaded we can perform inference and other checks on this file
	return fileStatus, nil
}

func (l *Loader) GetFileStatus(filePath string, importerPath string) *FileStatus {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// evaluate the canonical path
	_, canonicalPath, err := l.resolver.Resolve(importerPath, filePath, false)
	if err != nil {
		return nil
	}
	fileStatus, found := l.fileStatuses[canonicalPath]
	if !found {
		return nil
	}
	return fileStatus
}

// Validates a file
// Performs all kinds of static checks like type checking/inference etc
// If the file is not loaded it is also loaded.
// Validation for any imports from this file are also kicked off if they are not already validated.
func (l *Loader) Validate(fs *FileStatus) bool {
	// First the imports, then the components and then the system decls
	return l.validateFileDecl(fs, fs.FileDecl, map[string]bool{})
}

func (l *Loader) validateFileDecl(fs *FileStatus, fileDecl *decl.FileDecl, visitedFiles map[string]bool) bool {
	if len(fs.Errors) > 0 {
		log.Println("File has errors.  Cannot validate")
		return false
	}
	if visitedFiles[fs.FullPath] {
		log.Println("circular improt: ", fs)
		fs.AddErrors(fmt.Errorf("circular import"))
		return false
	}
	visitedFiles[fs.FullPath] = true
	defer func() { visitedFiles[fs.FullPath] = false }()

	imports, err := fileDecl.Imports()
	if err != nil {
		log.Println("Err: ", fs, err)
		fs.AddErrors(err)
		return false
	}

	for _, importDecl := range imports {
		importedFS := l.GetFileStatus(importDecl.Path.Value.Value.(string), fs.FullPath)
		if importedFS == nil || !l.Validate(importedFS) {
			log.Println("Validation for import failed: ", importDecl.Path.Value, importedFS)
			return false
		}
	}

	inferenceErrors := decl.InferTypesForFile(fileDecl) // Call the function from decl package
	fs.AddErrors(inferenceErrors...)

	// now validate components
	if len(inferenceErrors) == 0 {
		fs.LastValidated = time.Now()
	}

	return len(inferenceErrors) == 0
}
