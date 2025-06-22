# Connect-Go Integration

This document describes the Connect protocol integration in the SDL console package.

## Architecture

The SDL server supports both gRPC and Connect protocols on the same port (8080):
- gRPC-gateway handles REST/JSON requests at `/v1/`
- Connect handles RPC requests at `/sdl.v1.CanvasService/`

Both protocols share the same `CanvasService` instance to ensure consistency.

## Connect Adapter Pattern

Since Connect-Go expects different method signatures than gRPC-Go, we use an adapter pattern:

```go
// gRPC signature
func (s *CanvasService) GetCanvas(ctx context.Context, req *protos.GetCanvasRequest) (*protos.GetCanvasResponse, error)

// Connect signature
func (s *ConnectCanvasServiceAdapter) GetCanvas(ctx context.Context, req *connect.Request[protos.GetCanvasRequest]) (*connect.Response[protos.GetCanvasResponse], error)
```

The `ConnectCanvasServiceAdapter` wraps the gRPC service and translates between the two interfaces:
1. Unwraps Connect requests to get proto messages
2. Calls the underlying gRPC service method
3. Wraps responses back into Connect format

## Implementation

### api.go
The main API setup in `api.go`:
- Creates the Connect adapter if a CanvasService is provided
- Registers the Connect handler using generated `v1connect.NewCanvasServiceHandler`
- Mounts at the path returned by the handler (e.g., `/sdl.v1.CanvasService/`)

### ConnectCanvasServiceAdapter
Located in `console/connect_adapter.go`, this adapter:
- Implements the generated `v1connect.CanvasServiceHandler` interface
- Wraps each gRPC method with Connect request/response handling
- Preserves all error codes and messages from gRPC

## Benefits

1. **Protocol Flexibility**: Clients can choose gRPC, REST (via gateway), or Connect
2. **Code Reuse**: Single service implementation serves all protocols
3. **Type Safety**: Generated code ensures type correctness
4. **Web Compatibility**: Connect works better in browsers than gRPC-Web

## Future Improvements

1. **Code Generation**: The adapter could potentially be auto-generated
2. **Streaming Support**: Add server-streaming for real-time metrics
3. **Unified Transport**: Replace WebSocket with Connect streaming
4. **Error Mapping**: Enhanced error detail preservation between protocols