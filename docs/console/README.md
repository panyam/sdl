# SDL Console Documentation

This directory contains comprehensive documentation for SDL's client-server console architecture.

## Overview

SDL provides a professional client-server architecture that separates the interactive console from the simulation engine. This creates a clean, scalable, and powerful environment for system modeling and performance testing.

## Sequential Tutorial (Start Here!)

**New users should follow the sequential tutorial:**

1. **[Overview](01-OVERVIEW.md)** - Architecture and concepts
2. **[Getting Started](02-GETTING-STARTED.md)** - Your first server and console session  
3. **[Basic Commands](03-BASIC-COMMANDS.md)** - Core SDL operations
4. **[Traffic Generation](04-TRAFFIC-GENERATION.md)** - Simulating load with generators
5. **[Measurements](05-MEASUREMENTS.md)** - Collecting and analyzing metrics
6. **[Web Dashboard](06-WEB-DASHBOARD.md)** - Real-time visualization and control
7. **[Advanced Features](07-ADVANCED-FEATURES.md)** - History, tab completion, automation
8. **[Remote Access](08-REMOTE-ACCESS.md)** - Connecting to remote servers
9. **[Troubleshooting](09-TROUBLESHOOTING.md)** - Common issues and solutions

## Architecture

SDL uses a clean client-server split:

```
┌─────────────────┐    HTTP API    ┌─────────────────┐
│  sdl console    │ ◄─────────────► │   sdl serve     │
│  (Client)       │                │   (Server)      │
│                 │                │                 │
│ • Clean REPL    │                │ • Canvas Engine │
│ • Tab Complete  │                │ • Web Dashboard │
│ • History       │                │ • Logs & Stats  │
│ • No Logs       │                │ • REST API      │
└─────────────────┘                └─────────────────┘
```

## Quick Start

```bash
# Terminal 1: Start SDL server
./bin/sdl serve

# Terminal 2: Connect console client
./bin/sdl console

# Load and explore with tab completion
SDL> load examples/contacts/contacts.sdl
SDL> use <TAB>
SDL> use ContactsSystem
SDL[ContactsSystem]> gen add load1 server.HandleLookup 10
SDL[ContactsSystem]> gen start load1
```

## Key Features

1. **Client-Server Architecture** - Clean separation of concerns
2. **Real-time Web Dashboard** - Live visualization at http://localhost:8080
3. **Traffic Generation** - Sustained load simulation with generators
4. **Measurement System** - DuckDB-based metrics collection and analysis
5. **Remote Access** - Connect to servers on other machines
6. **Rich Tab Completion** - Context-aware command suggestions
7. **Persistent History** - Commands saved across sessions
8. **Shell Integration** - Execute shell commands with `!` prefix
9. **Recipe Automation** - Scripted command sequences

## Additional Documentation

### Legacy Feature Documentation
- **[CONSOLE_ENHANCEMENTS.md](CONSOLE_ENHANCEMENTS.md)** - Historical feature overview
- **[READLINE_COMPARISON.md](READLINE_COMPARISON.md)** - Go readline library analysis

### Testing and Development Guides  
- **[test_shell_commands.md](test_shell_commands.md)** - Shell integration testing
- **[test_history_features.md](test_history_features.md)** - History system testing
- **[test_dynamic_system_suggestions.md](test_dynamic_system_suggestions.md)** - Tab completion testing
- **[test_console_features.sh](test_console_features.sh)** - Automated test script

## Development

Core implementation files:
- `cmd/sdl/commands/console.go` - Console client implementation
- `cmd/sdl/commands/serve.go` - Server command implementation  
- `console/canvas_web.go` - Web API and WebSocket handling
- `console/canvas.go` - Core Canvas simulation engine

## Support

For issues or questions:
- Follow the **[Troubleshooting](09-TROUBLESHOOTING.md)** guide
- Check existing documentation in this directory
- Review server and console logs for error details