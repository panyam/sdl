//go:build js && wasm
// +build js,wasm

package singleton

import (
	"context"
	"fmt"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/loader"
)

// SingletonSystemsService provides system listing for WASM mode.
// It scans the filesystem for available SDL system definitions.
type SingletonSystemsService struct {
	fileSystem loader.FileSystem
}

// NewSingletonSystemsService creates a new singleton systems service
func NewSingletonSystemsService(fs loader.FileSystem) *SingletonSystemsService {
	return &SingletonSystemsService{
		fileSystem: fs,
	}
}

// ListSystems returns all available systems
func (s *SingletonSystemsService) ListSystems(ctx context.Context, req *protos.ListSystemsRequest) (*protos.ListSystemsResponse, error) {
	// List available systems from filesystem
	systems := []*protos.SystemInfo{}

	// Check stdlib directory
	stdlibFiles, err := s.fileSystem.ListFiles("@stdlib/")
	if err == nil {
		for _, file := range stdlibFiles {
			systems = append(systems, &protos.SystemInfo{
				Id:          "@stdlib/" + file,
				Name:        file,
				Description: "Standard library system",
				Category:    "stdlib",
			})
		}
	}

	// Check examples directory
	exampleFiles, err := s.fileSystem.ListFiles("/examples/")
	if err == nil {
		for _, file := range exampleFiles {
			systems = append(systems, &protos.SystemInfo{
				Id:          "/examples/" + file,
				Name:        file,
				Description: "Example system",
				Category:    "examples",
			})
		}
	}

	// Check workspace directory
	workspaceFiles, err := s.fileSystem.ListFiles("/workspace/")
	if err == nil {
		for _, file := range workspaceFiles {
			systems = append(systems, &protos.SystemInfo{
				Id:          "/workspace/" + file,
				Name:        file,
				Description: "User workspace system",
				Category:    "workspace",
			})
		}
	}

	return &protos.ListSystemsResponse{
		Systems: systems,
	}, nil
}

// GetSystem returns a specific system with metadata
func (s *SingletonSystemsService) GetSystem(ctx context.Context, req *protos.GetSystemRequest) (*protos.GetSystemResponse, error) {
	// Try to find the system by ID (which is the path)
	path := req.Id
	if path == "" {
		return nil, fmt.Errorf("system ID required")
	}

	// Check if file exists
	content, err := s.fileSystem.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("system not found: %s", path)
	}

	return &protos.GetSystemResponse{
		System: &protos.SystemProject{
			Id:          path,
			Name:        path,
			Description: fmt.Sprintf("System file (%d bytes)", len(content)),
		},
	}, nil
}

// GetSystemContent returns the SDL and recipe content for a system
func (s *SingletonSystemsService) GetSystemContent(ctx context.Context, req *protos.GetSystemContentRequest) (*protos.GetSystemContentResponse, error) {
	path := req.Id
	if path == "" {
		return nil, fmt.Errorf("system ID required")
	}

	content, err := s.fileSystem.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read system: %w", err)
	}

	return &protos.GetSystemContentResponse{
		SdlContent: string(content),
	}, nil
}
