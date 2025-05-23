# SDL Parser Package Summary (`parser` package)

**Purpose:**

This package is responsible for parsing the System Design Language (SDL) text format into an Abstract Syntax Tree (AST) representation. The AST nodes themselves are defined in the `sdl/decl` package. The parser bridges the gap between the human-readable DSL and these Go structures that the loader and type inferencer use.

**Core Components & Files:**

1.  **`grammar.y`**: Defines the context-free grammar for the SDL using `goyacc` syntax. It specifies the language rules, token definitions, operator precedence, and includes semantic actions (Go code within `{}`) that construct the AST nodes (from `sdl/decl`) during parsing. It relies on a lexer (defined by the `LexerInterface`) to provide tokens. Position tracking logic is embedded within the semantic actions, deriving node spans from the positions of consumed tokens/sub-rules.
2.  **`lexer.go`**: Implements the `Lexer` which conforms to the `LexerInterface`. It reads the input stream character by character, identifies tokens (keywords, identifiers, literals, operators, punctuation) based on defined patterns, handles whitespace and comments, and tracks byte offsets as well as line/column numbers for position reporting. For literal tokens (numbers, strings, booleans, durations) and identifiers, it creates the corresponding `decl.RuntimeValue` and wraps it in a `decl.LiteralExpr` or `decl.IdentifierExpr`, storing position info.
3.  **`parser.go`**: **Generated file** (created by running `goyacc -o parser.go -p "SDL" grammar.y`). Contains the actual LALR parsing tables and the `SDLParse` function implementing the state machine defined by `grammar.y`. *This file should not be edited manually.* It also contains the top-level `Parse(input io.Reader) (*Lexer, *FileDecl, error)` function which initializes the lexer, calls `SDLParse`, and returns the result or error.
4.  **`chainexpr.go`**: Contains the `ChainedExpr` AST node and its `Unchain` method. This is an internal mechanism used during parsing to handle sequences of binary operators at the same precedence level. The `Unchain` method converts this linear chain into a canonical tree of `decl.BinaryExpr` nodes, applying the correct associativity, before the final AST is passed to subsequent phases.
5.  **`llparser.go` and `lltest_utils.go`**: These files appear to be related to an **alternative or experimental LL parser implementation**. The primary parser used by `cmd/sdl/commands/validate.go` (via `loader`) is the `goyacc`-generated one. The LL parser might be for specific use cases or future development. *`(Self-correction: The main parsing is done by the yacc parser. These LL files might be for specific sub-parsing tasks or an alternative approach not currently integrated into the main loading flow.)`*
6.  **`utils.go`**: Contains helper functions used within the parser package, such as `newNodeInfo`, `newLiteralExpr`, `newIdentifierExpr`, `parseDuration`, and the `TokenNode` struct.
7.  **`imports.go`**: Provides type aliases for the AST node types defined in the `sdl/decl` package (e.g., `type FileDecl = decl.FileDecl`). This avoids circular dependencies.
8.  **`Makefile`**: Simple makefile to automate the `goyacc` generation step for `parser.go`.

**Process (for `goyacc` parser):**

1.  Input DSL text is read by the `Lexer` (`lexer.go`).
2.  The `Lexer` tokenizes the input, attaching position information and semantic values (like `decl.RuntimeValue` for literals) to tokens passed to the parser.
3.  The `SDLParse` function (in the generated `parser.go`) consumes tokens from the `Lexer`.
4.  As `SDLParse` recognizes grammar rules from `grammar.y`, it executes semantic actions.
5.  These actions construct AST nodes (from `sdl/decl`), setting their fields and position information.
6.  Intermediate `ChainedExpr` nodes are resolved into `BinaryExpr` trees via `Unchain`.
7.  The final result, a `*decl.FileDecl` node, is returned by the top-level `Parse` function.

**Current Status:**

*   The `goyacc`-based parser is functional and capable of parsing the defined SDL grammar into the AST structure specified in `sdl/decl`.
*   The lexer correctly identifies tokens and handles position tracking.
*   Semantic actions in the grammar build the AST.
*   The `ChainedExpr` mechanism handles operator precedence and associativity for binary expressions.
*   The role and integration status of the `llparser.go` components are less clear in the main file processing flow but exist within the package.

**Next Steps (for this package):**

*   Ensure `grammar.y` is up-to-date with any DSL syntax changes.
*   Thoroughly test parsing of all language constructs, including edge cases and error conditions.
*   Clarify the role or complete the integration of the LL parser components if they are intended for broader use.
