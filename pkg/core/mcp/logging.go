package mcp

import (
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"io"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ McpServerRef = &loggingTransportMcpServerRef{}

// WithLogging creates a logging transport MCP server reference that wraps another transport
func WithLogging(delegate mcp.Transport, writer io.Writer) (*loggingTransportMcpServerRef, error) {
	if delegate == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "delegate transport cannot be nil")
	}
	if writer == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "writer cannot be nil")
	}
	return &loggingTransportMcpServerRef{
		delegate: delegate,
		writer:   writer,
	}, nil
}

type loggingTransportMcpServerRef struct {
	delegate mcp.Transport
	writer   io.Writer
}

func (l *loggingTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewLoggingTransport(l.delegate, l.writer), nil
}
