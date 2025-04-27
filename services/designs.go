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
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// Add if using SectionMetadataToProto helper:
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// --- DesignService struct holds configuration and state ---
type DesignService struct {
	protos.UnimplementedDesignServiceServer
	BaseService
	clients *ClientMgr // Keep for potential auth needs

	idgen IDGen

	// Configuration / State moved from globals
	basePath string
	store    *DesignStore // Add the store instance
	mutexMap sync.Map     // Mutex map keyed by design ID
}

// --- NewDesignService Constructor ---
func NewDesignService(clients *ClientMgr, basePath string) *DesignService {
	if basePath == "" {
		basePath = defaultDesignsBasePath
	}
	store, err := NewDesignStore(basePath)
	if err != nil {
		// Fatal error if store cannot be initialized
		log.Fatalf("Could not initialize DesignStore at '%s': %v", basePath, err)
	}

	out := &DesignService{
		clients:  clients,
		basePath: basePath,
		store:    store, // Assign store
	}
	out.idgen = IDGen{
		NextIDFunc: (&SimpleIDGen{}).NextID,
		GetID:      func(kind, id string) (*GenID, error) { return out.getDesignId(id) },
	}
	if err := ensureDir(out.basePath); err != nil {
		log.Fatalf("Could not create base designs directory '%s': %v", out.basePath, err)
	}
	return out
}

