package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	// "strings" // Not strictly needed in tests after removing some checks
	"sync"
	"testing"
	"time"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata" // <-- Add gRPC metadata import
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// ****** ADDED IMPORT ******
	"github.com/google/go-cmp/cmp" // Correct import for cmp.Equal/Diff
)

// Helper to create a context with a logged-in user ID for testing
func testContextWithUser(userId string) context.Context {
	md := metadata.Pairs("loggedinuserid", userId)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return ctx
}

// Helper to create a DesignService instance using a temporary directory
func newTestDesignService(t *testing.T) (*DesignService, string) {
	t.Helper()

	tempDir := t.TempDir()
	t.Logf("Using temp directory for test service: %s", tempDir)

	service := NewDesignService(nil, tempDir)

	t.Cleanup(func() {
		service.mutexMap = sync.Map{}
	})

	return service, tempDir
}

// Helper to directly create a design metadata file for test setup
func createDesignDirectly(t *testing.T, basePath string, metadata Design) {
	t.Helper()
	designPath := filepath.Join(basePath, metadata.Id)
	metadataPath := filepath.Join(designPath, "design.json")
	sectionsPath := filepath.Join(designPath, "sections")

	err := os.MkdirAll(sectionsPath, 0755)
	require.NoError(t, err, "Failed to create directories for direct creation")

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err, "Failed to marshal metadata for direct creation")

	err = os.WriteFile(metadataPath, jsonData, 0644)
	require.NoError(t, err, "Failed to write metadata file for direct creation")
}

// Helper to read design metadata directly from file for verification
func readDesignMetadata(t *testing.T, basePath string, designId string) *Design {
	t.Helper()
	metadataPath := filepath.Join(basePath, designId, "design.json")
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		require.NoError(t, err, "Failed to read metadata file for verification")
	}

	var metadata Design
	err = json.Unmarshal(jsonData, &metadata)
	require.NoError(t, err, "Failed to unmarshal metadata for verification")
	return &metadata
}

// --- Test Cases ---

func TestCreateDesign(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("test-creator")

	t.Run("Create with Provided ID", func(t *testing.T) {
		req := &protos.CreateDesignRequest{
			Design: &protos.Design{
				Id:          "test-create-1",
				Name:        "Test Create Design 1",
				Description: "Description 1",
				Visibility:  "public",
			},
		}
		resp, err := service.CreateDesign(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Design)
		assert.Equal(t, "test-create-1", resp.Design.Id)
		assert.Equal(t, "Test Create Design 1", resp.Design.Name)
		assert.Equal(t, "Description 1", resp.Design.Description)
		assert.Equal(t, "public", resp.Design.Visibility)
		assert.NotEmpty(t, resp.Design.OwnerId)
		assert.Equal(t, "test-creator", resp.Design.OwnerId) // Verify owner ID
		assert.NotNil(t, resp.Design.CreatedAt)
		assert.NotNil(t, resp.Design.UpdatedAt)

		meta := readDesignMetadata(t, tempDir, "test-create-1")
		require.NotNil(t, meta)
		assert.Equal(t, "test-create-1", meta.Id)
		assert.Equal(t, "Test Create Design 1", meta.Name)
		assert.Equal(t, "public", meta.Visibility)
		assert.Equal(t, resp.Design.OwnerId, meta.OwnerId)
	})

	t.Run("Create with Auto-Generated ID", func(t *testing.T) {
		req := &protos.CreateDesignRequest{
			Design: &protos.Design{
				Name:       "Test Auto ID",
				Visibility: "private",
			},
		}
		resp, err := service.CreateDesign(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Design)
		// assert.Len(t, resp.Design.Id, service.idLength)
		assert.Regexp(t, `^[a-zA-Z0-9]+$`, resp.Design.Id)
		assert.Equal(t, "Test Auto ID", resp.Design.Name)
		assert.Equal(t, "private", resp.Design.Visibility)
		assert.Equal(t, "test-creator", resp.Design.OwnerId) // Verify owner ID

		generatedId := resp.Design.Id
		meta := readDesignMetadata(t, tempDir, generatedId)
		require.NotNil(t, meta)
		assert.Equal(t, generatedId, meta.Id)
		assert.Equal(t, "Test Auto ID", meta.Name)
		assert.Equal(t, "private", meta.Visibility)
	})

	t.Run("Fail on Empty Name", func(t *testing.T) {
		req := &protos.CreateDesignRequest{
			Design: &protos.Design{
				Id:   "test-empty-name",
				Name: "  ",
			},
		}
		_, err := service.CreateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "name cannot be empty")
	})

	t.Run("Fail on Already Exists (Provided Id)", func(t *testing.T) {
		createDesignDirectly(t, tempDir, Design{Id: "test-conflict", Name: "Existing"})

		req := &protos.CreateDesignRequest{
			Design: &protos.Design{
				Id:   "test-conflict",
				Name: "Trying to overwrite",
			},
		}
		_, err := service.CreateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})
}

