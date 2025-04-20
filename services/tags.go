package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	tspb "google.golang.org/protobuf/types/known/timestamppb"

	// "strings"

	fn "github.com/panyam/goutils/fn"
	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}
	return output
}

type TagService struct {
	protos.UnimplementedTagServiceServer
	BaseService
	clients *ClientMgr
}

func NewTagService(clients *ClientMgr) *TagService {
	return &TagService{
		clients: clients,
	}
}

// Create a new Tag
func (s *TagService) CreateTag(ctx context.Context, req *protos.CreateTagRequest) (resp *protos.CreateTagResponse, err error) {
	tag := req.Tag
	tag.FirstUserId, err = s.EnsureLoggedIn(ctx)
	if err != nil {
		return
	}
	resp = &protos.CreateTagResponse{}
	dsc := s.clients.GetTagDSClient()
	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		err = status.Error(codes.InvalidArgument, fmt.Sprintf("A tag name MUST be specified"))
	}

	// see if it already exists
	dsnot := Tag{}
	err = dsc.GetByID(tag.Name, &dsnot)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Tag with id '%s' already exists", tag.Name))
	}

	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		resp.FieldErrors["name"] = "Tag name cannot be empty"
		return nil, status.Error(codes.InvalidArgument, "Tag name cannot be empty")
	}

	tag.CreatedAt = tspb.New(time.Now())
	tag.UpdatedAt = tspb.New(time.Now())
	tag.NormalizedName = strings.ToLower(removeAccents(tag.Name))

	// Now save it
	dbnot := TagFromProto(tag)
	if _, err = dsc.SaveEntity(dbnot); err != nil {
		slog.Error("error saving tag: ", "err", err)
	}
	return resp, err
}

// Finds and retrieves tags matching the particular criteria.
func (s *TagService) ListTags(ctx context.Context, req *protos.ListTagsRequest) (resp *protos.ListTagsResponse, err error) {
	slog.Info("listing tags: ", "req: ", req)
	query := s.clients.GetTagDSClient().NewQuery().Offset(0).Limit(200)
	if req.UserId != "" {
		query = query.FilterField("userId", "=", req.UserId)
	}

	if req.OrderBy == "recent" {
		query = query.Order("-updatedAt")
	} else {
		query = query.Order("normalizedName")
	}
	results, err := s.clients.GetTagDSClient().Select(query)
	if err != nil {
		return nil, err
	}
	resp = &protos.ListTagsResponse{
		Tags: fn.Map(results, TagToProto),
	}
	return
}

func (s *TagService) GetTag(ctx context.Context, req *protos.GetTagRequest) (resp *protos.GetTagResponse, err error) {
	dsc := s.clients.GetTagDSClient()
	var tag Tag
	err = dsc.GetByID(req.Id, &tag)
	if err != nil {
		if err == ErrNoSuchEntity {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Tag with id '%s' not found", tag.Name))
		} else {
			return nil, err
		}
	}
	resp = &protos.GetTagResponse{
		Tag: TagToProto(&tag),
	}
	return
}

func (s *TagService) GetTags(ctx context.Context, req *protos.GetTagsRequest) (resp *protos.GetTagsResponse, err error) {
	return
}

// Update a new Tag
func (s *TagService) UpdateTag(ctx context.Context, req *protos.UpdateTagRequest) (resp *protos.UpdateTagResponse, err error) {
	_, err = s.EnsureLoggedIn(ctx)
	if err != nil {
		return
	}
	return resp, err
}

// Deletes an tag from our system.
func (s *TagService) DeleteTag(ctx context.Context, req *protos.DeleteTagRequest) (resp *protos.DeleteTagResponse, err error) {
	_, err = s.EnsureLoggedIn(ctx)
	if err == nil {
		resp = &protos.DeleteTagResponse{}
	}
	return
}