// --- Filesystem Path Helpers ---
func (s *DesignService) getDesignId(id string) (g *GenID, err error) {
	designPath := s.getDesignPath(id)
	_, err = os.Stat(designPath)
	if err == nil {
		g = &GenID{Id: id}
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	slog.Error("Error checking design path existence", "id", id, "path", designPath, "error", err)
	return nil, err
}

// Helper to get or create a mutex for a specific design ID
func (s *DesignService) getDesignMutex(designId string) *sync.Mutex {
	mutex, _ := s.mutexMap.LoadOrStore(designId, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

// Helper to get the path for a specific design
func (s *DesignService) getDesignPath(designId string) string {
	return filepath.Join(s.basePath, designId)
}

// Helper to get the path for the design metadata file
func (s *DesignService) getDesignMetadataPath(designId string) string {
	return filepath.Join(s.getDesignPath(designId), "design.json")
}

// Helper to get the base path for a design's sections
func (s *DesignService) getSectionsBasePath(designId string) string {
	return filepath.Join(s.getDesignPath(designId), "sections")
}

// Helper to get the path for a specific section's data file
func (s *DesignService) ensureSectionPath(designId, sectionId string) (string, error) {
	sectionPath := filepath.Join(s.getSectionsBasePath(designId), sectionId) // fmt.Sprintf("%s.json", sectionId))
	err := os.MkdirAll(sectionPath, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("Failed to create directory", "path", sectionPath, "error", err)
		return sectionPath, err
	}
	return filepath.Join(s.getSectionsBasePath(designId), sectionId, "main.json"), nil
}

func (s *DesignService) getSectionPath(designId, sectionId string) string {
	return filepath.Join(s.getSectionsBasePath(designId), sectionId, "main.json") // fmt.Sprintf("%s.json", sectionId))
}

func (s *DesignService) getSectionBasePath(designId, sectionId string) string {
	return filepath.Join(s.getSectionsBasePath(designId), sectionId)
}

// Helper to get the path for a specific prompt file within a section dir
func (s *DesignService) getSectionPromptPath(designId, sectionId, promptName string) (string, error) {
	// Basic sanitization - ensure promptName is simple (e.g., "get_answer.md", "verify.md")
	safePromptName := strings.ReplaceAll(promptName, "..", "") // Prevent directory traversal
	safePromptName = strings.ReplaceAll(safePromptName, "/", "")
	safePromptName = strings.ReplaceAll(safePromptName, "\\", "")
	if safePromptName == "" {
		return "", fmt.Errorf("invalid prompt name provided")
	}
	if !strings.HasSuffix(safePromptName, ".md") && !strings.HasSuffix(safePromptName, ".txt") {
		// Enforce markdown/text extension for safety/clarity, adjust if needed
		safePromptName += ".md"
	}

	promptsDir := filepath.Join(s.getSectionBasePath(designId, sectionId), "prompts")
	// Ensure the prompts directory exists for the section
	err := os.MkdirAll(promptsDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("Failed to create prompts directory", "path", promptsDir, "error", err)
		return "", fmt.Errorf("failed to ensure prompts directory: %w", err)
	}
	return filepath.Join(promptsDir, safePromptName), nil
}

func (s *DesignService) getContentPath(designId, sectionId, contentName string) string {
	safeName := strings.ReplaceAll(contentName, "..", "")
	safeName = strings.ReplaceAll(safeName, "/", "")
	safeName = strings.ReplaceAll(safeName, "\\", "")
	if safeName == "" {
		safeName = "default"
	}
	return filepath.Join(s.getSectionBasePath(designId, sectionId), fmt.Sprintf("content.%s", safeName))
}

// Helper to read prompt file content
func (s *DesignService) readPromptFile(designId, sectionId, promptName string) (string, error) {
	promptPath, err := s.getSectionPromptPath(designId, sectionId, promptName)
	if err != nil {
		return "", err // Error getting path (e.g., bad name, couldn't create dir)
	}
	contentBytes, err := os.ReadFile(promptPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoSuchEntity // Use specific error for not found
		}
		slog.Error("Failed to read prompt file", "path", promptPath, "error", err)
		return "", fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
	}
	return string(contentBytes), nil
}

// Helper to read and unmarshal section metadata (main.json)
func (s *DesignService) readSectionData(designId, sectionId string) (*Section, error) {
	sectionPath := s.getSectionPath(designId, sectionId)
	jsonData, err := os.ReadFile(sectionPath)
	if err != nil {
		// Distinguish between file not existing and other errors
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoSuchEntity // Use our custom error
		}
		return nil, fmt.Errorf("failed to read section file %s: %w", sectionPath, err)
	}
	var section Section
	if err := json.Unmarshal(jsonData, &section); err != nil {
		return nil, fmt.Errorf("failed to unmarshal section data from %s: %w", sectionPath, err)
	}
	return &section, nil
}

// --- Static/Utility Helpers (can remain outside the struct or moved) ---

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
		if genid == nil || err != nil {
			slog.Error("Failed to generate a unique design ID", "error", err)
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

	mutex := s.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex", "designId", designId)

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

	// Ensure base directory and sections subdir exist
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

	// --- Populate SectionsMetadata if requested ---
	if req.IncludeSectionMetadata {
		slog.Debug("IncludeSectionMetadata requested, fetching section metadata", "designId", designId, "sectionCount", len(metadata.SectionIds))
		sectionMetadataProtos := make([]*protos.Section, 0, len(metadata.SectionIds))

		for index, sectionId := range metadata.SectionIds {
			sectionData, readErr := s.readSectionData(designId, sectionId)
			if readErr != nil {
				// Log error but continue, maybe the file is missing/corrupt
				slog.Warn("Failed to read section data while fetching metadata",
					"designId", designId,
					"sectionId", sectionId,
					"error", readErr)
				// Optionally, add a placeholder section with an error indicator?
				// For now, we just skip it.
				continue
			}

			// --- Populate prompts by reading files ---
			if getPrompt, errRead := s.readPromptFile(designId, sectionId, "get_answer.md"); errRead == nil {
				sectionData.GetAnswerPrompt = getPrompt
			} // Ignore ErrNoSuchEntity
			if verifyPrompt, errRead := s.readPromptFile(designId, sectionId, "verify.md"); errRead == nil {
				sectionData.VerifyAnswerPrompt = verifyPrompt
			} // Ignore ErrNoSuchEntity
			// ---------------------------------------

			// Convert to proto
			sectionProto := SectionToProto(sectionData)
			// Set the order based on its index in the design's list
			sectionProto.Order = uint32(index)

			sectionMetadataProtos = append(sectionMetadataProtos, sectionProto)
		}
		resp.SectionsMetadata = sectionMetadataProtos
		slog.Debug("Finished fetching section metadata", "designId", designId, "metadataCount", len(resp.SectionsMetadata))
	}

	slog.Info("Successfully retrieved design metadata", "designId", designId)
	return resp, nil
}

// GetSection retrieves the content and metadata of a single section.
func (s *DesignService) GetSection(ctx context.Context, req *protos.GetSectionRequest) (resp *protos.Section, err error) {
	designId := req.DesignId
	sectionId := req.SectionId
	if designId == "" || sectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID and Section ID must be provided")
	}
	slog.Info("GetSection Request", "designId", designId, "sectionId", sectionId)

	// No need for mutex for reads usually, but consistency might be desired if reads/writes are frequent.
	// Let's skip the mutex for GetSection for now to optimize reads.

	sectionData, err := s.readSectionData(designId, sectionId)
	if err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Warn("Section not found", "designId", designId, "sectionId", sectionId)
			return nil, status.Errorf(codes.NotFound, "Section with id '%s' not found in design '%s'", sectionId, designId)
		}
		slog.Error("Failed to read section data", "designId", designId, "sectionId", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read section data")
	}

	// --- Populate prompts by reading files before converting ---
	if getPrompt, errRead := s.readPromptFile(designId, sectionId, "get_answer.md"); errRead == nil {
		sectionData.GetAnswerPrompt = getPrompt
	} // Ignore ErrNoSuchEntity
	if verifyPrompt, errRead := s.readPromptFile(designId, sectionId, "verify.md"); errRead == nil {
		sectionData.VerifyAnswerPrompt = verifyPrompt
	} // Ignore ErrNoSuchEntity
	// ----------------------------------------------------------

	// TODO: Add permission check - can the user view this design/section?

	// Note: Order is not stored in the section file, needs to be looked up if required.
	// The GetSection API doesn't necessarily need to return order, but the converter handles it.
	resp = SectionToProto(sectionData) // Convert to proto
	slog.Info("Successfully retrieved section", "designId", designId, "sectionId", sectionId)
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
		start = max(int(req.Pagination.PageOffset), 0)
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

