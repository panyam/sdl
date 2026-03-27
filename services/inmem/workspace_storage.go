//go:build !wasm
// +build !wasm

package inmem

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/services"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WorkspaceStorage is an in-memory implementation of WorkspaceStorageProvider.
// Can be seeded from an examples directory containing sdl.json manifests.
type WorkspaceStorage struct {
	mu             sync.RWMutex
	workspaces     map[string]*protos.Workspace
	designContents map[string]string // "wsId/designName" -> SDL content
}

func NewWorkspaceStorage() *WorkspaceStorage {
	return &WorkspaceStorage{
		workspaces:     make(map[string]*protos.Workspace),
		designContents: make(map[string]string),
	}
}

// SeedFromExamples scans a directory for sdl.json manifests and loads them.
func (s *WorkspaceStorage) SeedFromExamples(baseDir string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		log.Printf("Warning: could not read examples dir %s: %v", baseDir, err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(baseDir, entry.Name())
		manifestPath := filepath.Join(dir, "sdl.json")

		ws, err := services.LoadWorkspaceManifest(manifestPath)
		if err != nil {
			continue
		}

		ws.Id = entry.Name()
		ws.Dir = dir
		now := timestamppb.Now()
		ws.CreatedAt = now
		ws.UpdatedAt = now

		for _, design := range ws.Designs {
			sdlPath := filepath.Join(dir, design.File)
			if data, err := os.ReadFile(sdlPath); err == nil {
				key := ws.Id + "/" + design.Name
				s.designContents[key] = string(data)
			}
		}

		s.workspaces[ws.Id] = ws
		log.Printf("[WORKSPACE] Seeded: %s (%d designs)", ws.Name, len(ws.Designs))
	}
}

func (s *WorkspaceStorage) LoadWorkspace(_ context.Context, id string) (*protos.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return nil, fmt.Errorf("workspace %s not found", id)
	}
	return ws, nil
}

func (s *WorkspaceStorage) SaveWorkspace(_ context.Context, id string, ws *protos.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workspaces[id] = ws
	return nil
}

func (s *WorkspaceStorage) DeleteWorkspace(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.workspaces, id)
	return nil
}

func (s *WorkspaceStorage) ListWorkspaces(_ context.Context) ([]*protos.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*protos.Workspace
	for _, ws := range s.workspaces {
		out = append(out, ws)
	}
	return out, nil
}

func (s *WorkspaceStorage) LoadDesignContent(_ context.Context, workspaceId, designName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := workspaceId + "/" + designName
	content, ok := s.designContents[key]
	if !ok {
		return "", fmt.Errorf("design %s not found in workspace %s", designName, workspaceId)
	}
	return content, nil
}

func (s *WorkspaceStorage) SaveDesignContent(_ context.Context, workspaceId, designName, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := workspaceId + "/" + designName
	s.designContents[key] = content
	return nil
}

func (s *WorkspaceStorage) LoadAllDesignContents(_ context.Context, workspaceId string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ws, ok := s.workspaces[workspaceId]
	if !ok {
		return nil, fmt.Errorf("workspace %s not found", workspaceId)
	}

	contents := make(map[string]string)
	for _, d := range ws.Designs {
		key := workspaceId + "/" + d.Name
		if c, ok := s.designContents[key]; ok {
			contents[d.Name] = c
		}
	}
	return contents, nil
}
