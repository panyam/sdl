package services

import (
	"log"
	"strings"

	// "cloud.google.com/go/datastore" // Keep if used by other converters

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"google.golang.org/protobuf/types/known/structpb"
	tspb "google.golang.org/protobuf/types/known/timestamppb"
)

func DesignToProto(input *Design) (out *protos.Design) {
	out = &protos.Design{
		CreatedAt:   tspb.New(input.BaseModel.CreatedAt),
		UpdatedAt:   tspb.New(input.BaseModel.UpdatedAt),
		OwnerId:     input.OwnerId,
		VisibleTo:   input.VisibleTo,
		Visibility:  input.Visibility,
		SectionIds:  input.SectionIds,
		Id:          input.Id,
		Name:        input.Name,
		Description: input.Description,
	}
	if input.ContentMetadata.Properties != nil {
		if data, err := structpb.NewStruct(input.ContentMetadata.Properties); err != nil {
			log.Println("Error converting ContentMetadata: ", err)
		} else {
			out.ContentMetadata = data
		}
	}
	return
}

func DesignFromProto(input *protos.Design) (out *Design) {
	out = &Design{
		BaseModel: BaseModel{
			CreatedAt: input.CreatedAt.AsTime(),
			UpdatedAt: input.UpdatedAt.AsTime(),
		},
		Id:          input.Id,
		OwnerId:     input.OwnerId,
		VisibleTo:   input.VisibleTo,
		Visibility:  input.Visibility,
		SectionIds:  input.SectionIds,
		Name:        input.Name,
		Description: input.Description,
	}
	if input.ContentMetadata != nil {
		out.ContentMetadata.Properties = input.ContentMetadata.AsMap()
	}
	return
}

func TagToProto(input *Tag) (out *protos.Tag) {
	out = &protos.Tag{
		CreatedAt:      tspb.New(input.BaseModel.CreatedAt),
		UpdatedAt:      tspb.New(input.BaseModel.UpdatedAt),
		FirstUserId:    input.FirstUserId,
		Name:           input.Name,
		Description:    input.Description,
		NormalizedName: input.NormalizedName,
		ImageUrl:       input.ImageUrl,
		NumDesigns:     input.NumDesigns, // Added num_designs
	}
	return
}

func TagFromProto(input *protos.Tag) (out *Tag) {
	out = &Tag{
		BaseModel: BaseModel{
			CreatedAt: input.CreatedAt.AsTime(),
			UpdatedAt: input.UpdatedAt.AsTime(),
		},
		FirstUserId:    input.FirstUserId,
		Name:           input.Name,
		Description:    input.Description,
		NormalizedName: input.NormalizedName,
		ImageUrl:       input.ImageUrl,
		NumDesigns:     input.NumDesigns, // Added num_designs
	}
	return
}

// SectionToProto converts a services.Section struct (metadata only) to a protos.Section proto.
func SectionToProto(input *Section) (out *protos.Section) {
	if input == nil {
		return nil
	}
	out = &protos.Section{
		CreatedAt:          tspb.New(input.BaseModel.CreatedAt),
		UpdatedAt:          tspb.New(input.BaseModel.UpdatedAt),
		Id:                 input.Id,
		DesignId:           input.DesignId,
		Type:               SectionTypeFromString(input.Type), // Use helper
		Title:              input.Title,
		GetAnswerPrompt:    input.GetAnswerPrompt,    // Directly copy from struct field
		VerifyAnswerPrompt: input.VerifyAnswerPrompt, // Directly copy from struct field
	}
	// Content oneof is intentionally NOT populated.
	return out
}

// SectionFromProto converts a protos.Section proto (metadata only) to a services.Section struct.
func SectionFromProto(input *protos.Section) (out *Section) {
	if input == nil {
		return nil
	}
	out = &Section{
		BaseModel: BaseModel{
			CreatedAt: input.CreatedAt.AsTime(),
			UpdatedAt: input.UpdatedAt.AsTime(),
		},
		Id:       input.Id,
		DesignId: input.DesignId,
		Type:     SectionTypeToString(input.Type), // Use helper
		Title:    input.Title,
		// Order is not stored in the service struct
		// Content-related fields (ContentType, Format) from the proto are ignored here.
		GetAnswerPrompt:    input.GetAnswerPrompt,    // Directly copy from struct field
		VerifyAnswerPrompt: input.VerifyAnswerPrompt, // Directly copy from struct field
	}
	// Content field in the service struct is intentionally left nil/unset.
	return
}

// ContentMetadataToProto converts services.ContentMetadata to protos.Content
func ContentMetadataToProto(input *ContentMetadata) *protos.Content {
	if input == nil {
		return nil
	}
	return &protos.Content{
		CreatedAt: tspb.New(input.CreatedAt),
		UpdatedAt: tspb.New(input.UpdatedAt),
		Name:      input.Name,
		Type:      input.ContentType,
		Format:    input.Format,
		// Size is not in the proto, could be added if needed.
	}
}

// ContentMetadataFromProto converts protos.Content to services.ContentMetadata
func ContentMetadataFromProto(input *protos.Content) *ContentMetadata {
	if input == nil {
		return nil
	}
	return &ContentMetadata{
		BaseModel: BaseModel{
			CreatedAt: input.CreatedAt.AsTime(),
			UpdatedAt: input.UpdatedAt.AsTime(),
		},
		Name:        input.Name,
		ContentType: input.Type,
		Format:      input.Format,
	}
}

// Helper function to map proto enum to string
func SectionTypeToString(st protos.SectionType) string {
	switch st {
	case protos.SectionType_SECTION_TYPE_TEXT:
		return "text"
	case protos.SectionType_SECTION_TYPE_DRAWING:
		return "drawing"
	case protos.SectionType_SECTION_TYPE_PLOT:
		return "plot"
	default:
		return "unspecified"
	}
}

// Helper function to map string to proto enum
func SectionTypeFromString(s string) protos.SectionType {
	val, ok := protos.SectionType_value["SECTION_TYPE_"+strings.ToUpper(s)]
	if ok {
		return protos.SectionType(val)
	}
	return protos.SectionType_SECTION_TYPE_UNSPECIFIED
}
