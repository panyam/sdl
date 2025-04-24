// FILE: ./services/content_test.go
package services

import (
	"os"
	"path/filepath"
	"testing"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Helper to create section metadata (main.json) without content for content tests
func setupSectionMetaForContentTest(t *testing.T, basePath, designId, sectionId, ownerId string) {
	t.Helper()
	// Ensure design exists first
	createDesignDirectly(t, basePath, Design{Id: designId, OwnerId: ownerId, SectionIds: []string{sectionId}})
	// Create section metadata
	createSectionDirectly(t, basePath, designId, sectionId, Section{
		Id: sectionId, DesignId: designId, Type: "text", Title: "Content Test Section",
	})
}

// Helper to directly create a content file
func createContentDirectly(t *testing.T, basePath, designId, sectionId, contentName string, contentBytes []byte) {
	t.Helper()
	ds := &DesignService{basePath: basePath} // Temporary service just for path helper
	contentPath := ds.getContentPath(designId, sectionId, contentName)
	sectionDir := filepath.Dir(contentPath)
	err := os.MkdirAll(sectionDir, 0755) // Ensure section dir exists
	require.NoError(t, err, "Failed to create section directory for direct content creation")
	err = os.WriteFile(contentPath, contentBytes, 0644)
	require.NoError(t, err, "Failed to write content file")
}

// Helper to read content directly
func readContentDirectly(t *testing.T, basePath, designId, sectionId, contentName string) ([]byte, error) {
	t.Helper()
	ds := &DesignService{basePath: basePath} // Temporary service just for path helper
	contentPath := ds.getContentPath(designId, sectionId, contentName)
	return os.ReadFile(contentPath)
}

// --- ContentService Test Cases ---

func TestSetContent(t *testing.T) {
	designService, tempDir := newTestDesignService(t)
	contentService := NewContentService(designService)
	ctx := testContextWithUser("content-owner")
	designId := "design-set-content"
	sectionId := "sec-set-1"
	contentName := "main"

	// Setup design and section metadata (without content file initially)
	setupSectionMetaForContentTest(t, tempDir, designId, sectionId, "content-owner")

	t.Run("Set Initial Content", func(t *testing.T) {
		contentBytes := []byte("<p>Initial Content</p>")
		req := &protos.SetContentRequest{
			DesignId:  designId,
			SectionId: sectionId,
			Content: &protos.Content{ // Provide metadata hint
				Name: contentName,
				Type: "text/html",
			},
			ContentBytes: contentBytes,
			// UpdateMask: nil, // Implicitly update bytes if provided? Or require mask? Assume required for now.
			// UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"content_bytes", "content"}}, // Include 'content' if updating metadata too
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"content_bytes"}}, // Mask only bytes
		}

		resp, err := contentService.SetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Content)
		assert.Equal(t, contentName, resp.Content.Name)
		assert.Equal(t, "text/html", resp.Content.Type) // Type from request proto
		assert.NotNil(t, resp.Content.UpdatedAt)

		// Verify file content
		readBytes, errRead := readContentDirectly(t, tempDir, designId, sectionId, contentName)
		require.NoError(t, errRead)
		assert.Equal(t, contentBytes, readBytes)

		// Verify timestamps updated (check modification time of file or main.json)
		secMeta := readSectionDataDirectly(t, tempDir, designId, sectionId)
		require.NotNil(t, secMeta)
		assert.True(t, secMeta.UpdatedAt.After(secMeta.CreatedAt)) // Updated time should be later
	})

	t.Run("Update Existing Content", func(t *testing.T) {
		contentBytes := []byte("<p>Updated Content</p>")
		req := &protos.SetContentRequest{
			DesignId:     designId,
			SectionId:    sectionId,
			Content:      &protos.Content{Name: contentName}, // Name is key
			ContentBytes: contentBytes,
			UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"content_bytes"}},
		}

		resp, err := contentService.SetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify file content updated
		readBytes, errRead := readContentDirectly(t, tempDir, designId, sectionId, contentName)
		require.NoError(t, errRead)
		assert.Equal(t, contentBytes, readBytes)
	})

	t.Run("Set Content for Different Name", func(t *testing.T) {
		contentBytes := []byte(`{"elements":[]}`)
		req := &protos.SetContentRequest{
			DesignId:  designId,
			SectionId: sectionId,
			Content: &protos.Content{
				Name:   "diagram.excalidraw.json",
				Type:   "application/json",
				Format: "excalidraw/json",
			},
			ContentBytes: contentBytes,
			UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"content_bytes"}}, // Assume metadata update isn't implemented yet
		}
		_, err := contentService.SetContent(ctx, req)
		require.NoError(t, err)

		// Verify new file created
		readBytes, errRead := readContentDirectly(t, tempDir, designId, sectionId, "diagram.excalidraw.json")
		require.NoError(t, errRead)
		assert.Equal(t, contentBytes, readBytes)

		// Verify original content still exists
		_, errReadOrig := readContentDirectly(t, tempDir, designId, sectionId, "main")
		require.NoError(t, errReadOrig)
	})

	t.Run("Set Content for Non-Existent Section", func(t *testing.T) {
		req := &protos.SetContentRequest{
			DesignId:     designId,
			SectionId:    "non-existent-sec",
			Content:      &protos.Content{Name: "main"},
			ContentBytes: []byte("test"),
			UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"content_bytes"}},
		}
		_, err := contentService.SetContent(ctx, req)
		require.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code()) // Should fail because section dir doesn't exist
	})

	// Add tests for permission denied, metadata updates (when implemented)
}

