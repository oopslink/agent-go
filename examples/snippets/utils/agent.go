package utils

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	agentcontext "github.com/oopslink/agent-go/pkg/core/agent/context"
	agentstate "github.com/oopslink/agent-go/pkg/core/agent/state"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/core/memory"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"strings"
)

// HandleAgentEvent processes agent events
func HandleAgentEvent(event *eventbus.Event) (bool, error) {
	switch event.Topic {
	case agent.EventTypeAgentMessage:
		if messageEvent := agent.GetAgentMessageEventData(event); messageEvent != nil {
			// Extract text content from message parts
			var content strings.Builder
			for _, part := range messageEvent.Message.Parts {
				if textPart, ok := part.(*llms.TextPart); ok {
					content.WriteString(textPart.Text)
				}
			}
			fmt.Print(content.String())
		}
	case agent.EventTypeAgentResponseEnd:
		if endEvent := agent.GetAgentResponseEndEventData(event); endEvent != nil {
			if endEvent.Error == nil {
				return true, endEvent.Error
			}
			if endEvent.Abort {
				fmt.Printf("\n❌ Agent stopped: %v\n", endEvent.Error)
			} else {
				fmt.Printf("\n✅ Agent finished with reason: %s\n", endEvent.FinishReason)
			}
			return true, nil // Signal to stop processing
		}
	default:
		fmt.Println(fmt.Sprintf("[SKIP]: %v", event))
	}
	return false, nil
}

// CreateAgent creates a agent
func CreateAgent(agentId, systemPrompt string,
	apiKey, provider, modelName string,
	behavior agent.BehaviorPattern,
	knowledgeBases []knowledge.KnowledgeBase,
	toolRegistry *tools.ToolCollection,
	autoAddToolInstructions bool) (agent.Agent, error) {
	// Create memory
	mem := memory.NewInMemoryMemory()

	// Create LLM provider
	llmProvider, err := llms.NewChatProvider(llms.ModelProvider(provider),
		llms.WithAPIKey(apiKey))
	if err != nil {
		return nil, errors.Errorf(errors.InternalError, "failed to create LLM provider: %v", err)
	}

	// Get model
	model, _ := llms.GetModel(
		llms.ModelId{
			Provider: llms.ModelProvider(provider),
			ID:       modelName,
		})

	// Create default behavior
	if behavior == nil {
		behavior, _ = behavior_patterns.NewGenericPattern()
	}

	var autoTools []string
	if toolRegistry != nil {
		for idx := range toolRegistry.Tools {
			tool := toolRegistry.Tools[idx]
			autoTools = append(autoTools, tool.Descriptor().Name)
		}
	}

	// Create agent context
	agentCtx := agentcontext.NewRuleBaseContext(
		agentId,
		systemPrompt,
		llmProvider,
		model,
		behavior,
		mem, agentstate.NewInMemoryState(),
		knowledgeBases,
		toolRegistry,
		agentcontext.ContextRules{
			AutoAddToolInstructions: autoAddToolInstructions,
			UseKnowledgeAsTool:      false,
			AutoTools:               autoTools,
		},
	)

	// create agent
	return agent.NewGenericAgent(
		agentCtx, behavior, llmProvider, model,
		[]llms.ChatOption{
			llms.WithTemperature(0.7),
			llms.WithStreaming(true),
		})
}
