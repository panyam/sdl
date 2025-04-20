package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// "strings"

	fn "github.com/panyam/goutils/fn"
	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// Just for test
const ENFORCE_LOGIN = false
const FAKE_USER_ID = ""

type DesignService struct {
	protos.UnimplementedDesignServiceServer
	BaseService
	clients *ClientMgr
}

func NewDesignService(clients *ClientMgr) *DesignService {
	out := &DesignService{
		clients: clients,
	}
	return out
}

// Update a new Design
func (s *DesignService) UpdateDesign(ctx context.Context, req *protos.UpdateDesignRequest) (resp *protos.UpdateDesignResponse, err error) {
	log.Println("In design update: ", req)
	loggedInOwnerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		log.Println("Login Failed: ", err)
		return
	}
	resp = &protos.UpdateDesignResponse{}
	design := req.Design
	dsc := s.clients.GetDesignDSClient()
	currnot := Design{}
	err = dsc.GetByID(design.Id, &currnot)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", design.Id))
	}

	if ENFORCE_LOGIN && (loggedInOwnerId == "" || loggedInOwnerId != currnot.OwnerId) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("User '%s' cannot update Design '%s'", loggedInOwnerId, design.Id))
	}

	// Now update things based on mask
	log.Println("Loaded Design from DS: ", currnot, currnot.Visibility)
	update_mask := req.UpdateMask
	has_update_mask := update_mask != nil && len(update_mask.Paths) > 0
	if !has_update_mask {
		return nil, status.Error(codes.InvalidArgument,
			"update_mask should specify (nested) fields to update or provide a content")
	}

	if req.UpdateMask != nil {
		log.Println("Got update mask: ", req.UpdateMask)
		for _, path := range req.UpdateMask.Paths {
			switch path {
			case "visibility", "design.visibility":
				currnot.Visibility = req.Design.Visibility
			case "visible_to", "design.visible_to":
				currnot.VisibleTo = req.Design.VisibleTo
			case "name", "design.name":
				if strings.TrimSpace(req.Design.Name) == "" {
					return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
				}
				currnot.Name = req.Design.Name
			case "description", "design.description":
				currnot.Description = req.Design.Description
			case "content_metadata", "design.content_metadata":
				currnot.ContentMetadata.Properties = nil
				if req.Design.ContentMetadata != nil {
					currnot.ContentMetadata.Properties = req.Design.ContentMetadata.AsMap()
				}
			default:
				log.Println("Invalid Field Path: ", path)
				return nil, status.Errorf(codes.InvalidArgument, "Custom Error - UpdateDesign - update_mask contains invalid path: %s", path)
			}
		}
	}

	currnot.UpdatedAt = time.Now()
	log.Println("Finally Updated Design: ", currnot)
	if _, err = dsc.SaveEntity(&currnot); err != nil {
		slog.Error("error saving design: ", "err", err)
		return
	}

	resp.Design = design
	return resp, err
}

// Create a new Design
func (s *DesignService) CreateDesign(ctx context.Context, req *protos.CreateDesignRequest) (resp *protos.CreateDesignResponse, err error) {
	log.Println("In design creation: ", req)
	design := req.Design
	design.OwnerId, err = s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return
	}
	resp = &protos.CreateDesignResponse{}
	dsc := s.clients.GetDesignDSClient()
	if design.Id != "" {
		// see if it already exists
		dsnot := Design{}
		err := dsc.GetByID(design.Id, &dsnot)
		if err == nil {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Design with id '%s' already exists", design.Id))
		}
	} else {
		// For now enforce ID is provided by user
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("id MUST be provided '%s'"))

		/*
			for {
				var existing Design
				log.Println("IDGEN: ", s.idgen)
				if newid, err := s.idgen.NextID("", time.Time{}); err == nil {
					err := dsc.GetByID(newid.Id, &existing)
					if err == ErrNoSuchEntity {
						// Finally found an id that was not taken by another design
						design.Id = newid.Id
						break
					} else if err != nil {
						log.Println("Error: ", err)
						return nil, err
					}
				} else {
					return nil, err
				}
			}
		*/
	}
	// for indexed tags are user tags
	// TODO - treat x=y tags differently
	design.CreatedAt = tspb.New(time.Now())
	design.UpdatedAt = tspb.New(time.Now())

	title := strings.TrimSpace(design.Name)
	if title == "" {
		resp.FieldErrors["title"] = "Title cannot be empty"
		return
	} else {
		// Now save it
		dbnot := DesignFromProto(design)
		if _, err = dsc.SaveEntity(dbnot); err != nil {
			slog.Error("error saving design: ", "err", err)
		}
	}

	// Now save the contents
	resp.Design = design
	return resp, err
}

// Finds and retrieves designs matching the particular criteria.
func (s *DesignService) ListDesigns(ctx context.Context, req *protos.ListDesignsRequest) (resp *protos.ListDesignsResponse, err error) {
	if req.Pagination == nil {
		req.Pagination = &protos.Pagination{}
	}
	if req.Pagination.PageSize <= 0 {
		req.Pagination.PageSize = 200
	}
	query := s.clients.GetDesignDSClient().NewQuery().Offset(int(req.Pagination.PageOffset)).Limit(int(req.Pagination.PageSize) + 1)
	if req.OwnerId != "" {
		query = query.FilterField("userId", "=", req.OwnerId)
	}
	if req.LimitToPublic {
		query = query.FilterField("visibility", "=", "public")
	}

	if req.OrderBy == "recent" {
		query = query.Order("-updatedAt")
	}
	results, err := s.clients.GetDesignDSClient().Select(query)
	if err != nil {
		return nil, err
	}
	resp = &protos.ListDesignsResponse{
		Pagination: &protos.PaginationResponse{
			HasMore: len(results) > int(req.Pagination.PageSize),
		},
	}
	if resp.Pagination.HasMore {
		results = results[:req.Pagination.PageSize]
	}
	resp.Designs = fn.Map(results, DesignToProto)
	return
}

func (s *DesignService) GetDesign(ctx context.Context, req *protos.GetDesignRequest) (resp *protos.GetDesignResponse, err error) {
	dsc := s.clients.GetDesignDSClient()
	var design Design
	err = dsc.GetByID(req.Id, &design)
	if err != nil {
		if err == ErrNoSuchEntity {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", req.Id))
		} else {
			return nil, err
		}
	}
	resp = &protos.GetDesignResponse{
		Design: DesignToProto(&design),
	}
	return
}

func (s *DesignService) GetDesigns(ctx context.Context, req *protos.GetDesignsRequest) (resp *protos.GetDesignsResponse, err error) {
	return
}

// Deletes an design from our system.
func (s *DesignService) DeleteDesign(ctx context.Context, req *protos.DeleteDesignRequest) (resp *protos.DeleteDesignResponse, err error) {
	loggedInOwnerId, err := s.EnsureLoggedIn(ctx)
	if ENFORCE_LOGIN && err != nil {
		return
	}

	resp = &protos.DeleteDesignResponse{}
	dsc := s.clients.GetDesignDSClient()
	currnot := Design{}
	err = dsc.GetByID(req.Id, &currnot)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Design with id '%s' not found", req.Id))
	}
	log.Println("LoggedInUser: ", loggedInOwnerId)
	log.Println("DesignUser: ", currnot.OwnerId)
	if ENFORCE_LOGIN && (loggedInOwnerId == "" || loggedInOwnerId != currnot.OwnerId) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("User '%s' cannot delete Design '%s'", loggedInOwnerId, req.Id))
	}

	// delete content first
	err = dsc.DeleteByKey(req.Id)
	if err != nil {
		return
	}
	return
}
