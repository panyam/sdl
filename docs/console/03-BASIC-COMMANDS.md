# SDL Console & Server Tutorial - Basic Commands

This chapter covers the core SDL operations: loading files, using systems, setting parameters, and running simulations.

## Prerequisites

Make sure you have completed [Getting Started](02-GETTING-STARTED.md) and have:
- SDL server running (`sdl serve`)
- Console client connected (`sdl console`)

## Core Command Pattern

SDL follows a simple workflow:
1. **Load** an SDL file
2. **Use** a system from that file
3. **Set** parameters (optional)
4. **Run** simulations

## 1. Loading SDL Files

### Basic Load
```
SDL> load examples/contacts/contacts.sdl
✅ Loaded examples/contacts/contacts.sdl successfully
Available systems: ContactsSystem
```

### Tab Completion for Files
Use tab completion to browse files:
```
SDL> load examples/<TAB>
contacts/    users/    orders/

SDL> load examples/contacts/<TAB>
contacts.sdl

SDL> load examples/contacts/contacts.sdl
```

### Understanding Load Output
When you load a file:
- Console shows success message and available systems
- Server terminal shows detailed parsing logs
- Web dashboard updates system information

## 2. Using Systems

### Select a System
After loading, select which system to work with:
```
SDL> use ContactsSystem
✅ Now using system: ContactsSystem
SDL[ContactsSystem]> 
```

Notice the prompt changes to show the active system.

### Tab Completion for Systems
```
SDL> use <TAB>
ContactsSystem

SDL> use ContactsSystem
```

### System Information
```
SDL[ContactsSystem]> info
System: ContactsSystem
Components:
  - server
  - database  
  - cache
Available methods:
  - server.HandleLookup
  - server.HandleCreate
```

## 3. Setting Parameters

### View Current Parameters
```
SDL[ContactsSystem]> get
server.pool.ArrivalRate = 1.0
server.pool.ServiceTime = 0.1
database.connections = 10
```

### Set Individual Parameters
```
SDL[ContactsSystem]> set server.pool.ArrivalRate 10
✅ Set server.pool.ArrivalRate = 10

SDL[ContactsSystem]> set database.connections 20
✅ Set database.connections = 20
```

### Tab Completion for Parameters
SDL provides intelligent tab completion:
```
SDL[ContactsSystem]> set <TAB>
server.pool.ArrivalRate    server.pool.ServiceTime    database.connections

SDL[ContactsSystem]> set server.<TAB>
server.pool.ArrivalRate    server.pool.ServiceTime

SDL[ContactsSystem]> set server.pool.<TAB>
ArrivalRate    ServiceTime
```

## 4. Running Simulations

### Basic Run Command
```
SDL[ContactsSystem]> run test1 server.HandleLookup 100
✅ Running test1: server.HandleLookup (100 calls)
🎯 Execution completed: test1
📊 Duration: 1.23s
⚡ Throughput: 81.3 calls/sec
```

### Run with Different Call Counts
```
SDL[ContactsSystem]> run quick server.HandleLookup 10
SDL[ContactsSystem]> run medium server.HandleLookup 100  
SDL[ContactsSystem]> run load server.HandleLookup 1000
```

### Tab Completion for Methods
```
SDL[ContactsSystem]> run test2 <TAB>
server.HandleLookup    server.HandleCreate

SDL[ContactsSystem]> run test2 server.<TAB>
server.HandleLookup    server.HandleCreate
```

## 5. Viewing Results

### Execution History
```
SDL[ContactsSystem]> history
Recent executions:
1. test1: server.HandleLookup (100 calls) - 1.23s
2. quick: server.HandleLookup (10 calls) - 0.12s  
3. medium: server.HandleLookup (100 calls) - 1.18s
```

### Real-time Dashboard
While running commands, watch the web dashboard at http://localhost:8080:
- Live execution graphs
- System component status
- Performance metrics

## 6. Recipe Files (Automation)

### Create a Recipe File
Create `my-test.recipe`:
```
load examples/contacts/contacts.sdl
use ContactsSystem
set server.pool.ArrivalRate 10
run test1 server.HandleLookup 100
run test2 server.HandleCreate 50
```

### Execute Recipe
```
SDL> execute my-test.recipe
✅ Executing recipe: my-test.recipe
✅ Loaded examples/contacts/contacts.sdl successfully
✅ Now using system: ContactsSystem
✅ Set server.pool.ArrivalRate = 10
✅ Running test1: server.HandleLookup (100 calls)
✅ Running test2: server.HandleCreate (50 calls)
📊 Recipe completed successfully
```

## 7. Shell Commands

Execute shell commands with `!` prefix:
```
SDL[ContactsSystem]> !ls examples/
contacts/    users/    orders/

SDL[ContactsSystem]> !git status
On branch main
nothing to commit, working tree clean
```

## Command Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `load <file>` | Load SDL file | `load examples/contacts/contacts.sdl` |
| `use <system>` | Select system | `use ContactsSystem` |
| `set <param> <value>` | Set parameter | `set server.pool.ArrivalRate 10` |
| `get [param]` | View parameters | `get` or `get server.pool.ArrivalRate` |
| `run <name> <method> <calls>` | Run simulation | `run test1 server.HandleLookup 100` |
| `info` | System information | `info` |
| `history` | Execution history | `history` |
| `execute <file>` | Run recipe file | `execute test.recipe` |
| `help` | Show help | `help` |
| `exit` | Quit console | `exit` |

## What's Next?

Now that you understand basic operations, continue to **[Traffic Generation](04-TRAFFIC-GENERATION.md)** to learn about automated load simulation.

## Tips

1. **Use Tab Completion** - Press TAB frequently for suggestions
2. **Check the Dashboard** - Watch real-time updates at http://localhost:8080
3. **Monitor Server Logs** - The server terminal shows detailed execution information
4. **Save Recipes** - Create `.recipe` files for repeatable test sequences
5. **Experiment** - Try different parameter values and call counts