package agent

import (
	"context"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

type GenerateContextParams struct {
	UserRequest    *UserRequest
	ToolCallResult *llms.ToolCallResult
	ChatOptions    []llms.ChatOption
}

func (p *GenerateContextParams) ToMessages() []*llms.Message {
	var messages []*llms.Message
	if p.UserRequest != nil {
		messages = append(messages, llms.NewUserMessage(p.UserRequest.Message))
	}
	if p.ToolCallResult != nil {
		messages = append(messages, llms.NewToolCallResultMessage(p.ToolCallResult, time.Now()))
	}
	return messages
}

type GeneratedContext struct {
	Messages []*llms.Message
	Options  []llms.ChatOption
}

type Context interface {
	AgentId() string
	SystemPrompt() string

	GetModel() *llms.Model

	Generate(ctx context.Context, params *GenerateContextParams) (*GeneratedContext, error)

	UpdateMemory(ctx context.Context, messages ...*llms.Message) error

	GetState() AgentState

	CallTool(ctx context.Context, call *llms.ToolCall) (*llms.ToolCallResult, error)
	CanAutoCall(toolCall *llms.ToolCall) bool
	ValidateToolCall(call *llms.ToolCall) error
}