// AddSection creates a new section, saves its data, and adds its ID to the design's list.
func (s *DesignService) AddSection(ctx context.Context, req *protos.AddSectionRequest) (resp *protos.Section, err error) {
	designId := req.Section.GetDesignId() // designId is now part of the section proto
	sectionProto := req.Section

	if designId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID must be provided within the section payload")
	}
	if sectionProto == nil {
		return nil, status.Error(codes.InvalidArgument, "Section payload cannot be nil")
	}
	if sectionProto.Type == protos.SectionType_SECTION_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "Section type must be specified")
	}
	slog.Info("AddSection Request", "designId", designId, "type", sectionProto.Type, "relativeTo", req.RelativeSectionId, "position", req.Position)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	mutex := s.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for AddSection", "designId", designId)

	// 1. Read and check design metadata + permissions
	metadataPath := s.getDesignMetadataPath(designId)
	metadataJson, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Design not found for AddSection", "designId", designId)
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		slog.Error("Failed to read design metadata for AddSection", "designId", designId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata")
	}
	var metadata Design
	if err = json.Unmarshal(metadataJson, &metadata); err != nil {
		slog.Error("Failed to unmarshal design metadata for AddSection", "designId", designId, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata")
	}

	if ENFORCE_LOGIN && ownerId != metadata.OwnerId {
		slog.Warn("Permission denied for AddSection", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot add sections to this design")
	}

	// 2. Generate new Section ID
	genid, err := s.idgen.NextID("section", ownerId, time.Unix(0, 0)) // Using ownerId context for section ID generation too
	if err != nil || genid == nil {
		slog.Error("Failed to generate unique section ID", "designId", designId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to generate section ID")
	}
	newSectionId := genid.Id
	slog.Info("Generated new section ID", "designId", designId, "newSectionId", newSectionId)

	// 3. Prepare new Section data struct
	now := time.Now()
	newSectionData := SectionFromProto(sectionProto) // Use converter
	newSectionData.Id = newSectionId
	newSectionData.DesignId = designId // Ensure design ID is set
	newSectionData.CreatedAt = now
	newSectionData.UpdatedAt = now
	if providedTitle := sectionProto.GetTitle(); providedTitle != "" {
		newSectionData.Title = providedTitle
	} else if newSectionData.Title == "" { // Fallback to random only if proto title AND struct title are empty
		newSectionData.Title = randomSectionTitle(newSectionData.Type) // Assuming a helper function like randomDesignName
	}

	// 4. Determine insertion index in metadata.SectionIds
	insertIndex := len(metadata.SectionIds) // Default to end
	if req.RelativeSectionId != "" {
		found := false
		for i, id := range metadata.SectionIds {
			if id == req.RelativeSectionId {
				if req.Position == protos.PositionType_POSITION_TYPE_BEFORE {
					insertIndex = i
				} else { // Default to AFTER
					insertIndex = i + 1
				}
				found = true
				break
			}
		}
		if !found {
			slog.Warn("RelativeSectionId not found during AddSection", "designId", designId, "relativeId", req.RelativeSectionId)
			// Optional: return error or just append to end? Let's append.
			insertIndex = len(metadata.SectionIds)
		}
	} // If RelativeSectionId is empty, insertIndex remains len(metadata.SectionIds) -> append

	// 5. Write new section data file *first*
	sectionPath, err := s.ensureSectionPath(designId, newSectionId)
	if err != nil {
		slog.Error("Failed to get section path", "designId", designId, "sectionId", newSectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to ensure section path")
	}
	sectionJson, err := json.MarshalIndent(newSectionData, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal new section data", "designId", designId, "sectionId", newSectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to serialize section data")
	}
	if err = os.WriteFile(sectionPath, sectionJson, 0644); err != nil {
		slog.Error("Failed to write new section file", "designId", designId, "sectionPath", sectionPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to save section data")
	}
	slog.Info("Successfully wrote new section file", "path", sectionPath)

	// --- Write Initial Prompts (if provided) ---
	if req.InitialGetAnswerPrompt != "" {
		err = s.writePromptFile(designId, newSectionId, "get_answer.md", req.InitialGetAnswerPrompt)
		if err != nil {
			slog.Error("Failed to write initial get_answer prompt", "designId", designId, "sectionId", newSectionId, "error", err)
			// Decide: Fail the whole operation? Or just log warning? Let's log for now.
		}
	}
	if req.InitialVerifyPrompt != "" {
		err = s.writePromptFile(designId, newSectionId, "verify.md", req.InitialVerifyPrompt)
		if err != nil {
			slog.Error("Failed to write initial verify prompt", "designId", designId, "sectionId", newSectionId, "error", err)
		}
	}
	// ------------------------------------------

	// 6. Update and write design metadata
	metadata.SectionIds = append(metadata.SectionIds[:insertIndex], append([]string{newSectionId}, metadata.SectionIds[insertIndex:]...)...)
	metadata.UpdatedAt = now
	if err = s.writeDesignMetadata(designId, &metadata); err != nil {
		// Attempt to clean up orphaned section file, but prioritize reporting the metadata write error
		slog.Error("Failed to write updated design metadata after adding section", "designId", designId, "error", err)
		_ = os.Remove(sectionPath) // Best effort cleanup
		return nil, status.Error(codes.Internal, "Failed to update design metadata after adding section")
	}
	slog.Info("Successfully updated design metadata with new section ID", "designId", designId)

	// 7. Prepare and return response proto
	resp = SectionToProto(newSectionData)
	resp.Order = uint32(insertIndex) // Add the calculated order
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

// UpdateSection updates specific fields of an existing section.
func (s *DesignService) UpdateSection(ctx context.Context, req *protos.UpdateSectionRequest) (resp *protos.Section, err error) {
	designId := req.Section.DesignId
	sectionId := req.Section.Id
	updatesProto := req.Section // This contains the partial updates
	updateMask := req.UpdateMask

	if designId == "" || sectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID and Section ID must be provided")
	}
	if updatesProto == nil {
		return nil, status.Error(codes.InvalidArgument, "Section update payload cannot be nil")
	}
	if updateMask == nil || len(updateMask.Paths) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Update mask must be provided with at least one field path for section update")
	}
	slog.Info("UpdateSection Request", "designId", designId, "sectionId", sectionId, "mask", updateMask.Paths)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	mutex := s.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for UpdateSection", "designId", designId, "sectionId", sectionId)

	// 1. Check design exists and user has permission
	metadata, err := s.readDesignMetadata(designId)
	if err != nil {
		// Handles NotFound and other read errors
		return nil, err
	}
	if ENFORCE_LOGIN && ownerId != metadata.OwnerId {
		slog.Warn("Permission denied for UpdateSection", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot update sections in this design")
	}

	// 2. Check section exists within the design's list
	sectionIndex := -1
	for i, id := range metadata.SectionIds {
		if id == sectionId {
			sectionIndex = i
			break
		}
	}
	if sectionIndex == -1 {
		slog.Warn("Section ID not found in design metadata during update", "designId", designId, "sectionId", sectionId)
		return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
	}

	// 3. Read existing section data
	sectionData, err := s.readSectionData(designId, sectionId)
	if err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Error("Inconsistency: Section ID in metadata but file not found", "designId", designId, "sectionId", sectionId)
			return nil, status.Errorf(codes.Internal, "Section '%s' data file not found despite being listed in design '%s'", sectionId, designId)
		}
		slog.Error("Failed to read section data for update", "designId", designId, "sectionId", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read section data for update")
	}

	// 4. Apply updates based on mask
	updated := applySectionUpdates(sectionData, updatesProto, updateMask)

	// 5. Write updated section data if changed
	if updated {
		slog.Info("Section data changed, saving...", "designId", designId, "sectionId", sectionId)
		sectionData.UpdatedAt = time.Now()
		if err = s.writeSectionData(designId, sectionId, sectionData); err != nil {
			return nil, status.Error(codes.Internal, "Failed to write updated section data")
		}

		// 6. Update design metadata timestamp (only if section was updated)
		metadata.UpdatedAt = sectionData.UpdatedAt
		if err = s.writeDesignMetadata(designId, metadata); err != nil {
			// Log error, but proceed as section data *was* saved. Maybe return warning?
			slog.Error("Failed to update design metadata timestamp after section update", "designId", designId, "error", err)
		}
	} else {
		slog.Info("No changes detected for section based on update mask", "designId", designId, "sectionId", sectionId)
	}

	// 7. Return the potentially updated section
	resp = SectionToProto(sectionData)
	resp.Order = uint32(sectionIndex) // Add the order based on its position
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

// DeleteSection removes a section from the design's list and deletes its data file.
func (s *DesignService) DeleteSection(ctx context.Context, req *protos.DeleteSectionRequest) (resp *protos.DeleteSectionResponse, err error) {
	designId := req.DesignId
	sectionId := req.SectionId

	if designId == "" || sectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID and Section ID must be provided")
	}
	slog.Info("DeleteSection Request", "designId", designId, "sectionId", sectionId)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	mutex := s.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for DeleteSection", "designId", designId, "sectionId", sectionId)

	// 1. Read design metadata & check permissions
	metadata, err := s.readDesignMetadata(designId)
	if err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Warn("Design not found during DeleteSection", "designId", designId)
			// Design doesn't exist, so section implicitly doesn't. Idempotency.
			return &protos.DeleteSectionResponse{}, nil
		}
		slog.Error("Failed to read design metadata for DeleteSection", "designId", designId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata")
	}

	if ENFORCE_LOGIN && ownerId != metadata.OwnerId {
		slog.Warn("Permission denied for DeleteSection", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot delete sections from this design")
	}

	// 2. Find and remove section ID from metadata list
	foundIndex := -1
	originalLen := len(metadata.SectionIds)
	newSectionIds := make([]string, 0, originalLen)
	for i, id := range metadata.SectionIds {
		if id == sectionId {
			foundIndex = i
		} else {
			newSectionIds = append(newSectionIds, id)
		}
	}

	if foundIndex == -1 {
		slog.Warn("Section ID not found in metadata during DeleteSection (idempotent)", "designId", designId, "sectionId", sectionId)
		return &protos.DeleteSectionResponse{}, nil // Section already gone or never existed
	}

	// 3. Update and write metadata *first*
	metadata.SectionIds = newSectionIds
	metadata.UpdatedAt = time.Now()
	if err = s.writeDesignMetadata(designId, metadata); err != nil {
		slog.Error("Failed to write updated design metadata after removing section ID", "designId", designId, "error", err)
		// Don't delete the section file if metadata update failed
		return nil, status.Error(codes.Internal, "Failed to update design metadata before deleting section")
	}
	slog.Info("Successfully updated design metadata removing section ID", "designId", designId, "sectionId", sectionId)

	// 4. Delete section data file *second*
	sectionPath := s.getSectionPath(designId, sectionId)
	err = os.Remove(sectionPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// Log the error but consider the operation successful as metadata was updated
		slog.Error("Failed to delete section data file, but metadata was updated", "designId", designId, "sectionPath", sectionPath, "error", err)
	} else {
		slog.Info("Successfully deleted section data file (or it was already gone)", "path", sectionPath)
	}

	// Remove the folder too
	sectionDir := filepath.Join(s.getSectionsBasePath(designId), sectionId)
	err = os.RemoveAll(sectionDir)
	if err != nil {
		if _, statErr := os.Stat(sectionDir); errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Section directory already gone during delete", "designId", designId, "sectionId", sectionId, "path", sectionDir)
		} else {
			slog.Error("Failed to delete section directory", "designId", designId, "sectionId", sectionId, "path", sectionDir, "error", err)
			return nil, status.Error(codes.Internal, "Failed to delete section")
		}
	}

	return &protos.DeleteSectionResponse{}, nil
}

// MoveSection reorders a section within the design's section ID list.
func (s *DesignService) MoveSection(ctx context.Context, req *protos.MoveSectionRequest) (resp *protos.MoveSectionResponse, err error) {
	designId := req.DesignId
	sectionIdToMove := req.SectionId
	relativeSectionId := req.RelativeSectionId
	position := req.Position

	if designId == "" || sectionIdToMove == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID and Section ID to move must be provided")
	}
	// POSITION_TYPE_END doesn't make sense for move, only add.
	// relativeSectionId is required if position is BEFORE or AFTER.
	if position == protos.PositionType_POSITION_TYPE_UNSPECIFIED || position == protos.PositionType_POSITION_TYPE_END {
		return nil, status.Error(codes.InvalidArgument, "Position must be BEFORE or AFTER for move operation")
	}
	if relativeSectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "Relative Section ID must be provided for BEFORE/AFTER position")
	}
	slog.Info("MoveSection Request", "designId", designId, "sectionId", sectionIdToMove, "relativeTo", relativeSectionId, "position", position)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	mutex := s.getDesignMutex(designId)
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for MoveSection", "designId", designId, "sectionId", sectionIdToMove)

	// 1. Read metadata & check permissions
	metadata, err := s.readDesignMetadata(designId)
	if err != nil {
		return nil, err // Handles NotFound etc.
	}
	if ENFORCE_LOGIN && ownerId != metadata.OwnerId {
		slog.Warn("Permission denied for MoveSection", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot move sections in this design")
	}

	// 2. Find indices
	currentIndex := -1
	relativeIndex := -1
	for i, id := range metadata.SectionIds {
		if id == sectionIdToMove {
			currentIndex = i
		}
		if id == relativeSectionId {
			relativeIndex = i
		}
	}

	if currentIndex == -1 {
		return nil, status.Errorf(codes.NotFound, "Section to move ('%s') not found in design '%s'", sectionIdToMove, designId)
	}
	if relativeIndex == -1 {
		return nil, status.Errorf(codes.NotFound, "Relative section ('%s') not found in design '%s'", relativeSectionId, designId)
	}

	// 3. Calculate new position and perform move
	newSectionIds := make([]string, 0, len(metadata.SectionIds))
	// Copy elements before removal
	for i, id := range metadata.SectionIds {
		if i != currentIndex {
			newSectionIds = append(newSectionIds, id)
		}
	}

	// Adjust relativeIndex if removal affected it
	if currentIndex < relativeIndex {
		relativeIndex--
	}

	// Calculate insertion point
	insertIndex := 0
	if position == protos.PositionType_POSITION_TYPE_BEFORE {
		insertIndex = relativeIndex
	} else { // AFTER
		insertIndex = relativeIndex + 1
	}

	// Perform insertion
	finalSectionIds := append(newSectionIds[:insertIndex], append([]string{sectionIdToMove}, newSectionIds[insertIndex:]...)...)

	// 4. Update and write metadata
	metadata.SectionIds = finalSectionIds
	metadata.UpdatedAt = time.Now()
	if err = s.writeDesignMetadata(designId, metadata); err != nil {
		slog.Error("Failed to write updated design metadata after moving section", "designId", designId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to update design metadata after move")
	}
	slog.Info("Successfully updated design metadata after moving section", "designId", designId, "sectionId", sectionIdToMove)

	return &protos.MoveSectionResponse{}, nil
}

func (s *DesignService) GetDesigns(ctx context.Context, req *protos.GetDesignsRequest) (resp *protos.GetDesignsResponse, err error) {
	slog.Warn("GetDesigns called, which is inefficient for filesystem backend")
	return nil, status.Error(codes.Unimplemented, "GetDesigns is not efficiently implemented for the filesystem backend. Use GetDesign for individual IDs.")
}

/*
func (s *DesignService) SetContent(ctx context.Context, req *protos.SetContentRequest) (*protos.SetContentResponse, error) {
	designId := req.DesignId
	sectionId := req.SectionId

	if designId == "" || sectionId == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Design ID, Section ID, and Content with Name must be provided")
	}
	contentName := req.Name

	slog.Info("SetContent Request", "designId", designId, "sectionId", sectionId, "name", contentName)

	ownerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return nil, err
	}

	// --- Acquire Lock ---
	mutex := s.getDesignMutex(designId) // Use design-level lock
	mutex.Lock()
	defer mutex.Unlock()
	slog.Debug("Acquired mutex for SetContent", "designId", designId, "sectionId", sectionId, "name", contentName)

	// --- Check Permissions & Existence ---
	designMeta, err := s.readDesignMetadata(designId)
	if err != nil {
		return nil, err // Handles NotFound
	}
	if ENFORCE_LOGIN && ownerId != designMeta.OwnerId {
		slog.Warn("Permission denied for SetContent", "designId", designId, "user", ownerId)
		return nil, status.Error(codes.PermissionDenied, "User cannot set content in this design")
	}
	// Ensure section directory exists (section metadata doesn't need to be read here unless updating it)
	sectionPath := s.getSectionPath(designId, sectionId)
	if _, err := os.Stat(sectionPath); errors.Is(err, os.ErrNotExist) {
		// If AddSection didn't run or failed, the dir might not exist
		slog.Warn("Section directory not found during SetContent", "path", sectionPath)
		return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
	} else if err != nil {
		slog.Error("Error checking section directory existence", "path", sectionPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to access section directory")
	}

	// --- Handle Content Byte Update ---
	contentPath := s.getContentPath(designId, sectionId, contentName)
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

	metadataUpdated := false // Placeholder

	// --- Update Timestamps ---
	now := time.Now()
	if bytesUpdated || metadataUpdated {
		// Update section's main.json timestamp
		sectionMeta, err := s.readSectionData(designId, sectionId)
		if err == nil {
			sectionMeta.UpdatedAt = now
			if err_write := s.writeSectionData(designId, sectionId, sectionMeta); err_write != nil {
				slog.Error("Failed to update section metadata timestamp after SetContent", "path", s.getSectionPath(designId, sectionId), "error", err_write)
				// Continue, as content was potentially saved
			}
		} else {
			slog.Warn("Could not read section metadata to update timestamp", "designId", designId, "sectionId", sectionId, "error", err)
		}

		// Update design's design.json timestamp
		designMeta.UpdatedAt = now
		if err_write := s.writeDesignMetadata(designId, designMeta); err_write != nil {
			slog.Error("Failed to update design metadata timestamp after SetContent", "path", s.getDesignMetadataPath(designId), "error", err_write)
			// Continue
		}
	}

	// Return the metadata provided in the request (or read back if implemented)
	finalContentProto := &protos.Content{
		Name:      contentName,
		UpdatedAt: tspb.New(now),
		// CreatedAt needs proper tracking if important
	}

	resp := &protos.SetContentResponse{
		Content: finalContentProto,
	}
	return resp, nil
}
*/

// --- Internal Helpers ---

// Helper to read and unmarshal design metadata, handling errors.
func (s *DesignService) readDesignMetadata(designId string) (*Design, error) {
	metadataPath := s.getDesignMetadataPath(designId)
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		slog.Error("Failed to read design metadata", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read design metadata")
	}
	var metadata Design
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal design metadata", "designId", designId, "path", metadataPath, "error", err)
		return nil, status.Error(codes.DataLoss, "Failed to parse design metadata")
	}
	return &metadata, nil
}

