# SDL Web Dashboard

The SDL web dashboard provides a real-time visualization interface for system design simulations using TypeScript, Tailwind CSS, and Connect-Web for RPC communication.

## Architecture

### Frontend Stack
- **TypeScript** - Type-safe development
- **Tailwind CSS** - Utility-first styling
- **DockView** - Professional panel layout system
- **Chart.js** - Real-time performance metrics
- **Connect-Web v2** - Type-safe RPC client
- **Vite** - Fast build tooling

### RPC Integration
The dashboard uses Connect-Web (v2) to communicate with the SDL server:

- **Protocol Buffers** - Shared type definitions between frontend and backend
- **Connect-Web Transport** - HTTP/2 and HTTP/1.1 compatible RPC
- **Generated TypeScript** - Type-safe client from proto definitions
- **Real-time Updates** - WebSocket for live metrics (separate from RPC)

## Setup

### Prerequisites
- Node.js 18+ and npm
- SDL server running (`sdl serve`)

### Installation
```bash
cd web
npm install
```

### Development
```bash
# Start dev server with hot reload
npm run dev

# Run tests
npm test

# Run quick validation test
npm run dev-quick

# Run comprehensive dashboard test
npm run dev-test
```

### Production Build
```bash
npm run build
```

## Connect-Web Integration

### Generated Code
TypeScript client code is generated from protobuf definitions:

```bash
# From project root
buf generate
```

Generated files are placed in `web/src/gen/` to avoid relative import issues:
- `web/src/gen/sdl/v1/canvas_pb.ts` - Message types
- `web/src/gen/sdl/v1/models_pb.ts` - Data models
- `web/src/gen/sdl/v1/canvas_connect.ts` - Service client

### Client Configuration
The Canvas client is configured in `web/src/canvas-client.ts`:

```typescript
const transport = createConnectTransport({
  baseUrl: `${window.location.origin}/api`,
  useBinaryFormat: false, // JSON for browser compatibility
});
```

### Using Proto Types
The dashboard directly uses generated proto types:

```typescript
import type { SystemDiagram } from './gen/sdl/v1/canvas_pb.ts';
import type { Generator, Metric } from './gen/sdl/v1/models_pb.ts';
```

## Key Components

### CanvasClient (`canvas-client.ts`)
- Type-safe wrapper around Connect-Web client
- Handles all RPC communication with SDL server
- Default canvas ID management
- WebSocket connection for live updates

### Dashboard (`dashboard.ts`)
- Main application controller
- DockView panel management
- Real-time chart updates
- System diagram visualization

### Types
All types are generated from protobuf definitions - no manual type maintenance required.

## Development Workflow

1. **Proto Changes**: Edit `.proto` files and run `buf generate`
2. **Test Locally**: `npm run dev` and test with local SDL server
3. **Run Tests**: `npm test` for unit tests, `npm run dev-test` for integration
4. **Build**: `npm run build` creates production bundle in `dist/`

## Troubleshooting

### Connect-Web 404 Errors
If you see 404 errors for RPC calls:
1. Ensure SDL server is running with Connect handler
2. Check that Connect is mounted at `/api/sdl.v1.CanvasService/`
3. Verify no `/api/connect` prefix in client URLs

### Type Mismatches
Proto-generated types are Message classes, not plain objects:
- Use `response.data` directly instead of reconstructing objects
- Access fields with camelCase (e.g., `node.id` not `node.ID`)

### WebSocket Connection
WebSocket is separate from Connect-Web:
- RPC for control operations (load, use, add generator)
- WebSocket for real-time metrics streaming
- Both can work independently

## Architecture Decisions

### Why Connect-Web v2?
- Type-safe RPC with full TypeScript support
- Works with existing gRPC backend via Connect-Go
- Better than REST for complex operations
- Automatic request/response validation

### Why Generate in `web/src/gen/`?
- Avoids relative import path issues (`../../../gen`)
- Keeps generated code within web package
- Simplifies TypeScript configuration
- Better IDE support

### Proto-First Design
- Canvas returns proto types directly (no conversion)
- Viz package accepts proto types
- Single source of truth for data structures
- Eliminates marshaling overhead