package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
)

// --- DesignService struct holds configuration and state ---
type DesignService struct {
	protos.UnimplementedDesignServiceServer
	BaseService
	clients *ClientMgr // Keep for potential auth needs

	idgen IDGen

	// Configuration / State moved from globals
	basePath string
	mutexMap sync.Map // Mutex map keyed by design ID
}

// --- NewDesignService Constructor ---
func NewDesignService(clients *ClientMgr, basePath string) *DesignService {
	// Initialize with defaults
	if basePath == "" {
		basePath = defaultDesignsBasePath
	}
	out := &DesignService{
		clients:  clients,
		basePath: basePath,
	}
	out.idgen = IDGen{
		NextIDFunc: (&SimpleIDGen{}).NextID,
		GetID:      func(kind, id string) (*GenID, error) { return out.getDesignId(id) },
	}
	// Ensure base directory exists on startup using the struct's path
	if err := ensureDir(out.basePath); err != nil {
		// Use log.Fatalf for critical startup errors
		log.Fatalf("Could not create base designs directory '%s': %v", out.basePath, err)
	}
	return out
}

// --- Helper functions are now methods on DesignService ---
func (s *DesignService) getDesignId(id string) (g *GenID, err error) {
	designPath := s.getDesignPath(id) // Use s.getDesignPath
	_, err = os.Stat(designPath)
	if err == nil {
		g = &GenID{Id: id}
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // Directory does not exist
	}
	// Some other error occurred
	slog.Error("Error checking design path existence", "id", id, "path", designPath, "error", err)
	return nil, err
}

// Helper to get or create a mutex for a specific design ID
func (s *DesignService) getDesignMutex(designId string) *sync.Mutex {
	mutex, _ := s.mutexMap.LoadOrStore(designId, &sync.Mutex{}) // Use s.mutexMap
	return mutex.(*sync.Mutex)
}

// Helper to get the path for a specific design
func (s *DesignService) getDesignPath(designId string) string {
	return filepath.Join(s.basePath, designId) // Use s.basePath
}

// Helper to get the path for the design metadata file
func (s *DesignService) getDesignMetadataPath(designId string) string {
	// Internally calls s.getDesignPath
	return filepath.Join(s.getDesignPath(designId), "design.json")
}

// Helper to get the base path for a design's sections
func (s *DesignService) getSectionsBasePath(designId string) string {
	// Internally calls s.getDesignPath
	return filepath.Join(s.getDesignPath(designId), "sections")
}

// --- Static/Utility Helpers (can remain outside the struct or moved) ---

// ensureDir doesn't depend on service state, can stay as utility
func ensureDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("Failed to create directory", "path", path, "error", err)
		return err
	}
	return nil
}

func (s *DesignService) CreateDesign(ctx context.Context, req *protos.CreateDesignRequest) (resp *protos.CreateDesignResponse, err error) {
	slog.Info("CreateDesign Request", "req", req)
	designProto := req.Design
	if designProto == nil {
		return nil, status.Error(codes.InvalidArgument, "Design payload cannot be nil")
	}

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		slog.Error("Login enforcement failed", "error", err)
		return nil, err
	}
	if !ENFORCE_LOGIN && ownerId == "" {
		ownerId = "dev_user"
	}

	designId := designProto.Id

	if designId == "" {
		slog.Debug("Design ID not provided, attempting auto-generation")
		genid, err := s.idgen.NextID("design", ownerId, time.Unix(0, 0))
		if genid == nil {
			slog.Error("Failed to generate a unique design ID after multiple retries", "error", err)
			return nil, status.Error(codes.Internal, "Failed to generate unique design ID")
		}
		designId = genid.Id
		slog.Info("Generated unique Design ID", "id", designId)
	} else {
		slog.Debug("Using client-provided Design ID", "id", designId)
	}

	if strings.TrimSpace(designProto.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "Design name cannot be empty")
	}

	mutex := s.getDesignMutex(designId) // Use s.getDesignMutex
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex", "designId", designId)

	// Use helper methods for paths
	designPath := s.getDesignPath(designId)
	metadataPath := s.getDesignMetadataPath(designId)
	if _, err := os.Stat(designPath); err == nil {
		if _, err := os.Stat(metadataPath); err == nil {
			slog.Warn("Design already exists", "designId", designId)
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Design with id '%s' already exists", designId))
		}
		slog.Warn("Design directory exists but metadata missing", "designId", designId, "path", metadataPath)
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Design with id '%s' potentially exists but metadata is missing", designId))
	} else if !errors.Is(err, os.ErrNotExist) {
		slog.Error("Error checking design path", "designId", designId, "path", designPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to check design existence")
	}

	sectionsPath := s.getSectionsBasePath(designId)
	if err := ensureDir(sectionsPath); err != nil { // ensureDir is still static utility
		return nil, status.Error(codes.Internal, "Failed to create design directories")
	}
	slog.Debug("Ensured directories exist", "designPath", designPath, "sectionsPath", sectionsPath)

	now := time.Now()
	metadata := Design{
		BaseModel: BaseModel{
			CreatedAt: now,
			UpdatedAt: now,
		},
		Id:          designId,
		OwnerId:     ownerId,
		Name:        designProto.Name,
		Description: designProto.Description,
		Visibility:  designProto.Visibility,
		VisibleTo:   designProto.VisibleTo,
		SectionIds:  []string{},
	}
	if metadata.Visibility == "" {
		metadata.Visibility = "private"
	}

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal design metadata", "designId", designId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to serialize design metadata")
	}

	err = os.WriteFile(metadataPath, jsonData, 0644)
	if err != nil {
		slog.Error("Failed to write design metadata file", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to save design metadata")
	}
	slog.Info("Successfully created design metadata", "designId", designId, "path", metadataPath)

	createdDesignProto := DesignToProto(&metadata) // Static utility

	resp = &protos.CreateDesignResponse{
		Design: createdDesignProto,
	}
	return resp, nil
}

