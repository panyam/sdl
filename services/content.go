// FILE: ./services/content.go
package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// --- ContentService struct ---
// Needs access to design service helpers or similar logic.
// Let's assume it gets the DesignService instance or relevant parts.
type ContentService struct {
	protos.UnimplementedContentServiceServer
	BaseService
	designService *DesignService // Access to mutexes and path helpers
}

// --- NewContentService Constructor ---
func NewContentService(ds *DesignService) *ContentService {
	if ds == nil {
		panic("DesignService cannot be nil for ContentService")
	}
	return &ContentService{
		designService: ds,
	}
}

// --- ContentService RPC Implementations ---

func (s *ContentService) GetContent(ctx context.Context, req *protos.GetContentRequest) (*protos.GetContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId
	contentName := req.Name
	if designId == "" || sectionId == "" || contentName == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Content Name must be provided")
	}
	slog.Info("GetContent Request", "designId", designId, "sectionId", sectionId, "name", contentName)

	// Optional: Permission check - does user have read access to the design?
	// Could call designService.readDesignMetadata and check owner/visibility.

	contentPath := s.designService.getContentPath(designId, sectionId, contentName)
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Content file not found", "path", contentPath)
			return nil, status.Errorf(codes.NotFound, "Content '%s' not found for section '%s/%s'", contentName, designId, sectionId)
		}
		slog.Error("Failed to read content file", "path", contentPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read content")
	}

	// TODO: Read content metadata (type, format) if stored separately (e.g., in main.json).
	// For now, return minimal metadata based on common conventions or defaults.
	contentType := "application/octet-stream" // Default
	format := ""
	// Example: Infer type/format from name extension? (Less reliable)
	if strings.HasSuffix(contentName, ".json") {
		contentType = "application/json"
		if strings.Contains(contentName, "excalidraw") {
			format = "excalidraw/json"
		}
	} else if strings.HasSuffix(contentName, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(contentName, ".svg") {
		contentType = "image/svg+xml"
	}

	// Placeholder for reading actual metadata if implemented
	fileInfo, _ := os.Stat(contentPath) // Get file info for timestamps/size
	modTime := time.Now()
	if fileInfo != nil {
		modTime = fileInfo.ModTime()
	}

	resp := &protos.GetContentResponse{
		Content: &protos.Content{
			Name:      contentName,
			Type:      contentType,
			Format:    format,
			UpdatedAt: tspb.New(modTime),
			// CreatedAt would ideally be stored, default to UpdatedAt for now
			CreatedAt: tspb.New(modTime),
		},
		ContentBytes: contentBytes,
	}

	slog.Info("Successfully retrieved content", "path", contentPath, "size", len(contentBytes))
	return resp, nil
}

