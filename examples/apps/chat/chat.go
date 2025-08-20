package chat

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/core/tools"
	chores "github.com/oopslink/agent-go/pkg/core/tools/chores"
	duckduckgo "github.com/oopslink/agent-go/pkg/core/tools/duckduckgo"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/journal"

	"github.com/oopslink/agent-go-apps/chat/agent"
	"github.com/oopslink/agent-go-apps/chat/ui"
	"github.com/oopslink/agent-go-apps/pkg/config"
	"github.com/oopslink/agent-go-apps/pkg/knowledges"
	"github.com/oopslink/agent-go-apps/pkg/mcp"
)

//go:embed system_prompt.md
var _systemPromptDefault string

func RunChat(ctx context.Context, cfg *config.Config) error {

	// Init shutdown handler
	ctx, cancel := initShutdownHandler(ctx)
	defer cancel()

	// Create event bus
	eventBus := eventbus.NewEventBus()
	defer eventBus.Close()

	// Load knowledge bases if configured
	knowledgeBases, err := initKnowledge(cfg)
	if err != nil {
		return err
	}

	// Init tools
	toolRegistry, err := initTools(ctx)
	if err != nil {
		return err
	}

	// Create chat agent
	chatAgent, err := initChatAgent(cfg, eventBus, knowledgeBases, toolRegistry)
	if err != nil {
		return err
	}

	// Start the agent in a goroutine
	go func() {
		if err = chatAgent.Start(ctx); err != nil {
			journal.Error("chat", "main", "Agent stopped with error", "error", err)
			os.Exit(-1)
		}
	}()

	// Start the UI (this will block until the context is cancelled)
	chatUI := ui.NewChatUI(eventBus, chatAgent)
	if err = chatUI.Run(ctx); err != nil {
		journal.Error("chat", "main", "UI stopped with error", "error", err)
		return err
	}

	return nil
}

func initShutdownHandler(ctx context.Context) (context.Context, context.CancelFunc) {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	return ctx, cancel
}

func initChatAgent(cfg *config.Config, eventBus *eventbus.EventBus,
	knowledgeBases []knowledge.KnowledgeBase, toolRegistry *tools.ToolCollection) (*agent.ChatAgent, error) {
	systemPrompt := _systemPromptDefault
	if len(cfg.SystemPrompt) > 0 {
		systemPrompt = cfg.SystemPrompt
	}
	chatAgent, err := agent.NewChatAgent(cfg, eventBus, systemPrompt, knowledgeBases, toolRegistry)
	if err != nil {
		journal.Error("chat", "main", "Failed to create chat agent", "error", err)
		return nil, err
	}
	return chatAgent, nil
}

func initTools(ctx context.Context) (*tools.ToolCollection, error) {
	// Setup MCP tools if enabled
	mcpTools, err := mcp.SetupMcpTools(ctx)
	if err != nil {
		journal.Error("chat", "main", err.Error())
		return nil, err
	}
	// Setup other tools
	allTools := append(mcpTools,
		chores.NewSleepTool(time.Second),
		duckduckgo.NewDuckDuckGoTool(),
	)
	return tools.OfTools(allTools...), nil
}

func initKnowledge(cfg *config.Config) ([]knowledge.KnowledgeBase, error) {
	var knowledgeBases []knowledge.KnowledgeBase
	if cfg.VectorDB.Type != "mock" {
		var err error
		knowledgeBases, err = knowledges.LoadKnowledgeBases(cfg)
		if err != nil {
			journal.Warning("chat", "main", "Failed to load knowledge bases", "error", err)
			return nil, err
		}
	}
	return knowledgeBases, nil
}
