package tools

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeToolNotFound = errors.ErrorCode{
		Code:           20300,
		Name:           "ToolNotFound",
		DefaultMessage: "Tool not found",
	}
	ErrorCodeToolCallFailed = errors.ErrorCode{
		Code:           20301,
		Name:           "ToolCallFailed",
		DefaultMessage: "Failed to call tool",
	}
)
