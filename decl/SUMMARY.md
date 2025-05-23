# SDL Declarations Package Summary (`decl` package)

**Purpose:**

This package is central to the System Design Language (SDL) processing. It defines:
1.  The Go structures for all Abstract Syntax Tree (AST) nodes that represent the parsed SDL code.
2.  The SDL type system, including primitive types, composite types (List, Tuple, Outcomes), and how user-defined types like Enums and Components are represented.
3.  The logic for type inference and static semantic checking of the AST.
4.  Supporting structures like `TypeScope` for managing symbols during inference and `Env[Node]` for generic environment/symbol storage.

**Key Concepts & Components:**

1.  **AST Node Definitions (`ast.go` - *implicitly, as definitions are spread across files like `exprs.go`, `stmts.go`, `toplevel.go` within `decl`*):
    *   Defines interfaces (`Node`, `Expr`, `Stmt`, `TopLevelDeclaration`, etc.) and concrete structs (`ComponentDecl`, `SystemDecl`, `EnumDecl`, `ImportDecl`, `MethodDecl`, `ParamDecl`, `UsesDecl`, `InstanceDecl`, `AnalyzeDecl`, `LetStmt`, `IfStmt`, `ReturnStmt`, `CallExpr`, `MemberAccessExpr`, `LiteralExpr`, `IdentifierExpr`, `BinaryExpr`, `UnaryExpr`, `DistributeExpr`, `SampleExpr`, etc.) representing all language constructs.
    *   Each AST node includes `NodeInfo` for tracking its position (line, column) in the source file.
    *   Expressions implement the `Expr` interface, which includes methods for getting/setting declared and inferred types.

2.  **SDL Type System (`types.go`):
    *   `Type` struct: Represents a type in SDL (e.g., `Int`, `Bool`, `List[String]`, `Outcomes[AccessResult]`, an enum type like `HttpStatusCode`, or a component type like `MyCache`).
    *   Includes fields for `Name`, `ChildTypes` (for generics/composites), `IsEnum`, and importantly `OriginalDecl Node` which links a `Type` instance back to its defining AST node (e.g., the `*EnumDecl` for an enum type, or `*ComponentDecl` for a component type).
    *   Singleton instances for basic types (`IntType`, `StrType`, etc.) and factory functions (`ListType`, `OutcomesType`, `EnumType`, `ComponentTypeInstance`) for creating `Type` objects.

3.  **Type Inference (`infer.go`):
    *   `InferTypesForFile(file *FileDecl, rootEnv *Env[Node]) []error`: The main entry point for type checking an entire file's AST. It takes the `FileDecl` and a pre-populated `Env[Node]` (from the loader) containing all visible global and imported declarations.
    *   It performs multiple passes, first resolving types in signatures and parameter defaults, then inferring types for method bodies and system definitions.
    *   Relies on `TypeScope` to manage contextual symbol lookups.
    *   Recursive helper functions (`InferExprType`, `InferTypesForStmt`, etc.) traverse the AST, infer types for expressions, and check type compatibility for operations, assignments, function calls, etc.
    *   Errors encountered during inference are collected and returned.

4.  **Type Scope (`typescope.go`):
    *   `TypeScope` struct: Assists `infer.go` by providing a structured way to look up the type of identifiers.
    *   It uses an `env *Env[Node]` (passed from the loader via `InferTypesForFile` and pushed for lexical blocks) to find named declarations (local `let` variables, global/imported enums, components).
    *   It also uses `currentComponent` and `currentMethod` context to resolve `self`, method parameters, and component members (params/uses).
    *   `Get(name string) (*Type, bool)`: The primary lookup method.
    *   `Set(name string, identNode *IdentifierExpr, t *Type)`: Used to record the inferred type of `let`-defined variables into the current lexical `env`.

5.  **Generic Environment (`env.go`):
    *   `Env[T any]` struct: A general-purpose, nestable environment for storing key-value pairs, where keys are strings (names) and values are of generic type `T` (in type inference, `T` is `Node`).
    *   Supports `Get`, `Set`, and `Push` (for creating nested scopes).
    *   Used by the `loader` to build the initial symbol table for a file and by `TypeScope` to manage lexical scoping during inference.

6.  **File Declaration Resolution (`ast.go` - within `FileDecl` methods):
    *   `FileDecl.Resolve()`: Parses the `Declarations` list in a `FileDecl` to populate internal maps (`components`, `enums`, `imports`, `systems`). This makes looking up locally defined symbols efficient.
    *   `FileDecl.Get<Type>()` methods (e.g., `GetEnum`, `GetComponent`) provide access to these resolved local declarations.

**Current Status & Recent Work:**

*   The AST node definitions are largely in place, covering most of the intended SDL syntax.
*   The `Type` system and `Env` structure are defined.
*   **Significant Refactoring Underway/Completed for Type Inference:**
    *   `TypeScope` has been redesigned to work with an `Env[Node]` passed from the `loader`. This `Env` contains all global and imported symbols, correctly aliased.
    *   `TypeScope.Get` now correctly prioritizes lookups (locals, self, params, env) and can return `*Type` objects that include `OriginalDecl` for enums and components, enabling more accurate member access resolution.
    *   `TypeScope.Set` correctly registers `let`-defined variables in the lexical environment.
    *   `InferTypesForFile` has been updated to use this new `TypeScope` and `Env` structure.
    *   `TypeDecl` has been enhanced with `TypeUsingScope(*TypeScope) *Type` and `SetResolvedType(*Type)` to allow `TypeDecl` nodes (e.g., in parameter lists or return types) to resolve their names to full `*Type` objects within a given scope, caching the result.
    *   Parameter type inference (`InferTypesForParamDecl`) has been improved to infer a parameter's type from its default value if no explicit type is given.
*   The goal of these refactorings is to correctly handle type resolution for imported and aliased symbols, particularly for cases like `Http.StatusOk` where `Http` is an alias for an imported enum.

**Next Steps (for this package):**

*   Complete and thoroughly test the refactored type inference logic in `infer.go` to cover all language constructs and edge cases, ensuring it correctly utilizes the new `TypeScope` and `Env`.
*   Ensure `TypeDecl.TypeUsingScope()` is robust for all kinds of type declarations (primitives, generics, named types from scope).
*   Add any missing AST nodes or refine existing ones as DSL features evolve.
