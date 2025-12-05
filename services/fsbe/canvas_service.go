//go:build !wasm
// +build !wasm

package fsbe

import (
	"context"
	"fmt"
	"time"

	"github.com/turnforge/turnengine/engine/storage"
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

var CANVASES_STORAGE_DIR = ""

// FSCanvasService implements Canvas CRUD operations with file-based storage
type FSCanvasService struct {
	storage *storage.FileStorage
}

// NewFSCanvasService creates a new FSCanvasService
func NewFSCanvasService(storageDir string) *FSCanvasService {
	if storageDir == "" {
		if CANVASES_STORAGE_DIR == "" {
			CANVASES_STORAGE_DIR = DevDataPath("storage/canvases")
		}
		storageDir = CANVASES_STORAGE_DIR
	}
	return &FSCanvasService{
		storage: storage.NewFileStorage(storageDir),
	}
}

// ListCanvases returns all available canvases
func (s *FSCanvasService) ListCanvases(ctx context.Context, req *protos.ListCanvasesRequest) (*protos.ListCanvasesResponse, error) {
	resp := &protos.ListCanvasesResponse{
		Canvases: []*protos.Canvas{},
	}

	canvases, err := storage.ListFSEntities[*protos.Canvas](s.storage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list canvases: %w", err)
	}

	resp.Canvases = canvases
	return resp, nil
}

// GetCanvas returns a specific canvas by ID
func (s *FSCanvasService) GetCanvas(ctx context.Context, req *protos.GetCanvasRequest) (*protos.GetCanvasResponse, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("canvas ID is required")
	}

	canvas, err := storage.LoadFSArtifact[*protos.Canvas](s.storage, req.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("canvas not found: %w", err)
	}

	return &protos.GetCanvasResponse{
		Canvas: canvas,
	}, nil
}

// CreateCanvas creates a new canvas
func (s *FSCanvasService) CreateCanvas(ctx context.Context, req *protos.CreateCanvasRequest) (*protos.CreateCanvasResponse, error) {
	if req.Canvas == nil {
		return nil, fmt.Errorf("canvas data is required")
	}

	canvasId, err := s.storage.CreateEntity(req.Canvas.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to create canvas: %w", err)
	}
	req.Canvas.Id = canvasId

	now := time.Now()
	req.Canvas.CreatedAt = tspb.New(now)
	req.Canvas.UpdatedAt = tspb.New(now)

	if err := s.storage.SaveArtifact(req.Canvas.Id, "metadata", req.Canvas); err != nil {
		return nil, fmt.Errorf("failed to save canvas: %w", err)
	}

	return &protos.CreateCanvasResponse{
		Canvas: req.Canvas,
	}, nil
}

// UpdateCanvas updates an existing canvas
func (s *FSCanvasService) UpdateCanvas(ctx context.Context, req *protos.UpdateCanvasRequest) (*protos.UpdateCanvasResponse, error) {
	if req.Canvas == nil || req.Canvas.Id == "" {
		return nil, fmt.Errorf("canvas ID is required")
	}

	// Load existing canvas
	canvas, err := storage.LoadFSArtifact[*protos.Canvas](s.storage, req.Canvas.Id, "metadata")
	if err != nil {
		return nil, fmt.Errorf("canvas not found: %w", err)
	}

	// Update fields
	if req.Canvas.Name != "" {
		canvas.Name = req.Canvas.Name
	}
	if req.Canvas.Description != "" {
		canvas.Description = req.Canvas.Description
	}
	canvas.UpdatedAt = tspb.New(time.Now())

	if err := s.storage.SaveArtifact(req.Canvas.Id, "metadata", canvas); err != nil {
		return nil, fmt.Errorf("failed to update canvas: %w", err)
	}

	return &protos.UpdateCanvasResponse{
		Canvas: canvas,
	}, nil
}

// DeleteCanvas deletes a canvas
func (s *FSCanvasService) DeleteCanvas(ctx context.Context, req *protos.DeleteCanvasRequest) (*protos.DeleteCanvasResponse, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("canvas ID is required")
	}

	if err := s.storage.DeleteEntity(req.Id); err != nil {
		return nil, fmt.Errorf("failed to delete canvas: %w", err)
	}

	return &protos.DeleteCanvasResponse{}, nil
}