func TestGetDesign(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("test-creator")

	now := time.Now().UTC().Truncate(time.Second)
	setupMeta := Design{
		Id:          "test-get-1",
		OwnerId:     "test-owner",
		Name:        "Test Get Design",
		Description: "Get Description",
		Visibility:  "private",
		VisibleTo:   []string{"user1", "user2"},
		BaseModel: BaseModel{
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now,
		},
		SectionIds: []string{},
	}
	createDesignDirectly(t, tempDir, setupMeta)

	t.Run("Get Existing Design", func(t *testing.T) {
		req := &protos.GetDesignRequest{Id: "test-get-1"}
		resp, err := service.GetDesign(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Design)

		expectedProto := &protos.Design{
			Id:          setupMeta.Id,
			OwnerId:     setupMeta.OwnerId,
			Name:        setupMeta.Name,
			Description: setupMeta.Description,
			Visibility:  setupMeta.Visibility,
			VisibleTo:   setupMeta.VisibleTo,
			CreatedAt:   tspb.New(setupMeta.CreatedAt),
			UpdatedAt:   tspb.New(setupMeta.UpdatedAt),
			SectionIds:  []string{},
		}
		// ****** FIXED: Ensure cmp import is present ******
		assert.True(t, cmp.Equal(expectedProto, resp.Design, protocmp.Transform()), cmp.Diff(expectedProto, resp.Design, protocmp.Transform()))
	})

	t.Run("Get Non-Existent Design", func(t *testing.T) {
		req := &protos.GetDesignRequest{Id: "does-not-exist"}
		_, err := service.GetDesign(ctx, req)

		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Fail on Corrupted Metadata", func(t *testing.T) {
		designId := "corrupted-meta"
		designPath := filepath.Join(tempDir, designId)
		metadataPath := filepath.Join(designPath, "design.json")
		sectionsPath := filepath.Join(designPath, "sections")
		require.NoError(t, os.MkdirAll(sectionsPath, 0755))
		require.NoError(t, os.WriteFile(metadataPath, []byte("{ invalid json "), 0644))

		req := &protos.GetDesignRequest{Id: designId}
		_, err := service.GetDesign(ctx, req)

		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Contains(t, []codes.Code{codes.DataLoss, codes.Internal}, st.Code())
		assert.Equal(t, codes.DataLoss, st.Code()) // Be specific if possible
	})
}

