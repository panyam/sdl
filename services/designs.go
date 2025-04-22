package services

import (
	"context"
	"encoding/json" // Added
	"fmt"
	"log"
	"log/slog"
	"os"            // Added
	"path/filepath" // Added
	"strings"
	"sync" // Added
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// "strings"

	// fn "github.com/panyam/goutils/fn"
	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// Just for test
const ENFORCE_LOGIN = false
const FAKE_USER_ID = ""

// Filesystem storage constants
const designsBasePath = "./data/designs" // Base directory for all designs

// In-memory mutex map for design IDs
var designMutexes sync.Map // maps[string]*sync.Mutex

// Struct representing the metadata stored in design.json
type DesignMetadata struct {
	ID          string        `json:"id"`
	OwnerID     string        `json:"ownerId"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Visibility  string        `json:"visibility"`
	VisibleTo   []string      `json:"visibleTo"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Sections    []SectionMeta `json:"sections"` // List of section metadata
}

// Struct representing metadata for a single section within design.json
type SectionMeta struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"` // "text", "drawing", "plot"
	Order int    `json:"order"`
}

// Helper to get or create a mutex for a specific design ID
func getDesignMutex(designId string) *sync.Mutex {
	mutex, _ := designMutexes.LoadOrStore(designId, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

// Helper to get the path for a specific design
func getDesignPath(designId string) string {
	return filepath.Join(designsBasePath, designId)
}

// Helper to get the path for the design metadata file
func getDesignMetadataPath(designId string) string {
	return filepath.Join(getDesignPath(designId), "design.json")
}

// Helper to get the base path for a design's sections
func getSectionsBasePath(designId string) string {
	return filepath.Join(getDesignPath(designId), "sections")
}

// Helper to ensure a directory exists
func ensureDir(path string) error {
	err := os.MkdirAll(path, 0755) // Use 0755 for permissions
	if err != nil && !os.IsExist(err) {
		slog.Error("Failed to create directory", "path", path, "error", err)
		return err
	}
	return nil
}

// Helper function to convert DesignMetadata to proto (similar to DesignToProto)
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
		SectionIds:  make([]string, len(input.Sections)), // Populate Section IDs
	}
	for i, secMeta := range input.Sections {
		out.SectionIds[i] = secMeta.ID
	}
	// Note: The proto definition doesn't have explicit order or type for sections,
	// only the IDs. This is handled by the structure in design.json.
	// ContentMetadata from the original DB model is not in DesignMetadata.json yet.
	// We can add it later if needed.
	return out
}

// --- Rest of DesignService struct and NewDesignService function ---
type DesignService struct {
	protos.UnimplementedDesignServiceServer
	BaseService
	clients *ClientMgr // Keep clients for now, might need auth parts
}

func NewDesignService(clients *ClientMgr) *DesignService {
	out := &DesignService{
		clients: clients,
	}
	// Ensure base directory exists on startup
	if err := ensureDir(designsBasePath); err != nil {
		log.Fatalf("Could not create base designs directory '%s': %v", designsBasePath, err)
	}
	return out
}

// --- UpdateDesign - Keep placeholder for now ---
func (s *DesignService) UpdateDesign(ctx context.Context, req *protos.UpdateDesignRequest) (resp *protos.UpdateDesignResponse, err error) {
	log.Println("In design update: ", req)
	// ... (Implementation in next checkpoint) ...
	return nil, status.Errorf(codes.Unimplemented, "UpdateDesign not implemented yet")
}

