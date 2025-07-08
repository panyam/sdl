# SystemDetailTool

A comprehensive Go tool for SDL system compilation, recipe parsing, and execution with WASM browser integration. This tool demonstrates the "minitools" pattern for creating focused, environment-agnostic tools that work across CLI, server, and WASM contexts.

## Architecture Overview

The SystemDetailTool follows a modular architecture designed for maximum reusability:

```
tools/systemdetail/
├── tool.go              # Core tool implementation
├── sdl.go               # SDL compilation and @stdlib support
├── wasm/                # WASM bindings
│   └── main.go          # WASM entry point
├── cmd/                 # Command-line interfaces
│   └── testcli.go       # Interactive CLI for testing
├── Makefile             # Build automation
├── tool_test.go         # Comprehensive test suite
└── README.md            # This documentation
```

## Key Features

### 1. **Environment-Agnostic Design**
- Works in CLI, server, and WASM contexts
- No filesystem dependencies in WASM mode
- Consistent API across all environments

### 2. **SDL Compilation with @stdlib Support**
- Validates SDL syntax and imports
- Blocks local imports (`./file.sdl`) for security
- Allows @stdlib imports via memory filesystem
- Generates compilation results with system metadata

### 3. **Recipe Parsing and Execution**
- Parses `.recipe` files with TypeScript parity
- Supports `echo`, `sdl`, `pause`, `read` commands
- Validates recipe syntax and security constraints
- Tracks execution state for step-through debugging

### 4. **WASM Browser Integration**
- Dedicated WASM module (not monolithic)
- JavaScript callbacks for UI integration
- JSON-serialized results for web consumption
- Cache-busted loading for development

## Usage Patterns

### CLI Usage

```bash
# Build and test the tool
make test
make wasm

# Run interactive CLI
go run cmd/testcli.go

# Available commands:
# load-bitly      - Load Bitly SDL with @stdlib imports
# use <system>    - Activate a specific system
# recipe          - Load and parse recipe file
# info            - Show system information
# status          - Show current tool status
```

### Go API Usage

```go
// Create and initialize tool
tool := systemdetail.NewSystemDetailTool()

// Set up callbacks for output
tool.SetCallbacks(&systemdetail.Callbacks{
    OnError: func(msg string) { fmt.Printf("ERROR: %s\n", msg) },
    OnInfo:  func(msg string) { fmt.Printf("INFO: %s\n", msg) },
    OnSuccess: func(msg string) { fmt.Printf("SUCCESS: %s\n", msg) },
})

// Initialize with system data
err := tool.Initialize("bitly", sdlContent, recipeContent)

// Compile SDL
err = tool.SetSDLContent(sdlContent)
result := tool.GetCompileResult()

// Parse recipe
err = tool.SetRecipeContent(recipeContent)
execState := tool.GetExecState()

// Use a system
err = tool.UseSystem("Bitly")
```

### WASM Browser Usage

```typescript
// Load WASM module
const tool = new WASMSystemDetailTool();

// Set up callbacks
tool.setCallbacks({
  onError: (msg) => console.error(msg),
  onInfo: (msg) => console.info(msg),
  onSuccess: (msg) => console.log(msg)
});

// Initialize with system data
await tool.initialize("bitly", sdlContent, recipeContent);

// Compile SDL
await tool.setSDLContent(sdlContent);
const result = await tool.getCompileResult();

// Parse recipe
await tool.setRecipeContent(recipeContent);
const execState = await tool.getExecState();
```

## Core Components

### 1. **SystemDetailTool** (`tool.go`)

Main tool implementation with these key methods:

```go
type SystemDetailTool struct {
    systemID      string
    sdlContent    string
    recipeContent string
    callbacks     *Callbacks
    compileResult *CompileResult
    execState     *RecipeExecState
}

// Core methods
func (t *SystemDetailTool) Initialize(systemID, sdlContent, recipeContent string) error
func (t *SystemDetailTool) SetSDLContent(content string) error
func (t *SystemDetailTool) SetRecipeContent(content string) error
func (t *SystemDetailTool) GetCompileResult() *CompileResult
func (t *SystemDetailTool) GetExecState() *RecipeExecState
func (t *SystemDetailTool) UseSystem(systemName string) error
```

### 2. **SDL Compilation** (`sdl.go`)

Handles SDL compilation with @stdlib support:

```go
// Memory filesystem for @stdlib imports
func createStdlibFileSystem() loader.FileSystem

// Custom filesystem wrapper for @stdlib prefix handling
type StdlibPrefixFS struct {
    fs loader.FileSystem
}

// Security validation
func (t *SystemDetailTool) validateNoLocalImports(content string) error
```

### 3. **WASM Bindings** (`wasm/main.go`)

Exposes tool methods to JavaScript:

```go
// JavaScript-accessible functions
js.Global().Set("newSystemDetailTool", js.FuncOf(newSystemDetailTool))
js.Global().Set("setSDLContent", js.FuncOf(setSDLContent))
js.Global().Set("setRecipeContent", js.FuncOf(setRecipeContent))
js.Global().Set("getSystemInfo", js.FuncOf(getSystemInfo))
js.Global().Set("getCompileResult", js.FuncOf(getCompileResult))
js.Global().Set("getExecState", js.FuncOf(getExecState))
```

### 4. **Recipe Parser Integration**

Uses shared recipe parser from `tools/shared/recipe`:

