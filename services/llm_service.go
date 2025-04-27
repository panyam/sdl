// FILE: ./services/llm_service.go
package services

import (
	"context"
	"errors"
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
	BaseService                 // Inherit base service helpers if needed (like EnsureLoggedIn)
	llmClient   llm.LLMClient   // Use the interface for testability
	designSvc   *DesignService  // Need DesignService for titles/metadata
	contentSvc  *ContentService // Need ContentService for reading/writing content
	// Add other dependencies like DesignService/ContentService later if needed
}

// NewLlmService creates a new instance of the LlmService.
func NewLlmService(client llm.LLMClient, designSvc *DesignService, contentSvc *ContentService) *LlmService {
	if client == nil {
		slog.Warn("NewLlmService created with nil LLMClient")
	}
	if designSvc == nil {
		slog.Warn("NewLlmService created with nil DesignService")
	}
	if contentSvc == nil {
		slog.Warn("NewLlmService created with nil ContentService")
	}
	return &LlmService{
		llmClient:  client,
		designSvc:  designSvc,
		contentSvc: contentSvc,
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
	promptBuilder.WriteString("You are a helpful assistant for system design interviews.\n")
	promptBuilder.WriteString("Given the following existing section titles in a system design document:\n")
	if len(existingTitles) > 0 {
		for _, title := range existingTitles {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", title))
		}
	} else {
		promptBuilder.WriteString("(No sections added yet)\n")
	}
	promptBuilder.WriteString("\nSuggest 3 to 5 relevant sections to add next. For each suggestion, provide:\n")
	promptBuilder.WriteString("1. A concise Title.\n")
	promptBuilder.WriteString("2. A Type (must be one of: text, drawing, plot).\n")
	promptBuilder.WriteString("3. A brief Description (1 sentence max).\n")
	promptBuilder.WriteString("4. A Get Answer Prompt: A detailed prompt the user can later give an LLM to generate the content for this section.\n")
	promptBuilder.WriteString("5. A Verify Prompt: A detailed prompt the user can later give an LLM to review the content they wrote for this section.\n\n")
	promptBuilder.WriteString("Format each suggestion exactly like this, separated by '---':\n")
	promptBuilder.WriteString("Title: <Suggested Title>\n")
	promptBuilder.WriteString("Type: <text|drawing|plot>\n")
	promptBuilder.WriteString("Get Answer Prompt: <Prompt text for generation>\n")
	promptBuilder.WriteString("Verify Prompt: <Prompt text for verification>\n")
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
	for part := range strings.SplitSeq(strings.TrimSpace(rawResponse), "---") {
		suggestion := &protos.SuggestedSection{}
		for line := range strings.SplitSeq(strings.TrimSpace(part), "\n") {
			if strings.HasPrefix(line, "Title:") {
				suggestion.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			} else if strings.HasPrefix(line, "Get Answer Prompt:") { // Expecting these prefixes from LLM
				suggestion.GetAnswerPrompt = strings.TrimSpace(strings.TrimPrefix(line, "Get Answer Prompt:"))
			} else if strings.HasPrefix(line, "Type:") {
				// Basic validation for type
				typ := strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
				if typ == "text" || typ == "drawing" || typ == "plot" {
					suggestion.Type = typ
				} else {
					suggestion.Type = "text" // Default if invalid
					slog.Warn("LLM suggested invalid type, defaulting to text", "invalid_type", typ)
				}
			} else if strings.HasPrefix(line, "Verify Prompt:") { // Expecting these prefixes from LLM
				suggestion.VerifyAnswerPrompt = strings.TrimSpace(strings.TrimPrefix(line, "Verify Prompt:"))
			} else if strings.HasPrefix(line, "Description:") {
				suggestion.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			}
		}
		// Only add if we got a title and type (prompts are optional but requested)
		if suggestion.Title != "" && suggestion.Type != "" {
			suggestions = append(suggestions, suggestion)
		}
	}
	return suggestions
}

// GenerateTextContent handles requests to generate content for a text section.
func (s *LlmService) GenerateTextContent(ctx context.Context, req *protos.GenerateTextContentRequest) (*protos.GenerateTextContentResponse, error) {
	designId := req.GetDesignId()
	sectionId := req.GetSectionId()
	slog.Info("Handling GenerateTextContent", "design_id", designId, "section_id", sectionId)

	if s.llmClient == nil || s.designSvc == nil {
		slog.Error("LLM or Design service dependency is nil in GenerateTextContent")
		return nil, status.Error(codes.Internal, "LLM service dependencies not configured")
	}

	// TODO: Add permission check (can user access/edit this design?)

	// Get section metadata (mainly for title)
	sectionData, err := s.designSvc.readSectionData(designId, sectionId) // Use internal read method
	if err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Warn("Section not found for GenerateTextContent", "design_id", designId, "section_id", sectionId)
			return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
		}
		slog.Error("Failed to read section data for GenerateTextContent", "design_id", designId, "section_id", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read section metadata")
	}

	if sectionData.Type != "text" {
		return nil, status.Errorf(codes.InvalidArgument, "GenerateTextContent only supports 'text' sections, found '%s'", sectionData.Type)
	}

	// --- Read prompt from file ---
	prompt, err := s.designSvc.readPromptFile(designId, sectionId, "get_answer.md")
	if err != nil && !errors.Is(err, ErrNoSuchEntity) {
		// Log error reading but attempt fallback
		slog.Error("Failed to read get_answer prompt file, using default", "design_id", designId, "section_id", sectionId, "error", err)
	}

	if err != nil || prompt == "" { // If file not found OR read error OR file empty
		slog.Info("Get answer prompt file not found or empty, using default.", "design_id", designId, "section_id", sectionId)
		// Default prompt using the section title
		prompt = fmt.Sprintf("Generate concise HTML content for a system design document section titled '%s'. Focus on key concepts, potential trade-offs, and common patterns related to this topic. ONLY include relevant HTML tags like <p>, <ul>, <li>, <strong>, <h2>, <h3>. Do NOT include <html>, <head>, <body>, or <style> tags. Start the content directly, for example, with an <p> or a <h3> tag.", sectionData.Title)
	} else {
		slog.Debug("Using get_answer prompt read from file.", "design_id", designId, "section_id", sectionId)
	}

	// Call LLM
	generatedText, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		slog.Error("LLM query failed for GenerateTextContent", "error", err)
		return nil, status.Error(codes.Internal, "Failed to generate content via LLM")
	}

	resp := &protos.GenerateTextContentResponse{
		GeneratedText: generatedText,
	}
	slog.Info("GenerateTextContent successful")
	return resp, nil
}

