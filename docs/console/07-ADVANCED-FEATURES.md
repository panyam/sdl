# SDL Console & Server Tutorial - Advanced Features

This chapter covers advanced SDL console features that boost productivity and provide professional workflow capabilities.

## Prerequisites

Complete [Web Dashboard](06-WEB-DASHBOARD.md) and have:
- Working SDL server and console setup
- Understanding of basic commands and measurements
- Familiarity with traffic generation

## Command History

### Persistent History
SDL automatically saves your command history across sessions:

```
SDL> !ls ~/.sdl_history
/Users/username/.sdl_history
```

### Navigating History
Use arrow keys to navigate through previous commands:
- **Up Arrow** - Previous command
- **Down Arrow** - Next command
- **Ctrl+R** - Reverse search through history

### History Commands
```
SDL> history
Recent commands:
1. load examples/contacts/contacts.sdl
2. use ContactsSystem
3. set server.pool.ArrivalRate 10
4. run test1 server.HandleLookup 100
5. gen add load1 server.HandleLookup 20
6. gen start load1
7. measure add lat1 server.HandleLookup latency
```

### Search History
```
SDL> history | grep gen
5. gen add load1 server.HandleLookup 20
6. gen start load1

SDL> history | grep measure
7. measure add lat1 server.HandleLookup latency
```

## Advanced Tab Completion

### Context-Aware Completion
SDL provides intelligent completion based on current context:

#### File Completion
```
SDL> load <TAB>
examples/           tools/              docs/

SDL> load examples/<TAB>
contacts/           users/              orders/

SDL> load examples/contacts/<TAB>
contacts.sdl        README.md
```

#### System Completion
```
SDL> use <TAB>
ContactsSystem      UserSystem          OrderSystem

SDL> use Contact<TAB>
ContactsSystem
```

#### Parameter Completion
```
SDL[ContactsSystem]> set <TAB>
server.pool.ArrivalRate    server.pool.ServiceTime    database.connections

SDL[ContactsSystem]> set server.<TAB>
server.pool.ArrivalRate    server.pool.ServiceTime

SDL[ContactsSystem]> set server.pool.<TAB>
ArrivalRate    ServiceTime
```

#### Method Completion
```
SDL[ContactsSystem]> run test1 <TAB>
server.HandleLookup    server.HandleCreate    server.HandleUpdate

SDL[ContactsSystem]> run test1 server.<TAB>
server.HandleLookup    server.HandleCreate    server.HandleUpdate
```

### Dynamic System Discovery
SDL automatically extracts system names from loaded SDL files:

```
SDL> load complex_system.sdl
✅ Loaded complex_system.sdl successfully
Available systems: OrderProcessingSystem, PaymentSystem, InventorySystem

SDL> use <TAB>
OrderProcessingSystem    PaymentSystem    InventorySystem
```

## Shell Integration

### Execute Shell Commands
Use `!` prefix to run shell commands without leaving SDL:

```
SDL> !pwd
/Users/username/projects/sdl

SDL> !ls -la examples/
drwxr-xr-x  5 user  staff   160 Jun 15 10:30 contacts
drwxr-xr-x  3 user  staff    96 Jun 15 10:30 users
drwxr-xr-x  4 user  staff   128 Jun 15 10:30 orders

SDL> !git status
On branch main
Your branch is up to date with 'origin/main'.
nothing to commit, working tree clean
```

### Common Shell Integration Patterns

#### File Management
```
SDL> !mkdir my_tests
SDL> !touch my_tests/load_test.recipe
SDL> !ls my_tests/
load_test.recipe
```

#### Version Control
```
SDL> !git log --oneline -5
82ad5fe Update console documentation
6b5a43a Implement measurement system
d326268 Add traffic generators
0ca2c9d Add DuckDB integration
c674474 Initial console implementation
```

#### System Monitoring
```
SDL> !ps aux | grep sdl
user     12345   0.1  0.5  34567  8901 pts/0    S+   10:30   0:01 ./bin/sdl serve

SDL> !netstat -an | grep 8080
tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN
```

