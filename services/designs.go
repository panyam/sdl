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
	"google.golang.org/grpc/status" // Added for update mask processing
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// "strings"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// --- Constants, Structs, Helpers (Unchanged) ---
const designsBasePath = "./data/designs"

// Just for test
const ENFORCE_LOGIN = false
const FAKE_USER_ID = ""

var designMutexes sync.Map

type DesignMetadata struct {
	ID          string        `json:"id"`
	OwnerID     string        `json:"ownerId"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Visibility  string        `json:"visibility"`
	VisibleTo   []string      `json:"visibleTo"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Sections    []SectionMeta `json:"sections"`
}

type SectionMeta struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

func getDesignMutex(designId string) *sync.Mutex {
	mutex, _ := designMutexes.LoadOrStore(designId, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

func getDesignPath(designId string) string {
	return filepath.Join(designsBasePath, designId)
}

func getDesignMetadataPath(designId string) string {
	return filepath.Join(getDesignPath(designId), "design.json")
}

func getSectionsBasePath(designId string) string {
	return filepath.Join(getDesignPath(designId), "sections")
}

func ensureDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("Failed to create directory", "path", path, "error", err)
		return err
	}
	return nil
}

func DesignMetadataToProto(input *DesignMetadata) *protos.Design {
	out := &protos.Design{
		Id:          input.ID,
		OwnerId:     input.OwnerID,
		Name:        input.Name,
		Description: input.Description,
		Visibility:  input.Visibility,
		VisibleTo:   input.VisibleTo,
		CreatedAt:   tspb.New(input.CreatedAt),
		UpdatedAt:   tspb.New(input.UpdatedAt),
		SectionIds:  make([]string, len(input.Sections)),
	}
	for i, secMeta := range input.Sections {
		out.SectionIds[i] = secMeta.ID
	}
	// Note: ContentMetadata is not handled in this file-based metadata version yet.
	return out
}

// --- Service Struct and Constructor (Unchanged) ---
type DesignService struct {
	protos.UnimplementedDesignServiceServer
	BaseService
	clients *ClientMgr
}

func NewDesignService(clients *ClientMgr) *DesignService {
	out := &DesignService{
		clients: clients,
	}
	if err := ensureDir(designsBasePath); err != nil {
		log.Fatalf("Could not create base designs directory '%s': %v", designsBasePath, err)
	}
	return out
}

// --- CreateDesign (Unchanged) ---
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

	if designProto.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID must be provided during creation in this phase")
	}
	designId := designProto.Id
	if strings.TrimSpace(designProto.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "Design name cannot be empty")
	}

	mutex := getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex", "designId", designId)

	designPath := getDesignPath(designId)
	metadataPath := getDesignMetadataPath(designId)
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

	sectionsPath := getSectionsBasePath(designId)
	if err := ensureDir(sectionsPath); err != nil {
		return nil, status.Error(codes.Internal, "Failed to create design directories")
	}
	slog.Debug("Ensured directories exist", "designPath", designPath, "sectionsPath", sectionsPath)

	now := time.Now()
	metadata := DesignMetadata{
		ID:          designId,
		OwnerID:     ownerId,
		Name:        designProto.Name,
		Description: designProto.Description,
		Visibility:  designProto.Visibility,
		VisibleTo:   designProto.VisibleTo,
		CreatedAt:   now,
		UpdatedAt:   now,
		Sections:    []SectionMeta{},
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

	createdDesignProto := DesignMetadataToProto(&metadata)

	resp = &protos.CreateDesignResponse{
		Design: createdDesignProto,
	}
	return resp, nil
}

