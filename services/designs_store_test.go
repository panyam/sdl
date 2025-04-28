// FILE: ./services/designs_store_test.go
package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a DesignStore instance pointing to a temp directory.
func newTestDesignStore(t *testing.T) *DesignStore {
	t.Helper()
	tempDir := t.TempDir()
	store, err := NewDesignStore(tempDir)
	require.NoError(t, err, "Failed to create DesignStore for test")
	return store
}

// Helper to create a file with content, ensuring parent dirs exist.
func createTestFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "Failed to create directory for test file: %s", dir)
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "Failed to write test file: %s", path)
}

func TestNewDesignStore(t *testing.T) {
	t.Run("Success Case", func(t *testing.T) {
		tempDir := t.TempDir()
		store, err := NewDesignStore(tempDir)
		require.NoError(t, err)
		require.NotNil(t, store)
		assert.Equal(t, tempDir, store.basePath)

		// Verify directory was created
		_, statErr := os.Stat(tempDir)
		assert.NoError(t, statErr, "Base directory should exist")
	})

	t.Run("Success with Empty Path (Uses Default)", func(t *testing.T) {
		// Temporarily set current working directory to a temp dir to isolate default creation
		originalWd, _ := os.Getwd()
		tempWd := t.TempDir()
		require.NoError(t, os.Chdir(tempWd))
		defer os.Chdir(originalWd) // Change back

		store, err := NewDesignStore("") // Pass empty string
		require.NoError(t, err)
		require.NotNil(t, store)
		expectedDefaultPath := filepath.Join(tempWd, defaultDesignsBasePath)
		resolvedExpected, errExpected := filepath.EvalSymlinks(expectedDefaultPath)
		require.NoError(t, errExpected, "Failed to resolve expected path symlinks")
		resolvedActual, errActual := filepath.EvalSymlinks(store.basePath)
		require.NoError(t, errActual, "Failed to resolve actual path symlinks")

		assert.Equal(t, resolvedExpected, resolvedActual, "Resolved paths should match")

		// Verify default directory was created
		_, statErr := os.Stat(expectedDefaultPath)
		assert.NoError(t, statErr, "Default base directory should exist")
	})

	// Note: Testing failure due to permissions is hard in standard unit tests.
	// We trust os.MkdirAll handles most cases correctly.
}

func TestDesignStore_Paths(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "design123"
	sectionId := "sectionABC"
	promptName := "my_prompt"
	contentName := "main.svg"

	assert.Equal(t, filepath.Join(store.basePath, designId), store.getDesignPath(designId))
	assert.Equal(t, filepath.Join(store.basePath, designId, "design.json"), store.getDesignMetadataPath(designId))
	assert.Equal(t, filepath.Join(store.basePath, designId, "sections"), store.getSectionsBasePath(designId))
	assert.Equal(t, filepath.Join(store.basePath, designId, "sections", sectionId), store.GetSectionBasePath(designId, sectionId))
	assert.Equal(t, filepath.Join(store.basePath, designId, "sections", sectionId, "main.json"), store.getSectionDataPath(designId, sectionId))

	expectedPromptPath := filepath.Join(store.basePath, designId, "sections", sectionId, "prompts", promptName+".md")
	actualPromptPath, err := store.getSectionPromptPath(designId, sectionId, promptName)
	assert.NoError(t, err)
	assert.Equal(t, expectedPromptPath, actualPromptPath)
	_, statErr := os.Stat(filepath.Dir(expectedPromptPath)) // Check if prompts dir was created
	assert.NoError(t, statErr)

	expectedContentPath := filepath.Join(store.basePath, designId, "sections", sectionId, "content."+contentName)
	assert.Equal(t, expectedContentPath, store.GetContentPath(designId, sectionId, contentName))

	// Test prompt name sanitization and extension handling
	promptPathExt, _ := store.getSectionPromptPath(designId, sectionId, "prompt.txt")
	assert.Equal(t, filepath.Join(store.basePath, designId, "sections", sectionId, "prompts", "prompt.txt"), promptPathExt)

	// 1. Test empty prompt name -> Error
	_, errEmpty := store.getSectionPromptPath(designId, sectionId, "") // Test with strictly empty string
	require.Error(t, errEmpty, "Empty prompt name should return an error")
	assert.Contains(t, errEmpty.Error(), "invalid prompt name provided")

	// Test whitespace prompt name -> Error
	_, errWhitespace := store.getSectionPromptPath(designId, sectionId, "   ") // Test with whitespace
	require.Error(t, errWhitespace, "Whitespace prompt name should return an error")
	assert.Contains(t, errWhitespace.Error(), "invalid prompt name provided")

	// 2. Test path traversal attempt -> Sanitized path, NO error expected
	sanitizedPromptName := "invalid.md"
	expectedSanitizedPath := filepath.Join(store.basePath, designId, "sections", sectionId, "prompts", sanitizedPromptName)
	actualSanitizedPath, errSanitized := store.getSectionPromptPath(designId, sectionId, "../invalid")
	require.NoError(t, errSanitized, "Sanitization should prevent error")
	assert.Equal(t, expectedSanitizedPath, actualSanitizedPath, "Path should be sanitized")
}