// Create a new Design
func (s *DesignService) CreateDesign(ctx context.Context, req *protos.CreateDesignRequest) (resp *protos.CreateDesignResponse, err error) {
	slog.Info("CreateDesign Request", "req", req)
	designProto := req.Design
	if designProto == nil {
		return nil, status.Error(codes.InvalidArgument, "Design payload cannot be nil")
	}

	// 1. Get User ID
	// Using BaseService method, assuming metadata propagation works
	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		slog.Error("Login enforcement failed", "error", err)
		return nil, err // Return the unauthenticated error
	}
	// Allow override if not enforcing login (useful for initial testing)
	if !ENFORCE_LOGIN && ownerId == "" {
		ownerId = "dev_user" // Default owner for testing
	}

	// 2. Validate input
	if designProto.Id == "" {
		// In Phase 1, require ID from client. Later could generate UUIDs.
		return nil, status.Error(codes.InvalidArgument, "Design ID must be provided during creation in this phase")
	}
	designId := designProto.Id // Use the provided ID
	if strings.TrimSpace(designProto.Name) == "" {
		// Name is essential metadata
		return nil, status.Error(codes.InvalidArgument, "Design name cannot be empty")
	}

	// 3. Acquire mutex
	mutex := getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex", "designId", designId)

	// 4. Check if design already exists
	designPath := getDesignPath(designId)
	metadataPath := getDesignMetadataPath(designId)
	if _, err := os.Stat(designPath); err == nil {
		// Directory exists, check if design.json also exists to be sure
		if _, err := os.Stat(metadataPath); err == nil {
			slog.Warn("Design already exists", "designId", designId)
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Design with id '%s' already exists", designId))
		}
		// Directory exists but no metadata file? Could be corrupted state.
		// For now, let's treat it as existing to prevent overwriting potentially valid section data.
		slog.Warn("Design directory exists but metadata missing", "designId", designId, "path", metadataPath)
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Design with id '%s' potentially exists but metadata is missing", designId))

	} else if !os.IsNotExist(err) {
		// Some other error occurred checking the path
		slog.Error("Error checking design path", "designId", designId, "path", designPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to check design existence")
	}
	// Path does not exist, proceed with creation

	// 5. Create directories
	sectionsPath := getSectionsBasePath(designId)
	if err := ensureDir(sectionsPath); err != nil { // ensureDir creates parent dirs too
		return nil, status.Error(codes.Internal, "Failed to create design directories")
	}
	slog.Debug("Ensured directories exist", "designPath", designPath, "sectionsPath", sectionsPath)

	// 6. Create DesignMetadata struct
	now := time.Now()
	metadata := DesignMetadata{
		ID:          designId,
		OwnerID:     ownerId, // Use obtained ownerId
		Name:        designProto.Name,
		Description: designProto.Description,
		Visibility:  designProto.Visibility, // Ensure default is handled if empty
		VisibleTo:   designProto.VisibleTo,
		CreatedAt:   now,
		UpdatedAt:   now,
		Sections:    []SectionMeta{}, // Start with empty sections
	}
	// Set default visibility if not provided
	if metadata.Visibility == "" {
		metadata.Visibility = "private" // Default to private
	}

	// 7. Marshal DesignMetadata to JSON
	jsonData, err := json.MarshalIndent(metadata, "", "  ") // Use MarshalIndent for readability
	if err != nil {
		slog.Error("Failed to marshal design metadata", "designId", designId, "error", err)
		// Attempt cleanup? For now, just return error.
		// os.RemoveAll(designPath) // Optional: Attempt cleanup
		return nil, status.Error(codes.Internal, "Failed to serialize design metadata")
	}

	// 8. Write JSON to design.json
	err = os.WriteFile(metadataPath, jsonData, 0644) // Use 0644 permissions
	if err != nil {
		slog.Error("Failed to write design metadata file", "designId", designId, "path", metadataPath, "error", err)
		// Attempt cleanup?
		// os.RemoveAll(designPath) // Optional: Attempt cleanup
		return nil, status.Error(codes.Internal, "Failed to save design metadata")
	}
	slog.Info("Successfully created design metadata", "designId", designId, "path", metadataPath)

	// 9. Handle initial sections (Skip for this checkpoint)

	// 10. Convert metadata back to proto for response
	createdDesignProto := DesignMetadataToProto(&metadata)

	resp = &protos.CreateDesignResponse{
		Design: createdDesignProto,
	}
	return resp, nil
}

// --- ListDesigns - Keep placeholder for now ---
func (s *DesignService) ListDesigns(ctx context.Context, req *protos.ListDesignsRequest) (resp *protos.ListDesignsResponse, err error) {
	// ... (Implementation in next checkpoint) ...
	return nil, status.Errorf(codes.Unimplemented, "ListDesigns not implemented yet")
}

// --- GetDesign - Keep placeholder for now ---
func (s *DesignService) GetDesign(ctx context.Context, req *protos.GetDesignRequest) (resp *protos.GetDesignResponse, err error) {
	// ... (Implementation in next checkpoint) ...
	return nil, status.Errorf(codes.Unimplemented, "GetDesign not implemented yet")
}

// --- GetDesigns - Keep placeholder for now (Likely not needed if GetDesign handles single ID) ---
func (s *DesignService) GetDesigns(ctx context.Context, req *protos.GetDesignsRequest) (resp *protos.GetDesignsResponse, err error) {
	return nil, status.Errorf(codes.Unimplemented, "GetDesigns not implemented yet")
}

// --- DeleteDesign - Keep placeholder for now ---
func (s *DesignService) DeleteDesign(ctx context.Context, req *protos.DeleteDesignRequest) (resp *protos.DeleteDesignResponse, err error) {
	// ... (Implementation in next checkpoint) ...
	return nil, status.Errorf(codes.Unimplemented, "DeleteDesign not implemented yet")
}