func TestListDesigns(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("user1")

	ts := time.Now().UTC().Truncate(time.Second)
	meta1 := Design{Id: "list-aaa", OwnerId: "user1", Name: "AAA Design", Visibility: "public", BaseModel: BaseModel{UpdatedAt: ts.Add(-2 * time.Hour)}}
	meta2 := Design{Id: "list-ccc", OwnerId: "user2", Name: "CCC Design", Visibility: "private", BaseModel: BaseModel{UpdatedAt: ts}}
	meta3 := Design{Id: "list-bbb", OwnerId: "user1", Name: "BBB Design", Visibility: "public", BaseModel: BaseModel{UpdatedAt: ts.Add(-1 * time.Hour)}}
	createDesignDirectly(t, tempDir, meta1)
	createDesignDirectly(t, tempDir, meta2)
	createDesignDirectly(t, tempDir, meta3)
	require.NoError(t, os.Mkdir(filepath.Join(tempDir, "list-corrupt"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "list-corrupt", "design.json"), []byte("bad"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tempDir, "list-no-meta"), 0755))

	t.Run("List All (Default Order - Recent)", func(t *testing.T) {
		req := &protos.ListDesignsRequest{}
		resp, err := service.ListDesigns(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Designs, 3)
		assert.False(t, resp.Pagination.HasMore)
		assert.EqualValues(t, 3, resp.Pagination.TotalResults)
		assert.Equal(t, "list-ccc", resp.Designs[0].Id)
		assert.Equal(t, "list-bbb", resp.Designs[1].Id)
		assert.Equal(t, "list-aaa", resp.Designs[2].Id)
	})

	t.Run("List Sort by Name", func(t *testing.T) {
		req := &protos.ListDesignsRequest{OrderBy: "name"}
		resp, err := service.ListDesigns(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Designs, 3)
		assert.Equal(t, "list-aaa", resp.Designs[0].Id)
		assert.Equal(t, "list-bbb", resp.Designs[1].Id)
		assert.Equal(t, "list-ccc", resp.Designs[2].Id)
	})

	t.Run("List Filter by Owner", func(t *testing.T) {
		req := &protos.ListDesignsRequest{OwnerId: "user1"}
		resp, err := service.ListDesigns(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Designs, 2)
		assert.Equal(t, "list-bbb", resp.Designs[0].Id)
		assert.Equal(t, "list-aaa", resp.Designs[1].Id)
	})

	t.Run("List Filter by Public", func(t *testing.T) {
		req := &protos.ListDesignsRequest{LimitToPublic: true}
		resp, err := service.ListDesigns(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Designs, 2)
		assert.Equal(t, "list-bbb", resp.Designs[0].Id)
		assert.Equal(t, "list-aaa", resp.Designs[1].Id)
	})

	t.Run("List with Pagination", func(t *testing.T) {
		req1 := &protos.ListDesignsRequest{Pagination: &protos.Pagination{PageSize: 2, PageOffset: 0}}
		resp1, err1 := service.ListDesigns(ctx, req1)
		require.NoError(t, err1)
		require.NotNil(t, resp1)
		assert.Len(t, resp1.Designs, 2)
		assert.True(t, resp1.Pagination.HasMore)
		assert.EqualValues(t, 3, resp1.Pagination.TotalResults)
		assert.Equal(t, "list-ccc", resp1.Designs[0].Id)
		assert.Equal(t, "list-bbb", resp1.Designs[1].Id)

		req2 := &protos.ListDesignsRequest{Pagination: &protos.Pagination{PageSize: 2, PageOffset: 2}}
		resp2, err2 := service.ListDesigns(ctx, req2)
		require.NoError(t, err2)
		require.NotNil(t, resp2)
		assert.Len(t, resp2.Designs, 1)
		assert.False(t, resp2.Pagination.HasMore)
		assert.EqualValues(t, 3, resp2.Pagination.TotalResults)
		assert.Equal(t, "list-aaa", resp2.Designs[0].Id)

		req3 := &protos.ListDesignsRequest{Pagination: &protos.Pagination{PageSize: 2, PageOffset: 4}}
		resp3, err3 := service.ListDesigns(ctx, req3)
		require.NoError(t, err3)
		require.NotNil(t, resp3)
		assert.Empty(t, resp3.Designs)
		assert.False(t, resp3.Pagination.HasMore)
		assert.EqualValues(t, 3, resp3.Pagination.TotalResults)
	})
}