func TestDesignStore_DesignExists(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "exist-test"

	// 1. Test non-existent
	exists, err := store.DesignExists(designId)
	require.NoError(t, err)
	assert.False(t, exists, "Design should not exist initially")

	// 2. Create directory and test
	designPath := store.getDesignPath(designId)
	err = os.MkdirAll(designPath, 0755)
	require.NoError(t, err)

	exists, err = store.DesignExists(designId)
	require.NoError(t, err)
	assert.True(t, exists, "Design should exist after creating directory")
}

func TestDesignStore_WriteReadDesignMetadata(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "meta-test"
	now := time.Now().UTC().Truncate(time.Second) // Truncate for consistent comparison
	metadata := &Design{
		Id:          designId,
		Name:        "Metadata Test Design",
		Description: "A test description",
		OwnerId:     "owner1",
		Visibility:  "public",
		SectionIds:  []string{"sec1", "sec2"},
		BaseModel:   BaseModel{CreatedAt: now, UpdatedAt: now},
	}

	// 1. Write metadata
	err := store.WriteDesignMetadata(designId, metadata)
	require.NoError(t, err, "Failed to write design metadata")

	// 2. Verify file content directly
	metadataPath := store.getDesignMetadataPath(designId)
	_, statErr := os.Stat(metadataPath)
	require.NoError(t, statErr, "Metadata file should exist")
	jsonData, readErr := os.ReadFile(metadataPath)
	require.NoError(t, readErr)
	var readBackMeta Design
	jsonErr := json.Unmarshal(jsonData, &readBackMeta)
	require.NoError(t, jsonErr)
	assert.Equal(t, metadata.Id, readBackMeta.Id)
	assert.Equal(t, metadata.Name, readBackMeta.Name)
	assert.Equal(t, metadata.SectionIds, readBackMeta.SectionIds)
	// Compare time after marshalling/unmarshalling
	assert.WithinDuration(t, metadata.CreatedAt, readBackMeta.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, metadata.UpdatedAt, readBackMeta.UpdatedAt, time.Millisecond)

	// 3. Read metadata using the store method
	readMetaStore, err := store.ReadDesignMetadata(designId)
	require.NoError(t, err, "Failed to read metadata using store")
	require.NotNil(t, readMetaStore)
	assert.Equal(t, metadata.Id, readMetaStore.Id)
	assert.Equal(t, metadata.Name, readMetaStore.Name)
	assert.Equal(t, metadata.SectionIds, readMetaStore.SectionIds)
	assert.WithinDuration(t, metadata.CreatedAt, readMetaStore.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, metadata.UpdatedAt, readMetaStore.UpdatedAt, time.Millisecond)

	// 4. Test Read Not Found
	_, err = store.ReadDesignMetadata("non-existent-design")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchEntity), "Expected ErrNoSuchEntity for non-existent design")

	// 5. Test Read Corrupted Data
	corruptPath := store.getDesignMetadataPath(designId)
	createTestFile(t, corruptPath, "{ invalid json")
	_, err = store.ReadDesignMetadata(designId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse design metadata")

	// 6. Test Write with ID mismatch
	metadataWrongId := &Design{Id: "wrong-id", Name: "Mismatch"}
	err = store.WriteDesignMetadata(designId, metadataWrongId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match design ID")
}

func TestDesignStore_WriteReadSectionData(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "section-data-design"
	sectionId := "section-data-sec1"
	now := time.Now().UTC().Truncate(time.Second)
	sectionData := &Section{
		Id:        sectionId,
		DesignId:  designId,
		Type:      "text",
		Title:     "Section Data Test",
		BaseModel: BaseModel{CreatedAt: now, UpdatedAt: now},
	}

	// 1. Write section data
	err := store.WriteSectionData(designId, sectionId, sectionData)
	require.NoError(t, err, "Failed to write section data")

	// 2. Verify file content directly
	dataPath := store.getSectionDataPath(designId, sectionId)
	_, statErr := os.Stat(dataPath)
	require.NoError(t, statErr, "Section data file should exist")
	jsonData, readErr := os.ReadFile(dataPath)
	require.NoError(t, readErr)
	var readBackData Section
	jsonErr := json.Unmarshal(jsonData, &readBackData)
	require.NoError(t, jsonErr)
	assert.Equal(t, sectionData.Id, readBackData.Id)
	assert.Equal(t, sectionData.DesignId, readBackData.DesignId)
	assert.Equal(t, sectionData.Title, readBackData.Title)
	assert.WithinDuration(t, sectionData.CreatedAt, readBackData.CreatedAt, time.Millisecond)

	// 3. Read section data using store method
	readDataStore, err := store.ReadSectionData(designId, sectionId)
	require.NoError(t, err, "Failed to read section data using store")
	require.NotNil(t, readDataStore)
	assert.Equal(t, sectionData.Id, readDataStore.Id)
	assert.Equal(t, sectionData.Title, readDataStore.Title)
	assert.WithinDuration(t, sectionData.CreatedAt, readDataStore.CreatedAt, time.Millisecond)

	// 4. Test Read Not Found
	_, err = store.ReadSectionData(designId, "non-existent-sec")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchEntity), "Expected ErrNoSuchEntity for non-existent section")

	// 5. Test Read Corrupted Data
	corruptPath := store.getSectionDataPath(designId, sectionId)
	createTestFile(t, corruptPath, "} invalid {")
	_, err = store.ReadSectionData(designId, sectionId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse section metadata")

	// 6. Test Write with ID mismatch
	sectionDataWrongId := &Section{Id: "wrong-sec", DesignId: designId, Title: "Wrong Section ID"}
	err = store.WriteSectionData(designId, sectionId, sectionDataWrongId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sectionData ID")

	sectionDataWrongDesignId := &Section{Id: sectionId, DesignId: "wrong-design", Title: "Wrong Design ID"}
	err = store.WriteSectionData(designId, sectionId, sectionDataWrongDesignId)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sectionData design ID")
}

func TestDesignStore_WriteReadPromptFile(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "prompt-design"
	sectionId := "prompt-sec"
	promptName := "get_answer.md"
	content := "This is the prompt content."

	// 1. Write prompt
	err := store.WritePromptFile(designId, sectionId, promptName, content)
	require.NoError(t, err, "Failed to write prompt file")

	// 2. Verify file content directly
	promptPath, pathErr := store.getSectionPromptPath(designId, sectionId, promptName)
	require.NoError(t, pathErr)
	_, statErr := os.Stat(promptPath)
	require.NoError(t, statErr, "Prompt file should exist")
	readBytes, readErr := os.ReadFile(promptPath)
	require.NoError(t, readErr)
	assert.Equal(t, content, string(readBytes))

	// 3. Read prompt using store method
	readContentStore, err := store.ReadPromptFile(designId, sectionId, promptName)
	require.NoError(t, err, "Failed to read prompt using store")
	assert.Equal(t, content, readContentStore)

	// 4. Test Read Not Found
	_, err = store.ReadPromptFile(designId, sectionId, "non_existent_prompt.md")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchEntity), "Expected ErrNoSuchEntity for non-existent prompt")

	// 5. Test Write with empty name (should fail in path generation)
	err = store.WritePromptFile(designId, sectionId, "", "content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid prompt name")
}

