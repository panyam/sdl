package services

import (
	"log"
	"os"
	"path/filepath"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	svc "github.com/panyam/sdl/services"
)

// SystemCatalogService discovers example workspaces from sdl.json manifests.
// Works entirely with Workspace protos — no custom Go types.
type SystemCatalogService struct {
	workspaces map[string]*protos.Workspace
	// SDL content for each design, keyed by workspace_id/design_name
	designContents map[string]string
	basePath       string
}

func NewSystemCatalogService() *SystemCatalogService {
	service := &SystemCatalogService{
		workspaces:     make(map[string]*protos.Workspace),
		designContents: make(map[string]string),
		basePath:       "examples",
	}
	service.discoverWorkspaces()
	return service
}

// discoverWorkspaces scans basePath for directories containing sdl.json.
func (s *SystemCatalogService) discoverWorkspaces() {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		log.Printf("Warning: could not read examples dir %s: %v", s.basePath, err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(s.basePath, entry.Name())
		manifestPath := filepath.Join(dir, "sdl.json")

		ws, err := svc.LoadWorkspaceManifest(manifestPath)
		if err != nil {
			continue // No manifest, skip
		}

		ws.Id = entry.Name()
		ws.Dir = dir

		// Load SDL content for each design
		for _, design := range ws.Designs {
			sdlPath := filepath.Join(dir, design.File)
			if data, err := os.ReadFile(sdlPath); err == nil {
				key := ws.Id + "/" + design.Name
				s.designContents[key] = string(data)
			}
		}

		s.workspaces[ws.Id] = ws
		log.Printf("[CATALOG] Discovered workspace: %s (%d designs)", ws.Name, len(ws.Designs))
	}
}

// ListWorkspaces returns all discovered workspaces.
func (s *SystemCatalogService) ListWorkspaces() []*protos.Workspace {
	var out []*protos.Workspace
	for _, ws := range s.workspaces {
		out = append(out, ws)
	}
	return out
}

// GetWorkspace returns a workspace by ID.
func (s *SystemCatalogService) GetWorkspace(id string) *protos.Workspace {
	return s.workspaces[id]
}

// GetDesignContent returns the SDL content for a specific design within a workspace.
func (s *SystemCatalogService) GetDesignContent(workspaceId, designName string) string {
	return s.designContents[workspaceId+"/"+designName]
}

// GetAllDesignContents returns all SDL content for a workspace, keyed by design name.
func (s *SystemCatalogService) GetAllDesignContents(workspaceId string) map[string]string {
	out := make(map[string]string)
	ws := s.workspaces[workspaceId]
	if ws == nil {
		return out
	}
	for _, d := range ws.Designs {
		key := workspaceId + "/" + d.Name
		if content, ok := s.designContents[key]; ok {
			out[d.Name] = content
		}
	}
	return out
}
