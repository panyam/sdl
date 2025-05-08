# SDL Parser Package Summary (`sdl/parser`)

**Purpose:**

This package is responsible for parsing the System Design Language (SDL) text format into an Abstract Syntax Tree (AST) representation. It bridges the gap between the human-readable DSL and the Go structures defined in the `sdl/decl` package that the evaluator uses.

**Core Components & Files:**

1.  **`grammar.y`**: Defines the context-free grammar for the SDL using `goyacc` syntax. It specifies the language rules, token definitions, operator precedence, and includes semantic actions (Go code within `{}`) that construct the AST nodes during parsing. It relies on a lexer (defined by the `LexerInterface`) to provide tokens. Position tracking logic is embedded within the semantic actions, deriving node spans from the positions of consumed tokens/sub-rules.
2.  **`lexer.go`**: Implements the `Lexer` which conforms to the `LexerInterface`. It reads the input stream character by character, identifies tokens (keywords, identifiers, literals, operators, punctuation) based on defined patterns, handles whitespace and comments, and tracks byte offsets as well as line/column numbers for position reporting. For literal tokens (numbers, strings, booleans, durations) and identifiers, it creates the corresponding `decl.RuntimeValue` and wraps it in a `decl.LiteralExpr` or `decl.IdentifierExpr`, storing position info. It provides methods (`Pos`, `End`, `Position`, `Text`, `Error`) used by the parser and error reporting functions.
3.  **`parser.go`**: Contains the main entry point `Parse(input io.Reader) (*Lexer, *File, error)`. It initializes the `Lexer`, calls the `goyacc`-generated `yyParse` function, and returns the final parsed `*decl.File` AST (stored in the lexer instance during parsing) or an error if parsing fails. It also includes the `yyerror` function used by the parser for basic error reporting, which leverages the lexer's position information.
4.  **`sdl.go`**: **Generated file** (created by running `goyacc -o parser.go -p "yy" grammar.y`). Contains the actual LALR parsing tables and the `yyParse` function implementing the state machine defined by `grammar.y`. *This file should not be edited manually.*
5.  **`imports.go`**: Provides type aliases for the AST node types defined in the `sdl/decl` package (e.g., `type File = decl.FileDecl`, `type ComponentDecl = decl.ComponentDecl`). This avoids circular dependencies and keeps the parser focused on building the structure defined elsewhere.
6.  **`utils.go`**: Contains helper functions used within the parser package, such as `newNodeInfo`, `newLiteralExpr`, `newIdentifierExpr`, `parseDuration`, and the `TokenNode` struct used to pass simple token position information to the parser.
7.  **`lexer_test.go` / `parser_test.go`**: Unit tests verifying the lexer's tokenization and position tracking, and the parser's ability to correctly build the AST structure for various valid and invalid DSL inputs.
8.  **`Makefile`**: Simple makefile to automate the `goyacc` generation step.

**Process:**

1.  Input DSL text is read by the `Lexer`.
2.  The `Lexer` tokenizes the input, identifying keywords, literals, identifiers, operators, etc., and tracking their positions (byte offset, line, column). It attaches position information and semantic values (like the `RuntimeValue` for literals) to the tokens passed to the parser via the `lval *yySymType` argument.
3.  The `yyParse` function (in the generated `sdl.go`) consumes tokens from the `Lexer`.
4.  As `yyParse` recognizes grammar rules defined in `grammar.y`, it executes the corresponding semantic actions.
5.  These actions construct AST nodes (defined in `sdl/decl` via `imports.go`), setting their fields and combining position information derived from the consumed tokens/sub-nodes.
6.  The final result, a `*decl.File` node representing the entire input, is stored in the lexer instance and returned by the top-level `Parse` function.

**Current Status:**

*   The parser is functional and capable of parsing the defined SDL grammar into the AST structure specified in `sdl/decl`.
*   The lexer correctly identifies tokens, including literals (populating `RuntimeValue`), keywords, and operators.
*   Basic position tracking (byte offsets, line/column) is implemented in the lexer.
*   Semantic actions in the grammar are implemented to build the AST and attempt to propagate position information.
*   Unit tests exist for both the lexer and parser, covering various constructs and basic error cases.
*   **Refinement Needed:**
    *   Thorough validation and potential correction of position spans calculated in `grammar.y` actions.
    *   More robust error reporting from `yyerror`, potentially showing expected tokens.
    *   More comprehensive testing, especially for edge cases and complex position scenarios.
