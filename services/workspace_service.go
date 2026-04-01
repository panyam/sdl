package services

import (
	"context"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// WorkspaceCRUD defines the CRUD subset of WorkspaceService.
// Used by implementations that only handle metadata (BackendWorkspaceService, inmem).
// The full WorkspaceService interface is generated from the proto definition
// at gen/go/sdl/v1/services/workspace_grpc.pb.go (WorkspaceServiceServer).
type WorkspaceCRUD interface {
	CreateWorkspace(context.Context, *protos.CreateWorkspaceRequest) (*protos.CreateWorkspaceResponse, error)
	GetWorkspace(context.Context, *protos.GetWorkspaceRequest) (*protos.GetWorkspaceResponse, error)
	ListWorkspaces(context.Context, *protos.ListWorkspacesRequest) (*protos.ListWorkspacesResponse, error)
	DeleteWorkspace(context.Context, *protos.DeleteWorkspaceRequest) (*protos.DeleteWorkspaceResponse, error)
	UpdateWorkspace(context.Context, *protos.UpdateWorkspaceRequest) (*protos.UpdateWorkspaceResponse, error)
	GetDesignContent(context.Context, *protos.GetDesignContentRequest) (*protos.GetDesignContentResponse, error)
	GetAllDesignContents(context.Context, *protos.GetAllDesignContentsRequest) (*protos.GetAllDesignContentsResponse, error)
}
