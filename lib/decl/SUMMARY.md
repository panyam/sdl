# SDL Declarations Package Summary (`decl`)

**Purpose:**

This package defines the core data structures for representing the System Design Language (SDL) Abstract Syntax Tree (AST). It provides the building blocks for all language constructs, including type definitions, expressions, statements, and top-level declarations like components and systems. It also defines the `Value` type system used during runtime evaluation and type inference.

**Key Concepts & Components:**

*   **AST Node Hierarchy:**
    *   A comprehensive set of Go structs representing every element of the SDL grammar (e.g., `FileDecl`, `ComponentDecl`, `SystemDecl`, `MethodDecl`, `InstanceDecl`, `LetStmt`, `CallExpr`, `LiteralExpr`, `IdentifierExpr`, `BinaryExpr`, `IfStmt`, `ForStmt`, `DistributeExpr`, `SampleExpr`, `ImportDecl`, `EnumDecl`, etc.).
    *   Most AST nodes embed `NodeInfo` (or a similar structure) to track their source code position (line, column, byte offset) for error reporting and analysis.
    *   Nodes often have methods for resolving internal references (e.g., `FileDecl.Resolve()`, `FileDecl.GetSystem()`, `ComponentDecl.GetMethod()`) and for type inference support (e.g., `SetInferredType()`, `InferredType()`).

*   **Type System (`Type`, `TypeDecl`):**
    *   `TypeDecl`: Represents a type annotation as it appears in the source code (e.g., "Int", "MyComponent", "List[String]"). It can include arguments for generic-like types.
    *   `Type`: Represents the resolved, canonical type information. Includes a `Tag` (e.g., `TypeTagInt`, `TypeTagComponent`, `TypeTagEnum`, `TypeTagOutcomes`) and `Info` (holding specific details like the `ComponentDecl` for a component type, or the element type for `Outcomes`).
    *   Factory functions (e.g., `IntType`, `StrType`, `ComponentType`, `EnumType`, `OutcomesType`, `ListType`, `TupleType`, `RefType`) create canonical `Type` instances.

*   **Runtime Value System (`Value`):**
    *   The `Value` struct is the primary representation of data during SDL program evaluation (interpretation or simulation).
    *   It holds the actual Go data (`Value.Value`) and its corresponding SDL `Type` (`Value.Type`).
    *   Crucially, `Value` now includes a `Time float64` field, intended to accumulate or represent the latency/duration associated with the operation that produced this value. This is particularly important for native function calls and performance modeling.
    *   Helper functions (e.g., `IntValue()`, `BoolValue()`, `NewValue()`) for creating `Value` instances.

*   **Symbol Management (`Env[T]`):**
    *   A generic `Env[T]` type provides a hierarchical environment for managing named symbols (variables, parameters, component/enum definitions) and their associated data (e.g., `Node` for type scopes, `Value` for runtime scopes).
    *   Supports `Push()` for creating nested scopes and `Get()`/`Set()` for symbol lookup and definition.

*   **Interfaces & Utilities:**
    *   `Node` interface (likely): Common interface for all AST nodes, providing methods like `Pos()`, `End()`, `String()`, `PrettyPrint()`.
    *   Helper functions for AST manipulation and inspection.

**Role in the Project:**

*   **Parser (`parser` package):** The parser's primary output is an AST composed of nodes defined in this `decl` package.
*   **Loader (`loader` package):** The loader traverses and validates the AST. Type inference (in `loader/infer.go` and `loader/typescope.go`) heavily relies on and annotates the `Type` and `Value` information within these AST nodes. The loader also uses `Env[Node]` to build type scopes.
*   **Runtime (`runtime` package):** The runtime evaluator (e.g., `SimpleEval`) interprets the AST nodes. It uses the `Value` system for computations and `Env[Value]` for managing runtime variable states.
*   **Commands (`cmd/sdl/commands`):** CLI commands interact with these AST structures to perform validation, listing, description, and diagramming.

**Current Status:**

*   The `decl` package appears to be well-established and provides a comprehensive representation of the SDL.
*   The `Value` system, including the recent addition of the `Time` field, is central to enabling performance-aware evaluation.
*   The type system is sophisticated enough to support features like generics (via `TypeDecl` arguments) and reference types.

**Key Dependencies:**

*   Relied upon by `parser`, `loader`, `runtime`, and `cmd/sdl`.
*   Has minimal external dependencies itself, focusing on core language structure.

**Future Considerations:**

*   Further refinement of the visitor pattern for AST traversal if not already comprehensive.
*   Serialization/deserialization of ASTs (e.g., for caching parsed files or for language server protocols) might be a future need.
