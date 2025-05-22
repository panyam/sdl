# SDL Parser Package Summary (`sdl/parser`)

## **Purpose:**

This package is responsible for parsing the System Design Language (SDL) text format into an Abstract Syntax Tree (AST) representation. It bridges the gap between the human-readable DSL and the Go structures defined in the `sdl/decl` package that the evaluator uses.

## **Core Components & Files:**

1.  **`grammar.y`**: Defines the context-free grammar for the SDL using `goyacc` syntax. It specifies the language rules, token definitions, operator precedence, and includes semantic actions (Go code within `{}`) that construct the AST nodes during parsing. It relies on a lexer (defined by the `LexerInterface`) to provide tokens. Position tracking logic is embedded within the semantic actions, deriving node spans from the positions of consumed tokens/sub-rules.
2.  **`lexer.go`**: Implements the `Lexer` which conforms to the `LexerInterface`. It reads the input stream character by character, identifies tokens (keywords, identifiers, literals, operators, punctuation) based on defined patterns, handles whitespace and comments, and tracks byte offsets as well as line/column numbers for position reporting. For literal tokens (numbers, strings, booleans, durations) and identifiers, it creates the corresponding `decl.RuntimeValue` and wraps it in a `decl.LiteralExpr` or `decl.IdentifierExpr`, storing position info. It provides methods (`Pos`, `End`, `Position`, `Text`, `Error`) used by the parser and error reporting functions.
3.  **`parser.go`**: Contains the main entry point `Parse(input io.Reader) (*Lexer, *File, error)`. It initializes the `Lexer`, calls the `goyacc`-generated `yyParse` function, and returns the final parsed `*decl.File` AST (stored in the lexer instance during parsing) or an error if parsing fails. It also includes the `yyerror` function used by the parser for basic error reporting, which leverages the lexer's position information.
4.  **`sdl.go`**: **Generated file** (created by running `goyacc -o parser.go -p "yy" grammar.y`). Contains the actual LALR parsing tables and the `yyParse` function implementing the state machine defined by `grammar.y`. *This file should not be edited manually.*
5.  **`imports.go`**: Provides type aliases for the AST node types defined in the `sdl/decl` package (e.g., `type File = decl.FileDecl`, `type ComponentDecl = decl.ComponentDecl`). This avoids circular dependencies and keeps the parser focused on building the structure defined elsewhere.
6.  **`utils.go`**: Contains helper functions used within the parser package, such as `newNodeInfo`, `newLiteralExpr`, `newIdentifierExpr`, `parseDuration`, and the `TokenNode` struct used to pass simple token position information to the parser.
7.  **`lexer_test.go` / `parser_test.go`**: Unit tests verifying the lexer's tokenization and position tracking, and the parser's ability to correctly build the AST structure for various valid and invalid DSL inputs.
8.  **`Makefile`**: Simple makefile to automate the `goyacc` generation step.

## **Process:**

1.  Input DSL text is read by the `Lexer`.
2.  The `Lexer` tokenizes the input, identifying keywords, literals, identifiers, operators, etc., and tracking their positions (byte offset, line, column). It attaches position information and semantic values (like the `RuntimeValue` for literals) to the tokens passed to the parser via the `lval *yySymType` argument.
3.  The `yyParse` function (in the generated `sdl.go`) consumes tokens from the `Lexer`.
4.  As `yyParse` recognizes grammar rules defined in `grammar.y`, it executes the corresponding semantic actions.
5.  These actions construct AST nodes (defined in `sdl/decl` via `imports.go`), setting their fields and combining position information derived from the consumed tokens/sub-nodes.
6.  The final result, a `*decl.File` node representing the entire input, is stored in the lexer instance and returned by the top-level `Parse` function.

## ** Expression Chaining and Unchaining **

A key responsibility of the parser, when processing expressions with multiple operators, is to correctly interpret operator precedence and associativity. The `ChainedExpr` AST node and its `Unchain` method are internal mechanisms within the `parser` package to manage this. The goal is to ensure that the Abstract Syntax Tree (AST) passed to subsequent phases (like type inference and evaluation in the `decl` package) accurately represents the intended order of operations as a tree of `BinaryExpr` nodes.

**1. Chaining (During Parsing):**

