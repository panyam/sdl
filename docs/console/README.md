# SDL Console Documentation

This directory contains documentation and testing guides for the SDL interactive console.

## Overview

The SDL console provides a professional REPL (Read-Eval-Print Loop) environment for interactive SDL system modeling, with features like command history, tab completion, shell integration, and real-time web dashboard synchronization.

## Documentation Files

### Core Documentation
- **[CONSOLE_ENHANCEMENTS.md](CONSOLE_ENHANCEMENTS.md)** - Comprehensive overview of all console features
- **[READLINE_COMPARISON.md](READLINE_COMPARISON.md)** - Analysis of Go readline libraries and selection rationale

### Feature Testing Guides
- **[test_shell_commands.md](test_shell_commands.md)** - Shell command execution with `!` prefix
- **[test_history_features.md](test_history_features.md)** - Persistent command history and auto-suggest improvements  
- **[test_dynamic_system_suggestions.md](test_dynamic_system_suggestions.md)** - Dynamic system name extraction from SDL files

### Testing Scripts
- **[test_console_features.sh](test_console_features.sh)** - Automated testing script for console features

## Quick Start

```bash
# Start the SDL console
./sdl console --port 8080

# Load an SDL file and explore with tab completion
SDL> load examples/contacts/contacts.sdl
SDL> use <TAB>
SDL> use ContactsSystem
SDL[ContactsSystem]> set server.pool.<TAB>
SDL[ContactsSystem]> !git status
```

## Key Features

1. **Rich Tab Completion** - Context-aware completions for commands, files, systems, and parameters
2. **Persistent History** - Commands saved to `~/.sdl_history` across sessions
3. **Shell Integration** - Execute shell commands with `!` prefix
4. **Dynamic System Discovery** - Auto-completion extracts real system names from loaded SDL files
5. **Web Dashboard** - Real-time visualization at http://localhost:8080
6. **Recipe Execution** - Automated command sequences for demonstrations

## Architecture

The console is built using:
- **go-prompt** library for rich REPL experience
- **Canvas API** for SDL file loading and system management
- **WebSocket broadcasting** for real-time dashboard updates
- **Reflection-based** dynamic system discovery

## Development

For console development, see the source files:
- `cmd/sdl/commands/console.go` - Main console implementation
- `console/canvas_web.go` - Web API and WebSocket handling
- `console/canvas.go` - Core Canvas functionality

The testing guides in this directory provide step-by-step instructions for verifying each feature works correctly.