### Environment Integration
```
SDL> !echo $SDL_CONFIG_PATH
/Users/username/.sdl/config

SDL> !which sdl
/usr/local/bin/sdl
```

## Recipe Automation

### Advanced Recipe Files
Create sophisticated automation scripts:

#### Complete Load Test Recipe
Create `comprehensive_test.recipe`:
```
# Comprehensive load testing scenario
!echo "Starting comprehensive load test..."

# System setup
load examples/contacts/contacts.sdl
use ContactsSystem
set server.pool.ArrivalRate 50
set database.connections 20

# Measurement setup
measure add baseline_latency server.HandleLookup latency
measure add baseline_throughput server.HandleLookup throughput
measure add baseline_errors server.HandleLookup error_rate

# Baseline test
!echo "Phase 1: Baseline measurement"
run baseline server.HandleLookup 1000

# Light load test
!echo "Phase 2: Light sustained load"
gen add light_load server.HandleLookup 10
gen start light_load
!sleep 30
gen stop light_load

# Medium load test  
!echo "Phase 3: Medium sustained load"
gen add medium_load server.HandleLookup 25
gen start medium_load
!sleep 60
gen stop medium_load

# Heavy load test
!echo "Phase 4: Heavy sustained load"
gen add heavy_load server.HandleLookup 50
gen start heavy_load
!sleep 90
gen stop heavy_load

# Results summary
!echo "Test completed - check dashboard for results"
measure stats baseline_latency
measure stats baseline_throughput
measure stats baseline_errors
```

#### Multi-Method Testing Recipe
Create `method_comparison.recipe`:
```
# Compare performance across different methods
load examples/contacts/contacts.sdl
use ContactsSystem

# Set up measurements for each method
measure add lookup_latency server.HandleLookup latency
measure add create_latency server.HandleCreate latency
measure add update_latency server.HandleUpdate latency

# Parallel load on all methods
gen add lookup_gen server.HandleLookup 20
gen add create_gen server.HandleCreate 5
gen add update_gen server.HandleUpdate 8

# Start all generators
gen start

!echo "Running parallel load test for 2 minutes..."
!sleep 120

# Stop and analyze
gen stop
measure compare lookup_latency create_latency
measure compare lookup_latency update_latency
```

### Recipe Execution Patterns

#### Sequential Recipe Execution
```
SDL> execute baseline_test.recipe
SDL> execute stress_test.recipe  
SDL> execute cleanup.recipe
```

#### Conditional Recipe Execution
Create `conditional_test.recipe`:
```
# Run test only if previous test passed
run quick_test server.HandleLookup 10
!if [ $? -eq 0 ]; then echo "Quick test passed, running full test"; fi
run full_test server.HandleLookup 1000
```

## Advanced Measurement Features

### Batch Measurement Operations
```
# Create multiple related measurements
SDL> measure add read_latency server.HandleLookup latency
SDL> measure add read_throughput server.HandleLookup throughput
SDL> measure add read_errors server.HandleLookup error_rate
SDL> measure add write_latency server.HandleCreate latency
SDL> measure add write_throughput server.HandleCreate throughput
```

### Measurement Grouping
```
# Group measurements for analysis
SDL> measure group read_metrics read_latency read_throughput read_errors
SDL> measure group write_metrics write_latency write_throughput

# Analyze by group
SDL> measure stats read_metrics
SDL> measure compare read_metrics write_metrics
```

### Custom Time Windows
```
# Analyze specific time periods
SDL> measure data lat1 --since "2024-06-15 10:00:00" --until "2024-06-15 11:00:00"
SDL> measure stats lat1 --last 15m
SDL> measure export lat1 csv --last 1h data_last_hour.csv
```

## Advanced Generator Patterns

### Coordinated Generator Control
```
# Start generators in sequence
SDL> gen add phase1 server.HandleLookup 10
SDL> gen add phase2 server.HandleLookup 25  
SDL> gen add phase3 server.HandleLookup 50

# Orchestrated ramp-up
SDL> gen start phase1
SDL> !sleep 60
SDL> gen start phase2
SDL> !sleep 60  
SDL> gen start phase3
```