// ReviewTextContent handles requests to review existing text content.
func (s *LlmService) ReviewTextContent(ctx context.Context, req *protos.ReviewTextContentRequest) (*protos.ReviewTextContentResponse, error) {
	designId := req.GetDesignId()
	sectionId := req.GetSectionId()
	slog.Info("Handling ReviewTextContent", "design_id", designId, "section_id", sectionId)

	if s.llmClient == nil || s.contentSvc == nil || s.designSvc == nil {
		slog.Error("LLM, Content, or Design service dependency is nil in ReviewTextContent")
		return nil, status.Error(codes.Internal, "LLM service dependencies not configured")
	}

	// TODO: Add permission check

	// Get section metadata (for type check) - could combine with below read if needed
	sectionData, err := s.designSvc.readSectionData(designId, sectionId)
	if err != nil { // Handles NotFound
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Warn("Section not found for ReviewTextContent", "design_id", designId, "section_id", sectionId)
			return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
		}
		slog.Error("Failed to read section data for ReviewTextContent", "design_id", designId, "section_id", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read section metadata")
	}
	if sectionData.Type != "text" {
		return nil, status.Errorf(codes.InvalidArgument, "ReviewTextContent only supports 'text' sections, found '%s'", sectionData.Type)
	}

	// Get existing content
	contentBytes, err := s.contentSvc.GetContentBytes(ctx, designId, sectionId, "main")
	if err != nil && !errors.Is(err, ErrNoSuchEntity) { // Allow review even if content file doesn't exist yet
		slog.Error("Failed to read section content for ReviewTextContent", "design_id", designId, "section_id", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read existing section content")
	}
	existingContent := string(contentBytes) // Will be empty if file didn't exist

	var prompt string
	verifyPromptInstruction, err := s.designSvc.readPromptFile(designId, sectionId, "verify.md")
	if err == nil && verifyPromptInstruction != "" {
		slog.Debug("Using verify prompt read from file.", "design_id", designId, "section_id", sectionId)
		// Use the prompt from the file as instructions
		prompt = fmt.Sprintf("You are a senior software engineer reviewing a system design document section titled '%s'. Please review the following content based on the specific instructions provided below.\n\nSection Content:\n---\n%s\n---\n\nReview Instructions:\n%s\n\nReview:", sectionData.Title, existingContent, verifyPromptInstruction)
	} else {
		if !errors.Is(err, ErrNoSuchEntity) {
			// Log error reading file, but proceed with default
			slog.Error("Failed to read verify prompt file, using default", "design_id", designId, "section_id", sectionId, "error", err)
		} else {
			slog.Info("Verify prompt file not found or empty, using default.", "design_id", designId, "section_id", sectionId)
		}
		// Fallback to default review logic
		prompt = fmt.Sprintf("You are a senior software engineer reviewing a system design document section titled '%s'. Please review the following content for clarity, completeness, technical accuracy, potential missed edge cases, and overall quality. Provide constructive feedback.\n\nSection Content:\n---\n%s\n---\n\nReview:", sectionData.Title, existingContent)
		if existingContent == "" {
			prompt = fmt.Sprintf("The system design section titled '%s' is currently empty. What key points, trade-offs, or common patterns should be included in this section?", sectionData.Title)
		}
	}

	// Call LLM
	reviewText, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		slog.Error("LLM query failed for ReviewTextContent", "error", err)
		return nil, status.Error(codes.Internal, "Failed to get review via LLM")
	}

	resp := &protos.ReviewTextContentResponse{
		ReviewText: reviewText,
	}
	slog.Info("ReviewTextContent successful")
	return resp, nil
}

