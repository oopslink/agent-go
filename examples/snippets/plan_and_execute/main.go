package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	"github.com/oopslink/agent-go/pkg/core/tools"
	duckduckgo "github.com/oopslink/agent-go/pkg/core/tools/duckduckgo"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	u "github.com/oopslink/agent-go-snippets/utils"
)

func main() {
	// parse arguments
	question, apiKey, provider, modelName := u.ParseArgs(
		"Research and write a summary about the benefits of renewable energy")

	// Create plan and execute configuration
	config := &behavior_patterns.PlanAndExecuteConfig{
		RequirePlanConfirmation: true, // require user confirmation for plan
		RequireStepConfirmation: true, // require user confirmation for each step
	}

	// Create behavior pattern
	behavior, err := behavior_patterns.NewPlanExecutePattern(config)
	if err != nil {
		log.Fatalf("failed to create plan execute pattern: %v", err)
	}

	toolRegistry := tools.OfTools(
		u.NewWeatherTool(),
		duckduckgo.NewDuckDuckGoTool(),
	)
	theAgent, err := u.CreateAgent(
		"plan-and-execute-demo",
		"You are a helpful AI assistant",
		apiKey, provider, modelName,
		behavior, nil,
		toolRegistry, true)
	if err != nil {
		log.Fatalf("Failed to create PlanAndExecute agent: %v", err)
	}

	// Create run context
	runCtx := &agent.RunContext{
		SessionId: fmt.Sprintf("session-%d", time.Now().Unix()),
		Context:   context.Background(),
	}

	// Start agent
	inputChan, outputChan, err := theAgent.Run(runCtx)
	if err != nil {
		log.Fatalf("failed to start agent: %v", err)
	}

	// Send user request
	userRequest := &agent.UserRequest{
		Message: question,
	}

	inputChan <- agent.NewUserRequestEvent(userRequest)

	// Handle agent responses
	go func() {
		for event := range outputChan {
			handleEvent(event, inputChan, toolRegistry)
		}
	}()

	// Keep the program running
	select {}
}

func handleEvent(event *eventbus.Event, inputChan chan<- *eventbus.Event, registry *tools.ToolCollection) {
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
			fmt.Printf("< Agent:\n%s\n", content.String())
		}
		fmt.Println("---")

	case agent.EventTypeExternalAction:
		data := agent.GetExternalActionEventData(event)
		fmt.Println("< Agent:")
		if data.Message != "" {
			fmt.Println("  - request action: \n" + data.Message)
		}
		if data.ToolCall != nil {
			fmt.Println(fmt.Sprintf("  - request to use tool: %s\n%s",
				data.ToolCall.Name, data.ToolCall.MarshalJson()))
		}
		fmt.Println("---")
		simulateUserInteractions(inputChan, registry, data)

	case agent.EventTypeAgentResponseStart:
		fmt.Println("Agent response started")
		fmt.Println("---")

	case agent.EventTypeAgentResponseEnd:
		data := agent.GetAgentResponseEndEventData(event)
		if data.Error != nil {
			fmt.Printf("Agent response ended with error: %v\n", data.Error)
		} else {
			fmt.Printf("Agent response ended: %s\n", data.FinishReason)
		}
		fmt.Println("---")

		if data.Abort {
			os.Exit(0)
		}

	default:
		fmt.Printf("[skip] event: %s\n", event.Topic)
	}
}

func simulateUserInteractions(inputChan chan<- *eventbus.Event,
	registry *tools.ToolCollection, action *agent.ExternalAction) {
	defer func() {
		fmt.Println("---")
	}()
	// Wait a bit for the agent to process
	time.Sleep(2 * time.Second)

	if action.ToolCall != nil {
		fmt.Printf("> Me: \n exec tool [%s] with arguments [%s]\n",
			action.ToolCall.Name, action.ToolCall.MarshalJson())
		fmt.Println("---")

		result, err := registry.Call(context.TODO(), action.ToolCall)
		if err != nil {
			fmt.Printf(" * failed to exec tool, err: %s", err.Error())
			inputChan <- agent.NewFailedToolCallEvent(action.ToolCall.ToolCallId, action.ToolCall.Name, err)
			return
		}
		fmt.Printf("* tool executed, result: %s", result.MarshalJson())
		inputChan <- agent.NewToolCallResultEvent(result)
		return
	}

	confirmMessage := "I confirm the plan. Please proceed with execution."
	if len(action.Message) == 0 {
		confirmMessage = "no plan found, please re-plan."
	}

	fmt.Println("> Me: " + confirmMessage)
	inputChan <- agent.NewExternalActionResultEvent(confirmMessage)
	return
}