func (s *DesignService) GetDesign(ctx context.Context, req *protos.GetDesignRequest) (resp *protos.GetDesignResponse, err error) {
	slog.Info("GetDesign Request", "req", req)
	designId := req.Id
	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty")
	}

	metadataPath := s.getDesignMetadataPath(designId) // Use s.getDesignMetadataPath
	slog.Debug("Attempting to read metadata", "path", metadataPath)

	if _, err := os.Stat(metadataPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Metadata file not found", "designId", designId, "path", metadataPath)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", designId))
		}
		slog.Error("Error checking metadata file", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to access design metadata")
	}

	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		slog.Error("Failed to read metadata file", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata")
	}

	var metadata Design
	err = json.Unmarshal(jsonData, &metadata)
	if err != nil {
		slog.Error("Failed to unmarshal metadata JSON", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	// TODO: Permission Checks

	designProto := DesignToProto(&metadata) // Static utility
	resp = &protos.GetDesignResponse{
		Design: designProto,
	}
	slog.Info("Successfully retrieved design metadata", "designId", designId)
	return resp, nil
}

func (s *DesignService) ListDesigns(ctx context.Context, req *protos.ListDesignsRequest) (resp *protos.ListDesignsResponse, err error) {
	slog.Info("ListDesigns Request", "req", req)

	// Use s.basePath
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		slog.Error("Failed to read designs directory", "path", s.basePath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to list designs")
	}

	var allMetadata []*Design
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		designId := entry.Name()
		metadataPath := s.getDesignMetadataPath(designId) // Use s.getDesignMetadataPath

		if _, statErr := os.Stat(metadataPath); statErr == nil {
			jsonData, readErr := os.ReadFile(metadataPath)
			if readErr != nil {
				slog.Warn("Failed to read metadata file during list", "designId", designId, "path", metadataPath, "error", readErr)
				continue
			}

			var metadata Design
			unmarshalErr := json.Unmarshal(jsonData, &metadata)
			if unmarshalErr != nil {
				slog.Warn("Failed to unmarshal metadata JSON during list", "designId", designId, "path", metadataPath, "error", unmarshalErr)
				continue
			}
			if metadata.Id != designId {
				slog.Warn("Mismatch between directory name and metadata Id", "dirName", designId, "metadataId", metadata.Id)
				continue
			}
			allMetadata = append(allMetadata, &metadata)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Error stating metadata file during list", "designId", designId, "path", metadataPath, "error", statErr)
		}
	}
	slog.Debug("Collected metadata for designs", "count", len(allMetadata))

	// Filtering logic remains the same...
	var filteredMetadata []*Design
	loggedInUserId, _ := s.EnsureLoggedIn(ctx)
	for _, meta := range allMetadata {
		if req.OwnerId != "" && meta.OwnerId != req.OwnerId {
			continue
		}
		isOwner := (!ENFORCE_LOGIN && meta.OwnerId == "dev_user") || (ENFORCE_LOGIN && loggedInUserId == meta.OwnerId)
		isVisible := meta.Visibility == "public" || isOwner

		if req.LimitToPublic && meta.Visibility != "public" {
			continue
		}
		if !isVisible && req.OwnerId == "" && !req.LimitToPublic {
			// continue // Re-evaluate this permission logic if needed
		}
		filteredMetadata = append(filteredMetadata, meta)
	}
	slog.Debug("Filtered metadata for designs", "count", len(filteredMetadata))

	// Sorting logic remains the same...
	sort.Slice(filteredMetadata, func(i, j int) bool {
		m1 := filteredMetadata[i]
		m2 := filteredMetadata[j]
		switch req.OrderBy {
		case "recent", "":
			return m1.UpdatedAt.After(m2.UpdatedAt)
		case "name":
			return strings.ToLower(m1.Name) < strings.ToLower(m2.Name)
		case "created":
			return m1.CreatedAt.After(m2.CreatedAt)
		default:
			return m1.UpdatedAt.After(m2.UpdatedAt)
		}
	})
	slog.Debug("Sorted metadata")

	// Pagination logic remains the same...
	totalResults := len(filteredMetadata)
	start := 0
	end := totalResults
	pageSize := 20
	hasMore := false

	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
		start = int(req.Pagination.PageOffset)
		if start < 0 {
			start = 0
		}
		if start >= totalResults {
			filteredMetadata = []*Design{}
			start = 0
			end = 0
		} else {
			end = start + pageSize
			if end >= totalResults {
				end = totalResults
			} else {
				hasMore = true
			}
		}
		filteredMetadata = filteredMetadata[start:end]
	}
	slog.Debug("Applied pagination", "offset", start, "limit", pageSize, "actualCount", len(filteredMetadata), "hasMore", hasMore)

	finalProtos := make([]*protos.Design, len(filteredMetadata))
	for i, meta := range filteredMetadata {
		finalProtos[i] = DesignToProto(meta) // Static utility
	}

	resp = &protos.ListDesignsResponse{
		Designs: finalProtos,
		Pagination: &protos.PaginationResponse{
			HasMore:      hasMore,
			TotalResults: int32(totalResults),
		},
	}

	slog.Info("Successfully listed designs", "request", req, "responseCount", len(resp.Designs))
	return resp, nil
}

