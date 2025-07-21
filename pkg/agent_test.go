package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// MockLLM is a mock implementation of the llms.Model interface for testing.
type MockLLM struct {
	CallFn func(ctx context.Context, prompt string, options ...llms.CallOption) (string, error)
}

// Call implements the llms.Model interface.
func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return m.CallFn(ctx, prompt, options...)
}

// GenerateContent implements the llms.Model interface.
func (m *MockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Not needed for this test.
	return nil, nil
}

func TestAgent_Prompt(t *testing.T) {
	// Create a mock LLM that returns a dummy response.
	mockLLM := &MockLLM{
		CallFn: func(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
			return `{"name": "GetWeather", "arguments": "{\"city\": \"London\"}"}`, nil
		},
	}

	// Create a new agent with the mock LLM.
	agent := &Agent{llm: mockLLM}

	// Send a prompt to the agent.
	response, err := agent.Prompt("What is the weather in London?")
	assert.NoError(t, err)

	// Check that the response is what we expect.
	assert.Equal(t, `{"name": "GetWeather", "arguments": "{\"city\": \"London\"}"}`, response)
}
