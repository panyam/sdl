# SDL Testing Strategy

## Overview
This document outlines our comprehensive testing strategy for continuous development with fast feedback loops.

## Test Categories

### 1. Unit Tests
Fast, isolated tests for individual components.

#### Flow Strategy Tests (`runtime/flow_strategy_test.go`)
- Test strategy registration and retrieval
- Test RuntimeFlowStrategy.Evaluate() with mock data
- Test component name resolution
- Test flow data conversion

#### Flow Types Tests (`runtime/flow_types_test.go`)
- Test type conversions between runtime and API types
- Test serialization/deserialization

### 2. Integration Tests

#### Canvas Flow Tests (`console/canvas_flow_test.go`)
- Test EvaluateFlowWithStrategy
- Test ApplyFlowStrategy
- Test SetComponentArrivalRate
- Test manual rate override persistence

#### System Flow Tests (`runtime/floweval_runtime_test.go` - extend existing)
- Test flow evaluation with real components
- Test the cache hit rate bug specifically
- Test flow propagation through system

### 3. API Tests (`console/canvas_web_test.go`)
Test HTTP endpoints:
- GET /api/flows/strategies
- GET /api/flows/{strategy}/eval
- POST /api/flows/{strategy}/apply
- GET /api/flows/current
- PUT /api/components/{component}/methods/{method}/arrival-rate

### 4. CLI Tests (`cmd/sdl/commands/flows_test.go`)
Test command execution:
- `sdl flows list`
- `sdl flows eval [strategy]`
- `sdl flows apply [strategy]`
- `sdl flows show`
- `sdl flows set-rate component.method rate`

### 5. End-to-End Tests (`test/e2e/flows_test.go`)
Full workflow tests using real SDL files:
- Load contacts.sdl
- Apply flow strategy
- Verify rates are correct
- Test manual overrides
- Test strategy switching

## Test Data

### Test SDL Files
Create simplified test cases in `test/fixtures/`:
- `simple_cache.sdl` - Reproduces cache hit rate bug
- `complex_flow.sdl` - Tests multi-component flows
- `conditional_flow.sdl` - Tests early return handling

### Mock Components
Create mock components that:
- Have predictable flow patterns
- Support arrival rate setting
- Can verify expected rates

## Automation Strategy

### 1. Fast Unit Test Suite
```bash
# Run only unit tests (fast)
go test ./runtime ./console -short
```

### 2. Integration Test Suite
```bash
# Run integration tests (medium speed)
go test ./... -run Integration
```

### 3. Full Test Suite
```bash
# Run all tests including E2E
go test ./...
```

### 4. Continuous Testing
```bash
# Watch mode for development
air --build.cmd "go build ./cmd/sdl" --build.bin "./cmd/sdl" --build.exclude_dir "test"
```

### 5. Test Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Key Test Scenarios

### 1. Cache Hit Rate Bug Test
```go
// Test that 80% cache hit rate results in ~20 RPS to database
func TestCacheHitRatePropagation(t *testing.T) {
    // Load system with cache -> database
    // Set cache hit rate to 80%
    // Set generator at 100 RPS
    // Verify database sees ~20 RPS (not ~99 RPS)
}
```

### 2. Manual Override Test
```go
// Test that manual overrides persist and override flow calculations
func TestManualRateOverride(t *testing.T) {
    // Apply flow strategy
    // Manually set component rate
    // Re-evaluate flows
    // Verify manual rate is preserved
}
```

### 3. Strategy Switching Test
```go
// Test switching between strategies preserves system state
func TestStrategySwitch(t *testing.T) {
    // Apply runtime strategy
    // Switch to future capacity-aware strategy
    // Verify system remains functional
}
```

## Performance Benchmarks

### 1. Flow Evaluation Performance
```go
func BenchmarkFlowEvaluation(b *testing.B) {
    // Measure time to evaluate flows for systems of various sizes
}
```

### 2. API Response Time
```go
func BenchmarkFlowAPI(b *testing.B) {
    // Measure API endpoint response times
}
```

## Test Helpers

### 1. Test Canvas Builder
```go
func NewTestCanvas(t *testing.T) *Canvas {
    // Create canvas with test configuration
}
```

### 2. SDL File Loader
```go
func LoadTestSDL(t *testing.T, filename string) *SystemInstance {
    // Load SDL file and return system instance
}
```

### 3. Flow Assertion Helpers
```go
func AssertFlowRate(t *testing.T, component, method string, expected float64) {
    // Verify component has expected arrival rate
}
```

## Continuous Integration

### GitHub Actions Workflow
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
    - run: go test -v ./...
    - run: go test -race ./...
    - run: go test -coverprofile=coverage.out ./...
```

## Test Maintenance

### 1. Test Organization
- Keep tests close to code they test
- Use table-driven tests for multiple scenarios
- Share test fixtures and helpers

### 2. Test Naming
- Use descriptive names: `TestFlowStrategy_RuntimeEvaluation_CacheHitRate`
- Group related tests with subtests

### 3. Test Data Management
- Use golden files for complex expected outputs
- Version test fixtures with code
- Clean up test data after runs

## Success Metrics

1. **Coverage**: Aim for >80% code coverage
2. **Speed**: Unit tests complete in <1 second
3. **Reliability**: No flaky tests
4. **Clarity**: Test failures clearly indicate what's broken