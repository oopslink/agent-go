package mcp

import (
	"context"
	"testing"

	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestMcpTool_Descriptor(t *testing.T) {
	descriptor := &llms.ToolDescriptor{
		Name:        "test-tool",
		Description: "A test MCP tool",
		Parameters: &llms.Schema{
			Type: "object",
			Properties: map[string]*llms.Schema{
				"input": {
					Type:        "string",
					Description: "Input parameter",
				},
			},
		},
	}

	// Create in-memory transports for testing
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Create a test server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	// Create a client and connect
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	// Start server session (in background)
	go func() {
		serverSession, _ := server.Connect(context.Background(), serverTransport)
		defer serverSession.Close()
		serverSession.Wait()
	}()

	// Connect client
	session, err := client.Connect(context.Background(), clientTransport)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer session.Close()

	mcpTool, err := newMcpTool(descriptor, session)
	if err != nil {
		t.Fatalf("Failed to create MCP tool: %v", err)
	}

	result := mcpTool.Descriptor()
	if result == nil {
		t.Fatal("Descriptor should not be nil")
	}

	if result.Name != "test-tool" {
		t.Errorf("Expected name 'test-tool', got '%s'", result.Name)
	}

	if result.Description != "A test MCP tool" {
		t.Errorf("Expected description 'A test MCP tool', got '%s'", result.Description)
	}
}

func TestMcpTool_NewMcpTool_NilDescriptor(t *testing.T) {
	_, err := newMcpTool(nil, nil)
	if err == nil {
		t.Error("Expected error for nil descriptor")
	}
	expectedError := "[ERR,CreateMcpToolFailed]: tool descriptor cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected '%s', got '%s'", expectedError, err.Error())
	}
}

func TestMcpTool_NewMcpTool_NilSession(t *testing.T) {
	descriptor := &llms.ToolDescriptor{
		Name: "test-tool",
	}

	_, err := newMcpTool(descriptor, nil)
	if err == nil {
		t.Error("Expected error for nil session")
	}
	expectedError := "[ERR,CreateMcpToolFailed]: client session cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected '%s', got '%s'", expectedError, err.Error())
	}
}

func TestWithMcpTools_InvalidEndpoint(t *testing.T) {
	// Create a mock McpServerRef that will fail
	mockRef := &mockMcpServerRef{shouldFail: true}

	_, err := WithMcpTools(mockRef)
	if err == nil {
		t.Error("Expected error for failing transport")
	}
}

// mockMcpServerRef is a mock implementation for testing
type mockMcpServerRef struct {
	shouldFail bool
}

func (m *mockMcpServerRef) CreateTransport() (mcp.Transport, error) {
	if m.shouldFail {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "mock transport creation failed")
	}
	// Return a working in-memory transport for successful cases
	clientTransport, _ := mcp.NewInMemoryTransports()
	return clientTransport, nil
}

func TestConvertMcpToolToDescriptor(t *testing.T) {
	mcpTool := &mcp.Tool{
		Name:        "test-mcp-tool",
		Description: "Test MCP tool description",
	}

	descriptor := convertMcpToolToDescriptor(mcpTool)

	if descriptor.Name != "test-mcp-tool" {
		t.Errorf("Expected name 'test-mcp-tool', got '%s'", descriptor.Name)
	}

	if descriptor.Description != "Test MCP tool description" {
		t.Errorf("Expected description 'Test MCP tool description', got '%s'", descriptor.Description)
	}

	// Parameters should be nil when the MCP tool doesn't have an input schema
	if descriptor.Parameters != nil {
		t.Error("Expected parameters to be nil when MCP tool has no input schema")
	}
}
