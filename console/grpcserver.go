package console

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address       string
	CanvasService *CanvasService
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	// Use provided CanvasService or create a new one
	canvasSvc := s.CanvasService
	if canvasSvc == nil {
		canvasSvc = NewCanvasService() // Provide base path if needed, empty uses default
	}

	server := grpc.NewServer(
	// grpc.UnaryInterceptor(EnsureAccessToken), // Add interceptors if needed
	)

	// Register services
	v1.RegisterCanvasServiceServer(server, canvasSvc)

	l, err := net.Listen("tcp", s.Address)
	if err != nil {
		slog.Error("error in listening on port", "port", s.Address, "err", err)
		// Consider returning the error instead of fatal/panic in Start method
		return fmt.Errorf("failed to listen on %s: %w", s.Address, err)
	}

	// the gRPC server
	slog.Info("Starting grpc endpoint on: ", "addr", s.Address)
	reflection.Register(server)

	// Run server in a goroutine to allow graceful shutdown
	go func() {
		if err := server.Serve(l); err != nil && err != grpc.ErrServerStopped {
			log.Printf("grpc server failed to serve: %v", err)
			srvErr <- err // Send error to the main app routine
		}
	}()

	// Handle shutdown signal
	go func() {
		<-srvChan // Wait for shutdown signal from main app
		slog.Info("Shutting down gRPC server...")
		server.GracefulStop()
		slog.Info("gRPC server stopped.")
	}()

	return nil // Indicate successful start
}
