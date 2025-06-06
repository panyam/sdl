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
	ErrorCollector

	// The canonical path of a file
	FullPath string

	// The AST Node corresponding to this file
	FileDecl *decl.FileDecl

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
	if parser == nil {
		parser = &SDLParserAdapter{}
	}
	if resolver == nil {
		resolver = NewDefaultFileResolver()
	}
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
	fileDecl.FullPath = canonicalPath
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
		if importDecl.Path == nil || importDecl.Path.Value.IsNil() {
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
			err := fmt.Errorf("failed to load import '%s' from '%s': \n    %w", importPathStr, canonicalPath, err)
			fileStatus.Errors = append(fileStatus.Errors, err)
			return fileStatus, err
		}
		importDecl.ResolvedFullPath = importedFS.FullPath // Store the resolved path in the import declaration
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

// Validates an already loaded file
// Performs all kinds of static checks like type checking/inference etc
// If the file is not loaded it is also loaded.
// Validation for any imports from this file are also kicked off if they are not already validated.
func (l *Loader) Validate(fs *FileStatus) bool {
	// First the imports, then the components and then the system decls
	return l.validateFileDecl(fs, fs.FileDecl, make(map[string]bool))
}

func (l *Loader) validateFileDecl(fs *FileStatus, fileDecl *decl.FileDecl, visitedFiles map[string]bool) bool {
	if fs.HasErrors() {
		log.Printf("File %s has pre-existing errors. Cannot validate.", fs.FullPath)
		return false
	}
	if visitedFiles[fs.FullPath] {
		err := fmt.Errorf("circular import detected involving %s during validation", fs.FullPath)
		fs.AddErrors(err)
		log.Println(err)
		return false
	}
	visitedFiles[fs.FullPath] = true
	defer delete(visitedFiles, fs.FullPath)

	// Validate imported files first
	resolvedImports, err := fileDecl.Imports() // Returns map[alias]*ImportDecl
	if err != nil {
		fs.AddErrors(fmt.Errorf("in file %s: error getting imports for validation: %w", fs.FullPath, err))
		log.Println(fs.Errors[len(fs.Errors)-1])
		return false
	}

	for alias, importDeclNode := range resolvedImports {
		// Ensure import path is valid string
		importPathValue := importDeclNode.Path.Value
		if importPathValue.IsNil() {
			err := fmt.Errorf("in file %s: import for alias '%s' has nil path value", fs.FullPath, alias)
			fs.AddErrors(err)
			log.Println(err)
			// Potentially return false or continue to collect more errors
			continue
		}
		importPathStr, ok := importPathValue.Value.(string)
		if !ok {
			err := fmt.Errorf("in file %s: import path for alias '%s' is not a string: %T", fs.FullPath, alias, importPathValue.Value)
			fs.AddErrors(err)
			log.Println(err)
			continue
		}

		importedFS := l.GetFileStatus(importPathStr, fs.FullPath)
		if importedFS == nil {
			// This should ideally not happen if LoadFile worked, but defensive.
			err := fmt.Errorf("in file %s: could not find loaded file status for import '%s' (alias '%s')", fs.FullPath, importPathStr, alias)
			fs.AddErrors(err)
			log.Println(err)
			return false // Cannot proceed without the imported file's status
		}
		if !l.Validate(importedFS) { // Recursive call
			fs.AddErrors(fmt.Errorf("in file %s: validation failed for imported file '%s' (alias '%s')", fs.FullPath, importPathStr, alias))
			log.Printf("Validation for import %s (alias %s) in %s failed because %s failed validation.", importPathStr, alias, fs.FullPath, importedFS.FullPath)
			// Even if an import fails, continue validating other imports and then this file to collect all errors.
			// The final return false will propagate.
		}
	}

	// If any import validation failed and added errors to fs, return false now.
	if fs.HasErrors() {
		return false
	}

	// Create a new scope for the current file's inference.
	currentScope := decl.NewEnv[decl.Node](nil) // Our TypeScope, assuming Node can hold various decl types

	// First add imports
	l.AddImportedAliasesToScope(fs, currentScope)

	// then add fileDecl's internal decls
	errs := fileDecl.AddToScope(currentScope)
	fs.AddErrors(errs...)

	// If any errors were added during scope population, bail before inference.
	if fs.HasErrors() {
		return false
	}

	// Now, call InferTypesForFile with the populated scope
	// Assuming decl.InferTypesForFile signature: func(file *decl.FileDecl, typeEnv *decl.Env[decl.Node]) []error
	// PP(fileDecl)
	inf := NewInference(fs.FullPath, fileDecl)
	inf.MaxErrors = 1
	inf.Eval(currentScope)
	if inf.HasErrors() {
		fs.AddErrors(inf.Errors...)
	}

	if !fs.HasErrors() {
		fs.LastValidated = time.Now()
	}

	return !fs.HasErrors()
}

