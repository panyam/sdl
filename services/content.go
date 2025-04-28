// FILE: ./services/content.go
package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// ContentService manages the raw byte content of sections.
type ContentService struct {
	protos.UnimplementedContentServiceServer
	BaseService
	// designSvc *DesignService // No longer needed
	store *DesignStore // Use DesignStore directly for path logic
}

// NewContentService creates a new ContentService.
// It now depends on DesignStore instead of DesignService.
func NewContentService(store *DesignStore) *ContentService {
	if store == nil {
		slog.Error("Cannot create ContentService with a nil DesignStore")
		panic("Cannot create ContentService with a nil DesignStore")
	}
	return &ContentService{
		store: store,
	}
}

// GetContent retrieves raw content bytes for a named file within a section.
func (s *ContentService) GetContent(ctx context.Context, req *protos.GetContentRequest) (*protos.GetContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId
	contentName := req.Name
	if designId == "" || sectionId == "" || contentName == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Name must be provided")
	}
	slog.Info("GetContent Request", "designId", designId, "sectionId", sectionId, "name", contentName)

	// TODO: Add permission checks (can user view this design/section?)

	// Check if parent directories exist using the store
	designPath := s.store.getDesignPath(designId) // Use store's path method
	if _, err := os.Stat(designPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design directory not found for GetContent", "path", designPath)
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		slog.Error("Error checking design directory", "path", designPath, "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to access design path: %v", err)
	}
	sectionBasePath := s.store.GetSectionBasePath(designId, sectionId) // Use store's path method
	if _, err := os.Stat(sectionBasePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Section directory not found for GetContent", "path", sectionBasePath)
			return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
		}
		slog.Error("Error checking section directory", "path", sectionBasePath, "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to access section path: %v", err)
	}

	// Get content path using the store
	contentPath := s.store.GetContentPath(designId, sectionId, contentName)
	slog.Debug("Attempting to read content file", "path", contentPath)

	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		slog.Error("os.ReadFile failed", "path", contentPath, "error", err, "isNotExist", errors.Is(err, os.ErrNotExist))
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Content file not found", "path", contentPath)
			return nil, status.Errorf(codes.NotFound, "Content '%s' not found for section '%s/%s'", contentName, designId, sectionId)
		}
		return nil, status.Error(codes.Internal, "Failed to read content file")
	}

	// Get file info for metadata (like UpdatedAt)
	fileInfo, statErr := os.Stat(contentPath)
	if statErr != nil {
		slog.Warn("Failed to get file stats after reading content", "path", contentPath, "error", statErr)
	}

	resp := &protos.GetContentResponse{
		ContentBytes: contentBytes,
		Content: &protos.Content{
			Name: contentName,
			// Type and Format still need dedicated metadata storage.
			// Leaving them empty for now.
			UpdatedAt: nil,
		},
	}
	if statErr == nil {
		resp.Content.UpdatedAt = tspb.New(fileInfo.ModTime())
	}

	slog.Info("Successfully retrieved content", "path", contentPath, "size", len(contentBytes))
	return resp, nil
}

// SetContent saves or updates raw content bytes for a named file within a section.
// It NO LONGER updates design/section metadata timestamps.
func (s *ContentService) SetContent(ctx context.Context, req *protos.SetContentRequest) (*protos.SetContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId
	contentName := req.GetName()

	if designId == "" || sectionId == "" || contentName == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Name must be provided")
	}

	slog.Info("SetContent Request", "designId", designId, "sectionId", sectionId, "name", contentName)
	designPath := s.store.getDesignPath(designId)
	if _, err := os.Stat(designPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design directory not found for SetContent", "path", designPath)
			// Return NotFound if the design directory doesn't exist
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		// Handle other errors checking the design path (e.g., permissions)
		slog.Error("Error checking design directory for SetContent", "path", designPath, "error", err)
		return nil, status.Errorf(codes.Internal, "Failed to access design path: %v", err)
	}

	// --- Check Permissions (Simplified - Requires actual implementation) ---
	// _, err := s.EnsureLoggedIn(ctx) // Or check ownership based on design ID
	// if ENFORCE_LOGIN && err != nil {
	// 	return nil, err
	// }
	// Need to fetch design metadata to check owner - this adds back a dependency or requires passing owner info.
	// For now, skipping strict permission check within SetContent itself. Assume caller verified.

	// --- Ensure Parent Directories Exist using Store ---
	sectionBasePath := s.store.GetSectionBasePath(designId, sectionId)
	err := ensureDir(sectionBasePath) // Ensure section dir exists
	if err != nil {
		// This could fail due to permissions or issues creating the design dir if it's missing
		slog.Error("Failed to ensure section directory exists for SetContent", "path", sectionBasePath, "error", err)
		// Check if the design dir itself is missing
		if _, designStatErr := os.Stat(s.store.getDesignPath(designId)); errors.Is(designStatErr, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to ensure section directory: %v", err)
	}

	// --- Write Content Bytes ---
	contentPath := s.store.GetContentPath(designId, sectionId, contentName)
	slog.Debug("Attempting to write content file", "path", contentPath)
	err = os.WriteFile(contentPath, req.ContentBytes, 0644)
	if err != nil {
		slog.Error("Failed to write content bytes", "path", contentPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to save content")
	}
	slog.Info("Successfully wrote content bytes", "path", contentPath, "size", len(req.ContentBytes))

	// --- REMOVED Timestamp Update Logic ---
	// Timestamps (UpdatedAt in design.json and section's main.json)
	// should be updated by the caller via DesignService methods (e.g., UpdateSection)
	// after the content has been successfully set.

	// --- Construct Response ---
	// Get ModTime of the file we just wrote
	var modTime *tspb.Timestamp
	if fileInfo, statErr := os.Stat(contentPath); statErr == nil {
		modTime = tspb.New(fileInfo.ModTime())
	} else {
		slog.Warn("Could not stat file after writing for timestamp", "path", contentPath, "error", statErr)
		modTime = tspb.Now() // Fallback to current time
	}

	// Return minimal info - Name and UpdatedAt of the content file itself.
	// Type and Format are not stored/managed by this service currently.
	resp := &protos.SetContentResponse{
		Content: &protos.Content{
			Name:      contentName,
			UpdatedAt: modTime,
			// Type: req.ContentType, // Not stored/returned
			// Format: req.Format, // Not stored/returned
		},
	}
	return resp, nil
}

// GetContentBytes is a helper primarily used internally by other services (like LlmService).
// It returns raw bytes or ErrNoSuchEntity.
func (s *ContentService) GetContentBytes(ctx context.Context, designId, sectionId, contentName string) ([]byte, error) {
	// Basic validation
	if designId == "" || sectionId == "" || contentName == "" {
		return nil, fmt.Errorf("designId, sectionId, and contentName are required")
	}
	// TODO: Consider adding permission check here too?

	// Get path via store
	contentPath := s.store.GetContentPath(designId, sectionId, contentName)
	slog.Debug("GetContentBytes: Reading path", "path", contentPath)

	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("GetContentBytes: File not found", "path", contentPath)
			return nil, ErrNoSuchEntity // Return specific error
		}
		slog.Error("GetContentBytes: ReadFile failed", "path", contentPath, "error", err)
		return nil, fmt.Errorf("failed to read content bytes: %w", err) // Internal error
	}
	return contentBytes, nil
}
