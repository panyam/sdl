# SDL Command Line Interface (CLI) Package Summary (`cmd/sdl`)

**Purpose:**

This package implements the main command-line interface (CLI) tool for the System Design Language (SDL) project, named `sdl`. It allows users to interact with the SDL toolchain to parse, validate, inspect, and eventually analyze SDL model files.

**Core Structure & Files:**

*   **`main.go`**: The entry point for the CLI application. It simply calls `commands.Execute()` to run the root Cobra command.
*   **`commands/` (sub-package):** Contains the definitions for all CLI commands and their logic.
    *   **`root.go`**: Defines the root command (`sdl`) using the `github.com/spf13/cobra` library. It sets up persistent flags, such as `--file` (`-f`) for specifying the input DSL file path, which is stored in the global `dslFilePath` variable. It also provides an `AddCommand` helper for subcommands to register themselves.
    *   **`validate.go`**: Implements the `sdl validate <dsl_file_path...>` command. This command parses one or more DSL files and performs semantic checks, including type inference. It uses the `loader.Loader` (with `SDLParserAdapter` and `DefaultFileResolver`) to load and then validate the files. It reports success or lists errors.
    *   **`list.go`**: Implements the `sdl list <entity_type>` command. It lists defined entities (components, systems, analyses, enums) from a specified DSL file. It parses the file and iterates through the AST to find the requested entities. Supports JSON output.
    *   **`plot.go`**: Implements the `sdl plot <system_name> <component_name> <method_nam>` command. Intended to generate plots from making multiple calls to a given endpoint to observe its performance (latency etc).
    *   **`describe.go`**: Implements the `sdl describe <entity_type> <entity_name>` command. It shows detailed information about a specific component, system, or analysis from a DSL file. It parses the file and attempts to pretty-print or show JSON details of the requested AST node.
    *   **`diagram.go`**: (Currently Placeholder) Implements the `sdl diagram <diagram_type> <system_name>` command. Intended to generate static (component structure) or dynamic (interaction sequence) diagrams. Current implementation provides mock diagram content.
    *   **`run.go`**: (Currently Placeholder) Implements the `sdl run <system_name> [<analysis_name>]` command. Intended to execute analyses defined in a system. Current implementation provides mock results.
    *   **`trace.go`**: (Currently Placeholder) Implements the `sdl trace <system_name> <method_call_string>` command. Intended to trace the execution of a specific operation. Current implementation provides mock trace output.
    *   **`utils.go`**: Contains utility code for the commands, such as `SDLParserAdapter` which adapts the `parser.Parse` function to the `loader.Parser` interface required by the `loader.Loader`.

**Key Functionality & Features:**

*   **Modular Commands:** Uses Cobra for a structured command system (`validate`, `list`, `describe`, etc.).
*   **Global File Flag:** A persistent `-f/--file` flag is available for most commands to specify the primary DSL input file.
*   **Validation:** The `validate` command is the most developed functional command, performing parsing, import resolution, and semantic checks (including the ongoing type inference work).
*   **AST Inspection:** `list` and `describe` commands allow users to inspect the contents of their DSL files by parsing and navigating the AST.
*   **Placeholders for Advanced Features:** `diagram`, `plot`, `run`, and `trace` commands exist as placeholders, indicating future capabilities that will likely depend on the DSL VM/execution engine.

**Workflow for `validate` command:**

1.  User runs `sdl validate -f my_system.sdl another_file.sdl`.
2.  The `validateCmd.Run` function is executed.
3.  An `SDLParserAdapter` (wrapping `parser.Parse`) and a `DefaultFileResolver` are created.
4.  A `loader.NewLoader` is instantiated with the parser, resolver, and an import depth limit.
5.  `sdlLoader.LoadFiles(args...)` is called:
    *   This recursively parses `my_system.sdl` and `another_file.sdl` and all their imports, creating a graph of `FileStatus` objects, each containing a `*decl.FileDecl` (AST).
    *   Any parsing or file resolution errors are collected in the respective `FileStatus.Errors`.
6.  After loading, the `validate` command iterates through the results. For each successfully loaded root file, it would implicitly trigger the `loader.Validate(fileStatus)` call (as per the `validate` command's logic using `sdlLoader.LoadFiles` and then checking errors).
    *   `loader.Validate` orchestrates the scope creation (populating with local and aliased imported symbols) and then calls `decl.InferTypesForFile`.
7.  The command reports overall success or failure, listing any errors found during parsing, loading, or type inference.

**Current Status:**

*   The CLI framework using Cobra is well-established.
*   `validate`, `list`, and `describe` provide core functionality for working with SDL files at the AST level.
*   The `validate` command is a key driver for testing the loader and type inference capabilities.
*   Most analytical and visualization commands (`run`, `diagram`, `plot`, `trace`) are placeholders and their full implementation depends on the DSL execution engine.

**Next Steps (for this package):**

*   Integrate the DSL VM/execution engine into the `run`, `diagram`, `plot`, and `trace` commands to replace mock implementations with actual functionality.
*   Enhance error reporting for all commands to be more user-friendly and informative.
*   Add more sophisticated output formatting options (e.g., different diagram formats, plot customization).