func (l *Loader) AddImportedAliasesToScope(fs *FileStatus, currentScope *decl.Env[decl.Node]) {
	fileDecl := fs.FileDecl
	// 2. Add imported symbols to the scope, respecting aliases
	// Re-fetch resolvedImports in case of prior errors that might have prevented its population.
	resolvedImportsAfterLocal, err := fileDecl.Imports()
	if err != nil {
		fs.AddErrors(fmt.Errorf("in file %s: error re-getting resolved imports for scope population: %w", fs.FullPath, err))
	} else {
		for aliasName, importDeclNode := range resolvedImportsAfterLocal {
			importedItemOriginalName := importDeclNode.ImportedItem.Value

			importPathValue := importDeclNode.Path.Value
			importPathStr, _ := importPathValue.Value.(string) // Already checked validity

			importedFS := l.GetFileStatus(importPathStr, fs.FullPath)
			if importedFS == nil || importedFS.FileDecl == nil {
				fs.AddErrors(fmt.Errorf("in file %s: internal error during scope population - imported file %s (for alias '%s') not found or parsed",
					fs.FullPath, importPathStr, aliasName))
				continue
			}
			importedFileDecl := importedFS.FileDecl

			// Ensure the imported file itself is valid before trying to pull symbols from it
			if importedFS.HasErrors() {
				log.Printf("Skipping import from %s (alias %s) into %s because it has errors.", importedFS.FullPath, aliasName, fs.FullPath)
				fs.AddErrors(fmt.Errorf("in file %s: cannot import from %s (alias '%s') because it has errors", fs.FullPath, importedFS.FullPath, aliasName))
				continue
			}

			foundSymbol := false
			// Check Enums
			if enumDecl, err_e := importedFileDecl.GetEnum(importedItemOriginalName); err_e == nil && enumDecl != nil {
				if existingRef := currentScope.GetRef(aliasName); existingRef != nil {
					fs.AddErrors(fmt.Errorf("in file %s: import alias '%s' for enum '%s' from %s conflicts with an existing symbol",
						fs.FullPath, aliasName, importedItemOriginalName, importPathStr))
				} else {
					currentScope.Set(aliasName, enumDecl)
				}
				foundSymbol = true
			} else if err_e != nil {
				fs.AddErrors(fmt.Errorf("in file %s: error getting enum '%s' from imported file %s: %w",
					fs.FullPath, importedItemOriginalName, importedFS.FullPath, err_e))
			}

			// Check Aggregators (we have to move just "methods" or "functions")
			if aggDecl, err_e := importedFileDecl.GetAggregator(importedItemOriginalName); err_e == nil && aggDecl != nil {
				if existingRef := currentScope.GetRef(aliasName); existingRef != nil {
					fs.AddErrors(fmt.Errorf("in file %s: import alias '%s' for aggregator '%s' from %s conflicts with an existing symbol",
						fs.FullPath, aliasName, importedItemOriginalName, importPathStr))
				} else {
					currentScope.Set(aliasName, aggDecl)
				}
				foundSymbol = true
			} else if err_e != nil {
				fs.AddErrors(fmt.Errorf("in file %s: error getting aggregator '%s' from imported file %s: %w",
					fs.FullPath, importedItemOriginalName, importedFS.FullPath, err_e))
			}

			// Check Components - only if not already found as an enum
			if !foundSymbol {
				if compDecl, err_c := importedFileDecl.GetComponent(importedItemOriginalName); err_c == nil && compDecl != nil {
					if existingRef := currentScope.GetRef(aliasName); existingRef != nil {
						fs.AddErrors(fmt.Errorf("in file %s: import alias '%s' for component '%s' from %s conflicts with an existing symbol",
							fs.FullPath, aliasName, importedItemOriginalName, importPathStr))
					} else {
						currentScope.Set(aliasName, compDecl)
					}
					foundSymbol = true
				} else if err_c != nil {
					fs.AddErrors(fmt.Errorf("in file %s: error getting component '%s' from imported file %s: %w",
						fs.FullPath, importedItemOriginalName, importedFS.FullPath, err_c))
				}
			}

			// ... similar for other importable types ...

			if !foundSymbol && !fs.HasErrors() { // Only add "not found" if no other error occurred for this import
				// Check if the symbol has already been added to the scope to prevent duplicate error messages.
				// This might be tricky if AddError above was conditional.
				// A simple check: if after trying all types, currentScope.GetRef(aliasName) is still nil.
				if currentScope.GetRef(aliasName) == nil {
					fs.AddErrors(fmt.Errorf("in file %s: imported symbol '%s' (aliased as '%s') not found in %s or not an importable type",
						fs.FullPath, importedItemOriginalName, aliasName, importedFS.FullPath))
				}
			}
		}
	}
}
