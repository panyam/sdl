# Recipe Parser Documentation

## Overview

The SDL Recipe Parser (`recipe-runner.ts`) provides a limited shell script execution environment specifically designed for running SDL demonstrations. It supports a carefully curated subset of shell syntax to ensure safe and predictable execution.

## Supported Syntax

### SDL Commands
- `sdl load <file>` - Load an SDL file
- `sdl use <system>` - Use a specific system
- `sdl gen <generator> [rate]` - Create or update a generator
- `sdl metrics` - Show current metrics
- `sdl set <component.method> <param> <value>` - Set parameters
- `sdl canvas` - Show canvas state

### Built-in Commands
- `echo <text>` - Display text in console
- `read` - Pause execution and wait for user input (Step button)
- `# comment` - Comment lines (ignored during execution)

### Syntax Restrictions

The parser intentionally does NOT support:
- Variables (`$VAR`, `${VAR}`)
- Control flow (`if`, `for`, `while`, `case`)
- Functions or function definitions
- Pipes (`|`)
- Redirections (`>`, `>>`, `<`)
- Background execution (`&`)
- Command substitution (`` `cmd` ``, `$(cmd)`)
- Arithmetic (`$((expr))`)
- Arrays or associative arrays
- Shell expansions (`~`, `*`, `?`, `[...]`)
- Environment variable operations
- Most shell built-ins (`cd`, `pwd`, `export`, etc.)

## Syntax Validation

The parser performs comprehensive syntax validation and reports detailed errors:

```typescript
// Example error detection
if (line.match(/^\s*if\s+/)) {
  throw new Error(`Line ${lineNum}: if statements not supported`);
}
if (line.includes('|')) {
  throw new Error(`Line ${lineNum}: pipes not supported`);
}
```

### Error Categories
1. **Unsupported Control Flow**: if/then/else, for loops, while loops, case statements
2. **Unsupported Features**: pipes, redirections, background execution
3. **Variables**: Any form of variable usage or assignment
4. **Shell Features**: command substitution, arithmetic, expansions
5. **Dangerous Commands**: file operations, network commands, process control

## Monaco Editor Integration

The Recipe language is registered with Monaco editor to provide:

### Syntax Highlighting
- **Keywords** (blue): `sdl`, `echo`, `read`
- **Comments** (green): Lines starting with `#`
- **Strings** (orange): Quoted text
- **Numbers** (light green): Numeric values
- **Types** (cyan): Component.method patterns
- **Attributes** (light blue): Command flags like `--apply-flows`
- **Invalid** (bold red): Unsupported syntax like variables or pipes

### Language Definition
```typescript
monaco.languages.register({ id: 'recipe' });
monaco.languages.setMonarchTokensProvider('recipe', {
  keywords: ['sdl', 'echo', 'read'],
  tokenizer: {
    root: [
      [/^#.*$/, 'comment'],
      [/^sdl\s+(load|use|gen|metrics|set|canvas)/, 'keyword'],
      [/\$\w+/, 'invalid'],  // Variables shown as errors
      [/\|/, 'invalid'],     // Pipes shown as errors
      // ... more patterns
    ]
  }
});
```

## Example Files

### Valid Recipe (`valid-syntax.recipe`)
```bash
# Valid SDL Recipe Syntax Example

# Load an SDL file
sdl load examples/bookstore.sdl

# Use a system
sdl use BookstoreAPI

# Basic commands
echo "Starting traffic generation..."
sdl gen login.login 10

# Pause for user
echo "Press Step to continue..."
read

# Set parameters
sdl set login.authenticate latency 150ms
```

### Invalid Recipe (`test-syntax-errors.recipe`)
```bash
# This file demonstrates syntax errors

# Variables not supported
USER=test
echo "Hello $USER"

# Control flow not supported
if [ -f "test.sdl" ]; then
  sdl load test.sdl
fi

# Pipes not supported
sdl metrics | grep latency
```

## Error Reporting

When invalid syntax is detected:
1. Parser throws descriptive error with line number
2. Error appears in console panel with red highlighting
3. Monaco editor shows invalid syntax in bold red
4. Execution stops at the problematic line

## Implementation Details

The parser uses a line-by-line approach:
1. Skip empty lines and comments
2. Check against unsupported patterns
3. Parse supported SDL commands
4. Execute via SDL API or console output
5. Handle `read` commands by pausing execution

This design ensures recipes are:
- **Safe**: No shell injection or unintended commands
- **Predictable**: Limited syntax means consistent behavior
- **Educational**: Clear about what is and isn't supported
- **Debuggable**: Line-by-line execution with visual feedback