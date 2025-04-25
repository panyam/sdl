package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	v1 "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"github.com/panyam/leetcoach/services/llm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Address string
}

func (s *Server) Start(ctx context.Context, srvErr chan error, srvChan chan bool) error {
	clients := NewClientMgr(s.Address) // Assuming ClientMgr is still needed, maybe for other services

	// --- Create LLM Client ---
	llmClient, err := llm.NewOpenAIClient() // Use the factory function
	if err != nil {
		// If LLM is critical, fail startup. If optional, log and continue.
		// For now, let's treat it as critical.
		slog.Error("Failed to initialize LLM client", "error", err)
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}
	slog.Info("LLM Client initialized successfully")
	// -------------------------

	// Instantiate services
	designSvc := NewDesignService(clients, "") // Provide base path if needed, empty uses default
	contentSvc := NewContentService(designSvc) // ContentService needs DesignService (for paths/locks)
	tagSvc := NewTagService(clients)           // Instantiate TagService
	llmSvc := NewLlmService(llmClient, designSvc, contentSvc)

	server := grpc.NewServer(
	// grpc.UnaryInterceptor(EnsureAccessToken), // Add interceptors if needed
	)

	// Register services
	v1.RegisterContentServiceServer(server, contentSvc) // Register ContentService
	v1.RegisterDesignServiceServer(server, designSvc)
	v1.RegisterTagServiceServer(server, tagSvc) // Register TagService
	v1.RegisterLlmServiceServer(server, llmSvc) // *** Register LlmService ***

	if os.Getenv("LEETCOACH_ENV") == "dev" {
		adminSvc := NewAdminService(clients) // Instantiate AdminService
		v1.RegisterAdminServiceServer(server, adminSvc)
	}

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
