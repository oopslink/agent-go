package agent

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	agentcontext "github.com/oopslink/agent-go/pkg/core/agent/context"
	agentstate "github.com/oopslink/agent-go/pkg/core/agent/state"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/core/memory"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"

	"github.com/oopslink/agent-go-apps/pkg/config"
)

var (
	ErrorCodeInitAgentFailed = errors.ErrorCode{
		Code:           91000,
		Name:           "InitAgentFailed",
		DefaultMessage: "Failed to init agent",
	}

	ErrorCodeStartAgentFailed = errors.ErrorCode{
		Code:           91001,
		Name:           "StartAgentFailed",
		DefaultMessage: "Failed to start agent",
	}
)

type ChatAgentCtrl interface {
	ClearMemory() error
	RetrieveAllMemory(ctx context.Context) ([]memory.MemoryItem, error)
	GetInfo() string
	ExecTool(ctx context.Context, call *llms.ToolCall) error
}

// NewChatAgent creates a new chat agent instance
func NewChatAgent(
	cfg *config.Config,
	eventBus *eventbus.EventBus,
	systemPrompt string,
	knowledgeBases []knowledge.KnowledgeBase,
	toolRegistry *tools.ToolCollection) (*ChatAgent, error) {
	// 1. Create memory
	agentMemory := memory.NewInMemoryMemory()
	agentState := agentstate.NewInMemoryState()

	// 2. Create LLM provider
	llmProvider, err := llms.NewChatProvider(llms.ModelProvider(cfg.Provider),
		llms.WithAPIKey(cfg.APIKey),
		llms.WithBaseUrl(cfg.BaseURL))
	if err != nil {
		return nil, errors.Errorf(ErrorCodeInitAgentFailed, "failed to create LLM provider: %v", err)
	}

	// 3. Get model
	model, _ := llms.GetModel(
		llms.ModelId{
			Provider: llms.ModelProvider(cfg.Provider),
			ID:       cfg.ModelName,
		})

	// 4. Create behavior pattern
	behavior, err := behavior_patterns.NewGenericPattern()
	if err != nil {
		return nil, errors.Errorf(ErrorCodeInitAgentFailed, "failed to create behavior pattern: %v", err)
	}

	// Create agent context
	agentCtx := agentcontext.NewRuleBaseContext(
		"chat-demo",
		systemPrompt,
		llmProvider,
		model,
		behavior,
		agentMemory,
		agentState,
		knowledgeBases,
		toolRegistry,
		agentcontext.ContextRules{
			AutoAddToolInstructions: true,
			UseKnowledgeAsTool:      true,
			AutoTools:               []string{knowledge.KNOWLEDGE_RETRIVER_TOOL},
		},
	)

	// Create generic agent
	genericAgent, err := agent.NewGenericAgent(
		agentCtx, behavior, llmProvider, model,
		[]llms.ChatOption{
			llms.WithTemperature(0.8),
			llms.WithStreaming(true),
		})
	if err != nil {
		return nil, errors.Errorf(ErrorCodeInitAgentFailed, "failed to create generic agent: %v", err)
	}

	return &ChatAgent{
		config: cfg,
		model:  model,

		agent:              genericAgent,
		contextRuleUpdater: agentCtx.(agentcontext.ContextRuleUpdater),
		memory:             agentMemory,
		toolRegistry:       toolRegistry,
		knowledgeBases:     knowledgeBases,

		eventBus: eventBus,
	}, nil
}

var _ ChatAgentCtrl = &ChatAgent{}

// ChatAgent represents the chat agent implementation
type ChatAgent struct {
	config *config.Config
	model  *llms.Model

	agent              agent.Agent
	memory             memory.Memory
	toolRegistry       *tools.ToolCollection
	knowledgeBases     []knowledge.KnowledgeBase
	contextRuleUpdater agentcontext.ContextRuleUpdater

	inputChan chan<- *eventbus.Event
	eventBus  *eventbus.EventBus
}

// Start begins the agent processing loop
func (ca *ChatAgent) Start(ctx context.Context) error {
	// Subscribe to user input events
	_, err := ca.eventBus.Subscribe(EventTypeUserInput, ca.handleUserInput, false, 10)
	if err != nil {
		return errors.Errorf(ErrorCodeStartAgentFailed, "failed to subscribe to user input: %v", err)
	}

	// Create input and output channels for the agent
	inputChan, outputChan, err := ca.agent.Run(
		&agent.RunContext{
			SessionId: fmt.Sprintf("ctx-%s", utils.GenerateUUID()),
			Context:   ctx,
		})
	if err != nil {
		return errors.Errorf(ErrorCodeStartAgentFailed, "failed to start agent: %v", err)
	}

	ca.inputChan = inputChan

	// Start output processing goroutine
	go ca.processAgentResponse(ctx, outputChan)

	return nil
}

func (ca *ChatAgent) ClearMemory() error {
	return ca.memory.Reset()
}

func (ca *ChatAgent) RetrieveAllMemory(ctx context.Context) ([]memory.MemoryItem, error) {
	return ca.memory.Retrieve(ctx)
}

func (ca *ChatAgent) GetInfo() string {
	return fmt.Sprintf("Configuration Information:\n"+
		"- Provider: %s\n"+
		"- Model: %s\n"+
		"- Base URL: %s\n"+
		"- Knowledge Base: %s",
		ca.config.Provider,
		ca.config.ModelName,
		ca.config.BaseURL,
		func() string {
			if len(ca.knowledgeBases) > 0 {
				return "Available"
			}
			return "Not configured"
		}())
}

// processAgentResponse processes agent output events
func (ca *ChatAgent) processAgentResponse(ctx context.Context, outputChan <-chan *eventbus.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-outputChan:
			if !ok {
				return
			}
			_ = ca.eventBus.Publish(event)
		}
	}
}

// handleUserInput processes user input events
func (ca *ChatAgent) handleUserInput(ctx context.Context, event *eventbus.Event) error {
	if event == nil {
		return nil
	}

	userInput := event.Data.(*UserInput)
	if len(userInput.Message) > 0 {
		ca.askAgent(userInput)
	}

	if userInput.ToolCall != nil {
		return ca.handleUserToolCallConfirm(ctx, userInput)
	}

	return nil
}

func (ca *ChatAgent) ExecTool(ctx context.Context, toolCall *llms.ToolCall) error {
	return ca.handleUserToolCallConfirm(ctx,
		&UserInput{
			ToolCall: toolCall,
		})
}

func (ca *ChatAgent) askAgent(userInput *UserInput) {
	ca.inputChan <- agent.NewUserRequestEvent(
		&agent.UserRequest{
			Message: userInput.Message,
		})
}

func (ca *ChatAgent) handleUserToolCallConfirm(ctx context.Context, input *UserInput) error {
	toolCall := input.ToolCall
	if input.IgnoreToolCall {
		ca.inputChan <- agent.NewUserSkipToolCallEvent(toolCall.ToolCallId, toolCall.Name)
	} else {
		result, err := ca.toolRegistry.Call(ctx, toolCall)
		if err != nil {
			ca.inputChan <- agent.NewFailedToolCallEvent(toolCall.ToolCallId, toolCall.Name, err)
			return err
		}
		ca.inputChan <- agent.NewToolCallResultEvent(result)
	}
	return nil
}
