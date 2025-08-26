package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	"github.com/oopslink/agent-go/pkg/core/chat"
	"github.com/oopslink/agent-go/pkg/support/llms"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	u "github.com/oopslink/agent-go-snippets/utils"
)

// SimpleConversationHandler implements ConversationHandler interface
type SimpleConversationHandler struct{}

func (h *SimpleConversationHandler) OnResponse(ctx *chat.ConversationContext, agentResponse *chat.AgentResponse) error {
	if agentResponse.Message != nil {
		// Extract text content from message parts
		var content strings.Builder
		for _, part := range agentResponse.Message.Parts {
			if textPart, ok := part.(*llms.TextPart); ok {
				content.WriteString(textPart.Text)
			}
		}
		fmt.Print(content.String())
	}

	if len(agentResponse.ToolCalls) > 0 {
		for _, toolCall := range agentResponse.ToolCalls {
			fmt.Printf("\n🔧 Tool Call: %s\n", toolCall.Name)
		}
		// For demonstration, we'll auto-approve all tool calls
		// In a real application, you might want to ask for user confirmation
	}

	return nil
}

func main() {
	// Parse arguments
	question, apiKey, provider, modelName := u.ParseArgs("What is 15 * 23? Please show your work step by step.")

	// Create Chain of Thought behavior pattern
	behavior, err := behavior_patterns.NewChainOfThoughtPattern()
	if err != nil {
		log.Fatalf("failed to create behavior pattern: %v", err)
	}

	// Create agent instance
	agent, err := u.CreateAgent(
		"conversation-demo",
		"You are a helpful AI assistant that shows your reasoning step by step.",
		apiKey, provider, modelName,
		behavior, nil, nil, false)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create conversation
	conversation := chat.NewConversation(agent)

	// Create conversation handler
	handler := &SimpleConversationHandler{}

	// Run conversation
	fmt.Printf("Question: %s\n\n", question)
	fmt.Println("🤖 Agent thinking...")

	ctx := context.Background()
	if err := conversation.Ask(ctx, question, handler); err != nil {
		log.Fatalf("Failed to run conversation: %v", err)
	}

	fmt.Printf("\n✅ Conversation completed successfully!\n")
}