// GenerateDefaultPrompts generates and saves standard prompts for a section.
func (s *LlmService) GenerateDefaultPrompts(ctx context.Context, req *protos.GenerateDefaultPromptsRequest) (*protos.GenerateDefaultPromptsResponse, error) {
	designId := req.GetDesignId()
	sectionId := req.GetSectionId()
	slog.Info("Handling GenerateDefaultPrompts", "design_id", designId, "section_id", sectionId)

	if s.llmClient == nil || s.designSvc == nil {
		slog.Error("LLM or Design service dependency is nil in GenerateDefaultPrompts")
		return nil, status.Error(codes.Internal, "LLM service dependencies not configured")
	}

	// TODO: Permission check

	// Get section title
	sectionData, err := s.designSvc.readSectionData(designId, sectionId)
	if err != nil { // Handles NotFound
		slog.Error("Failed to read section data for GenerateDefaultPrompts", "design_id", designId, "section_id", sectionId, "error", err)
		// Distinguish between not found and other errors for status code
		if errors.Is(err, ErrNoSuchEntity) {
			return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
		}
		return nil, status.Error(codes.Internal, "Failed to read section metadata")
	}

	// Generate the prompts (using simple title-based logic for now, could use LLM later)
	// TODO: Use LLM to generate *better* default prompts if desired.
	defaultGetAnswerPrompt := fmt.Sprintf("Generate concise HTML content for a system design document section titled '%s'. Focus on key concepts, potential trade-offs, and common patterns related to this topic. ONLY include relevant HTML tags like <p>, <ul>, <li>, <strong>, <h2>, <h3>. Do NOT include <html>, <head>, <body>, or <style> tags. Start the content directly, for example, with an <p> or a <h3> tag.", sectionData.Title)
	defaultVerifyPrompt := fmt.Sprintf("Review the content of the section '%s' for clarity, completeness, technical accuracy, missed edge cases, and overall quality. Provide constructive feedback.", sectionData.Title)

	// Save the prompts to files
	err = s.designSvc.writePromptFile(designId, sectionId, "get_answer.md", defaultGetAnswerPrompt)
	if err != nil {
		// Log error but try to save the other prompt
		slog.Error("Failed to save generated get_answer.md prompt", "design_id", designId, "section_id", sectionId, "error", err)
	}

	errVerify := s.designSvc.writePromptFile(designId, sectionId, "verify.md", defaultVerifyPrompt)
	if errVerify != nil {
		slog.Error("Failed to save generated verify.md prompt", "design_id", designId, "section_id", sectionId, "error", errVerify)
		// If the first one failed too, return a general error
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to save generated prompts")
		}
	}

	// Return the generated prompts even if saving had issues (client can retry)
	resp := &protos.GenerateDefaultPromptsResponse{
		GetAnswerPrompt:    defaultGetAnswerPrompt,
		VerifyAnswerPrompt: defaultVerifyPrompt,
	}

	if err != nil || errVerify != nil {
		slog.Warn("GenerateDefaultPrompts completed but one or more prompt files failed to save.", "design_id", designId, "section_id", sectionId)
	} else {
		slog.Info("GenerateDefaultPrompts successful, prompts saved.", "design_id", designId, "section_id", sectionId)
	}
	return resp, nil
}