// --- GetDesign (Unchanged) ---
func (s *DesignService) GetDesign(ctx context.Context, req *protos.GetDesignRequest) (resp *protos.GetDesignResponse, err error) {
	slog.Info("GetDesign Request", "req", req)
	designId := req.Id
	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty")
	}

	metadataPath := getDesignMetadataPath(designId)
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

	var metadata DesignMetadata
	err = json.Unmarshal(jsonData, &metadata)
	if err != nil {
		slog.Error("Failed to unmarshal metadata JSON", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	// TODO: Permission Checks

	designProto := DesignMetadataToProto(&metadata)
	resp = &protos.GetDesignResponse{
		Design: designProto,
	}
	slog.Info("Successfully retrieved design metadata", "designId", designId)
	return resp, nil
}

// --- ListDesigns (Unchanged) ---
func (s *DesignService) ListDesigns(ctx context.Context, req *protos.ListDesignsRequest) (resp *protos.ListDesignsResponse, err error) {
	slog.Info("ListDesigns Request", "req", req)

	entries, err := os.ReadDir(designsBasePath)
	if err != nil {
		slog.Error("Failed to read designs directory", "path", designsBasePath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to list designs")
	}

	var allMetadata []*DesignMetadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		designId := entry.Name()
		metadataPath := getDesignMetadataPath(designId)

		if _, statErr := os.Stat(metadataPath); statErr == nil {
			jsonData, readErr := os.ReadFile(metadataPath)
			if readErr != nil {
				slog.Warn("Failed to read metadata file during list", "designId", designId, "path", metadataPath, "error", readErr)
				continue
			}

			var metadata DesignMetadata
			unmarshalErr := json.Unmarshal(jsonData, &metadata)
			if unmarshalErr != nil {
				slog.Warn("Failed to unmarshal metadata JSON during list", "designId", designId, "path", metadataPath, "error", unmarshalErr)
				continue
			}
			if metadata.ID != designId {
				slog.Warn("Mismatch between directory name and metadata ID", "dirName", designId, "metadataId", metadata.ID)
				continue
			}
			allMetadata = append(allMetadata, &metadata)
		} else if !errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Error stating metadata file during list", "designId", designId, "path", metadataPath, "error", statErr)
		}
	}
	slog.Debug("Collected metadata for designs", "count", len(allMetadata))

	var filteredMetadata []*DesignMetadata
	loggedInUserId, _ := s.EnsureLoggedIn(ctx)
	for _, meta := range allMetadata {
		if req.OwnerId != "" && meta.OwnerID != req.OwnerId {
			continue
		}
		isOwner := (!ENFORCE_LOGIN && meta.OwnerID == "dev_user") || (ENFORCE_LOGIN && loggedInUserId == meta.OwnerID)
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
			filteredMetadata = []*DesignMetadata{}
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
		finalProtos[i] = DesignMetadataToProto(meta)
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

// --- NEW IMPLEMENTATIONS ---

// UpdateDesign modifies design metadata based on the update mask.
func (s *DesignService) UpdateDesign(ctx context.Context, req *protos.UpdateDesignRequest) (resp *protos.UpdateDesignResponse, err error) {
	slog.Info("UpdateDesign Request", "req", req)
	designId := req.Design.GetId() // Get ID from nested design proto
	updateMask := req.UpdateMask

	// 1. Validate Input
	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty in update request")
	}
	if req.Design == nil {
		return nil, status.Error(codes.InvalidArgument, "Design payload cannot be nil in update request")
	}
	if updateMask == nil || len(updateMask.Paths) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Update mask must be provided with at least one field path")
	}

	// Ensure all paths in the mask are valid for this stage (metadata only)
	// We only support top-level fields of DesignMetadata here.
	validPaths := map[string]bool{
		"name":        true,
		"description": true,
		"visibility":  true,
		"visible_to":  true,
		// Add other metadata fields here if they become updatable
	}
	// FieldMask uses snake_case, Design proto uses snake_case, JSON uses camelCase. Map between them.
	protoToJsonPathMap := map[string]string{
		"name":        "Name",
		"description": "Description",
		"visibility":  "Visibility",
		"visible_to":  "VisibleTo",
	}

	for _, path := range updateMask.Paths {
		// Allow "design.field_name" as well for compatibility with some clients
		effectivePath := strings.TrimPrefix(path, "design.")
		if !validPaths[effectivePath] {
			slog.Warn("Invalid path in update mask for metadata update", "path", path)
			// In this phase, section updates aren't handled here.
			// If the path is for sections, we might ignore it or error later.
			// For now, error on any unexpected path.
			return nil, status.Errorf(codes.InvalidArgument, "Update mask contains invalid or unsupported path for metadata update: %s", path)
		}
	}

	// 2. Acquire Mutex
	mutex := getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for update", "designId", designId)

	// 3. Read existing metadata
	metadataPath := getDesignMetadataPath(designId)
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design not found for update", "designId", designId, "path", metadataPath)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", designId))
		}
		slog.Error("Failed to read metadata file for update", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata for update")
	}

	var metadata DesignMetadata
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal metadata JSON for update", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	// 4. Check Permissions
	loggedInUserId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN {
		if err != nil {
			return nil, err // Unauthenticated
		}
		if loggedInUserId != metadata.OwnerID {
			slog.Warn("Permission denied for update", "designId", designId, "loggedInUser", loggedInUserId, "owner", metadata.OwnerID)
			return nil, status.Error(codes.PermissionDenied, "User does not have permission to update this design")
		}
	}

	// 5. Apply Updates based on mask
	updated := false
	for _, path := range updateMask.Paths {
		effectivePath := strings.TrimPrefix(path, "design.")
		jsonField := protoToJsonPathMap[effectivePath] // Use the map

		switch jsonField { // Switch on the mapped JSON field name
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
			// Add validation for allowed visibility values if needed
			if metadata.Visibility != newValue {
				metadata.Visibility = newValue
				updated = true
			}
		case "VisibleTo":
			// Replace the whole slice
			newValue := req.Design.VisibleTo
			// Basic check if slices differ (more robust check might be needed)
			if len(metadata.VisibleTo) != len(newValue) { // || !slices.Equal(metadata.VisibleTo, newValue) { // Go 1.21+
				metadata.VisibleTo = newValue
				updated = true
			} else { // Check element by element if lengths are same (Go < 1.21)
				diff := false
				sort.Strings(metadata.VisibleTo) // Sort for consistent comparison
				sort.Strings(newValue)
				for i := range metadata.VisibleTo {
					if metadata.VisibleTo[i] != newValue[i] {
						diff = true
						break
					}
				}
				if diff {
					metadata.VisibleTo = newValue
					updated = true
				}
			}
		default:
			// This shouldn't be reached due to earlier validation, but safeguard anyway
			slog.Warn("Encountered unhandled valid path during update apply", "path", path)
		}
	}

	// 6. Save if updated
	if updated {
		metadata.UpdatedAt = time.Now()
		slog.Info("Metadata changed, attempting save", "designId", designId)

		updatedJsonData, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			slog.Error("Failed to marshal updated metadata", "designId", designId, "error", err)
			return nil, status.Error(codes.Internal, "Failed to serialize updated metadata")
		}

		// Write back to the file
		err = os.WriteFile(metadataPath, updatedJsonData, 0644)
		if err != nil {
			slog.Error("Failed to write updated metadata file", "designId", designId, "path", metadataPath, "error", err)
			return nil, status.Error(codes.Internal, "Failed to save updated design metadata")
		}
		slog.Info("Successfully updated design metadata file", "designId", designId, "path", metadataPath)
	} else {
		slog.Info("No changes detected based on update mask", "designId", designId)
	}

	// 7. Return response
	updatedProto := DesignMetadataToProto(&metadata) // Return the potentially updated state
	resp = &protos.UpdateDesignResponse{
		Design: updatedProto,
	}
	return resp, nil
}

