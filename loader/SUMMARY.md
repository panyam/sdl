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
    *   `validateFileDecl(...)`: The core recursive validation logic. It ensures imported files are validated first, then constructs a `*decl.Env[decl.Node]` (`currentScope`) for the file being validated. This scope is populated with:
        1.  Symbols from imported files, added under their specified aliases (handled by `AddImportedAliasesToScope`). Collision detection for aliases is performed.
        2.  Local symbols defined within the current file (handled by `fileDecl.AddToScope`). Collision detection with imported aliases or other local symbols is performed.
        It then calls `InferTypesForFile` (from `infer.go` within this package).
    *   `AddImportedAliasesToScope(...)`: Helper method that populates the `currentScope` with explicitly imported symbols from other files, respecting their aliases and checking for collisions. It can handle all importable definition types (`Component`, `Enum`, `Aggregator`, `Method`).

2.  **`interfaces.go`**:
    *   `Parser` interface: Defines the contract for parsing SDL content.
    *   `FileResolver` interface: Defines how import paths are resolved and file content is read.

3.  **`resolver.go`**:
    *   `DefaultFileResolver`: An implementation of `FileResolver` for the local filesystem.

4.  **`infer.go` (Type Inference Logic):**
    *   Contains the `Inference` struct and its methods, including the main entry point `Eval(rootEnv *Env[Node])`.
    *   It performs multiple passes, first resolving types in signatures and parameter defaults, then inferring types for method bodies and system definitions.
    *   Relies on `TypeScope` to manage contextual symbol lookups.
    *   Recursive helper functions (`EvalForExprType`, `EvalForStmt`, etc.) traverse the AST, infer types for expressions, and check type compatibility.
    *   Errors encountered during inference are collected.

5.  **`typescope.go` (Type Scope Management):**
    *   `TypeScope` struct: Assists `infer.go` by providing a structured way to look up the type of identifiers.
    *   It uses an `env *decl.Env[decl.Node]` (passed from the loader via `validateFileDecl` and pushed for lexical blocks) to find named declarations.
    *   It also uses `currentComponent` and `currentMethod` context to resolve `self`, method parameters, and component members.
    *   `Get(name string) (*decl.Type, bool)`: The primary lookup method.
    *   `Set(name string, identNode *decl.IdentifierExpr, t *decl.Type)`: Used to record the inferred type of `let`-defined variables.
    *   `ResolveType(td *decl.TypeDecl) *decl.Type`: Resolves a `TypeDecl` AST node to a canonical `*decl.Type`.

6.  **`errors.go` & `imports.go`**: Utilities for error handling and type aliasing from `decl`.

**Process Flow (Loading & Validation):**

1.  `LoadFile(filePath, ...)` is called for a root file.
2.  The `FileResolver` determines the canonical path and provides content.
3.  Cycle detection and max depth checks occur.
4.  The `Parser` parses the content into a `*decl.FileDecl` AST.
5.  `fileDecl.Resolve()` populates its internal maps of local symbols.
6.  Imports are recursively loaded.
7.  `Validate(fileStatus)` is called.
8.  `validateFileDecl` is invoked:
    *   Recursively validates imported files.
    *   Creates a `*decl.Env[decl.Node]` for the type scope.
    *   Calls `AddImportedAliasesToScope(...)`.
    *   Calls `fileDecl.AddToScope(...)`.
    *   If scope population is error-free, it creates an `Inference` object and calls its `Eval(scope)` method.
    *   Any errors are collected in `FileStatus.Errors`.

**Current Status & Recent Work:**

*   The loader can successfully parse files and their imports, resolve basic definitions, and detect cycles.
*   The type inference scoping mechanism (`validateFileDecl`, `AddImportedAliasesToScope`, `TypeScope`) is well-established to correctly construct and populate the environment for `infer.go`.
*   The import mechanism has been made more robust and can now correctly handle importing all definition types, including `native method` declarations.

**Next Steps (for this package):**

*   Thoroughly test the validation flow with various import scenarios.
*   Ensure error reporting is robust and provides clear context.
*   Consider if `FileStatus` needs more fine-grained status tracking.
