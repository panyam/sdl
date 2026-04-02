package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	v1services "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/panyam/sdl/services/devenvbe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address          string
	WorkspaceService *devenvbe.WorkspaceService
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	if s.WorkspaceService == nil {
		return fmt.Errorf("WorkspaceService is required")
	}

	server := grpc.NewServer()

	// Register WorkspaceService (replaces old CanvasService)
	v1services.RegisterWorkspaceServiceServer(server, s.WorkspaceService)

	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		slog.Error("error in listening on port", "port", s.Address, "err", err)
		return fmt.Errorf("failed to listen on %s: %w", s.Address, err)
	}

	slog.Info("Starting grpc endpoint on: ", "addr", s.Address)
	reflection.Register(server)

	go func() {
		if err := server.Serve(l); err != nil && err != grpc.ErrServerStopped {
			log.Printf("grpc server failed to serve: %v", err)
			srvErr <- err
		}
	}()

	go func() {
		<-srvChan
		slog.Info("Shutting down gRPC server...")
		server.GracefulStop()
		slog.Info("gRPC server stopped.")
	}()

	return nil
}
