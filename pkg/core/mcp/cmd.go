package mcp

import (
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oopslink/agent-go/pkg/commons/errors"
)

var _ McpServerRef = &commandTransportMcpServerRef{}

// WithCommand creates a command transport MCP server reference
func WithCommand(cmd *exec.Cmd) (*commandTransportMcpServerRef, error) {
	if cmd == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "command cannot be nil")
	}
	return &commandTransportMcpServerRef{cmd: cmd}, nil
}

type commandTransportMcpServerRef struct {
	cmd *exec.Cmd
}

func (c *commandTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewCommandTransport(c.cmd), nil
}
