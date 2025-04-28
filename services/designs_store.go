// FILE: ./services/designs_store.go
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	// Assuming 'Design' and 'Section' structs are still defined in models.go
	// We might need to import other relevant types if used here
)

// DesignStore handles the filesystem persistence logic for designs and sections.
type DesignStore struct {
	basePath string
}

// NewDesignStore creates a new filesystem store for designs.
func NewDesignStore(basePath string) (*DesignStore, error) {
	resolvedPath := basePath
	if resolvedPath == "" {
		resolvedPath = defaultDesignsBasePath
	}
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		slog.Error("Failed to resolve absolute path", "path", resolvedPath, "error", err)
		return nil, fmt.Errorf("could not resolve base designs path '%s': %w", resolvedPath, err)
	}
	resolvedPath = absPath
	if err := ensureDir(resolvedPath); err != nil { // Use local ensureDir helper
		return nil, fmt.Errorf("could not create base designs directory '%s': %w", resolvedPath, err)
	}
	return &DesignStore{
		basePath: resolvedPath,
	}, nil
}

// --- Path Generation Methods ---

func (ds *DesignStore) getDesignPath(designId string) string {
	return filepath.Join(ds.basePath, designId)
}

func (ds *DesignStore) getDesignMetadataPath(designId string) string {
	return filepath.Join(ds.getDesignPath(designId), "design.json")
}

func (ds *DesignStore) getSectionsBasePath(designId string) string {
	return filepath.Join(ds.getDesignPath(designId), "sections")
}

// Gets the base directory path for a specific section.
func (ds *DesignStore) GetSectionBasePath(designId, sectionId string) string {
	// Exposed for ContentService via DesignService
	return filepath.Join(ds.getSectionsBasePath(designId), sectionId)
}

// Gets the path for a specific section's metadata file (main.json)
func (ds *DesignStore) getSectionDataPath(designId, sectionId string) string {
	return filepath.Join(ds.GetSectionBasePath(designId, sectionId), "main.json")
}

// Gets the path for a specific prompt file within a section directory.
func (ds *DesignStore) getSectionPromptPath(designId, sectionId, promptName string) (string, error) {
	if strings.TrimSpace(promptName) == "" {
		slog.Warn("Prompt name provided is empty or whitespace, returning error")
		return "", fmt.Errorf("invalid prompt name provided: cannot be empty")
	}
	safePromptName := sanitizeFilename(promptName)
	if safePromptName == "" {
		return "", fmt.Errorf("invalid prompt name provided")
	}
	if !strings.HasSuffix(safePromptName, ".md") && !strings.HasSuffix(safePromptName, ".txt") {
		safePromptName += ".md"
	}

	promptsDir := filepath.Join(ds.GetSectionBasePath(designId, sectionId), "prompts")
	if err := ensureDir(promptsDir); err != nil {
		return "", fmt.Errorf("failed to ensure prompts directory: %w", err)
	}
	return filepath.Join(promptsDir, safePromptName), nil
}

// Gets the path for a named content file within a section directory.
func (ds *DesignStore) GetContentPath(designId, sectionId, contentName string) string {
	// Exposed for ContentService via DesignService
	safeName := sanitizeFilename(contentName)
	if safeName == "" {
		safeName = "default" // Or handle error?
	}
	// Assuming content files are directly under the section base path for now
	return filepath.Join(ds.GetSectionBasePath(designId, sectionId), fmt.Sprintf("content.%s", safeName))
}

// --- Filesystem Interaction Methods ---

// Checks if a design directory exists.
func (ds *DesignStore) DesignExists(designId string) (bool, error) {
	designPath := ds.getDesignPath(designId)
	_, err := os.Stat(designPath)
	if err == nil {
		return true, nil // Exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil // Doesn't exist
	}
	// Other error (e.g., permission denied)
	slog.Error("Error checking design path existence", "id", designId, "path", designPath, "error", err)
	return false, err
}

