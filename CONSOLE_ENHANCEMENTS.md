# SDL Console Enhancements with go-prompt

## Overview

The SDL console has been enhanced with the **go-prompt** library to provide a professional REPL experience with rich auto-completion, command history, and intelligent suggestions.

## New Features

### 1. **Rich Tab Completion**
- **Commands**: Type partial commands and press Tab to see all matching commands with descriptions
  ```
  SDL> lo<TAB>
  load    Load an SDL file
  ```

### 2. **Context-Aware Completions**
- **File paths**: When typing `load`, Tab shows `.sdl` files
  ```
  SDL> load ex<TAB>
  examples/           Directory
  examples/contacts/  Directory
  ```

- **System names**: After `use`, Tab shows available systems
  ```
  SDL> use <TAB>
  ContactsSystem    System definition
  KafkaSystem       System definition
  ```

- **Parameter paths**: When using `set`, Tab shows common parameters
  ```
  SDL> set server.<TAB>
  server.pool.ArrivalRate    Request arrival rate
  server.pool.Size           Connection pool size
  server.db.pool.Size        Database pool size
  ```

- **Parameter values**: Context-aware value suggestions
  ```
  SDL> set server.db.CacheHitRate <TAB>
  0.4    40% hit rate
  0.6    60% hit rate
  0.8    80% hit rate
  0.95   95% hit rate
  ```

### 3. **Dynamic Prompt**
The prompt changes to show the active system:
```
SDL> load examples/contacts/contacts.sdl
SDL> use ContactsSystem
SDL[ContactsSystem]> 
```

### 4. **Navigation Keys**
- **↑↓**: Navigate through command history
- **←→**: Move cursor within the line
- **Tab**: Auto-complete commands, paths, and parameters
- **Ctrl+A/E**: Jump to beginning/end of line
- **Ctrl+K/U**: Delete to end/beginning of line
- **Ctrl+W**: Delete word before cursor
- **Ctrl+D**: Exit console

### 5. **Enhanced Help**
The help command now includes navigation instructions:
```
SDL> help
Available commands:

  help                        Show this help message
  load <file_path>           Load an SDL file
  use <system_name>          Activate a system from loaded file
  set <path> <value>         Set parameter (e.g., server.pool.ArrivalRate 10)
  run <var> <target> [runs]  Run simulation (default 1000 runs)
  execute <recipe_file>      Execute commands from a recipe file
  state                      Show current Canvas state
  exit, quit                 Exit the console (or press Ctrl+D)

Navigation:
  ↑↓                         Navigate through command history
  ←→                         Move cursor within line
  Tab                        Auto-complete commands, paths, and parameters
  ...
```

## Usage Examples

### Basic Workflow with Tab Completion
```bash
# Start the console
./sdl console --port 8080

# Load a file (use Tab to navigate directories)
SDL> load ex<TAB>
SDL> load examples/<TAB>
SDL> load examples/contacts/<TAB>
SDL> load examples/contacts/contacts.sdl

# Activate a system
SDL> use <TAB>
SDL> use ContactsSystem

# Set parameters with intelligent suggestions
SDL[ContactsSystem]> set <TAB>
SDL[ContactsSystem]> set server.pool.<TAB>
SDL[ContactsSystem]> set server.pool.ArrivalRate <TAB>
SDL[ContactsSystem]> set server.pool.ArrivalRate 15

# Run simulations
SDL[ContactsSystem]> run <TAB>
SDL[ContactsSystem]> run latest <TAB>
SDL[ContactsSystem]> run latest server.HandleLookup <TAB>
SDL[ContactsSystem]> run latest server.HandleLookup 5000
```

### Recipe File Execution
```bash
SDL> execute <TAB>
SDL> execute examples/<TAB>
SDL> execute examples/demo_recipe.txt
```

## Technical Details

### Implementation
- Uses **github.com/c-bata/go-prompt** v0.2.6
- Actively maintained library with 5.3k GitHub stars
- Provides Emacs-like key bindings
- Cross-platform support (Windows, macOS, Linux)

### Custom Completers
The implementation includes several custom completion functions:
- `getCommandSuggestions()`: Main command completions
- `getFileSuggestions()`: File system navigation
- `getSystemSuggestions()`: System name completions
- `getParameterPathSuggestions()`: Common parameter paths
- `getValueSuggestions()`: Context-aware value suggestions
- `getTargetSuggestions()`: Simulation target completions

## Benefits

1. **Improved User Experience**: No more arrow key gibberish (^[[1;2D)
2. **Faster Command Entry**: Tab completion reduces typing
3. **Discovery**: Users can explore available options easily
4. **Professional Feel**: Suitable for conference demonstrations
5. **Reduced Errors**: Completions help avoid typos
6. **Context Awareness**: Smart suggestions based on current state

## Future Enhancements

Potential improvements that could be added:
- Syntax highlighting for SDL syntax
- Multi-line command support for complex operations
- Command aliases and macros
- Persistent history across sessions
- Custom color themes