*   **Parser's Role:** As the parser consumes tokens according to the grammar rules defined in `grammar.y`, it encounters sequences of operands and operators (e.g., `expr1 + expr2 - expr3`).
*   **Precedence Grouping:** The grammar itself is structured to handle different precedence levels. For operators at the *same* precedence level (e.g., `+` and `-`, or `*` and `/`), the parser collects a linear sequence of operands and the intervening operators into a `ChainedExpr` node.
    *   For instance, an input like `a + b - c` (where `+` and `-` are at the same precedence level) would be parsed into a single `ChainedExpr{ Children: [a, b, c], Operators: ["+", "-"] }`.
*   **Handling Different Precedences:** For expressions involving multiple precedence levels (e.g., `a * b + c`), the parser, guided by its grammar, would typically:
    1.  First parse the higher-precedence operations into `ChainedExpr` nodes (e.g., `a * b` becomes `chain_mul_ab`).
    2.  Call `Unchain()` on these higher-precedence `ChainedExpr` nodes immediately to resolve them into `BinaryExpr` trees (e.g., `chain_mul_ab.Unchain()` results in `binary_expr_ab`).
    3.  These resolved `BinaryExpr` nodes then become operands for the lower-precedence operations. For `binary_expr_ab + c`, a new `ChainedExpr` would be formed: `ChainedExpr{ Children: [binary_expr_ab, c], Operators: ["+"] }`.
    4.  Finally, `Unchain()` is called on this new `ChainedExpr` to produce the final `BinaryExpr` tree for the `+` operation.

**2. Unchaining (via `ChainedExpr.Unchain()` method):**

*   **Purpose:** The `Unchain` method is called by the parser immediately after a `ChainedExpr` node (representing operators of the *same* precedence level) is formed. Its responsibility is to transform this linear chain into a canonical tree of `BinaryExpr` nodes, applying the correct associativity (left, right, or non-associative) for that specific precedence level.
*   **Input:**
    *   A `ChainedExpr` instance (e.g., `Children: [a, b, c]`, `Operators: ["+", "-"]`).
    *   Information about the associativity of the operators in the chain (typically derived from the grammar rules or a precedence table lookup).
*   **Process:**
    *   **Left-Associative** (e.g., `a + b - c` becomes `((a + b) - c)`): The chain is processed from left to right, building nested `BinaryExpr` nodes.
    *   **Right-Associative** (e.g., `a = b = c` becomes `(a = (b = c))`): The chain is processed from right to left.
    *   **Non-Associative** (e.g., `a == b`): The chain must contain exactly one operator. If multiple non-associative operators are found in a single `ChainedExpr`, it's typically a parsing error (as they shouldn't be chained directly at the same level).
*   **Output:** The `ChainedExpr`'s `UnchainedExpr` field is populated with the root of the newly constructed `BinaryExpr` tree (or the single operand if the chain was trivial, e.g., just `a`). The `NodeInfo` for each new `BinaryExpr` is set to span from its leftmost operand to its rightmost operand.

3. **Parser's End Result:**

By the time the parser completes its work and produces the final AST for the `decl` package, all `ChainedExpr` nodes created during intermediate parsing steps should have been processed by `Unchain()`. The `decl` package therefore receives an AST where complex expressions are already structured as trees of `BinaryExpr` nodes, reflecting the correct operator precedence and associativity. This significantly simplifies the work required for type inference and evaluation in the `decl` package, as it doesn't need to re-implement precedence and associativity logic.

## **Current Status:**

*   The parser is functional and capable of parsing the defined SDL grammar into the AST structure specified in `sdl/decl`.
*   The lexer correctly identifies tokens, including literals (populating `RuntimeValue`), keywords, and operators.
*   Basic position tracking (byte offsets, line/column) is implemented in the lexer.
*   Semantic actions in the grammar are implemented to build the AST and attempt to propagate position information.
*   Unit tests exist for both the lexer and parser, covering various constructs and basic error cases.
*   **Refinement Needed:**
    *   Thorough validation and potential correction of position spans calculated in `grammar.y` actions.
    *   More robust error reporting from `yyerror`, potentially showing expected tokens.
    *   More comprehensive testing, especially for edge cases and complex position scenarios.
