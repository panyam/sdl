// FILE: ./services/llm_service.go
package services

import (
	"context"
	"encoding/json"
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
	// TODO: Add permission check

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
	promptBuilder.WriteString("Get Answer Prompt: <Prompt text for generation>\n") // Include prompt fields in suggestions
	promptBuilder.WriteString("Verify Prompt: <Prompt text for verification>\n")   // Include prompt fields in suggestions
	promptBuilder.WriteString("Description: <Brief explanation>\n")
	promptBuilder.WriteString("---\n") // Separator
	// Add an example with placeholders to reinforce the format
	promptBuilder.WriteString("Title: <Example Title>\nType: <text>\nGet Answer Prompt: <Prompt text...>\nVerify Prompt: <Prompt text...>\nDescription: <Example description.>\n---\n")

	prompt := promptBuilder.String()

	// Call the LLM
	llmResponse, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		slog.Error("LLM query failed for SuggestSections", "error", err)
		return nil, status.Error(codes.Internal, "Failed to get suggestions from LLM")
	}

	// Parse the LLM response (existing helper is fine for this structure)
	suggestions := parseSuggestions(llmResponse)
	if len(suggestions) == 0 {
		slog.Warn("LLM returned no parsable suggestions for SuggestSections", "raw_response", llmResponse)
		// Return empty list, maybe client can show a message
	}

	resp := &protos.SuggestSectionsResponse{
		Suggestions: suggestions,
	}
	slog.Info("SuggestSections successful", "suggestion_count", len(resp.Suggestions))
	return resp, nil
}

