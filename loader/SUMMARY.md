# SDL Loader Package Summary (`loader` package)

**Purpose:**

This package is responsible for managing the loading of SDL source files. Its key tasks include:
1.  Resolving import paths.
2.  Parsing SDL files into Abstract Syntax Trees (ASTs) defined in the `decl` package.
3.  Handling cyclic import dependencies.
4.  Caching loaded files to avoid redundant parsing.
5.  Orchestrating the validation process for loaded files, which includes semantic checks and type inference.

**Core Components & Files:**

1.  **`loader.go`**: Contains the main `Loader` struct and its methods.
    *   `Loader`: Holds the parser, file resolver, import depth limit, and internal state for loaded files (`fileStatuses`) and pending loads (`pending` for cycle detection).
    *   `FileStatus`: Stores the AST (`*decl.FileDecl`), error list, and metadata for each loaded file.
    *   `NewLoader(...)`: Constructor for the `Loader`.
    *   `LoadFile(...)`: Recursively loads a single file and its imports, handling parsing and basic resolution. It populates `FileDecl.resolved` maps for components, enums, etc., defined within that file.
    *   `LoadFiles(...)`: A convenience method to load multiple root files.
    *   `Validate(...)`: The main entry point for validating a loaded `FileStatus`. It orchestrates the validation of imports and then the file itself.
    *   `validateFileDecl(...)`: The core recursive validation logic. **Recently refactored** to correctly prepare the type inference scope. It ensures imported files are validated first, then constructs a `*decl.Env[decl.Node]` (`currentScope`) for the file being validated. This scope is populated with:
        1.  Symbols from imported files, added under their specified aliases (handled by `AddImportedAliasesToScope`). Collision detection for aliases is performed.
        2.  Local symbols defined within the current file (handled by `fileDecl.AddToScope`). Collision detection with imported aliases or other local symbols is performed.
        It then calls `decl.InferTypesForFile(fileDecl, currentScope)`.
    *   `AddImportedAliasesToScope(...)`: **New helper method** that populates the `currentScope` with explicitly imported symbols from other files, respecting their aliases and checking for collisions.

2.  **`interfaces.go`**:
    *   `Parser` interface: Defines the contract for parsing SDL content, making the loader independent of a specific parser implementation (e.g., the `goyacc` parser from the `parser` package).
    *   `FileResolver` interface: Defines how import paths are resolved and file content is read, allowing flexibility beyond the local filesystem.

3.  **`resolver.go`**:
    *   `DefaultFileResolver`: An implementation of `FileResolver` that works with the local filesystem, handling relative and absolute paths.

4.  **Type Inference (`infer.go`):
    *   `InferTypesForFile(file *FileDecl, rootEnv *Env[Node]) []error`: The main entry point for type checking an entire file's AST. It takes the `FileDecl` and a pre-populated `Env[Node]` (from the loader) containing all visible global and imported declarations.
    *   It performs multiple passes, first resolving types in signatures and parameter defaults, then inferring types for method bodies and system definitions.
    *   Relies on `TypeScope` to manage contextual symbol lookups.
    *   Recursive helper functions (`InferExprType`, `InferTypesForStmt`, etc.) traverse the AST, infer types for expressions, and check type compatibility for operations, assignments, function calls, etc.
    *   Errors encountered during inference are collected and returned.

5.  **Type Scope (`typescope.go`):
    *   `TypeScope` struct: Assists `infer.go` by providing a structured way to look up the type of identifiers.
    *   It uses an `env *Env[Node]` (passed from the loader via `InferTypesForFile` and pushed for lexical blocks) to find named declarations (local `let` variables, global/imported enums, components).
    *   It also uses `currentComponent` and `currentMethod` context to resolve `self`, method parameters, and component members (params/uses).
    *   `Get(name string) (*Type, bool)`: The primary lookup method.
    *   `Set(name string, identNode *IdentifierExpr, t *Type)`: Used to record the inferred type of `let`-defined variables into the current lexical `env`.

**Process Flow (Loading & Validation):**

1.  `LoadFile(filePath, ...)` is called for a root file.
2.  The `FileResolver` determines the canonical path and provides an `io.ReadCloser` for the file content.
3.  Cycle detection and max depth checks are performed.
4.  The `Parser` (e.g., from `parser` package) parses the content into a `*decl.FileDecl` AST.
5.  The `FileDecl.Resolve()` method is called to populate its internal maps of locally defined symbols (components, enums, etc.).
6.  Imports declared in the file are recursively loaded by calling `LoadFile` for each.
7.  This process builds up a map of `FileStatus` objects in the `Loader`.
8.  `Validate(fileStatus)` is called (e.g., by `cmd/sdl/commands/validate.go`).
9.  `validateFileDecl` is invoked:
    *   It recursively validates all imported files first.
    *   It creates a new `*decl.Env[decl.Node]` to serve as the `TypeScope`.
    *   It calls `l.AddImportedAliasesToScope(...)` to populate the scope with aliased symbols from direct imports.
    *   It calls `fileDecl.AddToScope(...)` to add local symbols to the scope, checking for collisions.
    *   If scope population is error-free, it calls `decl.InferTypesForFile(fileDecl, scope)`.
    *   Any errors (parsing, resolution, import, type inference) are collected in `FileStatus.Errors`.

**Current Status & Recent Work:**

*   The loader can successfully parse files and their imports, resolve basic definitions within each file, and detect cycles.
*   **Significant refactoring for Type Inference Scoping:** The `validateFileDecl` method and its new helper `AddImportedAliasesToScope` have been updated to correctly construct and populate a `decl.Env[Node]`. This environment is then used by `decl.TypeScope` (via `decl.InferTypesForFile`) to perform type inference with proper visibility of local, imported, and aliased symbols. This was a key step to fix issues with resolving symbols like aliased enum values (e.g., `Http.StatusOk`).

**Next Steps (for this package):**

*   Thoroughly test the validation flow with various import scenarios, including complex aliasing, potential name collisions, and error conditions in imported files.
*   Ensure error reporting is robust and provides clear context for users.
*   Consider if `FileStatus` needs more fine-grained status tracking (e.g., `Parsed`, `ImportsLoaded`, `ScopePopulated`, `Inferred`).