func TestUpdateDesign(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("updater")

	setupTime := time.Now().UTC().Truncate(time.Second)
	meta := Design{Id: "update-1", OwnerId: "updater", Name: "Original Name", Description: "Original Desc", Visibility: "private", BaseModel: BaseModel{CreatedAt: setupTime, UpdatedAt: setupTime}}
	createDesignDirectly(t, tempDir, meta)

	t.Run("Update Single Field (Name)", func(t *testing.T) {
		req := &protos.UpdateDesignRequest{
			Design: &protos.Design{
				Id:   "update-1",
				Name: "Updated Name",
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}
		resp, err := service.UpdateDesign(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "update-1", resp.Design.Id)
		assert.Equal(t, "Updated Name", resp.Design.Name)
		assert.Equal(t, "Original Desc", resp.Design.Description)
		assert.Equal(t, "private", resp.Design.Visibility)
		assert.True(t, resp.Design.UpdatedAt.AsTime().After(setupTime))

		updatedMeta := readDesignMetadata(t, tempDir, "update-1")
		require.NotNil(t, updatedMeta)
		assert.Equal(t, "Updated Name", updatedMeta.Name)
		assert.Equal(t, "Original Desc", updatedMeta.Description)
		assert.True(t, updatedMeta.UpdatedAt.After(setupTime))
		assert.Equal(t, setupTime, updatedMeta.CreatedAt)
	})

	t.Run("Update Multiple Fields (Desc, Visibility)", func(t *testing.T) {
		currentMeta := readDesignMetadata(t, tempDir, "update-1")
		require.NotNil(t, currentMeta)
		previousUpdateTime := currentMeta.UpdatedAt

		req := &protos.UpdateDesignRequest{
			Design: &protos.Design{
				Id:          "update-1",
				Name:        "Name From Request (ignored)",
				Description: "Updated Description",
				Visibility:  "public",
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description", "visibility"}},
		}
		resp, err := service.UpdateDesign(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Updated Name", resp.Design.Name)
		assert.Equal(t, "Updated Description", resp.Design.Description)
		assert.Equal(t, "public", resp.Design.Visibility)
		assert.True(t, resp.Design.UpdatedAt.AsTime().After(previousUpdateTime))

		updatedMeta := readDesignMetadata(t, tempDir, "update-1")
		require.NotNil(t, updatedMeta)
		assert.Equal(t, "Updated Name", updatedMeta.Name)
		assert.Equal(t, "Updated Description", updatedMeta.Description)
		assert.Equal(t, "public", updatedMeta.Visibility)
		assert.True(t, updatedMeta.UpdatedAt.After(previousUpdateTime))
	})

	t.Run("Fail on Not Found", func(t *testing.T) {
		req := &protos.UpdateDesignRequest{
			Design:     &protos.Design{Id: "not-found"},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}
		_, err := service.UpdateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Fail on Empty Name Update", func(t *testing.T) {
		req := &protos.UpdateDesignRequest{
			Design:     &protos.Design{Id: "update-1", Name: ""},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		}
		_, err := service.UpdateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "name cannot be updated to empty")
	})

	t.Run("Fail on Invalid Mask Path", func(t *testing.T) {
		req := &protos.UpdateDesignRequest{
			Design:     &protos.Design{Id: "update-1"},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"invalid_field"}},
		}
		_, err := service.UpdateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "invalid or unsupported path")
	})

	t.Run("Fail without Update Mask", func(t *testing.T) {
		req := &protos.UpdateDesignRequest{
			Design: &protos.Design{Id: "update-1", Name: "No Mask Update"},
		}
		_, err := service.UpdateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "Update mask must be provided")
	})
}

