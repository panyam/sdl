package services

import (
	"log"

	"cloud.google.com/go/datastore"
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

func EntityToStruct(input *datastore.Entity, output **structpb.Struct) {
	if len(input.Properties) > 0 {
		propmap := map[string]any{}
		for _, prop := range input.Properties {
			propmap[prop.Name] = prop.Value
		}
		if data, err := structpb.NewStruct(propmap); err != nil {
			log.Println("Error converting ContentMetadata: ", err)
		} else {
			*output = data
		}
	}
}

func StructToEntity(input *structpb.Struct, output *datastore.Entity) {
	if input != nil {
		asmap := input.AsMap()
		for key, value := range asmap {
			output.Properties = append(output.Properties, datastore.Property{
				Name:  key,
				Value: value,
			})
		}
	}
}
