package mcp

import (
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ McpServerRef = &sseServerTransportMcpServerRef{}
var _ McpServerRef = &sseClientTransportMcpServerRef{}

// WithSSEServer creates an SSE server transport MCP server reference
func WithSSEServer(endpoint string, writer http.ResponseWriter) (*sseServerTransportMcpServerRef, error) {
	if endpoint == "" {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "endpoint cannot be empty")
	}
	if writer == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "response writer cannot be nil")
	}
	return &sseServerTransportMcpServerRef{
		endpoint: endpoint,
		writer:   writer,
	}, nil
}

type sseServerTransportMcpServerRef struct {
	endpoint string
	writer   http.ResponseWriter
}

func (s *sseServerTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewSSEServerTransport(s.endpoint, s.writer), nil
}

// WithSSEClient creates an SSE client transport MCP server reference
func WithSSEClient(baseURL string, opts *mcp.SSEClientTransportOptions) (*sseClientTransportMcpServerRef, error) {
	if baseURL == "" {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "base URL cannot be empty")
	}
	return &sseClientTransportMcpServerRef{
		baseURL: baseURL,
		opts:    opts,
	}, nil
}

type sseClientTransportMcpServerRef struct {
	baseURL string
	opts    *mcp.SSEClientTransportOptions
}

func (s *sseClientTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return mcp.NewSSEClientTransport(s.baseURL, s.opts), nil
}