// Deletes a design and all its contents from the filesystem.
func (s *DesignService) DeleteDesign(ctx context.Context, req *protos.DeleteDesignRequest) (resp *protos.DeleteDesignResponse, err error) {
	slog.Info("DeleteDesign Request", "req", req)
	designId := req.Id

	// 1. Validate Input
	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID cannot be empty for delete request")
	}

	// 2. Acquire Mutex (prevents delete during create/update)
	mutex := getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for delete", "designId", designId)

	// 3. Check Existence and Read Metadata for Permissions
	metadataPath := getDesignMetadataPath(designId)
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design not found for delete", "designId", designId, "path", metadataPath)
			// Deleting something that doesn't exist is often considered idempotent success in REST/gRPC
			return &protos.DeleteDesignResponse{}, nil // Return success
		}
		slog.Error("Failed to read metadata file for delete permission check", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata for delete")
	}

	var metadata DesignMetadata
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal metadata JSON for delete permission check", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata (corrupted data?)")
	}

	// 4. Check Permissions
	loggedInUserId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN {
		if err != nil {
			return nil, err // Unauthenticated
		}
		if loggedInUserId != metadata.OwnerID {
			slog.Warn("Permission denied for delete", "designId", designId, "loggedInUser", loggedInUserId, "owner", metadata.OwnerID)
			return nil, status.Error(codes.PermissionDenied, "User does not have permission to delete this design")
		}
	}

	// 5. Delete the entire design directory
	designPath := getDesignPath(designId)
	slog.Info("Attempting to delete design directory", "designId", designId, "path", designPath)
	err = os.RemoveAll(designPath)
	if err != nil {
		// Check if it was already gone (e.g., race condition, though mutex should prevent)
		if _, statErr := os.Stat(designPath); errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Design directory already gone during delete", "designId", designId, "path", designPath)
			return &protos.DeleteDesignResponse{}, nil // Idempotent success
		}
		// Actual error during deletion
		slog.Error("Failed to delete design directory", "designId", designId, "path", designPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to delete design")
	}

	slog.Info("Successfully deleted design", "designId", designId, "path", designPath)

	// 6. Return success
	resp = &protos.DeleteDesignResponse{}
	return resp, nil
}

// GetDesigns is generally inefficient for a file-based backend.
func (s *DesignService) GetDesigns(ctx context.Context, req *protos.GetDesignsRequest) (resp *protos.GetDesignsResponse, err error) {
	// Reading multiple specific files based on IDs is inefficient.
	// Clients should typically call GetDesign multiple times if needed,
	// possibly in parallel.
	slog.Warn("GetDesigns called, which is inefficient for filesystem backend")
	return nil, status.Error(codes.Unimplemented, "GetDesigns is not efficiently implemented for the filesystem backend. Use GetDesign for individual IDs.")
}