// Reads and unmarshals the design metadata file (design.json).
func (ds *DesignStore) ReadDesignMetadata(designId string) (*Design, error) {
	metadataPath := ds.getDesignMetadataPath(designId)
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoSuchEntity // Use specific error
		}
		slog.Error("Failed to read design metadata", "designId", designId, "path", metadataPath, "error", err)
		return nil, fmt.Errorf("failed to read design metadata %s: %w", designId, err)
	}
	var metadata Design
	if err = json.Unmarshal(jsonData, &metadata); err != nil {
		slog.Error("Failed to unmarshal design metadata", "designId", designId, "path", metadataPath, "error", err)
		return nil, fmt.Errorf("failed to parse design metadata %s: %w", designId, err)
	}
	// Ensure ID from file matches expected ID
	if metadata.Id == "" {
		metadata.Id = designId // Populate if missing in file (older format?)
	} else if metadata.Id != designId {
		slog.Error("Design ID mismatch between directory and metadata file", "dirId", designId, "fileId", metadata.Id)
		return nil, fmt.Errorf("design ID mismatch for %s", designId)
	}
	return &metadata, nil
}

// Marshals and writes the design metadata file (design.json).
func (ds *DesignStore) WriteDesignMetadata(designId string, metadata *Design) error {
	// Ensure metadata ID matches the intended path
	if metadata.Id == "" {
		metadata.Id = designId
	} else if metadata.Id != designId {
		return fmt.Errorf("metadata ID '%s' does not match design ID '%s' for writing", metadata.Id, designId)
	}

	metadataPath := ds.getDesignMetadataPath(designId)
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal design metadata for writing", "designId", designId, "error", err)
		return fmt.Errorf("failed to serialize design metadata %s: %w", designId, err)
	}
	// Ensure parent directory exists before writing
	if err = ensureDir(filepath.Dir(metadataPath)); err != nil {
		return fmt.Errorf("failed to ensure directory for design metadata %s: %w", designId, err)
	}
	if err = os.WriteFile(metadataPath, jsonData, 0644); err != nil {
		slog.Error("Failed to write design metadata file", "designId", designId, "path", metadataPath, "error", err)
		return fmt.Errorf("failed to save design metadata %s: %w", designId, err)
	}
	return nil
}

// Reads and unmarshals section metadata (main.json).
func (ds *DesignStore) ReadSectionData(designId, sectionId string) (*Section, error) {
	sectionPath := ds.getSectionDataPath(designId, sectionId)
	jsonData, err := os.ReadFile(sectionPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoSuchEntity
		}
		slog.Error("Failed to read section metadata file", "path", sectionPath, "error", err)
		return nil, fmt.Errorf("failed to read section metadata %s/%s: %w", designId, sectionId, err)
	}
	var section Section
	if err := json.Unmarshal(jsonData, &section); err != nil {
		slog.Error("Failed to unmarshal section metadata", "path", sectionPath, "error", err)
		return nil, fmt.Errorf("failed to parse section metadata %s/%s: %w", designId, sectionId, err)
	}
	// Ensure IDs match
	if section.Id == "" {
		section.Id = sectionId
	} else if section.Id != sectionId {
		return nil, fmt.Errorf("section ID mismatch for %s/%s", designId, sectionId)
	}
	if section.DesignId == "" {
		section.DesignId = designId
	} else if section.DesignId != designId {
		return nil, fmt.Errorf("section design ID mismatch for %s/%s", designId, sectionId)
	}
	return &section, nil
}

