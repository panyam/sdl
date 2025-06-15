# SDL Console & Server Tutorial - Traffic Generation

This chapter covers SDL's powerful traffic generation system for automated load simulation.

## Prerequisites

Complete [Basic Commands](03-BASIC-COMMANDS.md) and have:
- SDL server running with a loaded system
- Understanding of `run` command basics

## What are Traffic Generators?

Traffic generators simulate continuous load against your system methods. Instead of running one-off simulations, generators create sustained traffic patterns that help you:

- Test system behavior under continuous load
- Observe performance over time
- Simulate realistic user patterns
- Generate data for measurements and analysis

## 1. Creating Traffic Generators

### Basic Generator Creation
```
SDL[ContactsSystem]> gen add load1 server.HandleLookup 10
✅ Generator 'load1' created
🎯 Target: server.HandleLookup
⚡ Rate: 10 calls/second  
🔄 Status: Stopped
```

### Generator Components
Each generator has:
- **Name** - Unique identifier (`load1`)
- **Target** - Method to call (`server.HandleLookup`)
- **Rate** - Calls per second (`10`)
- **Status** - Enabled/Disabled state

### Multiple Generators
Create multiple generators for different scenarios:
```
SDL[ContactsSystem]> gen add lookup_load server.HandleLookup 15
SDL[ContactsSystem]> gen add create_load server.HandleCreate 5
SDL[ContactsSystem]> gen add heavy_load server.HandleLookup 50
```

## 2. Managing Generators

### List All Generators
```
SDL[ContactsSystem]> gen list
Active Traffic Generators:
┌─────────────┬─────────────────────┬──────┬─────────┐
│ Name        │ Target              │ Rate │ Status  │
├─────────────┼─────────────────────┼──────┼─────────┤
│ load1       │ server.HandleLookup │ 10   │ Stopped │
│ lookup_load │ server.HandleLookup │ 15   │ Stopped │
│ create_load │ server.HandleCreate │ 5    │ Stopped │
│ heavy_load  │ server.HandleLookup │ 50   │ Stopped │
└─────────────┴─────────────────────┴──────┴─────────┘
```

### Modify Generator Settings
```
SDL[ContactsSystem]> gen set load1 rate 20
✅ Generator 'load1' rate updated to 20 calls/second

SDL[ContactsSystem]> gen set load1 target server.HandleCreate
✅ Generator 'load1' target updated to server.HandleCreate
```

### Remove Generators
```
SDL[ContactsSystem]> gen remove heavy_load
✅ Generator 'heavy_load' removed
```

## 3. Running Generators

### Start Individual Generators
```
SDL[ContactsSystem]> gen start load1
✅ Generator 'load1' started
🎯 Generating 20 calls/second to server.HandleCreate
```

Watch the server terminal - you'll see continuous execution logs:
```
✅ Generator load1: executed 20 calls to server.HandleCreate
✅ Generator load1: executed 20 calls to server.HandleCreate
✅ Generator load1: executed 20 calls to server.HandleCreate
```

### Start Multiple Generators
```
SDL[ContactsSystem]> gen start lookup_load
SDL[ContactsSystem]> gen start create_load
✅ Multiple generators now running
🎯 lookup_load: 15 calls/sec → server.HandleLookup
🎯 create_load: 5 calls/sec → server.HandleCreate
```

### Start All Generators
```
SDL[ContactsSystem]> gen start
✅ All generators started
🎯 Total load: 40 calls/second across 2 generators
```

## 4. Stopping Generators

### Stop Individual Generators
```
SDL[ContactsSystem]> gen stop load1
✅ Generator 'load1' stopped
🛑 Traffic generation halted
```

### Stop All Generators
```
SDL[ContactsSystem]> gen stop
✅ All generators stopped
🛑 All traffic generation halted
```

## 5. Real-time Monitoring

### Console Status Updates
While generators run, the console shows periodic updates:
```
SDL[ContactsSystem]> gen status
Generator Status (Live):
┌─────────────┬─────────────────────┬──────┬─────────┬───────────┐
│ Name        │ Target              │ Rate │ Status  │ Uptime    │
├─────────────┼─────────────────────┼──────┼─────────┼───────────┤
│ lookup_load │ server.HandleLookup │ 15   │ Running │ 00:02:34  │
│ create_load │ server.HandleCreate │ 5    │ Running │ 00:01:45  │
└─────────────┴─────────────────────┴──────┴─────────┴───────────┘
Total Active Load: 20 calls/second
```

