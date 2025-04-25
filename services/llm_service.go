// FILE: ./services/llm_service.go
package services

import (
	"context"
	"log/slog"

	protos "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	"github.com/panyam/leetcoach/services/llm" // Import the new llm package
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LlmService implements the gRPC service for LLM interactions.
type LlmService struct {
	protos.UnimplementedLlmServiceServer
	BaseService               // Inherit base service helpers if needed (like EnsureLoggedIn)
	llmClient   llm.LLMClient // Use the interface for testability
	// Add other dependencies like DesignService/ContentService later if needed
}

// NewLlmService creates a new instance of the LlmService.
func NewLlmService(client llm.LLMClient) *LlmService {
	if client == nil {
		// We could panic or return an error, but for now let's assume
		// the caller handles client creation properly.
		// In a real app, dependency injection framework would manage this.
		slog.Warn("NewLlmService created with nil LLMClient")
	}
	return &LlmService{
		llmClient: client,
	}
}

// SimpleLlmQuery handles the basic prompt request.
func (s *LlmService) SimpleLlmQuery(ctx context.Context, req *protos.SimpleLlmQueryRequest) (*protos.SimpleLlmQueryResponse, error) {
	// Optional: Check login status if needed for simple queries
	// _, err := s.EnsureLoggedIn(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	prompt := req.GetPrompt()
	if prompt == "" {
		return nil, status.Error(codes.InvalidArgument, "Prompt cannot be empty")
	}

	slog.Info("Handling SimpleLlmQuery", "design_id", req.GetDesignId(), "section_id", req.GetSectionId())

	if s.llmClient == nil {
		slog.Error("LLMClient is not initialized in LlmService")
		return nil, status.Error(codes.Internal, "LLM service not configured")
	}

	// Call the LLM client
	responseText, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		// Log the specific error from the client
		slog.Error("LLM query failed", "error", err)
		// Return a generic internal error to the user for now
		// We might map specific LLM errors (like rate limits) later
		return nil, status.Error(codes.Internal, "Failed to process LLM query")
	}

	resp := &protos.SimpleLlmQueryResponse{
		ResponseText: responseText,
	}

	slog.Info("SimpleLlmQuery successful")
	return resp, nil
}
