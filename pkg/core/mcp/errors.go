package mcp

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeCreateMcpToolFailed = errors.ErrorCode{
		Code:           20200,
		Name:           "CreateMcpToolFailed",
		DefaultMessage: "Failed to create mcp tool",
	}
)