```go
// Parse recipe content
parseResult := recipe.ParseRecipe(t.recipeContent)

// Convert to execution state
execState := &RecipeExecState{
    TotalSteps: len(executableCommands),
    Steps:      steps,
    Mode:       "step",
}
```

## Data Structures

### CompileResult
```go
type CompileResult struct {
    Success bool     `json:"success"`
    Errors  []string `json:"errors"`
    Systems []string `json:"systems"`
}
```

### RecipeExecState
```go
type RecipeExecState struct {
    IsRunning   bool         `json:"isRunning"`
    CurrentStep int          `json:"currentStep"`
    TotalSteps  int          `json:"totalSteps"`
    Steps       []RecipeStep `json:"steps"`
    Mode        string       `json:"mode"`
}
```

### RecipeStep
```go
type RecipeStep struct {
    Index      int      `json:"index"`
    LineNumber int      `json:"lineNumber"`
    Command    string   `json:"command"`
    Args       []string `json:"args"`
    Status     string   `json:"status"`
    Output     string   `json:"output,omitempty"`
}
```

## Security Model

### SDL Security
- **Local imports blocked**: Prevents `./file.sdl` and `../file.sdl` imports
- **@stdlib imports allowed**: Safe imports from standard library
- **Memory filesystem**: No access to local filesystem in WASM mode

### Recipe Security
- **Shell syntax blocked**: Prevents pipes, redirections, variables
- **Command whitelist**: Only `echo`, `sdl`, `pause`, `read` allowed
- **Argument validation**: Prevents command injection

## Testing

### Test Structure
```go
func TestNewSystemDetailTool(t *testing.T)           // Basic initialization
func TestSetSDLContent_Valid(t *testing.T)           // SDL compilation
func TestSetSDLContent_LocalImportsRejected(t *testing.T) // Security
func TestSetSDLContent_StdlibImportsAllowed(t *testing.T) // @stdlib support
func TestValidateNoLocalImports(t *testing.T)        // Import validation
func TestUseSystem(t *testing.T)                     // System activation
```

### Running Tests
```bash
# Run all tests
make test

# Run specific test
go test -v -run TestSetSDLContent_Valid

# Test recipe parser
go test -v ../shared/recipe
```

## Build System

### Makefile Targets
```makefile
wasm:     # Build WASM binary to web/dist/wasm/systemdetail.wasm
clean:    # Clean build artifacts
test:     # Run all tests
dev:      # Build with tests (development)
prod:     # Build for production
```

### Build Configuration
- **WASM Target**: `web/dist/wasm/systemdetail.wasm`
- **Go Version**: Compatible with Go 1.19+
- **Dependencies**: Uses `github.com/panyam/sdl/tools/shared/recipe`

## Creating New Tools

Use this SystemDetailTool as a template for new tools:

### 1. **Directory Structure**
```
tools/newtool/
├── tool.go              # Core tool implementation
├── wasm/main.go         # WASM bindings
├── cmd/testcli.go       # CLI interface
├── tool_test.go         # Test suite
├── Makefile             # Build automation
└── README.md            # Documentation
```

### 2. **Core Tool Pattern**
```go
type NewTool struct {
    // Tool-specific fields
}

func NewNewTool() *NewTool {
    return &NewTool{}
}

func (t *NewTool) Initialize(params...) error {
    // Initialize tool
}

func (t *NewTool) ProcessData(data string) error {
    // Process input data
}

func (t *NewTool) GetResult() *Result {
    // Return processed result
}
```

### 3. **WASM Bindings Pattern**
```go
func main() {
    tool := newtool.NewNewTool()
    
    js.Global().Set("newNewTool", js.FuncOf(newNewTool))
    js.Global().Set("processData", js.FuncOf(processData))
    js.Global().Set("getResult", js.FuncOf(getResult))
    
    select {} // Keep running
}
```

### 4. **Makefile Pattern**
```makefile
.PHONY: wasm clean test

wasm:
	@echo "Building NewTool WASM binary..."
	@mkdir -p ../../web/dist/wasm
	GOOS=js GOARCH=wasm go build -o ../../web/dist/wasm/newtool.wasm ./wasm/

test:
	@echo "Running NewTool tests..."
	go test -v .
```

## Best Practices

### 1. **Environment Agnostic**
- No filesystem access in core logic
- Use memory filesystems for WASM
- Consistent API across CLI/server/WASM

### 2. **Security First**
- Validate all inputs
- Block dangerous operations
- Use whitelists, not blacklists

### 3. **Testing**
- Comprehensive test coverage
- Test all public methods
- Include security test cases

### 4. **Documentation**
- Document all public APIs
- Include usage examples
- Explain security model

### 5. **Build Automation**
- Use Makefiles for consistency
- Include test and clean targets
- Target specific WASM locations

## Dependencies

- `github.com/panyam/sdl/tools/shared/recipe` - Recipe parsing
- `github.com/panyam/sdl/console` - SDL compilation
- `github.com/panyam/sdl/loader` - File system abstraction
- `syscall/js` - WASM JavaScript bindings

## Future Enhancements

- [ ] Real-time recipe execution with progress tracking
- [ ] Recipe step breakpoints and debugging
- [ ] System metrics collection during execution
- [ ] Recipe execution history and logging
- [ ] Performance profiling integration
- [ ] Multi-system orchestration support

## Contributing

When contributing to this tool:

1. Follow the existing patterns and architecture
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure WASM compatibility for new methods
5. Validate security implications of changes

This tool serves as a reference implementation for the "minitools" pattern and should maintain high code quality and comprehensive documentation standards.