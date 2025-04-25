// FILE: ./services/llm/client_test.go
package llm

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIClient(t *testing.T) {
	t.Run("Success Case", func(t *testing.T) {
		// Set a dummy key for this test
		originalKey, keyExisted := os.LookupEnv("OPENAI_API_KEY")
		os.Setenv("OPENAI_API_KEY", "test-key-123")
		defer func() {
			// Restore original environment state
			if keyExisted {
				os.Setenv("OPENAI_API_KEY", originalKey)
			} else {
				os.Unsetenv("OPENAI_API_KEY")
			}
		}()

		client, err := NewOpenAIClient()
		require.NoError(t, err)
		require.NotNil(t, client)

		// Check if it's the correct type (optional but good practice)
		_, ok := client.(*openaiClient)
		assert.True(t, ok, "Client should be of type *openaiClient")
	})

	t.Run("Failure Case - API Key Missing", func(t *testing.T) {
		// Ensure the key is NOT set
		originalKey, keyExisted := os.LookupEnv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_API_KEY")
		defer func() {
			// Restore original environment state
			if keyExisted {
				os.Setenv("OPENAI_API_KEY", originalKey)
			}
		}()

		client, err := NewOpenAIClient()
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrAPIKeyMissing, "Error should be ErrAPIKeyMissing")
	})

	t.Run("Model Defaulting", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "test-key-123")
		os.Unsetenv("OPENAI_MODEL") // Ensure model is not set
		defer os.Unsetenv("OPENAI_API_KEY")
		defer os.Unsetenv("OPENAI_MODEL")

		client, err := NewOpenAIClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		oaiClient, ok := client.(*openaiClient)
		require.True(t, ok)
		assert.NotEmpty(t, oaiClient.model, "Default model should be set")
		// You could assert the specific default model if needed, e.g., openai.GPT3Dot5Turbo
	})

	t.Run("Model From Environment", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "test-key-123")
		os.Setenv("OPENAI_MODEL", "test-model-from-env")
		defer os.Unsetenv("OPENAI_API_KEY")
		defer os.Unsetenv("OPENAI_MODEL")

		client, err := NewOpenAIClient()
		require.NoError(t, err)
		require.NotNil(t, client)
		oaiClient, ok := client.(*openaiClient)
		require.True(t, ok)
		assert.Equal(t, "test-model-from-env", oaiClient.model)
	})
}

func TestMockLLMClient(t *testing.T) {
	t.Run("Returns Response", func(t *testing.T) {
		mock := &MockLLMClient{
			ResponseToReturn: "Mock response",
			ErrorToReturn:    nil,
		}
		prompt := "Test prompt"
		resp, err := mock.SimpleQuery(context.Background(), prompt)

		require.NoError(t, err)
		assert.Equal(t, "Mock response", resp)
		assert.Equal(t, prompt, mock.ReceivedPrompt) // Verify prompt was captured
	})

	t.Run("Returns Error", func(t *testing.T) {
		mockErr := errors.New("mock LLM error")
		mock := &MockLLMClient{
			ErrorToReturn: mockErr,
		}
		prompt := "Another prompt"
		resp, err := mock.SimpleQuery(context.Background(), prompt)

		require.Error(t, err)
		assert.ErrorIs(t, err, mockErr)
		assert.Empty(t, resp)
		assert.Equal(t, prompt, mock.ReceivedPrompt) // Verify prompt was captured
	})

	t.Run("Uses Custom Func", func(t *testing.T) {
		expectedResp := "Response from func"
		expectedErr := errors.New("error from func")
		called := false
		mock := &MockLLMClient{
			SimpleQueryFunc: func(ctx context.Context, p string) (string, error) {
				called = true
				assert.Equal(t, "prompt for func", p)
				return expectedResp, expectedErr
			},
		}
		resp, err := mock.SimpleQuery(context.Background(), "prompt for func")
		assert.True(t, called, "SimpleQueryFunc should have been called")
		assert.Equal(t, expectedResp, resp)
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, "prompt for func", mock.ReceivedPrompt) // Verify prompt was captured by the struct field too
	})
}
