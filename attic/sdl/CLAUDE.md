# SDL Codebase Guide

## Build & Test Commands
- Run all tests: `go test ./...`
- Run specific test: `go test -run TestFunctionName`
- Continuous test on file change: `./livetest.sh`

## Code Style Guidelines
- **Formatting**: Follow standard Go formatting (gofmt)
- **Imports**: Standard library imports first, then third-party
- **Types**: Use generics with clear constraints (`[V any]`, `[V, U, Z any]`)
- **Naming**:
  - Variables: camelCase (outWeight, fracWeight)
  - Functions/Methods: PascalCase (MaxDuration, TotalWeight)
  - Types: PascalCase (Bucket, Outcomes, HeapFile)
- **Error Handling**: Use panic for invalid states, log.Fatalf for fatal errors
- **Patterns**:
  - Builder pattern with method chaining
  - Pointer receivers on most methods
  - Reducer pattern for data transformation
  - Init methods that return the receiver for chaining
  - Functional programming approach (functions as arguments)