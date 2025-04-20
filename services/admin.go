package services

import (

	// "strings"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// Just for test

type AdminService struct {
	protos.UnimplementedAdminServiceServer
	BaseService
	clients *ClientMgr
}

func NewAdminService(clients *ClientMgr) *AdminService {
	out := &AdminService{
		clients: clients,
	}
	return out
}
