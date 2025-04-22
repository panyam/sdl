package services

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	v1 "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address string
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	clients := NewClientMgr(s.Address)
	server := grpc.NewServer(
	// grpc.UnaryInterceptor(EnsureAccessToken),
	)
	v1.RegisterDesignServiceServer(server, NewDesignService(clients, ""))
	v1.RegisterTagServiceServer(server, NewTagService(clients))
	if os.Getenv("LEETCOACH_ENV") == "dev" {
		v1.RegisterAdminServiceServer(server, NewAdminService(clients))
	}
	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		slog.Error("error in listening on port", "port", s.Address, "err", err)
	} else {
		// the gRPC server
		slog.Info("Starting grpc endpoint on: ", "addr", s.Address)
		reflection.Register(server)
		err = server.Serve(l)
		if err != nil {
			log.Fatal("Unable to start grpc server", err)
		}
	}
	return err
}
