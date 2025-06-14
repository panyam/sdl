# SDL Console Generator Commands Test Guide

## Overview

The SDL console now includes comprehensive traffic generator management commands for creating, configuring, and controlling traffic generators in real-time.

## Generator Commands

### Core Generator Commands
- `gen-add <id> <target> <rate>` - Add a new traffic generator
- `gen-list` - List all configured traffic generators  
- `gen-remove <id>` - Remove a traffic generator
- `gen-start` - Start all traffic generators
- `gen-stop` - Stop all traffic generators
- `gen-pause <id>` - Pause a specific generator
- `gen-resume <id>` - Resume a specific generator
- `gen-modify <id> <field> <value>` - Modify generator properties

## Test Scenarios

### Scenario 1: Basic Generator Lifecycle
```bash
# Start console and load system
SDL> load examples/contacts/contacts.sdl
SDL> use ContactsSystem

# Check initial state (should be empty)
SDL> gen-list
📋 No traffic generators configured

# Add generators
SDL> gen-add load1 server.HandleLookup 10
✅ Added generator: load1 -> server.HandleLookup at 10 rps

SDL> gen-add load2 server.HandleCreate 5
✅ Added generator: load2 -> server.HandleCreate at 5 rps

# List generators
SDL> gen-list
📋 Traffic Generators (2):
  load1: Generator-load1 -> server.HandleLookup (10 rps) [▶️ running]
  load2: Generator-load2 -> server.HandleCreate (5 rps) [▶️ running]

# Start traffic generation
SDL> gen-start
✅ Started all traffic generators

# Stop traffic generation
SDL> gen-stop
✅ Stopped all traffic generators
```

### Scenario 2: Generator Modification
```bash
# Modify generator properties
SDL> gen-modify load1 rate 25
✅ Modified generator load1: rate = 25

SDL> gen-modify load1 target server.HandleUpdate
✅ Modified generator load1: target = server.HandleUpdate

SDL> gen-modify load1 name "High Load Test"
✅ Modified generator load1: name = High Load Test

SDL> gen-modify load1 enabled false
✅ Modified generator load1: enabled = false

# Verify changes
SDL> gen-list
📋 Traffic Generators (2):
  load1: High Load Test -> server.HandleUpdate (25 rps) [⏸️ paused]
  load2: Generator-load2 -> server.HandleCreate (5 rps) [▶️ running]
```

### Scenario 3: Individual Generator Control
```bash
# Pause specific generator
SDL> gen-pause load2
✅ Paused generator: load2

# Resume specific generator
SDL> gen-resume load2
✅ Resumed generator: load2

# Remove generator
SDL> gen-remove load1
✅ Removed generator: load1

# Verify removal
SDL> gen-list
📋 Traffic Generators (1):
  load2: Generator-load2 -> server.HandleCreate (5 rps) [▶️ running]
```

### Scenario 4: Tab Completion Testing
```bash
# Test command completion
SDL> gen-<TAB>
gen-add       Add traffic generator
gen-list      List all traffic generators
gen-modify    Modify traffic generator
gen-pause     Pause traffic generator
gen-remove    Remove traffic generator
gen-resume    Resume traffic generator
gen-start     Start all traffic generators
gen-stop      Stop all traffic generators

# Test generator ID completion
SDL> gen-remove <TAB>
load1    High Load Test -> server.HandleUpdate (paused)
load2    Generator-load2 -> server.HandleCreate (running)

# Test field completion
SDL> gen-modify load1 <TAB>
rate      Requests per second
target    Target component method
name      Generator display name
enabled   Enable/disable generator (true/false)

# Test target completion
SDL> gen-add newgen <TAB>
server.HandleLookup    Lookup handler latency
server.HandleCreate    Create handler latency
server.HandleUpdate    Update handler latency
db.Query              Database query latency
```

### Scenario 5: Error Handling
```bash
# Test invalid arguments
SDL> gen-add
❌ Error: usage: gen-add <id> <target> <rate>

SDL> gen-modify load1 rate invalid
❌ Error: invalid rate 'invalid': must be a number

SDL> gen-remove nonexistent
❌ Error: generator 'nonexistent' not found

SDL> gen-modify load1 invalid-field value
❌ Error: unknown field 'invalid-field'. Available fields: rate, target, name, enabled
```

## Integration with Dashboard

### Real-time Updates
When generator commands are executed, they should trigger WebSocket broadcasts to update the web dashboard in real-time:

1. **Generator Added**: Dashboard shows new generator in Traffic Generation panel
2. **Generator Started**: Dashboard indicates active traffic generation
3. **Generator Modified**: Dashboard reflects updated configuration
4. **Generator Removed**: Dashboard removes generator from display

### Verification Steps
1. Open dashboard at http://localhost:8080
2. Execute generator commands in console
3. Verify dashboard updates immediately
4. Check WebSocket messages in browser developer tools

## Expected Output Formats

### gen-list Output
```
📋 Traffic Generators (3):
  api-load: API Load Test -> server.HandleLookup (50 rps) [▶️ running]
  db-load: Database Load -> db.Query (25 rps) [⏸️ paused]
  burst-test: Burst Traffic -> server.HandleCreate (100 rps) [▶️ running]
```

### Success Messages
- ✅ Added generator: load1 -> server.HandleLookup at 10 rps
- ✅ Removed generator: load1
- ✅ Started all traffic generators
- ✅ Modified generator load1: rate = 25

### Error Messages
- ❌ Error: generator 'nonexistent' not found
- ❌ Error: invalid rate 'abc': must be a number
- ❌ Error: usage: gen-add <id> <target> <rate>

## Advanced Usage

### Recipe File Integration
Generators can be included in recipe files for automated testing:
```bash
# Load system
load examples/contacts/contacts.sdl
use ContactsSystem

# Configure traffic generators
gen-add baseline server.HandleLookup 10
gen-add spike server.HandleCreate 50

# Start load testing
gen-start
sleep 5s

# Increase load
gen-modify baseline rate 25
gen-modify spike rate 100
sleep 10s

# Stop testing
gen-stop
```

This comprehensive generator management system provides full control over traffic generation for performance testing and system analysis.