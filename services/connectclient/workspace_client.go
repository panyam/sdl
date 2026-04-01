package connectclient

import (
	"context"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1 "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// WorkspaceClient implements services.WorkspaceRuntime by wrapping a gRPC
// CanvasServiceClient. This is the remote-mode backend — requires a running
// `sdl serve`. Follows the lilbattle connectclient.GamesClient pattern.
//
// The server still uses Canvas/CanvasService internally. This client
// translates WorkspaceRuntime calls to the existing Canvas gRPC API.
type WorkspaceClient struct {
	conn     *grpc.ClientConn
	client   v1.CanvasServiceClient
	canvasID string
}

// NewWorkspaceClient creates a remote workspace client connected to the given gRPC address.
func NewWorkspaceClient(addr string, canvasID string) (*WorkspaceClient, error) {
	return NewWorkspaceClientWithAuth(addr, canvasID, "")
}

// NewWorkspaceClientWithAuth creates a remote workspace client with optional auth token.
// Follows the lilbattle connectclient pattern with authTransport for Bearer token injection.
func NewWorkspaceClientWithAuth(addr string, canvasID string, token string) (*WorkspaceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if token != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(authInterceptor(token)))
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %v", addr, err)
	}

	if canvasID == "" {
		canvasID = "default"
	}

	return &WorkspaceClient{
		conn:     conn,
		client:   v1.NewCanvasServiceClient(conn),
		canvasID: canvasID,
	}, nil
}

// authInterceptor returns a gRPC unary interceptor that injects a Bearer token.
func authInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *WorkspaceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// File and system management

func (c *WorkspaceClient) LoadFile(ctx context.Context, req *protos.LoadFileRequest) (*protos.LoadFileResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.LoadFile(ctx, req)
}

func (c *WorkspaceClient) UseSystem(ctx context.Context, req *protos.UseSystemRequest) (*protos.UseSystemResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.UseSystem(ctx, req)
}

// Generator management

func (c *WorkspaceClient) AddGenerator(ctx context.Context, req *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error) {
	return c.client.AddGenerator(ctx, req)
}

func (c *WorkspaceClient) UpdateGenerator(ctx context.Context, req *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error) {
	return c.client.UpdateGenerator(ctx, req)
}

func (c *WorkspaceClient) DeleteGenerator(ctx context.Context, req *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.DeleteGenerator(ctx, req)
}

func (c *WorkspaceClient) ListGenerators(ctx context.Context, req *protos.ListGeneratorsRequest) (*protos.ListGeneratorsResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.ListGenerators(ctx, req)
}

func (c *WorkspaceClient) StartGenerator(ctx context.Context, req *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.StartGenerator(ctx, req)
}

func (c *WorkspaceClient) StopGenerator(ctx context.Context, req *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.StopGenerator(ctx, req)
}

func (c *WorkspaceClient) StartAllGenerators(ctx context.Context, req *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.StartAllGenerators(ctx, req)
}

func (c *WorkspaceClient) StopAllGenerators(ctx context.Context, req *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.StopAllGenerators(ctx, req)
}

// Metric management

func (c *WorkspaceClient) AddMetric(ctx context.Context, req *protos.AddMetricRequest) (*protos.AddMetricResponse, error) {
	return c.client.AddMetric(ctx, req)
}

func (c *WorkspaceClient) DeleteMetric(ctx context.Context, req *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.DeleteMetric(ctx, req)
}

func (c *WorkspaceClient) ListMetrics(ctx context.Context, req *protos.ListMetricsRequest) (*protos.ListMetricsResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.ListMetrics(ctx, req)
}

// Parameters

func (c *WorkspaceClient) SetParameter(ctx context.Context, req *protos.SetParameterRequest) (*protos.SetParameterResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.SetParameter(ctx, req)
}

func (c *WorkspaceClient) GetParameters(ctx context.Context, req *protos.GetParametersRequest) (*protos.GetParametersResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.GetParameters(ctx, req)
}

// Diagram and flow analysis

func (c *WorkspaceClient) GetSystemDiagram(ctx context.Context, req *protos.GetSystemDiagramRequest) (*protos.GetSystemDiagramResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.GetSystemDiagram(ctx, req)
}

func (c *WorkspaceClient) EvaluateFlows(ctx context.Context, req *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error) {
	req.CanvasId = c.canvasID
	return c.client.EvaluateFlows(ctx, req)
}