func TestDesignStore_DeleteDesign(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "delete-design-test"
	designPath := store.getDesignPath(designId)
	sectionPath := store.GetSectionBasePath(designId, "sec1")

	// 1. Create structure to delete
	createTestFile(t, store.getDesignMetadataPath(designId), `{"id":"`+designId+`"}`)
	createTestFile(t, store.getSectionDataPath(designId, "sec1"), `{"id":"sec1"}`)

	// 2. Verify it exists
	_, statErr := os.Stat(designPath)
	require.NoError(t, statErr, "Design path should exist before delete")
	_, statErrSec := os.Stat(sectionPath)
	require.NoError(t, statErrSec, "Section path should exist before delete")

	// 3. Delete design
	err := store.DeleteDesign(designId)
	require.NoError(t, err, "Failed to delete design")

	// 4. Verify it's gone
	_, statErr = os.Stat(designPath)
	require.Error(t, statErr, "Design path should not exist after delete")
	assert.True(t, errors.Is(statErr, os.ErrNotExist))

	// 5. Delete non-existent (idempotency)
	err = store.DeleteDesign("non-existent-design-del")
	assert.NoError(t, err, "Deleting non-existent design should not error")

	// 6. Delete again (idempotency)
	err = store.DeleteDesign(designId)
	assert.NoError(t, err, "Deleting already deleted design should not error")
}

