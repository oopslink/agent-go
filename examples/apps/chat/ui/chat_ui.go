package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"

	chatagent "github.com/oopslink/agent-go-apps/chat/agent"
	"github.com/oopslink/agent-go-apps/pkg/ui"
)

var (
	ErrorCodeChatUIError = errors.ErrorCode{
		Code:           90010,
		Name:           "ChatUIError",
		DefaultMessage: "chat ui error",
	}
)

// ChatUI represents the command-line user interface
type ChatUI struct {
	eventBus *eventbus.EventBus
	uiChan   chan *eventbus.Event

	loadingMu     sync.Mutex
	loadingCancel context.CancelFunc

	chatAgentCtrl chatagent.ChatAgentCtrl

	autoTools []string
}

// NewChatUI creates a new chat UI instance
func NewChatUI(eventBus *eventbus.EventBus, chatAgentCtrl chatagent.ChatAgentCtrl) *ChatUI {
	uiChan := make(chan *eventbus.Event, 100)
	return &ChatUI{
		eventBus: eventBus,
		uiChan:   uiChan,

		chatAgentCtrl: chatAgentCtrl,
	}
}

// Run begins the UI loop
func (c *ChatUI) Run(ctx context.Context) error {
	// Subscribe to agent output events
	topics := []string{
		agent.EventTypeAgentMessage,
		agent.EventTypeExternalAction,
		agent.EventTypeAgentResponseStart,
		agent.EventTypeAgentResponseEnd,
	}
	for _, topic := range topics {
		_, err := c.eventBus.Subscribe(topic,
			func(ctx context.Context, event *eventbus.Event) error {
				c.uiChan <- event
				return nil
			},
			true, 10)
		if err != nil {
			return errors.Errorf(ErrorCodeChatUIError, "failed to subscribe [%s] to agent responses: %v", topic, err)
		}
	}

	// Main UI loop
	return c.runMainLoop(ctx)
}

func (c *ChatUI) runMainLoop(ctx context.Context) error {
	c.printWelcome()
	c.printHelp()
	ui.PrintSeparator("-")

	c.handleUserInput(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-c.uiChan:
			if err := c.handleAgentEvents(ctx, event); err != nil {
				return err
			}
		}
	}
}

func (c *ChatUI) handleAgentEvents(ctx context.Context, event *eventbus.Event) error {
	switch event.Topic {
	case agent.EventTypeAgentMessage:
		c.resetLoadingCancel(nil)
		msg := agent.GetAgentMessageEventData(event)
		c.printAgentMessagePart(msg.Message)
	case agent.EventTypeExternalAction:
		c.resetLoadingCancel(nil)
		externalAction := agent.GetExternalActionEventData(event)
		c.confirmExternalAction(ctx, externalAction)
	case agent.EventTypeAgentResponseStart:
		c.resetLoadingCancel(ui.Loading("Generating"))
	case agent.EventTypeAgentResponseEnd:
		c.resetLoadingCancel(nil)
		c.handleAgentResponseEnd(ctx, agent.GetAgentResponseEndEventData(event))
	default:
		return errors.Errorf(agent.ErrorCodeInvalidInputEvent, "invalid input event type: %s", event.Topic)
	}
	return nil
}

