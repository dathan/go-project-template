package pkg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

// Agent is a struct that represents an agent that can interact with different LLMs.
type Agent struct {
	llm llms.Model
}

// NewAgent creates a new Agent.
func NewAgent(provider string) (*Agent, error) {
	var llm llms.Model
	var err error

	switch provider {
	case "openai":
		llm, err = openai.New()
		if err != nil {
			return nil, err
		}
	case "claude":
		llm, err = anthropic.New()
		if err != nil {
			return nil, err
		}
	case "gemini":
		llm, err = googleai.New(context.Background())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return &Agent{llm: llm}, nil
}

// GetWeather is a function that returns the current temperature for a given city.
func GetWeather(city string) (string, error) {
	// In a real application, this would call a weather API.
	// For this example, we'll just return a dummy value.
	return fmt.Sprintf("The temperature in %s is 25°C", city), nil
}

// ToolCall is a struct that represents a tool call from the LLM.
type ToolCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Prompt sends a prompt to the LLM and returns the response.
// Returns an empty string and the error on failure so callers can decide how to surface it.
func (a *Agent) Prompt(prompt string) string {
	response, err := llms.GenerateFromSinglePrompt(context.Background(), a.llm, prompt,
		llms.WithTools([]llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "GetWeather",
					Description: "Gets the current weather for a given city.",
					Parameters:  json.RawMessage(`{"type": "object", "properties": {"city": {"type": "string", "description": "The city to get the weather for."}}, "required": ["city"]}`),
				},
			},
		}),
	)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return response
}

// Run demonstrates how to use the agent.
func Run() {
	agent, err := NewAgent("openai")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prompt := "What is the weather in San Francisco?"
	fmt.Println(agent.Prompt(prompt))
}