// Marshals and writes section metadata (main.json).
func (ds *DesignStore) WriteSectionData(designId, sectionId string, sectionData *Section) error {
	// Ensure IDs match
	if sectionData.Id == "" {
		sectionData.Id = sectionId
	} else if sectionData.Id != sectionId {
		return fmt.Errorf("sectionData ID '%s' does not match section ID '%s'", sectionData.Id, sectionId)
	}
	if sectionData.DesignId == "" {
		sectionData.DesignId = designId
	} else if sectionData.DesignId != designId {
		return fmt.Errorf("sectionData design ID '%s' does not match design ID '%s'", sectionData.DesignId, designId)
	}

	sectionPath := ds.getSectionDataPath(designId, sectionId)
	jsonData, err := json.MarshalIndent(sectionData, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal section metadata", "path", sectionPath, "error", err)
		return fmt.Errorf("failed to marshal section data %s/%s: %w", designId, sectionId, err)
	}
	// Ensure parent directory exists
	if err = ensureDir(filepath.Dir(sectionPath)); err != nil {
		return fmt.Errorf("failed to ensure directory for section metadata %s/%s: %w", designId, sectionId, err)
	}
	if err = os.WriteFile(sectionPath, jsonData, 0644); err != nil {
		slog.Error("Failed to write section metadata file", "path", sectionPath, "error", err)
		return fmt.Errorf("failed to write section file %s: %w", sectionPath, err)
	}
	slog.Info("Successfully wrote section metadata file", "path", sectionPath)
	return nil
}

// Reads prompt file content.
func (ds *DesignStore) ReadPromptFile(designId, sectionId, promptName string) (string, error) {
	promptPath, err := ds.getSectionPromptPath(designId, sectionId, promptName)
	if err != nil {
		return "", err // Error getting path
	}
	contentBytes, err := os.ReadFile(promptPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrNoSuchEntity
		}
		slog.Error("Failed to read prompt file", "path", promptPath, "error", err)
		return "", fmt.Errorf("failed to read prompt file %s: %w", promptPath, err)
	}
	return string(contentBytes), nil
}

// Writes prompt file content.
func (ds *DesignStore) WritePromptFile(designId, sectionId, promptName, content string) error {
	promptPath, err := ds.getSectionPromptPath(designId, sectionId, promptName)
	if err != nil {
		return err // Error getting path
	}
	// Ensure directory exists (getSectionPromptPath does this)
	err = os.WriteFile(promptPath, []byte(content), 0644)
	if err != nil {
		slog.Error("Failed to write prompt file", "path", promptPath, "error", err)
		return fmt.Errorf("failed to write prompt file %s: %w", promptPath, err)
	}
	slog.Info("Successfully wrote prompt file", "path", promptPath)
	return nil
}

// Deletes the entire directory for a design.
func (ds *DesignStore) DeleteDesign(designId string) error {
	designPath := ds.getDesignPath(designId)
	slog.Info("Attempting to delete design directory", "designId", designId, "path", designPath)
	err := os.RemoveAll(designPath)
	if err != nil {
		// Check if it already doesn't exist (idempotency)
		if _, statErr := os.Stat(designPath); errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Design directory already gone during delete", "designId", designId, "path", designPath)
			return nil // Not an error in this case
		}
		// Other error during deletion
		slog.Error("Failed to delete design directory", "designId", designId, "path", designPath, "error", err)
		return fmt.Errorf("failed to delete design %s: %w", designId, err)
	}
	slog.Info("Successfully deleted design", "designId", designId, "path", designPath)
	return nil
}

// Deletes the directory for a specific section.
func (ds *DesignStore) DeleteSection(designId, sectionId string) error {
	sectionDir := ds.GetSectionBasePath(designId, sectionId)
	slog.Info("Attempting to delete section directory", "designId", designId, "sectionId", sectionId, "path", sectionDir)
	err := os.RemoveAll(sectionDir)
	if err != nil {
		if _, statErr := os.Stat(sectionDir); errors.Is(statErr, os.ErrNotExist) {
			slog.Warn("Section directory already gone during delete", "designId", designId, "sectionId", sectionId, "path", sectionDir)
			return nil // Idempotent
		}
		slog.Error("Failed to delete section directory", "designId", designId, "sectionId", sectionId, "path", sectionDir, "error", err)
		return fmt.Errorf("failed to delete section %s/%s: %w", designId, sectionId, err)
	}
	slog.Info("Successfully deleted section", "designId", designId, "sectionId", sectionId)
	return nil
}

// Creates a directory if it doesn't exist.
func ensureDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("Failed to create directory", "path", path, "error", err)
		return err
	}
	return nil
}

// Basic filename sanitizer (improve as needed)
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	// Add more replacements if needed (e.g., for spaces, special chars)
	return name
}