func TestDesignStore_DeleteSection(t *testing.T) {
	store := newTestDesignStore(t)
	designId := "delete-section-design"
	secIdToDelete := "sec-to-delete"
	secIdToKeep := "sec-to-keep"

	designMetaPath := store.getDesignMetadataPath(designId)
	sectionPathDel := store.GetSectionBasePath(designId, secIdToDelete)
	sectionPathKeep := store.GetSectionBasePath(designId, secIdToKeep)

	// 1. Create structure
	createTestFile(t, designMetaPath, `{"id":"`+designId+`"}`)
	createTestFile(t, store.getSectionDataPath(designId, secIdToDelete), `{"id":"`+secIdToDelete+`"}`)
	createTestFile(t, store.getSectionDataPath(designId, secIdToKeep), `{"id":"`+secIdToKeep+`"}`)

	// 2. Verify existence
	_, statErrDel := os.Stat(sectionPathDel)
	require.NoError(t, statErrDel, "Section to delete path should exist before delete")
	_, statErrKeep := os.Stat(sectionPathKeep)
	require.NoError(t, statErrKeep, "Section to keep path should exist before delete")

	// 3. Delete section
	err := store.DeleteSection(designId, secIdToDelete)
	require.NoError(t, err, "Failed to delete section")

	// 4. Verify deleted section is gone
	_, statErrDel = os.Stat(sectionPathDel)
	require.Error(t, statErrDel, "Deleted section path should not exist after delete")
	assert.True(t, errors.Is(statErrDel, os.ErrNotExist))

	// 5. Verify other parts remain
	_, statErrKeep = os.Stat(sectionPathKeep)
	assert.NoError(t, statErrKeep, "Kept section path should still exist")
	_, statErrDesign := os.Stat(designMetaPath)
	assert.NoError(t, statErrDesign, "Design metadata should still exist")

	// 6. Delete non-existent (idempotency)
	err = store.DeleteSection(designId, "non-existent-sec-del")
	assert.NoError(t, err, "Deleting non-existent section should not error")

	// 7. Delete again (idempotency)
	err = store.DeleteSection(designId, secIdToDelete)
	assert.NoError(t, err, "Deleting already deleted section should not error")
}
