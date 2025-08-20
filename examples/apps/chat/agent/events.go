package agent

import (
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	EventTypeUserInput = "chat:user_input"
)

type UserInput struct {
	Message string

	ToolCall       *llms.ToolCall
	IgnoreToolCall bool
}

func NewUserMessageEvent(message string) *eventbus.Event {
	return eventbus.NewEvent(EventTypeUserInput, &UserInput{
		Message: message,
	})
}

func NewIgnoreToolCallEvent(toolCall *llms.ToolCall) *eventbus.Event {
	return eventbus.NewEvent(EventTypeUserInput, &UserInput{
		ToolCall:       toolCall,
		IgnoreToolCall: true,
	})
}
