package services

import (
	"context"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// WorkspaceCRUD defines workspace metadata CRUD operations.
// This is the subset that BackendWorkspaceService and inmem implement.
type WorkspaceCRUD interface {
	CreateWorkspace(context.Context, *protos.CreateWorkspaceRequest) (*protos.CreateWorkspaceResponse, error)
	GetWorkspace(context.Context, *protos.GetWorkspaceRequest) (*protos.GetWorkspaceResponse, error)
	ListWorkspaces(context.Context, *protos.ListWorkspacesRequest) (*protos.ListWorkspacesResponse, error)
	DeleteWorkspace(context.Context, *protos.DeleteWorkspaceRequest) (*protos.DeleteWorkspaceResponse, error)
	UpdateWorkspace(context.Context, *protos.UpdateWorkspaceRequest) (*protos.UpdateWorkspaceResponse, error)
	GetDesignContent(context.Context, *protos.GetDesignContentRequest) (*protos.GetDesignContentResponse, error)
	GetAllDesignContents(context.Context, *protos.GetAllDesignContentsRequest) (*protos.GetAllDesignContentsResponse, error)
}

// WorkspaceRuntime defines runtime/simulation operations on a workspace.
// This is what CLI commands and presenters use for simulation control.
type WorkspaceRuntime interface {
	// File and system management
	LoadFile(context.Context, *protos.LoadFileRequest) (*protos.LoadFileResponse, error)
	UseSystem(context.Context, *protos.UseSystemRequest) (*protos.UseSystemResponse, error)
	GetCanvas(context.Context, *protos.GetCanvasRequest) (*protos.GetCanvasResponse, error)

	// Generator management
	AddGenerator(context.Context, *protos.AddGeneratorRequest) (*protos.AddGeneratorResponse, error)
	UpdateGenerator(context.Context, *protos.UpdateGeneratorRequest) (*protos.UpdateGeneratorResponse, error)
	DeleteGenerator(context.Context, *protos.DeleteGeneratorRequest) (*protos.DeleteGeneratorResponse, error)
	ListGenerators(context.Context, *protos.ListGeneratorsRequest) (*protos.ListGeneratorsResponse, error)
	StartGenerator(context.Context, *protos.StartGeneratorRequest) (*protos.StartGeneratorResponse, error)
	StopGenerator(context.Context, *protos.StopGeneratorRequest) (*protos.StopGeneratorResponse, error)
	StartAllGenerators(context.Context, *protos.StartAllGeneratorsRequest) (*protos.StartAllGeneratorsResponse, error)
	StopAllGenerators(context.Context, *protos.StopAllGeneratorsRequest) (*protos.StopAllGeneratorsResponse, error)

	// Metric management
	AddMetric(context.Context, *protos.AddMetricRequest) (*protos.AddMetricResponse, error)
	DeleteMetric(context.Context, *protos.DeleteMetricRequest) (*protos.DeleteMetricResponse, error)
	ListMetrics(context.Context, *protos.ListMetricsRequest) (*protos.ListMetricsResponse, error)

	// Parameters
	SetParameter(context.Context, *protos.SetParameterRequest) (*protos.SetParameterResponse, error)
	GetParameters(context.Context, *protos.GetParametersRequest) (*protos.GetParametersResponse, error)

	// Diagram and flow analysis
	GetSystemDiagram(context.Context, *protos.GetSystemDiagramRequest) (*protos.GetSystemDiagramResponse, error)
	EvaluateFlows(context.Context, *protos.EvaluateFlowsRequest) (*protos.EvaluateFlowsResponse, error)
}

// WorkspaceService is the full interface combining CRUD and runtime operations.
// Follows the lilbattle GamesService pattern: one interface covering everything,
// with proto request/response types throughout.
//
// Implementations:
//   - devenvbe.WorkspaceService — local mode, wraps DevEnv (no server needed)
//   - connectclient.WorkspaceClient — remote mode, wraps gRPC client
type WorkspaceService interface {
	WorkspaceCRUD
	WorkspaceRuntime
}
