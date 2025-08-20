package agent

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeInvalidInputEvent = errors.ErrorCode{
		Code:           20000,
		Name:           "InvalidInputEvent",
		DefaultMessage: "invalid input event type",
	}
	ErrorCodeChatSessionAbort = errors.ErrorCode{
		Code:           20001,
		Name:           "ChatSessionAbort",
		DefaultMessage: "Chat session abort",
	}
	ErrorCodeChatSessionFailed = errors.ErrorCode{
		Code:           20002,
		Name:           "ChatSessionFailed",
		DefaultMessage: "Chat session failed",
	}
	ErrorCodeGenerateContextFailed = errors.ErrorCode{
		Code:           20003,
		Name:           "GenerateContextFailed",
		DefaultMessage: "Failed to generate context",
	}
	ErrorCodeInvalidToolCall = errors.ErrorCode{
		Code:           20004,
		Name:           "InvalidToolCall",
		DefaultMessage: "Tool call is not invalid",
	}
	ErrorCodeLoadPlanFailed = errors.ErrorCode{
		Code:           20005,
		Name:           "LoadPlanFailed",
		DefaultMessage: "Failed to load plan",
	}
)
