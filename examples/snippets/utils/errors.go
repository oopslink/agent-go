package utils

import "github.com/oopslink/agent-go/pkg/commons/errors"

var ErrorCodeSetupMcpToolsFailed = errors.ErrorCode{
	Code:           93000,
	Name:           "SetupMcpToolsFailed",
	DefaultMessage: "Failed to setup mcp tools",
}
