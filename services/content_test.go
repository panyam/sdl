// FILE: ./services/content_test.go
package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helper to create a ContentService instance with a DesignStore pointing to a temp directory.
func newTestContentService(t *testing.T) (*ContentService, *DesignStore) {
	t.Helper()
	// Use the existing helper from designs_store_test.go to create the store
	store := newTestDesignStore(t) // This now returns *DesignStore
	t.Logf("Using temp directory for test service: %s", store.basePath)

	service := NewContentService(store)                                     // Pass the store directly
	require.NotNil(t, service.store, "Service store should be initialized") // Add check
	return service, store
}

// Helper to create necessary directories for tests
func setupTestDirs(t *testing.T, store *DesignStore, designId, sectionId string) {
	t.Helper()
	sectionPath := store.GetSectionBasePath(designId, sectionId)
	err := os.MkdirAll(sectionPath, 0755)
	require.NoError(t, err, "Failed to create test directories")
}

// Helper to write content directly for setup
func createTestContentFile(t *testing.T, store *DesignStore, designId, sectionId, contentName, content string) string {
	t.Helper()
	contentPath := store.GetContentPath(designId, sectionId, contentName)
	// Ensure parent dirs exist first (WriteFile doesn't create intermediate dirs)
	err := os.MkdirAll(filepath.Dir(contentPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(contentPath, []byte(content), 0644)
	require.NoError(t, err)
	return contentPath
}

// --- Test Cases ---

func TestSetContent(t *testing.T) {
	service, store := newTestContentService(t)
	ctx := context.Background() // No auth needed directly in service for now
	designId := "set-content-design"
	sectionId := "set-content-sec"
	contentName := "main"
	initialContent := "<p>Initial</p>"
	updatedContent := "<p>Updated</p>"

	// Setup: Ensure base directories exist for the section
	setupTestDirs(t, store, designId, sectionId)
	// setupTime := time.Now() // Not needed for timestamp checks anymore

	t.Run("Set Initial Content", func(t *testing.T) {
		req := &protos.SetContentRequest{
			DesignId:     designId,
			SectionId:    sectionId,
			Name:         contentName,
			ContentBytes: []byte(initialContent),
			// ContentType and Format are passed but not used/stored by SetContent currently
		}
		resp, err := service.SetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp, "Response should not be nil")
		require.NotNil(t, resp.Content, "Response Content proto should not be nil")

		assert.Equal(t, contentName, resp.Content.Name)
		assert.NotNil(t, resp.Content.UpdatedAt) // Should get timestamp from the file system
		assert.Empty(t, resp.Content.Type, "ContentType is not expected in the SetContentResponse")
		assert.Empty(t, resp.Content.Format, "Format is not expected in the SetContentResponse")

		// Verify file content directly
		contentPath := store.GetContentPath(designId, sectionId, contentName)
		readBytes, readErr := os.ReadFile(contentPath)
		require.NoError(t, readErr)
		assert.Equal(t, initialContent, string(readBytes))

		// --- REMOVED Timestamp checks for design.json / main.json ---
	})

	t.Run("Update Existing Content", func(t *testing.T) {
		// Ensure initial content exists from previous step or setup here
		contentPath := createTestContentFile(t, store, designId, sectionId, contentName, initialContent)
		// updateStartTime := time.Now()

		req := &protos.SetContentRequest{
			DesignId:     designId,
			SectionId:    sectionId,
			Name:         contentName,
			ContentBytes: []byte(updatedContent),
		}
		resp, err := service.SetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Content)
		assert.Equal(t, contentName, resp.Content.Name)
		assert.NotNil(t, resp.Content.UpdatedAt)

		// Verify file content directly
		readBytes, readErr := os.ReadFile(contentPath)
		require.NoError(t, readErr)
		assert.Equal(t, updatedContent, string(readBytes))

		// --- REMOVED Timestamp checks for design.json / main.json ---
	})

	t.Run("Set Content for New Name", func(t *testing.T) {
		newName := "preview.svg"
		newContent := "<svg>...</svg>"
		req := &protos.SetContentRequest{
			DesignId:     designId,
			SectionId:    sectionId,
			Name:         newName,
			ContentBytes: []byte(newContent),
		}
		_, err := service.SetContent(ctx, req)
		require.NoError(t, err)

		// Verify new file exists
		contentPath := store.GetContentPath(designId, sectionId, newName)
		readBytes, readErr := os.ReadFile(contentPath)
		require.NoError(t, readErr)
		assert.Equal(t, newContent, string(readBytes))
	})

	t.Run("Fail on Missing Design Directory", func(t *testing.T) {
		req := &protos.SetContentRequest{
			DesignId:     "non-existent-design", // This design dir doesn't exist
			SectionId:    sectionId,
			Name:         contentName,
			ContentBytes: []byte(initialContent),
		}
		_, err := service.SetContent(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code()) // Should be NotFound because design doesn't exist
	})

	t.Run("Fail on Missing Section Directory (if creation fails)", func(t *testing.T) {
		// Simulate a scenario where section dir creation fails (hard to do reliably,
		// but test the case where ensureDir returns an error other than NotFound for design)
		// For now, assume if design exists, section dir creation should work unless permissions fail.
		// We can test the InvalidArgument cases.
	})

	t.Run("Fail on Invalid Args", func(t *testing.T) {
		testCases := []*protos.SetContentRequest{
			{SectionId: sectionId, Name: contentName, ContentBytes: []byte("a")},  // Missing DesignId
			{DesignId: designId, Name: contentName, ContentBytes: []byte("a")},    // Missing SectionId
			{DesignId: designId, SectionId: sectionId, ContentBytes: []byte("a")}, // Missing Name
			// {DesignId: designId, SectionId: sectionId, Name: contentName}, // Missing ContentBytes - Allowed now? Yes.
		}
		for _, tc := range testCases {
			_, err := service.SetContent(ctx, tc)
			assert.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		}
	})
}