### Server Statistics
The server terminal shows live statistics:
```
Canvas Statistics (updates every 30s):
🎯 Systems Loaded: 1
⚡ Active Generators: 2  
📊 Active Measurements: 0
🔥 Current Load: 20 calls/second
📈 Total Calls: 3,450
```

### Web Dashboard
The web dashboard at http://localhost:8080 shows:
- Real-time traffic generation charts
- Individual generator status
- Performance metrics
- System load visualization

## 6. Advanced Generator Patterns

### Load Testing Pattern
```
# Gradual ramp-up
SDL[ContactsSystem]> gen add ramp1 server.HandleLookup 10
SDL[ContactsSystem]> gen start ramp1
# ... wait and observe ...
SDL[ContactsSystem]> gen set ramp1 rate 25
# ... wait and observe ...
SDL[ContactsSystem]> gen set ramp1 rate 50
```

### Mixed Workload Pattern
```
# Simulate realistic traffic mix
SDL[ContactsSystem]> gen add reads server.HandleLookup 80    # 80% reads
SDL[ContactsSystem]> gen add writes server.HandleCreate 20   # 20% writes
SDL[ContactsSystem]> gen start
```

### Burst Testing Pattern
```
# Normal load with periodic bursts
SDL[ContactsSystem]> gen add baseline server.HandleLookup 10
SDL[ContactsSystem]> gen add burst server.HandleLookup 100
SDL[ContactsSystem]> gen start baseline
# ... run for a while ...
SDL[ContactsSystem]> gen start burst
# ... observe burst behavior ...
SDL[ContactsSystem]> gen stop burst
```

## 7. Generator Recipes

### Create Generator Recipe
Create `load-test.recipe`:
```
# Load test scenario
load examples/contacts/contacts.sdl
use ContactsSystem

# Configure system for load
set server.pool.ArrivalRate 50
set database.connections 25

# Set up traffic generators
gen add steady_reads server.HandleLookup 30
gen add steady_writes server.HandleCreate 10
gen add burst_reads server.HandleLookup 100

# Start baseline load
gen start steady_reads
gen start steady_writes

# Note: Start burst_reads manually when ready to test peaks
```

### Execute Generator Recipe
```
SDL> execute load-test.recipe
✅ Executing recipe: load-test.recipe
✅ Loaded examples/contacts/contacts.sdl successfully
✅ Now using system: ContactsSystem
✅ Set server.pool.ArrivalRate = 50
✅ Set database.connections = 25
✅ Generator 'steady_reads' created
✅ Generator 'steady_writes' created  
✅ Generator 'burst_reads' created
✅ Generator 'steady_reads' started
✅ Generator 'steady_writes' started
📊 Recipe completed - baseline load active
```

## Command Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `gen add <name> <method> <rate>` | Create generator | `gen add load1 server.HandleLookup 10` |
| `gen list` | Show all generators | `gen list` |
| `gen start [name]` | Start generator(s) | `gen start load1` or `gen start` |
| `gen stop [name]` | Stop generator(s) | `gen stop load1` or `gen stop` |
| `gen status` | Live generator status | `gen status` |
| `gen set <name> <property> <value>` | Modify generator | `gen set load1 rate 20` |
| `gen remove <name>` | Delete generator | `gen remove load1` |

## What's Next?

Now that you can generate sustained traffic, continue to **[Measurements](05-MEASUREMENTS.md)** to learn how to collect and analyze performance metrics.

## Best Practices

1. **Start Small** - Begin with low rates and increase gradually
2. **Monitor Resources** - Watch CPU/memory usage on server terminal
3. **Use Meaningful Names** - Name generators to reflect their purpose
4. **Mix Workloads** - Use multiple generators for realistic patterns
5. **Save Recipes** - Document load test scenarios for repeatability
6. **Watch the Dashboard** - Use real-time visualization for insights