func (s *ContentService) SetContent(ctx context.Context, req *protos.SetContentRequest) (*protos.SetContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId
	contentProto := req.Content // Contains metadata like name, type, format

	if designId == "" || sectionId == "" || contentProto == nil || contentProto.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Content with Name must be provided")
	}
	contentName := contentProto.Name
	updateMask := req.UpdateMask

	slog.Info("SetContent Request", "designId", designId, "sectionId", sectionId, "name", contentName, "mask", updateMask)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	// --- Acquire Lock ---
	mutex := s.designService.getDesignMutex(designId) // Use design-level lock
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for SetContent", "designId", designId, "sectionId", sectionId, "name", contentName)

	// --- Check Permissions & Existence ---
	designMeta, err := s.designService.readDesignMetadata(designId)
	if err != nil {
		return nil, err // Handles NotFound
	}
	if ENFORCE_LOGIN && ownerId != designMeta.OwnerId {
		slog.Warn("Permission denied for SetContent", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot set content in this design")
	}
	// Ensure section directory exists (section metadata doesn't need to be read here unless updating it)
	sectionPath := s.designService.getSectionPath(designId, sectionId)
	if _, err := os.Stat(sectionPath); errors.Is(err, os.ErrNotExist) {
		// If AddSection didn't run or failed, the dir might not exist
		slog.Warn("Section directory not found during SetContent", "path", sectionPath)
		return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
	} else if err != nil {
		slog.Error("Error checking section directory existence", "path", sectionPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to access section directory")
	}

	// --- Handle Content Byte Update ---
	contentPath := s.designService.getContentPath(designId, sectionId, contentName)
	bytesUpdated := false
	// Check if content_bytes is explicitly in the mask OR if the mask is nil/empty (implying full replace)
	// Or adopt a convention: if content_bytes is provided, always write it.
	// Let's assume if bytes are provided, we write them.
	if req.ContentBytes != nil {
		err = os.WriteFile(contentPath, req.ContentBytes, 0644)
		if err != nil {
			slog.Error("Failed to write content bytes", "path", contentPath, "error", err)
			return nil, status.Error(codes.Internal, "Failed to save content")
		}
		bytesUpdated = true
		slog.Info("Successfully wrote content bytes", "path", contentPath, "size", len(req.ContentBytes))
	}

	// --- Handle Metadata Update (TBD - if storing Content metadata) ---
	// Example: If Content metadata were stored in section's main.json:
	// sectionMeta, err := s.designService.readSectionData(designId, sectionId)
	// if err == nil {
	//   updated := false
	//   if sectionMeta.ContentMetadata == nil { sectionMeta.ContentMetadata = make(...) }
	//   if mask includes "type" && contentProto.Type != ... { update; updated = true }
	//   if mask includes "format" && contentProto.Format != ... { update; updated = true }
	//   if updated { s.designService.writeSectionMetadata(...) }
	// } else { log error reading section meta }
	metadataUpdated := false // Placeholder

	// --- Update Timestamps ---
	now := time.Now()
	if bytesUpdated || metadataUpdated {
		// Update section's main.json timestamp
		sectionMeta, err := s.designService.readSectionData(designId, sectionId)
		if err == nil {
			sectionMeta.UpdatedAt = now
			if err_write := s.designService.writeSectionMetadata(designId, sectionId, sectionMeta); err_write != nil {
				slog.Error("Failed to update section metadata timestamp after SetContent", "path", s.designService.getSectionMetadataPath(designId, sectionId), "error", err_write)
				// Continue, as content was potentially saved
			}
		} else {
			slog.Warn("Could not read section metadata to update timestamp", "designId", designId, "sectionId", sectionId, "error", err)
		}

		// Update design's design.json timestamp
		designMeta.UpdatedAt = now
		if err_write := s.designService.writeDesignMetadata(designId, designMeta); err_write != nil {
			slog.Error("Failed to update design metadata timestamp after SetContent", "path", s.designService.getDesignMetadataPath(designId), "error", err_write)
			// Continue
		}
	}

	// Return the metadata provided in the request (or read back if implemented)
	finalContentProto := &protos.Content{
		Name:      contentName,
		Type:      contentProto.GetType(), // Use values from request proto
		Format:    contentProto.GetFormat(),
		UpdatedAt: tspb.New(now),
		// CreatedAt needs proper tracking if important
	}

	resp := &protos.SetContentResponse{
		Content: finalContentProto,
	}
	return resp, nil
}

func (s *ContentService) DeleteContent(ctx context.Context, req *protos.DeleteContentRequest) (*protos.DeleteContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId
	contentName := req.Name // Corrected field name from proto

	if designId == "" || sectionId == "" || contentName == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Content Name must be provided")
	}
	slog.Info("DeleteContent Request", "designId", designId, "sectionId", sectionId, "name", contentName)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	// --- Acquire Lock ---
	mutex := s.designService.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for DeleteContent", "designId", designId, "sectionId", sectionId, "name", contentName)

	// --- Check Permissions ---
	designMeta, err := s.designService.readDesignMetadata(designId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			slog.Warn("Design not found during DeleteContent", "designId", designId)
			return &protos.DeleteContentResponse{}, nil // Design gone, content implicitly gone
		}
		return nil, err // Other read error
	}
	if ENFORCE_LOGIN && ownerId != designMeta.OwnerId {
		slog.Warn("Permission denied for DeleteContent", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot delete content from this design")
	}

	// --- Delete File ---
	contentPath := s.designService.getContentPath(designId, sectionId, contentName)
	err = os.Remove(contentPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Content file not found during delete (idempotent)", "path", contentPath)
			return &protos.DeleteContentResponse{}, nil // Already deleted
		}
		slog.Error("Failed to delete content file", "path", contentPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to delete content file")
	}
	slog.Info("Successfully deleted content file", "path", contentPath)

	// --- Update Timestamps ---
	now := time.Now()
	// Update section's main.json timestamp
	sectionMeta, err := s.designService.readSectionData(designId, sectionId)
	if err == nil {
		sectionMeta.UpdatedAt = now
		if err_write := s.designService.writeSectionMetadata(designId, sectionId, sectionMeta); err_write != nil {
			slog.Error("Failed to update section metadata timestamp after DeleteContent", "error", err_write)
			// Continue, main action succeeded
		}
	} else {
		slog.Warn("Could not read section metadata to update timestamp after delete", "error", err)
	}
	// Update design's design.json timestamp
	designMeta.UpdatedAt = now
	if err_write := s.designService.writeDesignMetadata(designId, designMeta); err_write != nil {
		slog.Error("Failed to update design metadata timestamp after DeleteContent", "error", err_write)
		// Continue
	}

	return &protos.DeleteContentResponse{}, nil
}