func (s *DesignService) UpdateDesign(ctx context.Context, req *protos.UpdateDesignRequest) (resp *protos.UpdateDesignResponse, err error) {
	slog.Info("UpdateDesign Request", "req", req)
	designId := req.Design.GetId()
	updateMask := req.UpdateMask

	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty in update request")
	}
	if req.Design == nil {
		return nil, status.Error(codes.InvalidArgument, "Design payload cannot be nil in update request")
	}
	if updateMask == nil || len(updateMask.Paths) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Update mask must be provided with at least one field path")
	}

	validPaths := map[string]bool{
		"name":        true,
		"description": true,
		"visibility":  true,
		"visible_to":  true,
	}
	protoToJsonPathMap := map[string]string{
		"name":        "Name",
		"description": "Description",
		"visibility":  "Visibility",
		"visible_to":  "VisibleTo",
	}

	for _, path := range updateMask.Paths {
		effectivePath := strings.TrimPrefix(path, "design.")
		if !validPaths[effectivePath] {
			slog.Warn("Invalid path in update mask for metadata update", "path", path)
			return nil, status.Errorf(codes.InvalidArgument, "Update mask contains invalid or unsupported path for metadata update: %s", path)
		}
	}

	mutex := s.getDesignMutex(designId) // Use s.getDesignMutex
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for update", "designId", designId)

	metadataPath := s.getDesignMetadataPath(designId) // Use s.getDesignMetadataPath
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design not found for update", "designId", designId, "path", metadataPath)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", designId))
		}
		slog.Error("Failed to read metadata file for update", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata for update")
	}

	var metadata Design
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal metadata JSON for update", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	loggedInUserId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN {
		if err != nil {
			return nil, err
		}
		if loggedInUserId != metadata.OwnerId {
			slog.Warn("Permission denied for update", "designId", designId, "loggedInUser", loggedInUserId, "owner", metadata.OwnerId)
			return nil, status.Error(codes.PermissionDenied, "User does not have permission to update this design")
		}
	}

	updated := false
	for _, path := range updateMask.Paths {
		effectivePath := strings.TrimPrefix(path, "design.")
		jsonField := protoToJsonPathMap[effectivePath]

		switch jsonField {
		case "Name":
			newValue := req.Design.Name
			if strings.TrimSpace(newValue) == "" {
				return nil, status.Error(codes.InvalidArgument, "Design name cannot be updated to empty")
			}
			if metadata.Name != newValue {
				metadata.Name = newValue
				updated = true
			}
		case "Description":
			newValue := req.Design.Description
			if metadata.Description != newValue {
				metadata.Description = newValue
				updated = true
			}
		case "Visibility":
			newValue := req.Design.Visibility
			if metadata.Visibility != newValue {
				metadata.Visibility = newValue
				updated = true
			}
		case "VisibleTo":
			newValue := req.Design.VisibleTo
			var diff bool // Declare diff here
			if len(metadata.VisibleTo) != len(newValue) {
				diff = true // Set diff to true if lengths differ
			} else {
				diff = false // Initialize diff to false
				sort.Strings(metadata.VisibleTo)
				sort.Strings(newValue)
				for i := range metadata.VisibleTo {
					if metadata.VisibleTo[i] != newValue[i] {
						diff = true
						break
					}
				}
			}
			if diff { // Now diff is in the correct scope
				metadata.VisibleTo = newValue
				updated = true
			}
		default:
			slog.Warn("Encountered unhandled valid path during update apply", "path", path)
		}
	}

	if updated {
		metadata.UpdatedAt = time.Now()
		slog.Info("Metadata changed, attempting save", "designId", designId)

		updatedJsonData, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal updated metadata", "designId", designId, "error", err)
			return nil, status.Error(codes.Internal, "Failed to serialize updated metadata")
		}

		err = os.WriteFile(metadataPath, updatedJsonData, 0644)
		if err != nil {
			slog.Error("Failed to write updated metadata file", "designId", designId, "path", metadataPath, "error", err)
			return nil, status.Error(codes.Internal, "Failed to save updated design metadata")
		}
		slog.Info("Successfully updated design metadata file", "designId", designId, "path", metadataPath)
	} else {
		slog.Info("No changes detected based on update mask", "designId", designId)
	}

	updatedProto := DesignToProto(&metadata) // Static utility
	resp = &protos.UpdateDesignResponse{
		Design: updatedProto,
	}
	return resp, nil
}

