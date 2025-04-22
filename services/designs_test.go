package services

import (
	"context"
	"encoding/json"
	"math/rand"
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
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// ****** ADDED IMPORT ******
	"github.com/google/go-cmp/cmp" // Correct import for cmp.Equal/Diff
)

// Helper to create a DesignService instance using a temporary directory
func newTestDesignService(t *testing.T) (*DesignService, string) {
	t.Helper()

	tempDir := t.TempDir()
	t.Logf("Using temp directory for test service: %s", tempDir)

	service := &DesignService{
		basePath:        tempDir,
		idLength:        defaultGeneratedIDLength,
		idChars:         defaultIDChars,
		maxIDGenRetries: defaultMaxIDGenerationRetries,
		randSource:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	err := ensureDir(service.basePath)
	require.NoError(t, err, "Failed to create test base directory")

	t.Cleanup(func() {
		service.mutexMap = sync.Map{}
	})

	return service, tempDir
}

// Helper to directly create a design metadata file for test setup
func createDesignDirectly(t *testing.T, basePath string, metadata DesignMetadata) {
	t.Helper()
	designPath := filepath.Join(basePath, metadata.ID)
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
func readDesignMetadata(t *testing.T, basePath string, designId string) *DesignMetadata {
	t.Helper()
	metadataPath := filepath.Join(basePath, designId, "design.json")
	jsonData, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		require.NoError(t, err, "Failed to read metadata file for verification")
	}

	var metadata DesignMetadata
	err = json.Unmarshal(jsonData, &metadata)
	require.NoError(t, err, "Failed to unmarshal metadata for verification")
	return &metadata
}

// --- Test Cases ---

func TestCreateDesign(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := context.Background()

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
		assert.NotNil(t, resp.Design.CreatedAt)
		assert.NotNil(t, resp.Design.UpdatedAt)

		meta := readDesignMetadata(t, tempDir, "test-create-1")
		require.NotNil(t, meta)
		assert.Equal(t, "test-create-1", meta.ID)
		assert.Equal(t, "Test Create Design 1", meta.Name)
		assert.Equal(t, "public", meta.Visibility)
		assert.Equal(t, resp.Design.OwnerId, meta.OwnerID)
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
		assert.Len(t, resp.Design.Id, service.idLength)
		assert.Regexp(t, `^[a-zA-Z0-9]+$`, resp.Design.Id)
		assert.Equal(t, "Test Auto ID", resp.Design.Name)
		assert.Equal(t, "private", resp.Design.Visibility)

		generatedId := resp.Design.Id
		meta := readDesignMetadata(t, tempDir, generatedId)
		require.NotNil(t, meta)
		assert.Equal(t, generatedId, meta.ID)
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
		// ****** FIXED: Removed unused resp variable ******
		_, err := service.CreateDesign(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "name cannot be empty")
	})

	t.Run("Fail on Already Exists (Provided ID)", func(t *testing.T) {
		createDesignDirectly(t, tempDir, DesignMetadata{ID: "test-conflict", Name: "Existing"})

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
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	setupMeta := DesignMetadata{
		ID:          "test-get-1",
		OwnerID:     "test-owner",
		Name:        "Test Get Design",
		Description: "Get Description",
		Visibility:  "private",
		VisibleTo:   []string{"user1", "user2"},
		CreatedAt:   now.Add(-1 * time.Hour),
		UpdatedAt:   now,
		Sections:    []SectionMeta{},
	}
	createDesignDirectly(t, tempDir, setupMeta)

	t.Run("Get Existing Design", func(t *testing.T) {
		req := &protos.GetDesignRequest{Id: "test-get-1"}
		resp, err := service.GetDesign(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Design)

		expectedProto := &protos.Design{
			Id:          setupMeta.ID,
			OwnerId:     setupMeta.OwnerID,
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
	})
}

func TestListDesigns(t *testing.T) {
	service, tempDir := newTestDesignService(t)
	ctx := context.Background()

	ts := time.Now().UTC().Truncate(time.Second)
	meta1 := DesignMetadata{ID: "list-aaa", OwnerID: "user1", Name: "AAA Design", Visibility: "public", UpdatedAt: ts.Add(-2 * time.Hour)}
	meta2 := DesignMetadata{ID: "list-ccc", OwnerID: "user2", Name: "CCC Design", Visibility: "private", UpdatedAt: ts}
	meta3 := DesignMetadata{ID: "list-bbb", OwnerID: "user1", Name: "BBB Design", Visibility: "public", UpdatedAt: ts.Add(-1 * time.Hour)}
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
	ctx := context.Background()

	setupTime := time.Now().UTC().Truncate(time.Second)
	meta := DesignMetadata{ID: "update-1", OwnerID: "updater", Name: "Original Name", Description: "Original Desc", Visibility: "private", CreatedAt: setupTime, UpdatedAt: setupTime}
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
		// ****** FIXED: Removed unused resp variable ******
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
	ctx := context.Background()

	meta := DesignMetadata{ID: "delete-1", OwnerID: "deleter", Name: "To Be Deleted", CreatedAt: time.Now(), UpdatedAt: time.Now()}
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
	ctx := context.Background()
	req := &protos.GetDesignsRequest{Ids: []string{"some-id"}}

	_, err := service.GetDesigns(ctx, req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
}
