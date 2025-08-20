package utils

import (
	"fmt"
	"log"
	"os"
)

func ParseArgs(questionSample string) (string, string, string, string) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go \"your question here\"")
		fmt.Println(fmt.Sprintf("Example: go run main.go \"%s\"", questionSample))
		os.Exit(1)
	}

	question := os.Args[1]

	// Get API key from environment
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	// Get provider from environment or use default
	provider := os.Getenv("PROVIDER")
	if provider == "" {
		provider = "openai"
	}

	// Get model from environment or use default
	modelName := os.Getenv("MODEL")
	if modelName == "" {
		switch provider {
		case "openai":
			modelName = "gpt-4o-mini"
		case "anthropic":
			modelName = "claude-3-5-sonnet-20241022"
		case "gemini":
			modelName = "gemini-2.5-flash"
		default:
			modelName = "gpt-4o-mini"
		}
	}
	return question, apiKey, provider, modelName
}
