package services

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BaseService struct {
}

func (b *BaseService) GetLoggedInUser(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Println("Ok, MD: ", ok, md)
		log.Println("Logged in user: ", md["loggedinuserid"])
		if t, ok := md["loggedinuserid"]; ok {
			return t[0]
		}
	}
	return ""
}

func (b *BaseService) EnsureLoggedIn(ctx context.Context) (loggedInUserId string, err error) {
	loggedInUserId = b.GetLoggedInUser(ctx)
	if loggedInUserId == "" {
		err = status.Error(codes.Unauthenticated, fmt.Sprintf("Login first"))
	}
	return
}
