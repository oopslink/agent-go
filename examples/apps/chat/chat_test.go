package chat

import (
	"context"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/eventbus"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	"github.com/oopslink/agent-go-apps/chat/agent"
	"github.com/oopslink/agent-go-apps/chat/ui"
	"github.com/oopslink/agent-go-apps/pkg/config"
)

func TestChatComponents(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Provider:            "openai",
		APIKey:              "test-key",
		BaseURL:             "http://localhost:8080",
		ModelName:           "gpt-3.5-turbo",
		SystemPrompt:        "You are a helpful assistant.",
		RuntimeDir:          "./test-runtime",
		OpenAICompatibility: false,
	}

	// Create event bus
	eventBus := eventbus.NewEventBus()
	defer eventBus.Close()

	// Test agent creation (with nil knowledge base for simplicity)
	chatAgent, err := agent.NewChatAgent(cfg, eventBus, cfg.SystemPrompt, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create chat agent: %v", err)
	}

	if chatAgent == nil {
		t.Fatal("Chat agent is nil")
	}

	// Test UI creation
	chatUI := ui.NewChatUI(eventBus, chatAgent)
	if chatUI == nil {
		t.Fatal("Failed to create chat UI")
	}

	// Test context creation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test that components can be started (they should exit quickly due to timeout)
	go func() {
		if err := chatAgent.Start(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("Agent start error: %v", err)
		}
	}()

	go func() {
		if err := chatUI.Run(ctx); err != nil && err != context.DeadlineExceeded {
			t.Errorf("UI start error: %v", err)
		}
	}()

	// Wait for context to be cancelled
	<-ctx.Done()
}
