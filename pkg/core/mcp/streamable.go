package mcp

import (
	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ McpServerRef = &streamableServerTransportMcpServerRef{}
var _ McpServerRef = &streamableClientTransportMcpServerRef{}

// WithStreamableServer creates a streamable server transport MCP server reference
func WithStreamableServer(sessionID string) (*streamableServerTransportMcpServerRef, error) {
	if sessionID == "" {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "session ID cannot be empty")
	}
	return &streamableServerTransportMcpServerRef{sessionID: sessionID}, nil
}

type streamableServerTransportMcpServerRef struct {
	sessionID string
}

func (s *streamableServerTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewStreamableServerTransport(s.sessionID), nil
}

// WithStreamableClient creates a streamable client transport MCP server reference
func WithStreamableClient(url string, opts *mcp.StreamableClientTransportOptions) (*streamableClientTransportMcpServerRef, error) {
	if url == "" {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "URL cannot be empty")
	}
	return &streamableClientTransportMcpServerRef{
		url:  url,
		opts: opts,
	}, nil
}

type streamableClientTransportMcpServerRef struct {
	url  string
	opts *mcp.StreamableClientTransportOptions
}

func (s *streamableClientTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewStreamableClientTransport(s.url, s.opts), nil
}
