package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

const weatherResponse = `{"name": "GetWeather", "arguments": "{\"city\": \"London\"}"}`

// MockLLM is a mock implementation of the llms.Model interface for testing.
// GenerateFromSinglePrompt calls GenerateContent internally, so that is the
// method we need to stub.
type MockLLM struct{}

func (m *MockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return weatherResponse, nil
}

func (m *MockLLM) GenerateContent(_ context.Context, _ []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: weatherResponse},
		},
	}, nil
}

func TestAgent_Prompt(t *testing.T) {
	agent := &Agent{llm: &MockLLM{}}
	response := agent.Prompt("What is the weather in London?")
	assert.Equal(t, weatherResponse, response)
}
