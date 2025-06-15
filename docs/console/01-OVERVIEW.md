# SDL Console & Server Tutorial - Overview

Welcome to the SDL (System Design Language) console and server tutorial! This guide will walk you through the complete client-server architecture that provides a clean, professional experience for interactive system modeling.

## What You'll Learn

This tutorial is organized into sequential chapters:

1. **[Overview](01-OVERVIEW.md)** - Architecture and concepts (this page)
2. **[Getting Started](02-GETTING-STARTED.md)** - Your first server and console session
3. **[Basic Commands](03-BASIC-COMMANDS.md)** - Core SDL operations (load, use, set, run)
4. **[Traffic Generation](04-TRAFFIC-GENERATION.md)** - Simulating load with generators
5. **[Measurements](05-MEASUREMENTS.md)** - Collecting and analyzing metrics
6. **[Web Dashboard](06-WEB-DASHBOARD.md)** - Real-time visualization and control
7. **[Advanced Features](07-ADVANCED-FEATURES.md)** - History, tab completion, shell commands
8. **[Remote Access](08-REMOTE-ACCESS.md)** - Connecting to remote servers
9. **[Troubleshooting](09-TROUBLESHOOTING.md)** - Common issues and solutions

## Architecture Overview

SDL uses a clean client-server architecture that separates concerns:

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

### SDL Server (`sdl serve`)
The server component hosts:
- **Canvas Simulation Engine** - Loads and executes SDL models
- **Web Dashboard** - Real-time visualization at http://localhost:8080
- **REST API** - HTTP endpoints for all Canvas operations
- **WebSocket Updates** - Live data streaming to dashboard
- **Logging & Statistics** - Server activity and performance metrics

### SDL Console (`sdl console`)
The console component provides:
- **Clean REPL Interface** - No server logs cluttering output
- **Tab Completion** - Smart suggestions for commands, files, systems
- **Command History** - Persistent history across sessions (saved to ~/.sdl_history)
- **Shell Integration** - Execute shell commands with `!` prefix
- **API Communication** - All operations via HTTP to server

## Key Benefits

1. **Clean Experience** - Server logs don't interrupt your REPL workflow
2. **Remote Access** - Console can connect to servers on other machines
3. **Multiple Clients** - Several console clients can share one server
4. **Professional UI** - Web dashboard for visualization and control
5. **Scalable** - Server handles multiple clients and web connections

## Prerequisites

- SDL binary built and available (`go build -o bin/sdl cmd/sdl/main.go`)
- Examples directory with SDL files (e.g., `examples/contacts/contacts.sdl`)
- Two terminal windows (one for server, one for console)

## What's Different from Other Tools

Unlike traditional monolithic CLI tools, SDL's architecture provides:
- **Separation of Concerns** - Display logic separate from simulation engine
- **Multiple Interfaces** - REPL console AND web dashboard simultaneously
- **Real-time Updates** - Changes in console immediately appear in dashboard
- **Remote Capability** - Run simulations on powerful servers, control from laptop

Ready to get started? Continue to **[Getting Started](02-GETTING-STARTED.md)** to launch your first SDL session!
