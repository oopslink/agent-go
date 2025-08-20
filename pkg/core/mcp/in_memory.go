package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ McpServerRef = &inMemoryTransportMcpServerRef{}

// WithInMemory creates an in-memory transport MCP server reference
// Note: This returns a pair of transports - one for client, one for server
func WithInMemory() (*inMemoryTransportMcpServerRef, *inMemoryTransportMcpServerRef, error) {
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	clientRef := &inMemoryTransportMcpServerRef{transport: clientTransport}
	serverRef := &inMemoryTransportMcpServerRef{transport: serverTransport}

	return clientRef, serverRef, nil
}

type inMemoryTransportMcpServerRef struct {
	transport *mcp.InMemoryTransport
}

func (i *inMemoryTransportMcpServerRef) CreateTransport() (mcp.Transport, error) {
	return i.transport, nil
}
