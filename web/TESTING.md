# SDL Web Dashboard Testing Strategy

## Overview

This document outlines the testing strategy for the SDL Web Dashboard, which has been refactored to use a component-based architecture. The strategy focuses on comprehensive unit testing of individual components while maintaining existing end-to-end tests.

## Testing Framework

### Unit Testing
- **Framework**: Vitest
- **Advantages**: 
  - Native Vite integration
  - Fast execution with HMR support
  - TypeScript support out-of-the-box
  - Compatible test API with Jest

### End-to-End Testing
- **Framework**: Playwright (existing)
- **Purpose**: Full user workflow validation

## Directory Structure

```
web/
├── src/
│   ├── __tests__/           # Unit tests
│   │   ├── core/           # Core module tests
│   │   ├── services/       # Service layer tests
│   │   ├── panels/         # Panel component tests
│   │   ├── components/     # UI component tests
│   │   ├── utils/          # Test utilities
│   │   ├── mocks/          # Mock implementations
│   │   └── setup.ts        # Test setup file
│   └── [source files]
└── tests/                   # Playwright E2E tests
```

## Test Scripts

```bash
# Run unit tests
npm run test:unit

# Run unit tests with UI
npm run test:unit:ui

# Run unit tests with coverage
npm run test:unit:coverage

# Run E2E tests
npm run test:e2e

# Run all tests
npm run test:all
```

## Testing Approach

### 1. Unit Testing Components

#### Core Modules
- **EventBus**: Test event emission, subscription, and cleanup
- **AppStateManager**: Test state management, persistence, and subscriptions
- **Services**: Test API interactions with mocked responses

#### Panel Components
- Test rendering logic
- Test user interactions
- Test event handling
- Test state updates

#### Example Test Structure
```typescript
describe('ComponentName', () => {
  let component: ComponentName;
  let mockDependency: MockType;

  beforeEach(() => {
    mockDependency = createMockDependency();
    component = new ComponentName(mockDependency);
  });

  describe('feature', () => {
    it('should handle specific case', () => {
      // Arrange
      const input = prepareInput();
      
      // Act
      const result = component.method(input);
      
      // Assert
      expect(result).toEqual(expectedOutput);
    });
  });
});
```

### 2. Mock Strategy

#### Canvas API Mocks
```typescript
// Mock Canvas API responses
export const mockSystemConfig = {
  name: 'test-system',
  definition: 'system TestSystem { use app App }',
  trafficEnabled: false
};

// Mock WebSocket for real-time features
export const createMockWebSocket = () => ({
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  // ... other methods
});
```

#### DOM Mocks
- Monaco Editor mocks for code editor tests
- Chart.js mocks for metrics visualization
- ResizeObserver/IntersectionObserver for layout tests

### 3. Test Utilities

#### Custom Matchers
```typescript
// Canvas-specific assertions
expect(eventSpies).toHaveEmittedEvent('system:config:changed');
expect(eventSpies).toHaveEmittedEventWith('metrics:updated', data);
```

#### Test Helpers
```typescript
// Wait for async conditions
await waitFor(() => component.isReady());

// Wait for specific events
const data = await waitForEvent(eventBus, 'data:loaded');

// Create test fixtures
const fileStructure = createMockFileStructure([
  { name: 'test.sdl', content: 'system Test {}' },
  { name: 'examples', isDirectory: true, children: [...] }
]);
```

## Coverage Goals

### Initial Phase (Current)
- Core modules: 80%+ coverage
- Service layer: 70%+ coverage
- Critical paths: 90%+ coverage

### Target Coverage
- Overall: 80%+ line coverage
- Critical components: 90%+ coverage
- UI components: 70%+ coverage

## Best Practices

### 1. Test Isolation
- Each test should be independent
- Use `beforeEach` for fresh state
- Clean up subscriptions/timers in `afterEach`

### 2. Meaningful Test Names
```typescript
// Good
it('should emit file:saved event when save is successful')

// Bad
it('should work')
```

### 3. Test Behavior, Not Implementation
```typescript
// Good - tests behavior
it('should notify subscribers when state changes')

// Bad - tests implementation
it('should call _notifySubscribers method')
```

### 4. Use Test Doubles Appropriately
- **Mocks**: For external dependencies (API, WebSocket)
- **Stubs**: For simple return values
- **Spies**: For verifying interactions

### 5. Async Testing
```typescript
// Always await async operations
it('should load data', async () => {
  const result = await service.loadData();
  expect(result).toBeDefined();
});
```

## Component-Specific Testing

### RecipeRunner Tests
- Recipe parsing validation
- Syntax error detection
- Execution flow control
- State management during execution

### FileExplorer Tests
- File tree rendering
- User interactions (click, drag, drop)
- Context menu operations
- File operation integration

### Metrics Panel Tests
- Data visualization updates
- Real-time data handling
- Chart interactions
- Export functionality

## CI/CD Integration

### Pre-commit
```bash
# Run affected unit tests
npm run test:unit -- --changed
```

### Pull Request
```bash
# Full test suite
npm run test:all
```

### Deployment
```bash
# Full test suite with coverage
npm run test:unit:coverage
npm run test:e2e
```

## Debugging Tests

### Vitest UI
```bash
npm run test:unit:ui
```

### Debug in VS Code
```json
{
  "type": "node",
  "request": "launch",
  "name": "Debug Vitest Tests",
  "program": "${workspaceFolder}/node_modules/vitest/vitest.mjs",
  "args": ["--run", "${file}"],
  "console": "integratedTerminal"
}
```

## Future Enhancements

1. **Visual Regression Testing**
   - Add screenshot comparison for UI components
   - Integrate with Playwright for visual tests

2. **Performance Testing**
   - Add benchmarks for critical paths
   - Monitor component render times

3. **Mutation Testing**
   - Validate test quality with mutation testing
   - Identify gaps in test coverage

4. **Contract Testing**
   - Validate API contracts between frontend and backend
   - Use generated proto types for validation

## Migration from E2E to Unit Tests

For components previously tested only through E2E:

1. Identify core functionality
2. Extract business logic into testable units
3. Write unit tests for extracted logic
4. Keep E2E tests for user workflows
5. Remove redundant E2E tests

## Troubleshooting

### Common Issues

1. **Monaco Editor Errors**
   - Ensure monaco mocks are loaded in setup.ts
   - Use `vi.mock('monaco-editor')` at module level

2. **WebSocket Tests**
   - Always mock WebSocket globally
   - Simulate events through mock methods

3. **Async Timeouts**
   - Increase timeout for slow operations
   - Use `waitFor` helper for conditions

### Getting Help

- Check test examples in `__tests__` directory
- Review mock implementations in `mocks/`
- Consult Vitest documentation for advanced features