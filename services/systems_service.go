package services

import (
	"context"
	"strings"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SystemsServiceImpl implements the SystemsService gRPC interface.
// Now backed by workspace-based catalog instead of individual system entries.
type SystemsServiceImpl struct {
	catalog *SystemCatalogService
}

func NewSystemsService() *SystemsServiceImpl {
	return &SystemsServiceImpl{
		catalog: NewSystemCatalogService(),
	}
}

// ListSystems returns all designs across all workspaces as SystemInfo entries.
func (s *SystemsServiceImpl) ListSystems(ctx context.Context, req *v1.ListSystemsRequest) (*v1.ListSystemsResponse, error) {
	var protoSystems []*v1.SystemInfo

	for _, ws := range s.catalog.ListWorkspaces() {
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

	return &v1.ListSystemsResponse{
		Systems: protoSystems,
	}, nil
}

// GetSystem returns a workspace by ID.
func (s *SystemsServiceImpl) GetSystem(ctx context.Context, req *v1.GetSystemRequest) (*v1.GetSystemResponse, error) {
	// ID format: "workspace_id" or "workspace_id/design_name"
	parts := strings.SplitN(req.Id, "/", 2)
	wsId := parts[0]

	ws := s.catalog.GetWorkspace(wsId)
	if ws == nil {
		return nil, status.Errorf(codes.NotFound, "workspace %s not found", wsId)
	}

	// Return workspace-level info
	return &v1.GetSystemResponse{
		System: &v1.SystemProject{
			Id:          ws.Id,
			Name:        ws.Name,
			Description: ws.Description,
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
		content := s.catalog.GetDesignContent(wsId, designName)
		if content == "" {
			return nil, status.Errorf(codes.NotFound, "design %s/%s not found", wsId, designName)
		}
		return &v1.GetSystemContentResponse{
			SdlContent: content,
		}, nil
	}

	// No design specified — return all combined
	allContents := s.catalog.GetAllDesignContents(wsId)
	combined := ""
	for _, c := range allContents {
		combined += c + "\n\n"
	}
	return &v1.GetSystemContentResponse{
		SdlContent: combined,
	}, nil
}