// Helper to marshal and write design metadata.
func (s *DesignService) writeDesignMetadata(designId string, metadata *Design) error {
	metadataPath := s.getDesignMetadataPath(designId)
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal design metadata for writing", "designId", designId, "error", err)
		return status.Error(codes.Internal, "Failed to serialize design metadata")
	}
	if err = os.WriteFile(metadataPath, jsonData, 0644); err != nil {
		slog.Error("Failed to write design metadata file", "designId", designId, "path", metadataPath, "error", err)
		return status.Error(codes.Internal, "Failed to save design metadata")
	}
	return nil
}

// Helper to marshal and write section data.
func (s *DesignService) writeSectionData(designId, sectionId string, sectionData *Section) error {
	sectionPath := s.getSectionPath(designId, sectionId)
	jsonData, err := json.MarshalIndent(sectionData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal section data %s/%s: %w", designId, sectionId, err)
	}
	if err = os.WriteFile(sectionPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write section file %s: %w", sectionPath, err)
	}
	slog.Info("Successfully wrote section file", "path", sectionPath)
	return nil
}

// Helper to write prompt file content
func (s *DesignService) writePromptFile(designId, sectionId, promptName, content string) error {
	promptPath, err := s.getSectionPromptPath(designId, sectionId, promptName)
	if err != nil {
		return err // Error getting path
	}
	err = os.WriteFile(promptPath, []byte(content), 0644)
	if err != nil {
		slog.Error("Failed to write prompt file", "path", promptPath, "error", err)
		return fmt.Errorf("failed to write prompt file %s: %w", promptPath, err)
	}
	slog.Info("Successfully wrote prompt file", "path", promptPath)
	return nil
}

// Helper function to apply updates to section data based on field mask
func applySectionUpdates(section *Section, updates *protos.Section, mask *fieldmaskpb.FieldMask) (updated bool) {
	for _, path := range mask.Paths {
		switch path {
		case "title":
			newVal := updates.GetTitle()
			if section.Title != newVal {
				// Add validation if needed (e.g., title cannot be empty)
				if strings.TrimSpace(newVal) == "" {
					slog.Warn("Attempted to update section title to empty, ignoring", "sectionId", section.Id)
					continue // Or return error
				}
				section.Title = newVal
				updated = true
			}
		}
	}
	return updated
}
