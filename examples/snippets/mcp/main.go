package main

import (
	"context"
	"log"

	"github.com/oopslink/agent-go/pkg/core/tools"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	mcp "github.com/oopslink/agent-go-snippets/mcp/mcp"
	u "github.com/oopslink/agent-go-snippets/utils"
)

func main() {
	// parse arguments
	question, apiKey, provider, modelName := u.ParseArgs("What time is it?")

	// setup MCP tools
	mcpTools, err := mcp.SetupMcpTools(context.Background())
	if err != nil {
		log.Fatalf("Failed to setup MCP tools: %v", err)
	}

	// create tool collection
	toolCollection := tools.OfTools(mcpTools...)

	// create MCP Agent instance
	mcpAgent, err := u.CreateAgent(
		"mcp-demo",
		"You are a helpful AI assistant with access to time tools. You can get the current time in various formats.",
		apiKey, provider, modelName,
		nil, nil, toolCollection, false)
	if err != nil {
		log.Fatalf("Failed to create MCP agent: %v", err)
	}

	// run single conversation
	if err = u.RunSingleConversation(mcpAgent,
		"🤖 Agent using MCP time tools to get current time...", question); err != nil {
		log.Fatalf("Failed to run conversation: %v", err)
	}
}
