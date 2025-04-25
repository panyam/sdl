// FILE: ./services/llm_service.go
package services

import (
	"context"
	"fmt"
	"log/slog"

	"strings"

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

// SuggestSections generates section suggestions based on existing titles.
func (s *LlmService) SuggestSections(ctx context.Context, req *protos.SuggestSectionsRequest) (*protos.SuggestSectionsResponse, error) {
	designId := req.GetDesignId()
	existingTitles := req.GetExistingSectionTitles()
	slog.Info("Handling SuggestSections", "design_id", designId, "existing_titles_count", len(existingTitles))

	if s.llmClient == nil {
		slog.Error("LLMClient is not initialized in LlmService for SuggestSections")
		return nil, status.Error(codes.Internal, "LLM service not configured")
	}

	// Construct the prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are a helpful assistant for system design interviews.")
	promptBuilder.WriteString("Given the following existing section titles in a system design document:\n")
	if len(existingTitles) > 0 {
		for _, title := range existingTitles {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", title))
		}
	} else {
		promptBuilder.WriteString("(No sections added yet)\n")
	}
	promptBuilder.WriteString("\nSuggest 3 to 5 relevant sections to add next. For each suggestion, provide a concise Title, a Type (must be one of: text, drawing, plot), and a brief Description (1 sentence max).\n")
	promptBuilder.WriteString("Format each suggestion exactly like this, separated by '---':\n")
	promptBuilder.WriteString("Title: <Suggested Title>\n")
	promptBuilder.WriteString("Type: <text|drawing|plot>\n")
	promptBuilder.WriteString("Description: <Brief explanation>\n")
	promptBuilder.WriteString("---\n") // Separator
	promptBuilder.WriteString("Title: ...\nType: ...\nDescription: ...\n")

	prompt := promptBuilder.String()

	// Call the LLM
	llmResponse, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		slog.Error("LLM query failed for SuggestSections", "error", err)
		return nil, status.Error(codes.Internal, "Failed to get suggestions from LLM")
	}

	// Parse the LLM response
	suggestions := parseSuggestions(llmResponse)
	if len(suggestions) == 0 {
		slog.Warn("LLM returned no parsable suggestions", "raw_response", llmResponse)
		// Return empty list, maybe client can show a message
	}

	resp := &protos.SuggestSectionsResponse{
		Suggestions: suggestions,
	}
	slog.Info("SuggestSections successful", "suggestion_count", len(resp.Suggestions))
	return resp, nil
}

// Helper to parse the structured LLM response
func parseSuggestions(rawResponse string) []*protos.SuggestedSection {
	var suggestions []*protos.SuggestedSection
	parts := strings.Split(strings.TrimSpace(rawResponse), "---")
	for _, part := range parts {
		suggestion := &protos.SuggestedSection{}
		lines := strings.Split(strings.TrimSpace(part), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Title:") {
				suggestion.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			} else if strings.HasPrefix(line, "Type:") {
				// Basic validation for type
				typ := strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
				if typ == "text" || typ == "drawing" || typ == "plot" {
					suggestion.Type = typ
				} else {
					suggestion.Type = "text" // Default if invalid
					slog.Warn("LLM suggested invalid type, defaulting to text", "invalid_type", typ)
				}
			} else if strings.HasPrefix(line, "Description:") {
				suggestion.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			}
		}
		// Only add if we got a title and type
		if suggestion.Title != "" && suggestion.Type != "" {
			suggestions = append(suggestions, suggestion)
		}
	}
	return suggestions
}
