package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/dathan/go-project-template/pkg"
)

func main() {
	provider := flag.String("provider", "openai", "The LLM provider to use (openai, claude, or gemini)")
	prompt := flag.String("prompt", "What is the weather in San Francisco?", "The prompt to send to the LLM")
	flag.Parse()

	agent, err := pkg.NewAgent(*provider)
	if err != nil {
		log.Fatalf("Error creating agent: %v", err)
	}

	response, err := agent.Prompt(*prompt)
	if err != nil {
		log.Fatalf("Error getting response from agent: %v", err)
	}

	fmt.Println(response)
}
