package services

import (
	"context"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SystemsServiceImpl implements the SystemsService gRPC interface
type SystemsServiceImpl struct {
	catalog *SystemCatalogService
}

// NewSystemsService creates a new SystemsService implementation
func NewSystemsService() *SystemsServiceImpl {
	return &SystemsServiceImpl{
		catalog: NewSystemCatalogService(),
	}
}

// ListSystems returns all available systems
func (s *SystemsServiceImpl) ListSystems(ctx context.Context, req *v1.ListSystemsRequest) (*v1.ListSystemsResponse, error) {
	systems := s.catalog.ListSystems()

	protoSystems := make([]*v1.SystemInfo, len(systems))
	for i, system := range systems {
		protoSystems[i] = &v1.SystemInfo{
			Id:          system.ID,
			Name:        system.Name,
			Description: system.Description,
			Category:    system.Category,
			Difficulty:  system.Difficulty,
			Tags:        system.Tags,
			Icon:        system.Icon,
			LastUpdated: system.LastUpdated,
		}
	}

	return &v1.ListSystemsResponse{
		Systems: protoSystems,
	}, nil
}

// GetSystem returns a specific system with metadata
func (s *SystemsServiceImpl) GetSystem(ctx context.Context, req *v1.GetSystemRequest) (*v1.GetSystemResponse, error) {
	system := s.catalog.GetSystem(req.Id)
	if system == nil {
		return nil, status.Errorf(codes.NotFound, "system %s not found", req.Id)
	}

	// Convert versions
	protoVersions := make(map[string]*v1.SystemVersion)
	for versionKey, version := range system.Versions {
		protoVersions[versionKey] = &v1.SystemVersion{
			Sdl:    version.SDL,
			Recipe: version.Recipe,
			Readme: version.Readme,
		}
	}

	protoSystem := &v1.SystemProject{
		Id:             system.ID,
		Name:           system.Name,
		Description:    system.Description,
		Category:       system.Category,
		Difficulty:     system.Difficulty,
		Tags:           system.Tags,
		Icon:           system.Icon,
		Versions:       protoVersions,
		DefaultVersion: system.DefaultVersion,
		LastUpdated:    system.LastUpdated,
	}

	return &v1.GetSystemResponse{
		System: protoSystem,
	}, nil
}

// GetSystemContent returns the SDL and recipe content for a system
func (s *SystemsServiceImpl) GetSystemContent(ctx context.Context, req *v1.GetSystemContentRequest) (*v1.GetSystemContentResponse, error) {
	system := s.catalog.GetSystem(req.Id)
	if system == nil {
		return nil, status.Errorf(codes.NotFound, "system %s not found", req.Id)
	}

	// Use requested version or default
	version := req.Version
	if version == "" {
		version = system.DefaultVersion
	}

	versionData, exists := system.Versions[version]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "version %s not found for system %s", version, req.Id)
	}

	return &v1.GetSystemContentResponse{
		SdlContent:    versionData.SDL,
		RecipeContent: versionData.Recipe,
		ReadmeContent: versionData.Readme,
	}, nil
}