func (s *DesignService) DeleteDesign(ctx context.Context, req *protos.DeleteDesignRequest) (resp *protos.DeleteDesignResponse, err error) {
	slog.Info("DeleteDesign Request", "req", req)
	designId := req.Id

	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty for delete request")
	}

	mutex := s.getDesignMutex(designId) // Use s.getDesignMutex
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for delete", "designId", designId)

	metadataPath := s.getDesignMetadataPath(designId) // Use s.getDesignMetadataPath
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design not found for delete", "designId", designId, "path", metadataPath)
			return &protos.DeleteDesignResponse{}, nil
		}
		slog.Error("Failed to read metadata file for delete permission check", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata for delete")
	}

	var metadata Design
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal metadata JSON for delete permission check", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	loggedInUserId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN {
		if err != nil {
			return nil, err
		}
		if loggedInUserId != metadata.OwnerId {
			slog.Warn("Permission denied for delete", "designId", designId, "loggedInUser", loggedInUserId, "owner", metadata.OwnerId)
			return nil, status.Error(codes.PermissionDenied, "User does not have permission to delete this design")
		}
	}

	designPath := s.getDesignPath(designId) // Use s.getDesignPath
	slog.Info("Attempting to delete design directory", "designId", designId, "path", designPath)
	err = os.RemoveAll(designPath)
	if err != nil {
		if _, statErr := os.Stat(designPath); errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Design directory already gone during delete", "designId", designId, "path", designPath)
			return &protos.DeleteDesignResponse{}, nil
		}
		slog.Error("Failed to delete design directory", "designId", designId, "path", designPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to delete design")
	}

	slog.Info("Successfully deleted design", "designId", designId, "path", designPath)

	resp = &protos.DeleteDesignResponse{}
	return resp, nil
}

func (s *DesignService) GetDesigns(ctx context.Context, req *protos.GetDesignsRequest) (resp *protos.GetDesignsResponse, err error) {
	slog.Warn("GetDesigns called, which is inefficient for filesystem backend")
	return nil, status.Error(codes.Unimplemented, "GetDesigns is not efficiently implemented for the filesystem backend. Use GetDesign for individual IDs.")
}
