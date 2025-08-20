package agent

import (
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	EventTypeUserRequest          = "agent:user_request"
	EventTypeAgentMessage         = "agent:agent_message"
	EventTypeExternalAction       = "agent:external_action"
	EventTypeExternalActionResult = "agent:external_action_result"
	EventTypeAgentResponseStart   = "agent:agent_response_start"
	EventTypeAgentResponseEnd     = "agent:agent_response_end"
)

func NewUserRequestEvent(userRequest *UserRequest) *eventbus.Event {
	return eventbus.NewEvent(EventTypeUserRequest, userRequest)
}

func GetUserRequestEventData(event *eventbus.Event) *UserRequest {
	return event.Data.(*UserRequest)
}

func NewAgentMessageEvent(traceId string, message *llms.Message) *eventbus.Event {
	return eventbus.NewEvent(EventTypeAgentMessage,
		&AgentMessage{
			TraceId: traceId,
			Message: message,
		})
}

func GetAgentMessageEventData(event *eventbus.Event) *AgentMessage {
	return event.Data.(*AgentMessage)
}

func NewExternalActionEvent(message string) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalAction,
		&ExternalAction{Message: message})
}

func NewToolCallEvent(toolCall *llms.ToolCall) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalAction,
		&ExternalAction{ToolCall: toolCall})
}

func GetToolCallEventData(event *eventbus.Event) *llms.ToolCall {
	if result, ok := event.Data.(*ExternalAction); ok {
		return result.ToolCall
	}
	return nil
}

func GetExternalActionEventData(event *eventbus.Event) *ExternalAction {
	return event.Data.(*ExternalAction)
}

func NewExternalActionResultEvent(message string) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalActionResult,
		&ExternalActionResult{Message: message})
}

func NewToolCallResultEvent(toolCallResult *llms.ToolCallResult) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalActionResult,
		&ExternalActionResult{ToolCallResult: toolCallResult})
}

func NewUserSkipToolCallEvent(toolCallId, toolName string) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalActionResult,
		&ExternalActionResult{
			ToolCallResult: &llms.ToolCallResult{
				ToolCallId: toolCallId,
				Name:       toolName,
				Result: map[string]any{
					"state":  "UserSkipped",
					"reason": "the tool not called, rejected by user",
				},
			},
		},
	)
}

func NewFailedToolCallEvent(toolCallId, toolName string, err error) *eventbus.Event {
	return eventbus.NewEvent(EventTypeExternalActionResult,
		&ExternalActionResult{
			ToolCallResult: &llms.ToolCallResult{
				ToolCallId: toolCallId,
				Name:       toolName,
				Result: map[string]any{
					"state":  "InvokeFailed",
					"reason": "failed to call the tool, cause: " + err.Error(),
				},
			},
		},
	)
}

func GetToolCallResultEventData(event *eventbus.Event) *llms.ToolCallResult {
	if result, ok := event.Data.(*ExternalActionResult); ok {
		return result.ToolCallResult
	}
	return nil
}

func NewAgentResponseStartEvent(traceId string) *eventbus.Event {
	return eventbus.NewEvent(EventTypeAgentResponseStart, &AgentResponseStart{TraceId: traceId})
}

func NewAgentResponseEndEvent(traceId string, end *AgentResponseEnd) *eventbus.Event {
	end.TraceId = traceId
	return eventbus.NewEvent(EventTypeAgentResponseEnd, end)
}

func GetAgentResponseEndEventData(event *eventbus.Event) *AgentResponseEnd {
	return event.Data.(*AgentResponseEnd)
}

type UserRequest struct {
	Message string
	Options []llms.ChatOption
}

type ExternalActionResult struct {
	Message        string
	ToolCallResult *llms.ToolCallResult
}

type ExternalAction struct {
	Message  string
	ToolCall *llms.ToolCall
}

type AgentMessage struct {
	TraceId string
	Message *llms.Message
}

type AgentResponseStart struct {
	TraceId string
}

type AgentResponseEnd struct {
	TraceId      string
	Error        error
	Abort        bool
	FinishReason llms.FinishReason
}