// Helper to parse the structured LLM response (used by SuggestSections)
func parseSuggestions(rawResponse string) []*protos.SuggestedSection {
	var suggestions []*protos.SuggestedSection
	parts := strings.Split(strings.TrimSpace(rawResponse), "---") // Split by ---
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue // Skip empty parts
		}

		suggestion := &protos.SuggestedSection{}
		lines := strings.Split(part, "\n")
		currentKey := "" // Track the current field being parsed (e.g., for multi-line content)
		currentValue := &strings.Builder{}

		processCurrentField := func() {
			content := strings.TrimSpace(currentValue.String())
			switch currentKey {
			case "Title":
				suggestion.Title = content
			case "Type":
				// Basic validation for type
				if content == "text" || content == "drawing" || content == "plot" {
					suggestion.Type = content
				} else {
					suggestion.Type = "text" // Default if invalid
					slog.Warn("LLM suggested invalid type, defaulting to text", "invalid_type", content)
				}
			case "Get Answer Prompt": // Match the prompt key from the LLM instruction
				suggestion.GetAnswerPrompt = content
			case "Verify Prompt": // Match the prompt key from the LLM instruction
				suggestion.VerifyAnswerPrompt = content
			case "Description":
				suggestion.Description = content
			}
			currentValue.Reset() // Clear for the next field
		}

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue // Skip empty lines within a block
			}

			// Check for known field prefixes
			if strings.HasPrefix(line, "Title:") {
				processCurrentField() // Process previous field if any
				currentKey = "Title"
				currentValue.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "Title:")))
			} else if strings.HasPrefix(line, "Type:") {
				processCurrentField()
				currentKey = "Type"
				currentValue.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "Type:")))
			} else if strings.HasPrefix(line, "Get Answer Prompt:") { // Match the prompt key
				processCurrentField()
				currentKey = "Get Answer Prompt"
				currentValue.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "Get Answer Prompt:")))
			} else if strings.HasPrefix(line, "Verify Prompt:") { // Match the prompt key
				processCurrentField()
				currentKey = "Verify Prompt"
				currentValue.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "Verify Prompt:")))
			} else if strings.HasPrefix(line, "Description:") {
				processCurrentField()
				currentKey = "Description"
				currentValue.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "Description:")))
			} else if currentKey != "" {
				// If no known prefix, append to the current value (handles multi-line)
				if currentValue.Len() > 0 {
					currentValue.WriteString("\n") // Add newline for multi-line content
				}
				currentValue.WriteString(line)
			}
		}
		processCurrentField() // Process the last field

		// Only add if we got a title and type
		if suggestion.Title != "" && suggestion.Type != "" {
			suggestions = append(suggestions, suggestion)
		} else {
			slog.Warn("Skipping suggestion due to missing Title or Type", "part", part)
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

	// Get section metadata (mainly for type/title)
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

	if err != nil || strings.TrimSpace(prompt) == "" { // If file not found OR read error OR file empty/whitespace
		slog.Info("Get answer prompt file not found or empty, using default.", "design_id", designId, "section_id", sectionId)
		// Default prompt using the section title (This fallback prompt is still hardcoded but less generic)
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

	// Get section metadata (for type check/title)
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
	existingContent := string(contentBytes) // Will be empty string if file didn't exist or was empty

	var prompt string
	verifyPromptInstruction, err := s.designSvc.readPromptFile(designId, sectionId, "verify.md")
	if err == nil && strings.TrimSpace(verifyPromptInstruction) != "" { // Use file content if exists and not empty
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
		if strings.TrimSpace(existingContent) == "" { // Special case for empty content
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

// --- New Helper to parse LLM's JSON response for prompts ---
type generatedPrompts struct {
	GetAnswerPrompt string `json:"get_answer_prompt"`
	VerifyPrompt    string `json:"verify_prompt"`
}

func parseGeneratedPromptsJSON(rawResponse string) (*generatedPrompts, error) {
	var prompts generatedPrompts
	err := json.Unmarshal([]byte(rawResponse), &prompts)
	if err != nil {
		slog.Error("Failed to unmarshal LLM JSON response", "error", err, "raw_response", rawResponse)
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	return &prompts, nil
}

// --- End New Helper ---

// GenerateDefaultPrompts generates and saves standard prompts for a section using LLM.
func (s *LlmService) GenerateDefaultPrompts(ctx context.Context, req *protos.GenerateDefaultPromptsRequest) (*protos.GenerateDefaultPromptsResponse, error) {
	designId := req.GetDesignId()
	sectionId := req.GetSectionId()
	slog.Info("Handling GenerateDefaultPrompts (LLM based)", "design_id", designId, "section_id", sectionId)

	if s.llmClient == nil || s.designSvc == nil {
		slog.Error("LLM or Design service dependency is nil in GenerateDefaultPrompts")
		return nil, status.Error(codes.Internal, "LLM service dependencies not configured")
	}

	// TODO: Permission check (Can user edit this design/section?)

	// Get section metadata (type, title, description)
	sectionData, err := s.designSvc.readSectionData(designId, sectionId)
	if err != nil {
		if errors.Is(err, ErrNoSuchEntity) {
			slog.Warn("Section not found for GenerateDefaultPrompts", "design_id", designId, "section_id", sectionId)
			return nil, status.Errorf(codes.NotFound, "Section '%s' not found in design '%s'", sectionId, designId)
		}
		slog.Error("Failed to read section data for GenerateDefaultPrompts", "design_id", designId, "section_id", sectionId, "error", err)
		return nil, status.Error(codes.Internal, "Failed to read section metadata")
	}

	// Get design metadata (design title)
	designMetadata, err := s.designSvc.readDesignMetadata(designId)
	if err != nil {
		slog.Error("Failed to read design metadata for GenerateDefaultPrompts", "design_id", designId, "error", err)
		// Decide if this is a fatal error or if we can proceed without design title
		// Let's make it fatal for now, as design context is important.
		if errors.Is(err, ErrNoSuchEntity) {
			return nil, status.Errorf(codes.NotFound, "Design '%s' not found", designId)
		}
		return nil, status.Error(codes.Internal, "Failed to read design metadata")
	}
	designTitle := designMetadata.Name

	// Construct the prompt for the LLM
	// Ask for both prompts in JSON format
	var promptBuilder strings.Builder
	promptBuilder.WriteString("You are an expert system design interviewer and assistant.\n")
	promptBuilder.WriteString(fmt.Sprintf("You are helping a user generate default LLM prompts for a specific section within a system design document.\n"))
	promptBuilder.WriteString(fmt.Sprintf("The overall document is titled: '%s'.\n", designTitle))
	promptBuilder.WriteString(fmt.Sprintf("The current section is titled: '%s'.\n", sectionData.Title))
	if sectionData.Description != "" {
		promptBuilder.WriteString(fmt.Sprintf("Its description is: '%s'.\n", sectionData.Description))
	}
	promptBuilder.WriteString(fmt.Sprintf("The section type is: '%s'.\n", sectionData.Type))

	promptBuilder.WriteString("\nGenerate two distinct prompts based on this context:\n")
	promptBuilder.WriteString("1.  **'Get Answer Prompt'**: This prompt will be used by the user to ask an LLM to *generate* content for this section.\n")
	promptBuilder.WriteString("    It should be a detailed instruction to the LLM on what kind of content to produce, considering the section's topic, type, and the overall design context. It should specify the desired output format (e.g., HTML for text sections, JSON config structure for others).\n")
	promptBuilder.WriteString("2.  **'Verify Prompt'**: This prompt will be used by the user to ask an LLM to *review* existing content within this section.\n")
	promptBuilder.WriteString("    It should instruct the LLM on how to evaluate the content for accuracy, completeness, clarity, edge cases, etc., based on the section's topic and the overall design.\n")

	promptBuilder.WriteString("\nProvide the output as a JSON object with the keys `get_answer_prompt` and `verify_prompt`. Ensure the output is valid JSON and contains ONLY the JSON object.\n")

	// Example of expected JSON structure (helpful for the LLM)
	promptBuilder.WriteString("\nExample Output:\n")
	promptBuilder.WriteString(`{"get_answer_prompt": "Generate detailed steps for setting up a distributed cache system...", "verify_prompt": "Review the following content about distributed caching for key concepts, performance considerations, and common pitfalls..."}`)
	promptBuilder.WriteString("\n\nRespond ONLY with the JSON object.\n")

	prompt := promptBuilder.String()

	// Call the LLM
	llmResponse, err := s.llmClient.SimpleQuery(ctx, prompt)
	if err != nil {
		slog.Error("LLM query failed for GenerateDefaultPrompts", "error", err)
		return nil, status.Error(codes.Internal, "Failed to generate prompts via LLM")
	}

	// Parse the LLM response (expecting JSON)
	generated, err := parseGeneratedPromptsJSON(llmResponse)
	if err != nil {
		slog.Error("Failed to parse LLM response for prompts", "raw_response", llmResponse, "error", err)
		// Decide on error code - DataLoss if format is bad, Internal if unexpected
		return nil, status.Error(codes.DataLoss, "Failed to parse LLM response into expected format")
	}

	// Save the generated prompts to files
	errGet := s.designSvc.writePromptFile(designId, sectionId, "get_answer.md", generated.GetAnswerPrompt)
	errVerify := s.designSvc.writePromptFile(designId, sectionId, "verify.md", generated.VerifyPrompt)

	if errGet != nil || errVerify != nil {
		slog.Error("Failed to save one or both generated prompt files", "design_id", designId, "section_id", sectionId, "error_get", errGet, "error_verify", errVerify)
		// Return an error, even if LLM call succeeded, because saving failed.
		return nil, status.Error(codes.Internal, "Failed to save generated prompts")
	}

	// Return the generated prompts
	resp := &protos.GenerateDefaultPromptsResponse{
		GetAnswerPrompt:    generated.GetAnswerPrompt,
		VerifyAnswerPrompt: generated.VerifyPrompt,
	}

	slog.Info("GenerateDefaultPrompts successful, prompts generated and saved.", "design_id", designId, "section_id", sectionId)
	return resp, nil
}
