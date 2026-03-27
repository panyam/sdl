//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WorkspaceStorageProvider is implemented by concrete backends (in-memory, fsbe, DB)
// to provide raw storage operations.
type WorkspaceStorageProvider interface {
	LoadWorkspace(ctx context.Context, id string) (*protos.Workspace, error)
	SaveWorkspace(ctx context.Context, id string, ws *protos.Workspace) error
	DeleteWorkspace(ctx context.Context, id string) error
	ListWorkspaces(ctx context.Context) ([]*protos.Workspace, error)

	// Design content storage
	LoadDesignContent(ctx context.Context, workspaceId, designName string) (string, error)
	SaveDesignContent(ctx context.Context, workspaceId, designName, content string) error
	LoadAllDesignContents(ctx context.Context, workspaceId string) (map[string]string, error)
}

// BackendWorkspaceService wraps a WorkspaceStorageProvider with common logic.
type BackendWorkspaceService struct {
	storage WorkspaceStorageProvider
}

func NewBackendWorkspaceService(storage WorkspaceStorageProvider) *BackendWorkspaceService {
	return &BackendWorkspaceService{storage: storage}
}

func (s *BackendWorkspaceService) CreateWorkspace(ctx context.Context, req *protos.CreateWorkspaceRequest) (*protos.CreateWorkspaceResponse, error) {
	ws := req.Workspace
	if ws == nil {
		return nil, fmt.Errorf("workspace is required")
	}
	if ws.Id == "" {
		return nil, fmt.Errorf("workspace ID is required")
	}

	// Check if already exists
	existing, _ := s.storage.LoadWorkspace(ctx, ws.Id)
	if existing != nil {
		return nil, fmt.Errorf("workspace %s already exists", ws.Id)
	}

	now := timestamppb.Now()
	ws.CreatedAt = now
	ws.UpdatedAt = now

	if err := s.storage.SaveWorkspace(ctx, ws.Id, ws); err != nil {
		return nil, fmt.Errorf("failed to save workspace: %w", err)
	}

	return &protos.CreateWorkspaceResponse{Workspace: ws}, nil
}

func (s *BackendWorkspaceService) GetWorkspace(ctx context.Context, req *protos.GetWorkspaceRequest) (*protos.GetWorkspaceResponse, error) {
	ws, err := s.storage.LoadWorkspace(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &protos.GetWorkspaceResponse{Workspace: ws}, nil
}

func (s *BackendWorkspaceService) ListWorkspaces(ctx context.Context, req *protos.ListWorkspacesRequest) (*protos.ListWorkspacesResponse, error) {
	workspaces, err := s.storage.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	return &protos.ListWorkspacesResponse{Workspaces: workspaces}, nil
}

func (s *BackendWorkspaceService) DeleteWorkspace(ctx context.Context, req *protos.DeleteWorkspaceRequest) (*protos.DeleteWorkspaceResponse, error) {
	if err := s.storage.DeleteWorkspace(ctx, req.Id); err != nil {
		return nil, err
	}
	return &protos.DeleteWorkspaceResponse{}, nil
}

func (s *BackendWorkspaceService) UpdateWorkspace(ctx context.Context, req *protos.UpdateWorkspaceRequest) (*protos.UpdateWorkspaceResponse, error) {
	ws := req.Workspace
	if ws == nil || ws.Id == "" {
		return nil, fmt.Errorf("workspace with ID is required")
	}
	ws.UpdatedAt = timestamppb.Now()
	if err := s.storage.SaveWorkspace(ctx, ws.Id, ws); err != nil {
		return nil, fmt.Errorf("failed to save workspace: %w", err)
	}
	return &protos.UpdateWorkspaceResponse{Workspace: ws}, nil
}

func (s *BackendWorkspaceService) GetDesignContent(ctx context.Context, req *protos.GetDesignContentRequest) (*protos.GetDesignContentResponse, error) {
	content, err := s.storage.LoadDesignContent(ctx, req.WorkspaceId, req.DesignName)
	if err != nil {
		return nil, err
	}
	return &protos.GetDesignContentResponse{
		SdlContent: content,
		DesignName: req.DesignName,
	}, nil
}

func (s *BackendWorkspaceService) GetAllDesignContents(ctx context.Context, req *protos.GetAllDesignContentsRequest) (*protos.GetAllDesignContentsResponse, error) {
	contents, err := s.storage.LoadAllDesignContents(ctx, req.WorkspaceId)
	if err != nil {
		return nil, err
	}
	return &protos.GetAllDesignContentsResponse{Contents: contents}, nil
}
