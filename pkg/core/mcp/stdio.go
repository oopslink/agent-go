package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ McpServerRef = &stdioTransportMcpServerRef{}

// WithStdio creates a stdio transport MCP server reference
func WithStdio() (*stdioTransportMcpServerRef, error) {
	return &stdioTransportMcpServerRef{}, nil
}

type stdioTransportMcpServerRef struct {
}

func (s *stdioTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewStdioTransport(), nil
}