func TestGetContent(t *testing.T) {
	designService, tempDir := newTestDesignService(t)
	contentService := NewContentService(designService)
	ctx := testContextWithUser("content-owner")
	designId := "design-get-content"
	sectionId := "sec-get-1"
	contentNameMain := "main"
	contentNameDiagram := "diagram.json"
	mainBytes := []byte("<p>Main HTML</p>")
	diagramBytes := []byte(`{"key":"value"}`)

	// Setup
	setupSectionMetaForContentTest(t, tempDir, designId, sectionId, "content-owner")
	createContentDirectly(t, tempDir, designId, sectionId, contentNameMain, mainBytes)
	createContentDirectly(t, tempDir, designId, sectionId, contentNameDiagram, diagramBytes)

	t.Run("Get Existing Content (main)", func(t *testing.T) {
		req := &protos.GetContentRequest{DesignId: designId, SectionId: sectionId, Name: contentNameMain}
		resp, err := contentService.GetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Content)
		assert.Equal(t, mainBytes, resp.ContentBytes)
		assert.Equal(t, contentNameMain, resp.Content.Name)
		assert.Equal(t, "application/octet-stream", resp.Content.Type) // Default guess for now
	})

	t.Run("Get Existing Content (diagram)", func(t *testing.T) {
		req := &protos.GetContentRequest{DesignId: designId, SectionId: sectionId, Name: contentNameDiagram}
		resp, err := contentService.GetContent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Content)
		assert.Equal(t, diagramBytes, resp.ContentBytes)
		assert.Equal(t, contentNameDiagram, resp.Content.Name)
		assert.Equal(t, "application/json", resp.Content.Type) // Inferred from extension
	})

	t.Run("Get Non-Existent Content Name", func(t *testing.T) {
		req := &protos.GetContentRequest{DesignId: designId, SectionId: sectionId, Name: "nonexistent.png"}
		_, err := contentService.GetContent(ctx, req)
		require.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Get Content from Non-Existent Section", func(t *testing.T) {
		req := &protos.GetContentRequest{DesignId: designId, SectionId: "no-such-section", Name: contentNameMain}
		_, err := contentService.GetContent(ctx, req)
		require.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code()) // Should fail because content file path doesn't exist
	})
}

func TestDeleteContent(t *testing.T) {
	designService, tempDir := newTestDesignService(t)
	contentService := NewContentService(designService)
	ctx := testContextWithUser("content-owner")
	designId := "design-delete-content"
	sectionId := "sec-delete-1"
	contentNameMain := "main.html"
	contentNameOther := "other.data"
	mainBytes := []byte("<p>Main HTML</p>")
	otherBytes := []byte{1, 2, 3}

	// Setup
	setupSectionMetaForContentTest(t, tempDir, designId, sectionId, "content-owner")
	createContentDirectly(t, tempDir, designId, sectionId, contentNameMain, mainBytes)
	createContentDirectly(t, tempDir, designId, sectionId, contentNameOther, otherBytes)

	t.Run("Delete Existing Content", func(t *testing.T) {
		req := &protos.DeleteContentRequest{DesignId: designId, SectionId: sectionId, Name: contentNameMain} // Use Name field

		// Verify exists before
		_, errRead := readContentDirectly(t, tempDir, designId, sectionId, contentNameMain)
		require.NoError(t, errRead)

		_, err := contentService.DeleteContent(ctx, req)
		require.NoError(t, err)

		// Verify deleted
		_, errRead = readContentDirectly(t, tempDir, designId, sectionId, contentNameMain)
		require.Error(t, errRead)
		assert.True(t, os.IsNotExist(errRead))

		// Verify other content still exists
		readOtherBytes, errReadOther := readContentDirectly(t, tempDir, designId, sectionId, contentNameOther)
		require.NoError(t, errReadOther)
		assert.Equal(t, otherBytes, readOtherBytes)
	})

	t.Run("Delete Non-Existent Content (Idempotent)", func(t *testing.T) {
		req := &protos.DeleteContentRequest{DesignId: designId, SectionId: sectionId, Name: "nonexistent.bin"}
		_, err := contentService.DeleteContent(ctx, req)
		require.NoError(t, err) // Should succeed idempotently

		// Verify other content still exists
		_, errReadOther := readContentDirectly(t, tempDir, designId, sectionId, contentNameOther)
		require.NoError(t, errReadOther)
	})

	t.Run("Delete Already Deleted Content (Idempotent)", func(t *testing.T) {
		req := &protos.DeleteContentRequest{DesignId: designId, SectionId: sectionId, Name: contentNameMain}
		_, err := contentService.DeleteContent(ctx, req)
		require.NoError(t, err) // Should succeed idempotently
	})

	// Add tests for permissions, deleting from non-existent section etc.
}
