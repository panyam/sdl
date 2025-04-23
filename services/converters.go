package services

import (
	"log"
	"strings"

	// "cloud.google.com/go/datastore"

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
	// EntityToStruct(&input.ContentMetadata, &out.ContentMetadata)
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
	// StructToEntity(input.ContentMetadata, &out.ContentMetadata)
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
	}
	return
}

// SectionToProto converts a services.Section struct to a protos.Section proto.
// Note: Order is typically calculated contextually and added after this conversion.
func SectionToProto(input *Section) (out *protos.Section) {
	out = &protos.Section{
		CreatedAt:   tspb.New(input.BaseModel.CreatedAt),
		UpdatedAt:   tspb.New(input.BaseModel.UpdatedAt),
		Id:          input.Id,
		DesignId:    input.DesignId,
		Type:        SectionTypeFromString(input.Type), // Needs mapping
		Title:       input.Title,
		ContentType: input.ContentType,
		Format:      input.Format,
		// Order: is set contextually based on design.SectionIds index
	}

	// Handle content based on type
	switch out.Type {
	case protos.SectionType_SECTION_TYPE_TEXT:
		if contentStr, ok := input.Content.(string); ok {
			out.Content = &protos.Section_TextContent{
				TextContent: &protos.TextSectionContent{HtmlContent: contentStr},
			}
		}
	case protos.SectionType_SECTION_TYPE_DRAWING:
		// Assuming content is stored as stringified JSON for drawing/plot
		contentBytes, ok := input.Content.([]byte)
		if !ok {
			// Log an error if the content isn't []byte for a drawing type
			log.Printf("Error: Expected []byte content for drawing section %s, but got %T", input.Id, input.Content)
		} else {
			out.Content = &protos.Section_DrawingContent{
				DrawingContent: &protos.DrawingSectionContent{Data: contentBytes},
			}
		}
	case protos.SectionType_SECTION_TYPE_PLOT:
		contentBytes, ok := input.Content.([]byte)
		if !ok {
			// Log an error if the content isn't []byte for a plot type
			log.Printf("Error: Expected []byte content for plot section %s, but got %T", input.Id, input.Content)
		} else {
			out.Content = &protos.Section_PlotContent{
				PlotContent: &protos.PlotSectionContent{Data: contentBytes},
			}
		}
	}
	return out
}

// SectionFromProto converts a protos.Section proto to a services.Section struct.
func SectionFromProto(input *protos.Section) (out *Section) {
	out = &Section{
		BaseModel: BaseModel{
			CreatedAt: input.CreatedAt.AsTime(),
			UpdatedAt: input.UpdatedAt.AsTime(),
		},
		Id:          input.Id,
		DesignId:    input.DesignId,
		Type:        SectionTypeToString(input.Type), // Needs mapping
		Title:       input.Title,
		ContentType: input.ContentType,
		Format:      input.Format,
		// Order is not stored in the struct
	}

	// Extract content from the oneof field
	switch c := input.Content.(type) {
	case *protos.Section_TextContent:
		out.Content = c.TextContent.GetHtmlContent() // Store string
	case *protos.Section_DrawingContent:
		out.Content = c.DrawingContent.GetData() // Store []byte
	case *protos.Section_PlotContent:
		out.Content = c.PlotContent.GetData() // Store []byte
	default:
		// Handle case where content is nil or unknown type
		log.Printf("Warning: Unknown or nil content type (%T) for section %s", c, input.Id)
		out.Content = nil // Or perhaps empty string/byte slice depending on expected behavior
	}
	return
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