func TestDeleteDesign(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("deleter")

	meta := Design{Id: "delete-1", OwnerId: "deleter", Name: "To Be Deleted", BaseModel: BaseModel{CreatedAt: time.Now(), UpdatedAt: time.Now()}}
	createDesignDirectly(t, tempDir, meta)

	t.Run("Delete Existing", func(t *testing.T) {
		req := &protos.DeleteDesignRequest{Id: "delete-1"}

		_, err := os.Stat(filepath.Join(tempDir, "delete-1"))
		require.NoError(t, err, "Design should exist before delete")

		resp, err := service.DeleteDesign(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		_, err = os.Stat(filepath.Join(tempDir, "delete-1"))
		require.Error(t, err, "Design should not exist after delete")
		assert.True(t, os.IsNotExist(err), "Error should be os.ErrNotExist")

		_, getErr := service.GetDesign(ctx, &protos.GetDesignRequest{Id: "delete-1"})
		require.Error(t, getErr)
		st, ok := status.FromError(getErr)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Delete Non-Existent (Idempotency)", func(t *testing.T) {
		req := &protos.DeleteDesignRequest{Id: "does-not-exist"}
		resp, err := service.DeleteDesign(ctx, req)
		require.NoError(t, err, "Deleting non-existent should not error")
		require.NotNil(t, resp)
	})

	t.Run("Delete Already Deleted (Idempotency)", func(t *testing.T) {
		req := &protos.DeleteDesignRequest{Id: "delete-1"}
		resp, err := service.DeleteDesign(ctx, req)
		require.NoError(t, err, "Deleting already deleted should not error")
		require.NotNil(t, resp)
	})
}

func TestGetDesigns(t *testing.T) {
	service, _ := newTestDesignService(t)
	ctx := testContextWithUser("test-user")
	req := &protos.GetDesignsRequest{Ids: []string{"some-id"}}

	_, err := service.GetDesigns(ctx, req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}

// --- NEW Section Tests ---

// Helper to create a section file directly
func createSectionDirectly(t *testing.T, basePath, designId, sectionId string, sectionData Section) {
	t.Helper()
	sectionsDir := filepath.Join(basePath, designId, "sections", sectionId)
	err := os.MkdirAll(sectionsDir, 0755) // Ensure sections dir exists
	require.NoError(t, err, "Failed to create sections directory for direct section creation")

	sectionPath := filepath.Join(sectionsDir, "main.json")
	jsonData, err := json.MarshalIndent(sectionData, "", "  ")
	require.NoError(t, err, "Failed to marshal section data for direct creation")
	err = os.WriteFile(sectionPath, jsonData, 0644)
	require.NoError(t, err, "Failed to write section file for direct creation")
}

// Helper to read section data directly
func readSectionDataDirectly(t *testing.T, basePath, designId, sectionId string) *Section {
	t.Helper()
	sectionsDir := filepath.Join(basePath, designId, "sections", sectionId)
	sectionPath := filepath.Join(sectionsDir, "main.json")
	jsonData, err := os.ReadFile(sectionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		require.NoError(t, err, "Failed to read section file for verification")
	}
	var sectionData Section
	err = json.Unmarshal(jsonData, &sectionData)
	require.NoError(t, err, "Failed to unmarshal section data for verification")
	return &sectionData
}

func TestGetSection(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := testContextWithUser("section-owner")
	designId := "design-for-sections"
	sectionId := "sec-1"

	// Setup Design
	createDesignDirectly(t, tempDir, Design{
		Id:         designId,
		OwnerId:    "section-owner",
		Name:       "Design with Sections",
		SectionIds: []string{sectionId}, // Ensure design metadata lists the section
	})

	// Setup Section
	now := time.Now()
	sectionData := Section{
		Id:        sectionId,
		DesignId:  designId,
		Type:      "text",
		Title:     "Test Section Title",
		Content:   "<p>Hello Section</p>",
		BaseModel: BaseModel{CreatedAt: now, UpdatedAt: now},
	}
	createSectionDirectly(t, tempDir, designId, sectionId, sectionData)

	t.Run("Get Existing Section", func(t *testing.T) {
		req := &protos.GetSectionRequest{DesignId: designId, SectionId: sectionId}
		resp, err := service.GetSection(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, sectionId, resp.Id)
		assert.Equal(t, designId, resp.DesignId)
		assert.Equal(t, protos.SectionType_SECTION_TYPE_TEXT, resp.Type)
		assert.Equal(t, "Test Section Title", resp.Title)
		require.NotNil(t, resp.GetTextContent(), "TextContent should not be nil")
		assert.Equal(t, "<p>Hello Section</p>", resp.GetTextContent().HtmlContent)
		// assert.EqualValues(t, 0, resp.Order) // Order is not directly returned by GetSection in this implementation
	})

	t.Run("Get Non-Existent Section", func(t *testing.T) {
		req := &protos.GetSectionRequest{DesignId: designId, SectionId: "non-existent-sec"}
		_, err := service.GetSection(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("Get Section from Non-Existent Design", func(t *testing.T) {
		req := &protos.GetSectionRequest{DesignId: "non-existent-design", SectionId: sectionId}
		_, err := service.GetSection(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		// Should be NotFound because the design path doesn't exist
		assert.Equal(t, codes.NotFound, st.Code())
	})

}

func TestAddSection(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ownerId := "add-sec-owner"
	ctx := testContextWithUser(ownerId)
	designId := "design-add-sec"

	// Setup Design
	createDesignDirectly(t, tempDir, Design{
		Id:         designId,
		OwnerId:    ownerId,
		Name:       "Design for Adding Sections",
		SectionIds: []string{"sec-a"}, // Start with one section
	})
	createSectionDirectly(t, tempDir, designId, "sec-a", Section{Id: "sec-a", DesignId: designId, Type: "text", Title: "Section A"})

	t.Run("Add Section to End", func(t *testing.T) {
		req := &protos.AddSectionRequest{
			Section: &protos.Section{
				DesignId: designId, // Important: Pass design ID in section proto now
				Type:     protos.SectionType_SECTION_TYPE_DRAWING,
				Title:    "New Drawing Section",
				Content:  &protos.Section_DrawingContent{DrawingContent: &protos.DrawingSectionContent{Data: []byte(`{"key":"val"}`)}},
			},
			Position: protos.PositionType_POSITION_TYPE_END, // Explicitly add to end
		}
		resp, err := service.AddSection(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		newSectionId := resp.Id
		assert.NotEmpty(t, newSectionId)
		assert.Equal(t, protos.SectionType_SECTION_TYPE_DRAWING, resp.Type)
		assert.Equal(t, "New Drawing Section", resp.Title)
		assert.EqualValues(t, 1, resp.Order) // Should be order 1 (after sec-a at index 0)

		// Verify metadata
		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		assert.Equal(t, []string{"sec-a", newSectionId}, meta.SectionIds)

		// Verify section file
		secData := readSectionDataDirectly(t, tempDir, designId, newSectionId)
		require.NotNil(t, secData)
		assert.Equal(t, "drawing", secData.Type)
		assert.Equal(t, "New Drawing Section", secData.Title)
		// Content is []byte, which json marshals to base64
		_, ok := secData.Content.(string) // Check if it's string (base64)
		assert.True(t, ok, "Drawing content should be stored as base64 string in JSON")

	})

	t.Run("Add Section After", func(t *testing.T) {
		req := &protos.AddSectionRequest{
			Section: &protos.Section{
				DesignId: designId,
				Type:     protos.SectionType_SECTION_TYPE_TEXT,
				Title:    "Section B",
				Content:  &protos.Section_TextContent{TextContent: &protos.TextSectionContent{HtmlContent: "<p>B</p>"}},
			},
			RelativeSectionId: "sec-a",
			Position:          protos.PositionType_POSITION_TYPE_AFTER,
		}
		resp, err := service.AddSection(ctx, req)
		require.NoError(t, err)
		newSectionIdB := resp.Id
		assert.EqualValues(t, 1, resp.Order) // Should be inserted at index 1

		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		// Order depends on previous test: ["sec-a", newDrawingId, newSectionIdB] ?
		// Let's assume the drawing section from previous test is newId1
		require.GreaterOrEqual(t, len(meta.SectionIds), 3) // A, Drawing, B
		foundA := false
		// foundB := false
		for i, id := range meta.SectionIds {
			if id == "sec-a" {
				foundA = true
				// B should be immediately after A if no other sections were added between tests
				// This assertion is fragile if tests run concurrently or state persists.
				// Better check: Ensure B exists and A exists before it in the final list.
				require.Contains(t, meta.SectionIds, newSectionIdB)
				// Find B's index relative to A's index
				indexB := -1
				for j, bid := range meta.SectionIds {
					if bid == newSectionIdB {
						indexB = j
						break
					}
				}
				require.NotEqual(t, -1, indexB)
				require.Greater(t, indexB, i) // B's index must be after A's index
				break
			}
		}
		require.True(t, foundA, "Section A should be in the list")

		secDataB := readSectionDataDirectly(t, tempDir, designId, newSectionIdB)
		require.NotNil(t, secDataB)
		assert.Equal(t, "text", secDataB.Type)
	})

	// Add more tests: Add Before, Add without type, Add to non-existent design, Permission denied
}

func TestUpdateSection(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ownerId := "update-sec-owner"
	ctx := testContextWithUser(ownerId)
	designId := "design-update-sec"
	secIdText := "update-text-1"
	secIdDraw := "update-draw-1"

	// Setup Design & Sections
	createDesignDirectly(t, tempDir, Design{
		Id:         designId,
		OwnerId:    ownerId,
		SectionIds: []string{secIdText, secIdDraw},
	})
	createSectionDirectly(t, tempDir, designId, secIdText, Section{Id: secIdText, DesignId: designId, Type: "text", Title: "Original Text", Content: "<p>Original</p>"})
	createSectionDirectly(t, tempDir, designId, secIdDraw, Section{Id: secIdDraw, DesignId: designId, Type: "drawing", Title: "Original Draw", Content: []byte(`{"a":1}`)}) // Store []byte

	t.Run("Update Text Content", func(t *testing.T) {
		newContent := "<p>Updated Content</p>"
		req := &protos.UpdateSectionRequest{
			Section: &protos.Section{ // Payload only needs updated fields + ID/DesignID maybe? Check API impl. No, handled by paths.
				DesignId: designId,
				Id:       secIdText,
				Type:     protos.SectionType_SECTION_TYPE_TEXT, // Good practice to send type
				Content:  &protos.Section_TextContent{TextContent: &protos.TextSectionContent{HtmlContent: newContent}},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"text_content"}},
		}
		resp, err := service.UpdateSection(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Original Text", resp.Title) // Title should not change
		require.NotNil(t, resp.GetTextContent())
		assert.Equal(t, newContent, resp.GetTextContent().HtmlContent)

		// Verify file
		secData := readSectionDataDirectly(t, tempDir, designId, secIdText)
		require.NotNil(t, secData)
		assert.Equal(t, newContent, secData.Content)
		assert.Equal(t, "Original Text", secData.Title)
	})

	t.Run("Update Drawing Content", func(t *testing.T) {
		newContentBytes := []byte(`{"a":2, "b": "new"}`)
		req := &protos.UpdateSectionRequest{
			Section: &protos.Section{
				DesignId: designId,
				Id:       secIdDraw,
				Type:     protos.SectionType_SECTION_TYPE_DRAWING,
				Content:  &protos.Section_DrawingContent{DrawingContent: &protos.DrawingSectionContent{Data: newContentBytes}},
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"drawing_content"}},
		}
		resp, err := service.UpdateSection(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Original Draw", resp.Title) // Title should not change
		require.NotNil(t, resp.GetDrawingContent())
		assert.Equal(t, newContentBytes, resp.GetDrawingContent().Data) // Compare bytes

		// Verify file
		secData := readSectionDataDirectly(t, tempDir, designId, secIdDraw)
		require.NotNil(t, secData)
		// JSON stores bytes as base64 string
		_ /*contentStr*/, ok := secData.Content.(string)
		require.True(t, ok, "Content should be base64 string")
		// Decode and compare underlying data if needed, or just check type for now
		// For simplicity, we assume if it's a string, it's the base64 from json marshal
		// A more robust test would decode base64 and unmarshal the JSON inside
		assert.Equal(t, "Original Draw", secData.Title)
	})

	t.Run("Update Title Only", func(t *testing.T) {
		newTitle := "Updated Text Title"
		req := &protos.UpdateSectionRequest{
			Section: &protos.Section{
				DesignId: designId,
				Id:       secIdText,
				Title:    newTitle,
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title"}},
		}
		resp, err := service.UpdateSection(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, newTitle, resp.Title)
		require.NotNil(t, resp.GetTextContent())
		assert.Equal(t, "<p>Updated Content</p>", resp.GetTextContent().HtmlContent) // Content from previous test run

		// Verify file
		secData := readSectionDataDirectly(t, tempDir, designId, secIdText)
		require.NotNil(t, secData)
		assert.Equal(t, newTitle, secData.Title)
		assert.Equal(t, "<p>Updated Content</p>", secData.Content)
	})

	// Add more tests: Update non-existent, permission denied, no mask, invalid mask
}

func TestDeleteSection(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ownerId := "delete-sec-owner"
	ctx := testContextWithUser(ownerId)
	designId := "design-delete-sec"
	secA := "del-a"
	secB := "del-b"
	secC := "del-c"

	// Setup Design & Sections
	createDesignDirectly(t, tempDir, Design{
		Id:         designId,
		OwnerId:    ownerId,
		SectionIds: []string{secA, secB, secC},
	})
	createSectionDirectly(t, tempDir, designId, secA, Section{Id: secA, DesignId: designId})
	createSectionDirectly(t, tempDir, designId, secB, Section{Id: secB, DesignId: designId})
	createSectionDirectly(t, tempDir, designId, secC, Section{Id: secC, DesignId: designId})

	t.Run("Delete Existing Section", func(t *testing.T) {
		req := &protos.DeleteSectionRequest{DesignId: designId, SectionId: secB}
		_, err := service.DeleteSection(ctx, req)
		require.NoError(t, err)

		// Verify metadata
		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		assert.Equal(t, []string{secA, secC}, meta.SectionIds)

		// Verify files
		assert.Nil(t, readSectionDataDirectly(t, tempDir, designId, secB))    // Should be gone
		assert.NotNil(t, readSectionDataDirectly(t, tempDir, designId, secA)) // Should exist
		assert.NotNil(t, readSectionDataDirectly(t, tempDir, designId, secC)) // Should exist
	})

	t.Run("Delete Non-Existent Section (Idempotent)", func(t *testing.T) {
		req := &protos.DeleteSectionRequest{DesignId: designId, SectionId: "non-existent"}
		_, err := service.DeleteSection(ctx, req)
		require.NoError(t, err)

		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		assert.Equal(t, []string{secA, secC}, meta.SectionIds) // Should be unchanged from previous test
	})

	// Add tests: Delete from non-existent design, permission denied
}

func TestMoveSection(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ownerId := "move-sec-owner"
	ctx := testContextWithUser(ownerId)
	designId := "design-move-sec"
	secA, secB, secC, secD := "mv-a", "mv-b", "mv-c", "mv-d"

	// Setup Design & Sections
	createDesignDirectly(t, tempDir, Design{
		Id:         designId,
		OwnerId:    ownerId,
		SectionIds: []string{secA, secB, secC, secD},
	})
	// No need to create section files for move test, only metadata matters

	t.Run("Move B After C", func(t *testing.T) {
		req := &protos.MoveSectionRequest{
			DesignId:          designId,
			SectionId:         secB,
			RelativeSectionId: secC,
			Position:          protos.PositionType_POSITION_TYPE_AFTER,
		}
		_, err := service.MoveSection(ctx, req)
		require.NoError(t, err)
		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		assert.Equal(t, []string{secA, secC, secB, secD}, meta.SectionIds)
	})

	t.Run("Move D Before C", func(t *testing.T) {
		// State after previous test: [A, C, B, D]
		req := &protos.MoveSectionRequest{
			DesignId:          designId,
			SectionId:         secD,
			RelativeSectionId: secC,
			Position:          protos.PositionType_POSITION_TYPE_BEFORE,
		}
		_, err := service.MoveSection(ctx, req)
		require.NoError(t, err)
		meta := readDesignMetadata(t, tempDir, designId)
		require.NotNil(t, meta)
		assert.Equal(t, []string{secA, secD, secC, secB}, meta.SectionIds) // A, D, C, B
	})

	// Add more tests: Move non-existent section, non-existent relative, permission denied, invalid position
}
