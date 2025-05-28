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
    *   Includes fields for `Tag`, `Info` for type tag and type info.  Type Tag tells what kind of type it i, (eg:
        * SimpleType for int, bool, float, string etc
        * ListType for lists (.Info would be the type of item in the list)
        * ComponentType for denoting component types - where Info would be the corresponding ComponentDecl
        * EnumType - Info is the EnumDecl
        * so on
    *   Singleton instances for basic types (`IntType`, `StrType`, etc.) and factory functions (`ListType`, `OutcomesType`, `EnumType`, `ComponentType`, `EnumType`) for creating `Type` objects.

3.  **Generic Environment (`env.go`):
    *   `Env[T any]` struct: A general-purpose, nestable environment for storing key-value pairs, where keys are strings (names) and values are of generic type `T` (in type inference, `T` is `Node`).
    *   Supports `Get`, `Set`, and `Push` (for creating nested scopes).
    *   Used by the `loader` to build the initial symbol table for a file and by `TypeScope` to manage lexical scoping during inference.

4.  **File Declaration Resolution (`ast.go` - within `FileDecl` methods):
    *   `FileDecl.Resolve()`: Parses the `Declarations` list in a `FileDecl` to populate internal maps (`components`, `enums`, `imports`, `systems`). This makes looking up locally defined symbols efficient.
    *   `FileDecl.Get<Type>()` methods (e.g., `GetEnum`, `GetComponent`) provide access to these resolved local declarations.

**Current Status & Recent Work:**

*   The AST node definitions are largely in place, covering most of the intended SDL syntax.
*   The `Type` system and `Env` structure are defined.

**Next Steps (for this package):**

*   Complete and thoroughly test the refactored type inference logic in `infer.go` to cover all language constructs and edge cases, ensuring it correctly utilizes the new `TypeScope` and `Env`.
*   Ensure `TypeDecl.TypeUsingScope()` is robust for all kinds of type declarations (primitives, generics, named types from scope).
*   Add any missing AST nodes or refine existing ones as DSL features evolve.
