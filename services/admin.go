package services

import (
	"log"

	// "strings"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	// tspb "google.golang.org/protobuf/types/known/timestamppb"
)

// Just for test

type AdminService struct {
	protos.UnimplementedAdminServiceServer
	BaseService
	clients *ClientMgr
	idgen   *IDGen
}

func NewAdminService(clients *ClientMgr) *AdminService {
	out := &AdminService{
		idgen:   NewIDGen("Scores"),
		clients: clients,
	}
	out.idgen.NextIDFunc = SimpleIDFunc(nil, 5)
	log.Println("IDG: ", out.idgen)
	return out
}