func TestGetContent(t *testing.T) {
	service, store := newTestContentService(t)
	ctx := context.Background()
	designId := "get-content-design"
	sectionId := "get-content-sec"
	contentName := "main"
	contentData := "<p>Hello World</p>"

	// Setup: Create design/section dirs and the content file
	setupTestDirs(t, store, designId, sectionId)
	contentPath := createTestContentFile(t, store, designId, sectionId, contentName, contentData)
	fileInfo, _ := os.Stat(contentPath) // Get mod time for comparison
	expectedModTime := fileInfo.ModTime()

	t.Run("Get Existing Content", func(t *testing.T) {
		req := &protos.GetContentRequest{
			DesignId:  designId,
			SectionId: sectionId,
			Name:      contentName,
		}
		resp, err := service.GetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Content)

		assert.Equal(t, contentData, string(resp.ContentBytes))
		assert.Equal(t, contentName, resp.Content.Name)
		assert.NotNil(t, resp.Content.UpdatedAt)
		// Compare timestamps (allow for slight differences due to system time precision)
		assert.WithinDuration(t, expectedModTime, resp.Content.UpdatedAt.AsTime(), time.Second)
		assert.Empty(t, resp.Content.Type)   // Type is not stored/returned
		assert.Empty(t, resp.Content.Format) // Format is not stored/returned
	})

	t.Run("Get Non-Existent Content Name", func(t *testing.T) {
		req := &protos.GetContentRequest{
			DesignId:  designId,
			SectionId: sectionId,
			Name:      "non-existent-name",
		}
		_, err := service.GetContent(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code()) // Expect NotFound
	})

	t.Run("Get Content from Non-Existent Section", func(t *testing.T) {
		req := &protos.GetContentRequest{
			DesignId:  designId,
			SectionId: "non-existent-section",
			Name:      contentName,
		}
		_, err := service.GetContent(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code()) // Expect NotFound
	})

	t.Run("Get Content from Non-Existent Design", func(t *testing.T) {
		req := &protos.GetContentRequest{
			DesignId:  "non-existent-design",
			SectionId: sectionId,
			Name:      contentName,
		}
		_, err := service.GetContent(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code()) // Expect NotFound
	})

	t.Run("Fail on Invalid Args", func(t *testing.T) {
		testCases := []*protos.GetContentRequest{
			{SectionId: sectionId, Name: contentName},  // Missing DesignId
			{DesignId: designId, Name: contentName},    // Missing SectionId
			{DesignId: designId, SectionId: sectionId}, // Missing Name
		}
		for _, tc := range testCases {
			_, err := service.GetContent(ctx, tc)
			assert.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
		}
	})
}

func TestGetContentBytes(t *testing.T) {
	service, store := newTestContentService(t)
	ctx := context.Background()
	designId := "getbytes-design"
	sectionId := "getbytes-sec"
	contentName := "raw.bin"
	contentData := []byte{0x01, 0x02, 0x03, 0x04}

	// Setup
	setupTestDirs(t, store, designId, sectionId)
	_ = createTestContentFile(t, store, designId, sectionId, contentName, string(contentData)) // Use helper, convert bytes to string for helper

	t.Run("Get Existing Bytes", func(t *testing.T) {
		bytes, err := service.GetContentBytes(ctx, designId, sectionId, contentName)
		require.NoError(t, err)
		assert.Equal(t, contentData, bytes)
	})

	t.Run("Get Non-Existent Bytes", func(t *testing.T) {
		_, err := service.GetContentBytes(ctx, designId, sectionId, "no-such-bytes")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNoSuchEntity)) // Expect specific error
	})

	t.Run("Get Bytes Invalid Args", func(t *testing.T) {
		_, err := service.GetContentBytes(ctx, "", sectionId, contentName)
		assert.Error(t, err)
		_, err = service.GetContentBytes(ctx, designId, "", contentName)
		assert.Error(t, err)
		_, err = service.GetContentBytes(ctx, designId, sectionId, "")
		assert.Error(t, err)
	})
}
