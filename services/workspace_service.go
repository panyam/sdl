package services

import (
	"context"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// WorkspaceService defines the interface for workspace CRUD operations.
// Implementations: in-memory (seeded from examples), file-based, DB-backed,
// IndexedDB (WASM singleton).
type WorkspaceService interface {
	CreateWorkspace(context.Context, *protos.CreateWorkspaceRequest) (*protos.CreateWorkspaceResponse, error)
	GetWorkspace(context.Context, *protos.GetWorkspaceRequest) (*protos.GetWorkspaceResponse, error)
	ListWorkspaces(context.Context, *protos.ListWorkspacesRequest) (*protos.ListWorkspacesResponse, error)
	DeleteWorkspace(context.Context, *protos.DeleteWorkspaceRequest) (*protos.DeleteWorkspaceResponse, error)
	UpdateWorkspace(context.Context, *protos.UpdateWorkspaceRequest) (*protos.UpdateWorkspaceResponse, error)
	GetDesignContent(context.Context, *protos.GetDesignContentRequest) (*protos.GetDesignContentResponse, error)
	GetAllDesignContents(context.Context, *protos.GetAllDesignContentsRequest) (*protos.GetAllDesignContentsResponse, error)
}
