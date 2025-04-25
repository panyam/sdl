// FILE: ./services/llm/client.go
package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// ErrAPIKeyMissing indicates the OpenAI API key was not found in the environment.
var ErrAPIKeyMissing = errors.New("OpenAI API key not found in environment variable OPENAI_API_KEY")

// LLMClient defines the interface for interacting with an LLM.
type LLMClient interface {
	// SimpleQuery sends a single prompt to the LLM and returns the text response.
	SimpleQuery(ctx context.Context, prompt string) (string, error)
	// TODO: Add methods for chat history, system prompts, etc. later if needed.
}

// openaiClient implements LLMClient using the OpenAI API.
type openaiClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new client for interacting with the OpenAI API.
// It reads the API key from the OPENAI_API_KEY environment variable and
// the model name from OPENAI_MODEL (defaulting to gpt-3.5-turbo).
func NewOpenAIClient() (LLMClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		slog.Error("OpenAI API key missing")
		return nil, ErrAPIKeyMissing
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT4o // Default model
		slog.Info("OPENAI_MODEL not set, defaulting", "model", model)
	} else {
		slog.Info("Using OpenAI model from environment", "model", model)
	}

	client := openai.NewClient(apiKey)
	return &openaiClient{
		client: client,
		model:  model,
	}, nil
}

// SimpleQuery implements the LLMClient interface for OpenAI.
func (c *openaiClient) SimpleQuery(ctx context.Context, prompt string) (string, error) {
	slog.Debug("Sending simple query to OpenAI", "model", c.model, "prompt_length", len(prompt))

	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		// Add other parameters like Temperature, MaxTokens if needed
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Error("OpenAI API call failed", "error", err)
		return "", fmt.Errorf("LLM API request failed: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		slog.Warn("OpenAI response missing choices or content")
		// Consider retrying or returning a more specific error
		return "", errors.New("LLM returned empty response")
	}

	responseContent := resp.Choices[0].Message.Content
	slog.Debug("Received response from OpenAI", "response_length", len(responseContent))
	return responseContent, nil
}

// --- Mock Client for Testing ---

// MockLLMClient provides a mock implementation for testing purposes.
type MockLLMClient struct {
	SimpleQueryFunc  func(ctx context.Context, prompt string) (string, error) // Allow custom logic
	ResponseToReturn string
	ErrorToReturn    error
	ReceivedPrompt   string // Store the received prompt for assertion
}

// SimpleQuery implements the LLMClient interface for the mock.
func (m *MockLLMClient) SimpleQuery(ctx context.Context, prompt string) (string, error) {
	m.ReceivedPrompt = prompt // Record the prompt received
	if m.SimpleQueryFunc != nil {
		return m.SimpleQueryFunc(ctx, prompt)
	}
	// Default behavior if SimpleQueryFunc is not set
	if m.ErrorToReturn != nil {
		return "", m.ErrorToReturn
	}
	return m.ResponseToReturn, nil
}
