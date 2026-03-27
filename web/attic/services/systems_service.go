package services

import (
	"context"
	"strings"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SystemsServiceImpl implements the SystemsService gRPC interface.
// Delegates to WorkspaceService for data.
type SystemsServiceImpl struct {
	workspaceSvc WorkspaceService
}

func NewSystemsService(workspaceSvc WorkspaceService) *SystemsServiceImpl {
	return &SystemsServiceImpl{workspaceSvc: workspaceSvc}
}

// ListSystems returns all designs across all workspaces as SystemInfo entries.
func (s *SystemsServiceImpl) ListSystems(ctx context.Context, req *v1.ListSystemsRequest) (*v1.ListSystemsResponse, error) {
	listResp, err := s.workspaceSvc.ListWorkspaces(ctx, &v1.ListWorkspacesRequest{})
	if err != nil {
		return nil, err
	}

	var protoSystems []*v1.SystemInfo
	for _, ws := range listResp.Workspaces {
		for _, design := range ws.Designs {
			protoSystems = append(protoSystems, &v1.SystemInfo{
				Id:          ws.Id + "/" + design.Name,
				Name:        design.Name,
				Description: design.Description,
				Difficulty:  design.Difficulty,
				Tags:        design.Tags,
				Category:    design.Category,
			})
		}
	}

	return &v1.ListSystemsResponse{Systems: protoSystems}, nil
}

// GetSystem returns a workspace by ID.
func (s *SystemsServiceImpl) GetSystem(ctx context.Context, req *v1.GetSystemRequest) (*v1.GetSystemResponse, error) {
	parts := strings.SplitN(req.Id, "/", 2)
	wsResp, err := s.workspaceSvc.GetWorkspace(ctx, &v1.GetWorkspaceRequest{Id: parts[0]})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "workspace %s not found", parts[0])
	}

	return &v1.GetSystemResponse{
		System: &v1.SystemProject{
			Id:          wsResp.Workspace.Id,
			Name:        wsResp.Workspace.Name,
			Description: wsResp.Workspace.Description,
		},
	}, nil
}

// GetSystemContent returns the SDL content for a design.
func (s *SystemsServiceImpl) GetSystemContent(ctx context.Context, req *v1.GetSystemContentRequest) (*v1.GetSystemContentResponse, error) {
	parts := strings.SplitN(req.Id, "/", 2)
	wsId := parts[0]
	designName := ""
	if len(parts) > 1 {
		designName = parts[1]
	}

	if designName != "" {
		resp, err := s.workspaceSvc.GetDesignContent(ctx, &v1.GetDesignContentRequest{
			WorkspaceId: wsId,
			DesignName:  designName,
		})
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "design %s/%s not found", wsId, designName)
		}
		return &v1.GetSystemContentResponse{SdlContent: resp.SdlContent}, nil
	}

	// No design specified — return all combined
	resp, err := s.workspaceSvc.GetAllDesignContents(ctx, &v1.GetAllDesignContentsRequest{WorkspaceId: wsId})
	if err != nil {
		return nil, err
	}
	combined := ""
	for _, c := range resp.Contents {
		combined += c + "\n\n"
	}
	return &v1.GetSystemContentResponse{SdlContent: combined}, nil
}
