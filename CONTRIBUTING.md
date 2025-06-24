# Contributing to SDL

Thank you for your interest in contributing to SDL! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive criticism and solutions
- Respect differing viewpoints and experiences

## Getting Started

### Prerequisites

Before contributing, ensure you have:

1. Go 1.24 or later installed
2. Node.js 18+ and npm installed
3. goyacc installed: `go install golang.org/x/tools/cmd/goyacc@latest`
4. buf installed: `go install github.com/bufbuild/buf/cmd/buf@latest`
5. Git configured with your GitHub account

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/sdl.git
   cd sdl
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/anthropics/sdl.git
   ```

4. Install dependencies:
   ```bash
   make deps
   npm install
   ```

5. Build the project:
   ```bash
   make
   ```

6. Run tests to verify setup:
   ```bash
   make test
   ```

## Development Process

### Branch Naming

Use descriptive branch names:
- `feature/add-component-x` - New features
- `fix/issue-123-description` - Bug fixes
- `docs/improve-readme` - Documentation
- `refactor/simplify-parser` - Code refactoring

### Commit Messages

Follow these guidelines for commit messages:

1. Use present tense ("Add feature" not "Added feature")
2. Use imperative mood ("Move cursor to..." not "Moves cursor to...")
3. Limit first line to 72 characters
4. Reference issues and PRs in the body

Example:
```
Add support for distributed tracing in components

- Implement TraceContext propagation
- Add span creation for method calls
- Update documentation with examples

Fixes #123
```

### Development Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature
   ```

2. Make your changes following code style guidelines

3. Add or update tests as needed

4. Run tests and linting:
   ```bash
   make test
   make lint
   ```

5. Commit your changes with a descriptive message

6. Push to your fork and create a pull request

## Code Style Guidelines

### Go Code Style

We follow standard Go conventions with some specific guidelines:

1. **Formatting**: Use `gofmt` and `goimports`
   ```bash
   gofmt -w .
   goimports -w .
   ```

2. **Naming Conventions**:
   - Exported functions/types: `PascalCase`
   - Unexported functions/types: `camelCase`
   - Constants: `UPPER_SNAKE_CASE` for groups, `PascalCase` for single values
   - Interfaces: Should end with `-er` suffix when possible

3. **Error Handling**:
   ```go
   // Preferred
   if err != nil {
       return fmt.Errorf("failed to process request: %w", err)
   }
   
   // Not preferred
   if err != nil {
       return err
   }
   ```

4. **Comments**:
   - All exported types and functions must have godoc comments
   - Comments should start with the name being described
   ```go
   // ComponentInstance represents a runtime instance of a component.
   type ComponentInstance struct {
       // ...
   }
   ```

5. **Testing**:
   - Test files should be in the same package
   - Use table-driven tests where appropriate
   - Test names should be descriptive: `TestComponentInstance_SetParameter_InvalidType`

### TypeScript/JavaScript Code Style

For web dashboard code:

1. **Formatting**: Use Prettier with project configuration
   ```bash
   npm run format
   ```

2. **TypeScript Usage**:
   - Prefer interfaces over type aliases for objects
   - Use strict mode
   - Avoid `any` type - use `unknown` if type is truly unknown

3. **React Guidelines**:
   - Use functional components with hooks
   - Props interfaces should be named `{ComponentName}Props`
   - Keep components focused and small

4. **Naming**:
   - Components: `PascalCase`
   - Functions/variables: `camelCase`
   - Constants: `UPPER_SNAKE_CASE`
   - CSS classes: `kebab-case`

### SDL Language Style

When writing SDL examples or tests:

1. Use meaningful component and method names
2. Add comments explaining modeling decisions
3. Format consistently with 4-space indentation
4. Group related parameters together

Example:
```sdl
component DatabaseService {
    // Connection pool configuration
    uses pool ResourcePool(
        Size = 20,
        ArrivalRate = 100.0,
        AvgHoldTime = 50ms
    )
    
    // Cache configuration
    uses cache Cache(HitRate = 0.85)
    
    method Query() Bool {
        // Try cache first for better performance
        let hit = self.cache.Read()
        if hit {
            return true
        }
        
        // Fall back to database
        return self.pool.Acquire()
    }
}
```