func (c *ChatUI) printAgentMessagePart(message *llms.Message) {
	if message == nil {
		return
	}
	for _, part := range message.Parts {
		text, ok := part.(*llms.TextPart)
		if ok {
			for _, w := range text.Text {
				ui.Print(fmt.Sprintf("%c", w))
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (c *ChatUI) confirmExternalAction(ctx context.Context, externalAction *agent.ExternalAction) {
	toolCall := externalAction.ToolCall
	if toolCall != nil {
		c.confirmToolCall(ctx, toolCall)
		return
	}
	if len(externalAction.Message) > 0 {
		c.confirmAgentMessage(ctx, externalAction.Message)
		return
	}

	event := chatagent.NewUserMessageEvent("empty action received, no response.")
	if err := c.eventBus.Publish(event); err != nil {
		ui.PrintWarning(fmt.Sprintf("cannot confirm external action (tool), err: %s", err.Error()))
	}
}

func (c *ChatUI) confirmAgentMessage(ctx context.Context, message string) {
	input := ui.Input(fmt.Sprintf(`
Required to confirm agent messages:

<message>
%s
</message>

What do you think?

>>> `, message))
	event := chatagent.NewUserMessageEvent(input)
	if err := c.eventBus.Publish(event); err != nil {
		ui.PrintWarning(fmt.Sprintf("cannot confirm external action (message), err: %s", err.Error()))
	}
}

func (c *ChatUI) confirmToolCall(ctx context.Context, toolCall *llms.ToolCall) {

	userChoice := -1
	if utils.ContainString(c.autoTools, toolCall.Name) {
		userChoice = 1
	} else {
		userChoice = ui.Confirm(fmt.Sprintf(`
Required to call the tool: 
-      tool: [%s]
- arguments:
   %s
---
Options:
`, toolCall.Name, toolCall.MarshalJson()), []string{
			"Yes, execute once",
			"Yes, always",
			"No, ignore",
		})
	}

	var event *eventbus.Event
	switch userChoice {
	case 1, 2:
		if userChoice == 2 {
			c.addAutoTools(toolCall)
		}
		cancel := ui.Loading("Executing")
		if err := c.chatAgentCtrl.ExecTool(ctx, toolCall); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to execute tool, err: %s", err.Error()))
		}
		cancel()
	case 3:
		event = chatagent.NewIgnoreToolCallEvent(toolCall)
		if err := c.eventBus.Publish(event); err != nil {
			ui.PrintWarning(fmt.Sprintf("cannot confirm tool call, err: %s", err.Error()))
		}
	default:
		ui.PrintWarning("invalid input")
		return
	}
}

func (c *ChatUI) handleAgentResponseEnd(ctx context.Context, data *agent.AgentResponseEnd) {
	if data.FinishReason != llms.FinishReasonNormalEnd && data.FinishReason != llms.FinishReasonToolUse {
		ui.PrintWarning(fmt.Sprintf("Agent is not normal end: %s", data.FinishReason))
	}
	if data.Error != nil {
		ui.PrintWarning(fmt.Sprintf("[ERROR]: %s", data.Error.Error()))
	}
	if data.Abort {
		c.quit(data.Error)
	}

	if data.FinishReason == llms.FinishReasonNormalEnd {
		c.handleUserInput(ctx)
	}
}

func (c *ChatUI) handleUserInput(ctx context.Context) {
	for {
		input := ui.Input(">>>")
		if strings.HasPrefix(input, ":") {
			parts := strings.Fields(input)
			if len(parts) == 0 {
				return
			}
			_ = c.handleCommand(ctx, strings.ToLower(parts[0]))
			continue
		} else {
			c.resetLoadingCancel(ui.Loading("Pending"))
			event := chatagent.NewUserMessageEvent(input)
			if err := c.eventBus.Publish(event); err != nil {
				ui.PrintWarning(fmt.Sprintf("cannot publish event: [%s:%v], err: %s",
					event.Topic, event.Data, err.Error()))
			}
			break
		}
	}
}

// printWelcome prints the welcome message
func (c *ChatUI) printWelcome() {
	ui.Print(`
-----------------------------------------------------------
            🤖 Welcome to Agent-Go Chat!
   > Type your message or use commands starting with ':'
-----------------------------------------------------------`)
}

// printHelp prints the help message
func (c *ChatUI) printHelp() {
	ui.Print(`
📋 Available Commands:
    :help, :h            - Show this help
    :clear               - Clear conversation history
    :history             - Show conversation history
    :info                - Show configuration information
    :quit, :exit, :q, :x - Exit the application
💬 Just type your message to chat with the agent!
`)
}

func (c *ChatUI) quit(err error) {
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Quit, error: %s", err.Error()))
		os.Exit(-1)
	} else {
		ui.Print("Goodbye!\n")
		os.Exit(0)
	}
}

func (c *ChatUI) handleCommand(ctx context.Context, command string) (err error) {
	switch command {
	case ":clear":
		return c.handleClearCommand(ctx)
	case ":history":
		return c.handleHistoryCommand(ctx)
	case ":info":
		return c.handleInfoCommand(ctx)
	case ":quit", ":exit", ":q", ":x":
		c.quit(nil)
	case ":help", ":h":
		c.printHelp()
	default:
		ui.PrintWarning(fmt.Sprintf("unknown command: %s", command))
	}
	return nil
}

// handleClearCommand clears conversation history
func (c *ChatUI) handleClearCommand(ctx context.Context) error {
	if err := c.chatAgentCtrl.ClearMemory(); err != nil {
		return err
	}
	ui.Print("\nConversation history cleared.\n")
	return nil
}

// handleHistoryCommand shows conversation history
func (c *ChatUI) handleHistoryCommand(ctx context.Context) error {
	memoryItems, err := c.chatAgentCtrl.RetrieveAllMemory(ctx)
	if err != nil {
		return errors.Errorf(ErrorCodeChatUIError, "failed to get conversation history: %v", err)
	}

	if len(memoryItems) == 0 {
		ui.Print("\nNo conversation history found.\n")
	}

	// Format history
	var history strings.Builder
	history.WriteString("Conversation History:\n")
	for i, item := range memoryItems {
		if msg, ok := item.AsMessage(); ok && msg != nil {
			role := msg.Creator.Role

			// Extract text content from message parts
			var content string
			for _, part := range msg.Parts {
				if textPart, ok := part.(*llms.TextPart); ok {
					content += textPart.Text
				}
				if toolCallPart, ok := part.(*llms.ToolCall); ok {
					content += utils.Idents(fmt.Sprintf(`
<tool>
  <name>%s</name>
  <arguments>
  %s
  </arguments>
</tool>
`, toolCallPart.Name, toolCallPart.MarshalJson()), 4)
				}
				if toolCallResult, ok := part.(*llms.ToolCallResult); ok {
					content += utils.Idents(fmt.Sprintf(`
<tool>
  <name>%s</name>
  <result>
  %s
  </result>
</tool>
`, toolCallResult.Name, toolCallResult.MarshalJson()), 4)
				}
			}

			history.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, role, content))
		}
	}

	ui.Print(history.String())
	return nil
}

// handleInfoCommand shows configuration information
func (c *ChatUI) handleInfoCommand(ctx context.Context) error {
	info := c.chatAgentCtrl.GetInfo()
	ui.Print("\n" + info + "\n")

	chatUsage := journal.GetUsage("chat")
	if len(chatUsage) > 0 {
		ui.PrintKVs("chat", chatUsage)
	}
	embeddingUsage := journal.GetUsage("embedding")
	if len(chatUsage) > 0 {
		ui.PrintKVs("embedding", embeddingUsage)
	}
	return nil
}

func (c *ChatUI) addAutoTools(toolCall *llms.ToolCall) {
	ui.Print(fmt.Sprintf("    > Add [%s] to auto tools", toolCall.Name))
	c.autoTools = append(c.autoTools, toolCall.Name)
}

func (c *ChatUI) resetLoadingCancel(loadingCancel context.CancelFunc) {
	c.loadingMu.Lock()
	defer c.loadingMu.Unlock()

	if c.loadingCancel != nil {
		c.loadingCancel()
	}
	c.loadingCancel = loadingCancel
	// leave a small span for clean
	time.Sleep(200 * time.Millisecond)
}
