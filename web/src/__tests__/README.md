# SDL Web Dashboard Unit Tests

This directory contains unit tests for the SDL Web Dashboard components using Vitest.

## Running Tests

```bash
# Run all unit tests
npm run test:unit

# Run tests in watch mode
npm run test

# Run tests with UI
npm run test:unit:ui

# Run tests with coverage
npm run test:unit:coverage

# Run a specific test file
npm run test:unit -- src/__tests__/core/event-bus.test.ts
```

## Test Structure

```
__tests__/
├── core/           # Core module tests (EventBus, AppStateManager)
├── services/       # Service layer tests (CanvasClient, etc.)
├── panels/         # Panel component tests
├── components/     # UI component tests
├── utils/          # Test utilities and helpers
├── mocks/          # Mock implementations
└── setup.ts        # Global test setup
```

## Writing Tests

### Basic Test Structure

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ComponentName } from '../../path/to/component';

describe('ComponentName', () => {
  let component: ComponentName;
  
  beforeEach(() => {
    // Setup before each test
    component = new ComponentName();
  });
  
  describe('feature', () => {
    it('should do something', () => {
      // Arrange
      const input = 'test';
      
      // Act
      const result = component.method(input);
      
      // Assert
      expect(result).toBe('expected');
    });
  });
});
```

### Using Mocks

The test infrastructure provides several pre-configured mocks:

- **Monaco Editor**: Mocked via alias in vitest.config.ts
- **Chart.js**: Mocked in setup.ts
- **WebSocket**: Global mock for real-time features
- **Canvas API**: Mock implementations in mocks/canvas-api.ts

### Test Utilities

Available utilities in `utils/test-helpers.ts`:

- `waitFor()`: Wait for async conditions
- `waitForEvent()`: Wait for EventBus events
- `createMockElement()`: Create mock DOM elements
- `spyOnEvents()`: Spy on EventBus emissions
- Custom matchers for Canvas-specific assertions

## Coverage

Run coverage reports with:
```bash
npm run test:unit:coverage
```

Coverage reports are generated in:
- Terminal output
- `coverage/` directory (HTML report)

## Best Practices

1. **Test behavior, not implementation**
2. **Keep tests isolated and independent**
3. **Use descriptive test names**
4. **Clean up resources in afterEach**
5. **Mock external dependencies**
6. **Test error cases and edge conditions**

## Common Issues

### Monaco Editor Errors
The Monaco editor is mocked via Vite alias. If you encounter issues, check:
- The mock file at `mocks/monaco-editor.ts`
- The alias configuration in `vitest.config.ts`

### Async Test Timeouts
For slow async operations, increase the timeout:
```typescript
it('should handle slow operation', async () => {
  // test code
}, 10000); // 10 second timeout
```

### WebSocket Testing
Use the mock WebSocket from setup.ts:
```typescript
const mockWs = createMockWebSocket();
mockWs.dispatchEvent('message', { data: JSON.stringify(testData) });
```