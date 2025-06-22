# SDL Logging Guide

## Overview
SDL uses a structured logging system that provides clean test output by default while allowing detailed debugging when needed.

## Quick Start

### Running Tests Quietly (Default)
```bash
go test ./...
```

### Running Tests with Debug Output
```bash
# Enable debug logs
SDL_LOG_LEVEL=debug go test ./... -v

# Enable specific log levels
SDL_LOG_LEVEL=info go test ./...
SDL_LOG_LEVEL=warn go test ./...
SDL_LOG_LEVEL=error go test ./...
```

### Disabling Logs Completely
```bash
SDL_LOG_LEVEL=off go test ./...
```

## For Developers

### Using the Logger

#### In Runtime Package
```go
import "github.com/panyam/sdl/runtime"

// Log at different levels
runtime.Debug("Component %s processing at %.2f RPS", name, rate)
runtime.Info("System initialized")
runtime.Warn("Cache hit rate below threshold: %.2f", hitRate)
runtime.Error("Failed to load component: %v", err)
```

#### In Console Package
```go
import "github.com/panyam/sdl/console"

// Standard logs
console.Debug("Generator state: %v", state)
console.Info("WebSocket connected")
console.Warn("Measurement storage slow")
console.Error("API request failed: %v", err)

// Event logs with emojis (automatic in console)
console.Start("Generator started: %s @ %d RPS", name, rate)
console.Stop("Generator stopped: %s", name)
console.Success("Operation completed")
console.Failure("Operation failed: %v", err)
```

### In Tests

#### Silence All Logs
```go
func TestSomething(t *testing.T) {
    defer runtime.QuietTest(t)()
    // Your test code - no logs will appear
}
```

#### Capture Logs for Assertions
```go
func TestLogging(t *testing.T) {
    buffer, cleanup := runtime.CaptureLog(t, runtime.LogLevelInfo)
    defer cleanup()
    
    // Your code that produces logs
    runtime.Info("Expected message")
    
    // Assert on captured logs
    logs := buffer.String()
    if !strings.Contains(logs, "Expected message") {
        t.Error("Expected log not found")
    }
}
```

#### Enable Debug for Verbose Tests
```go
func TestDebugOutput(t *testing.T) {
    defer runtime.VerboseTest(t)()
    // Debug logs will appear when running with -v flag
}
```

## Environment Variables

- `SDL_LOG_LEVEL` - Set the minimum log level (debug, info, warn, error, off)
- `SDL_NO_EMOJI` - Disable emoji output in console logs

## Log Levels

1. **DEBUG** - Detailed information for debugging (flow rates, method calls)
2. **INFO** - General information (connections, initialization)
3. **WARN** - Warning conditions (performance issues, fallbacks)
4. **ERROR** - Error conditions (failures, missing components)
5. **OFF** - No logging

## Best Practices

1. **Use Debug for Flow Analysis**
   ```go
   Debug("FlowEval: %s.%s @ %.1f RPS -> %v", component, method, rate, flows)
   ```

2. **Use Info for Lifecycle Events**
   ```go
   Info("System %s initialized", systemName)
   ```

3. **Use Warn for Degraded Performance**
   ```go
   Warn("Cache hit rate low: %.2f%%, expected > 80%%", hitRate*100)
   ```

4. **Use Error for Failures**
   ```go
   Error("Component %s not found in system", componentName)
   ```

5. **In Tests, Always Silence or Capture**
   ```go
   defer QuietTest(t)() // Most common pattern
   ```

## Migration from log.Printf

Replace existing logging:

```go
// Before
log.Printf("Processing %s at %d RPS", name, rate)
fmt.Printf("✅ Success: %s\n", message)

// After
Debug("Processing %s at %d RPS", name, rate)
Success("Success: %s", message)
```