### Dynamic Load Adjustment
```
# Create load adjustment recipe
SDL> gen add dynamic_load server.HandleLookup 10
SDL> gen start dynamic_load

# Adjust load based on conditions
SDL> measure data latency | grep "95th" > temp_stats.txt
SDL> !if grep -q "100ms" temp_stats.txt; then echo "High latency detected"; fi
SDL> gen set dynamic_load rate 5  # Reduce load
```

## Configuration Management

### SDL Configuration File
Create `~/.sdl/config.yaml`:
```yaml
# SDL Console Configuration
console:
  history_size: 1000
  auto_complete: true
  color_output: true
  
server:
  default_url: "http://localhost:8080"
  timeout: 30s
  
measurements:
  default_retention: "24h"
  auto_export: true
  export_format: "csv"
  
generators:
  default_rate: 10
  max_rate: 1000
  safety_checks: true
```

### Environment Variables
```bash
# Set default server URL
export SDL_SERVER_URL="http://production-server:8080"

# Set measurement retention
export SDL_MEASUREMENT_RETENTION="7d"

# Set default data directory
export SDL_DATA_DIR="/opt/sdl/data"
```

## Workspace Management

### Project-Specific Configurations
Create `.sdlrc` in project directory:
```bash
# Project-specific SDL settings
SDL_DEFAULT_SYSTEM="ContactsSystem"
SDL_DEFAULT_MEASUREMENT_DIR="./measurements"
SDL_DEFAULT_RECIPE_DIR="./recipes"

# Auto-load commands
load examples/contacts/contacts.sdl
use ContactsSystem
set server.pool.ArrivalRate 25
```

### Session Management
```
# Save current session state
SDL> session save production_test
✅ Session saved as 'production_test'

# List saved sessions  
SDL> session list
Available sessions:
- production_test (2024-06-15 10:30)
- baseline_config (2024-06-14 15:45)
- stress_test_setup (2024-06-13 09:15)

# Load previous session
SDL> session load production_test
✅ Session 'production_test' loaded
✅ System: ContactsSystem
✅ Generators: load1 (stopped), stress_gen (stopped)
✅ Measurements: lat1, throughput1, errors1
```

## Command Aliases

### Built-in Aliases
```
SDL> l          # alias for 'load'
SDL> u          # alias for 'use'  
SDL> r          # alias for 'run'
SDL> g          # alias for 'gen'
SDL> m          # alias for 'measure'
```

### Custom Aliases
```
# Define custom aliases
SDL> alias quick_baseline "run baseline server.HandleLookup 100"
SDL> alias setup_monitoring "measure add lat server.HandleLookup latency; measure add tput server.HandleLookup throughput"

# Use custom aliases
SDL> quick_baseline
✅ Running baseline: server.HandleLookup (100 calls)

SDL> setup_monitoring  
✅ Measurement 'lat' created
✅ Measurement 'tput' created
```

## Performance Optimization

### Console Performance
```
# Optimize for large systems
SDL> set console.buffer_size 10000
SDL> set console.completion_limit 100
SDL> set console.history_search_limit 500
```

### Measurement Performance
```
# Optimize measurement collection
SDL> set measurement.batch_size 1000
SDL> set measurement.flush_interval 5s
SDL> set measurement.compression true
```

## What's Next?

Continue to **[Remote Access](08-REMOTE-ACCESS.md)** to learn how to connect to SDL servers on remote machines and manage distributed testing scenarios.

## Advanced Tips

1. **Master Tab Completion** - Use TAB extensively to discover available options
2. **Leverage Shell Integration** - Combine SDL commands with shell scripts
3. **Create Recipe Libraries** - Build reusable test scenarios
4. **Use History Search** - Ctrl+R for quick command recall
5. **Automate Common Tasks** - Create aliases for frequent operations
6. **Monitor Resource Usage** - Use shell commands to track system performance
7. **Save Important Sessions** - Preserve complex setups for future use
8. **Document Your Workflows** - Create recipes that serve as documentation