package main

import (
	"log"

	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	"github.com/oopslink/agent-go/pkg/core/tools"
	duckduckgo "github.com/oopslink/agent-go/pkg/core/tools/duckduckgo"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	u "github.com/oopslink/agent-go-snippets/utils"
)

func main() {
	// parse arguments
	question, apiKey, provider, modelName := u.ParseArgs("What is the current weather in New York?")

	// create ReAct agent with tool support
	// 1. create ReAct behavior pattern
	behavior, err := behavior_patterns.NewReActPattern(10)
	if err != nil {
		log.Fatalf("failed to create behavior pattern: %v", err)
	}
	// 2. create ReAct agent instance with tools
	reactAgent, err := u.CreateAgent(
		"react-demo",
		"You are a helpful AI assistant that can reason and act to solve problems",
		apiKey, provider, modelName,
		behavior, nil,
		tools.OfTools(
			u.NewWeatherTool(),
			duckduckgo.NewDuckDuckGoTool(),
		),
		true,
	)
	if err != nil {
		log.Fatalf("Failed to create ReAct agent: %v", err)
	}

	// run single conversation
	if err = u.RunSingleConversation(reactAgent,
		"🤖 Agent reasoning and acting with tool calls...", question); err != nil {
		log.Fatalf("Failed to run conversation: %v", err)
	}
}