## Testing Requirements

### Test Coverage

- All new features must include tests
- Bug fixes should include a test that would have caught the bug
- Aim for >80% code coverage for new code
- Critical paths should have >90% coverage

### Types of Tests

1. **Unit Tests**: Test individual functions and methods
   ```bash
   go test ./...
   ```

2. **Integration Tests**: Test component interactions
   ```bash
   go test ./tests/integration/...
   ```

3. **End-to-End Tests**: Test complete workflows
   ```bash
   ./test_metrics_e2e.sh
   ```

4. **Browser Tests**: Test web dashboard
   ```bash
   npm run test
   ```

### Writing Tests

1. **Go Tests**:
   ```go
   func TestResourcePool_Acquire_HighLoad(t *testing.T) {
       pool := NewResourcePool(10, 100.0, 50*time.Millisecond)
       
       // Test behavior under high load
       results := make([]bool, 1000)
       for i := range results {
           results[i] = pool.Acquire()
       }
       
       // Verify some requests were rejected
       rejected := countFalse(results)
       assert.Greater(t, rejected, 0, "Expected some rejections under high load")
   }
   ```

2. **Recipe Tests**: Create `.recipe` files for complex scenarios
   ```bash
   # test_feature.recipe
   echo "=== Testing new feature ==="
   sdl load examples/test.sdl
   sdl use TestSystem
   sdl run test.Method 1000
   ```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./runtime/...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...

# Run browser tests
npm run test
```

## Pull Request Process

### Before Submitting

1. Ensure all tests pass
2. Update documentation for any API changes
3. Add entries to RELEASE_NOTES.md for significant changes
4. Verify no hardcoded paths or personal data

### PR Guidelines

1. **Title**: Clear, concise description of changes
   - Good: "Add support for Kafka component"
   - Bad: "Updates"

2. **Description Template**:
   ```markdown
   ## Summary
   Brief description of what this PR does.
   
   ## Motivation
   Why are these changes needed?
   
   ## Changes
   - List of specific changes
   - Breaking changes marked with ⚠️
   
   ## Testing
   How has this been tested?
   
   ## Screenshots
   (If applicable)
   
   Fixes #issue_number
   ```

3. **PR Size**: Keep PRs focused and reasonably sized
   - Large features should be split into multiple PRs
   - Refactoring should be separate from feature changes

### Review Process

1. All PRs require at least one review
2. Address review feedback promptly
3. Re-request review after making changes
4. Squash commits before merging if requested

### Merge Requirements

- All CI checks must pass
- No merge conflicts
- Approved by at least one maintainer
- Documentation updated if needed

## Issue Reporting

### Bug Reports

Use the bug report template and include:

1. SDL version (`sdl --version`)
2. Operating system and version
3. Steps to reproduce
4. Expected behavior
5. Actual behavior
6. Relevant logs or error messages
7. Minimal SDL file that reproduces the issue

### Feature Requests

Use the feature request template and include:

1. Clear use case description
2. Proposed solution
3. Alternative solutions considered
4. Examples of how it would be used

### Good First Issues

Look for issues labeled `good first issue` if you're new to the project. These are typically:
- Documentation improvements
- Simple bug fixes
- Small feature additions
- Test coverage improvements

## Documentation

### Documentation Standards

1. All public APIs must be documented
2. Include examples for complex features
3. Keep documentation up-to-date with code changes
4. Use clear, concise language

### Types of Documentation

1. **Code Comments**: Inline documentation for developers
2. **API Documentation**: Generated from code comments
3. **User Guide**: How-to guides and tutorials
4. **Architecture Docs**: System design and decisions

### Writing Documentation

- Use active voice
- Include code examples
- Explain the "why" not just the "what"
- Test all code examples

## Community

### Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues before creating new ones

### Providing Help

- Answer questions in issues and discussions
- Review pull requests
- Improve documentation
- Share your SDL use cases

### Recognition

We value all contributions! Contributors are recognized in:
- Release notes
- Contributors file
- Project statistics

## License

By contributing to SDL, you agree that your contributions will be licensed under the Apache License 2.0.

## Thank You!

Your contributions make SDL better for everyone. Whether it's fixing a typo, adding a feature, or helping others, every contribution matters.

Happy contributing!