package main

import (
	"log"

	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	u "github.com/oopslink/agent-go-snippets/utils"
)

func main() {
	// parse arguments
	question, apiKey, provider, modelName := u.ParseArgs("What is 15 * 23?")

	// create CoT agent
	// 1. create CoT behavior pattern
	behavior, err := behavior_patterns.NewChainOfThoughtPattern()
	if err != nil {
		log.Fatalf("failed to create behavior pattern: %v", err)
	}

	// 2. create CoT agent instance
	cotAgent, err := u.CreateAgent(
		"cot-demo",
		"You are a helpful AI assistant",
		apiKey, provider, modelName,
		behavior, nil, nil, false)
	if err != nil {
		log.Fatalf("Failed to create CoT agent: %v", err)
	}

	// run single conversation
	if err = u.RunSingleConversation(cotAgent,
		"🤖 Agent thinking...", question); err != nil {
		log.Fatalf("Failed to run conversation: %v", err)
	}
}
