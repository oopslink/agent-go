package mcp

import (
	"context"
	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

var (
	_ McpServerRef = &commandTransportMcpServerRef{}
	_ McpServerRef = &sseServerTransportMcpServerRef{}
	_ McpServerRef = &sseClientTransportMcpServerRef{}
	_ McpServerRef = &streamableServerTransportMcpServerRef{}
	_ McpServerRef = &streamableClientTransportMcpServerRef{}
	_ McpServerRef = &stdioTransportMcpServerRef{}
	_ McpServerRef = &inMemoryTransportMcpServerRef{}
	_ McpServerRef = &loggingTransportMcpServerRef{}
)

type McpServerRef interface {
	CreateTransport() (mcp.Transport, error)
}

func WithMcpTools(mcpServerRef McpServerRef) ([]tools.Tool, error) {
	// Create a new MCP client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "agent-go-client",
		Version: "1.0.0",
	}, nil)

	// Create a transport based on the endpoint
	transport, err := mcpServerRef.CreateTransport()
	if err != nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed,
			"failed to create transport: %s", err.Error())
	}

	// Connect to the MCP server
	ctx := context.Background()
	session, err := client.Connect(ctx, transport)
	if err != nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed,
			"failed to connect to MCP server: %s", err.Error())
	}

	// Use the session-based function to create tools
	mcpTools, err := WithMcpToolsFromSession(session)
	if err != nil {
		session.Close()
		return nil, err
	}

	return mcpTools, nil
}

// WithMcpToolsFromSession creates MCP tools from an existing MCP client session
// This is useful when you already have an established MCP connection
func WithMcpToolsFromSession(session *mcp.ClientSession) ([]tools.Tool, error) {
	if session == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "session cannot be nil")
	}

	// List available tools from the MCP server
	ctx := context.Background()
	toolsResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed,
			"failed to list tools from MCP server: %s", err.Error())
	}

	// Convert MCP tools to our McpTool instances
	var mcpTools []tools.Tool
	for _, toolDescription := range toolsResult.Tools {
		descriptor := convertMcpToolToDescriptor(toolDescription)
		mcpTool, err := newMcpTool(descriptor, session)
		if err != nil {
			continue // Skip tools that can't be converted
		}
		mcpTools = append(mcpTools, mcpTool)
	}

	return mcpTools, nil
}

// convertMcpToolToDescriptor converts an MCP Tool to our ToolDescriptor format
func convertMcpToolToDescriptor(mcpTool *mcp.Tool) *llms.ToolDescriptor {
	descriptor := &llms.ToolDescriptor{
		Name:        mcpTool.Name,
		Description: mcpTool.Description,
	}

	// Convert MCP input schema to our schema format if available
	if mcpTool.InputSchema != nil {
		descriptor.Parameters = convertMcpSchema(mcpTool.InputSchema)
	}

	return descriptor
}

// convertMcpSchema converts MCP jsonschema.Schema to our llms.Schema
func convertMcpSchema(mcpSchema *jsonschema.Schema) *llms.Schema {
	if mcpSchema == nil {
		return nil
	}

	schema := &llms.Schema{}

	// Convert basic type information
	if mcpSchema.Type != "" {
		schema.Type = convertMcpType(mcpSchema.Type)
	} else if len(mcpSchema.Types) > 0 && len(mcpSchema.Types) == 1 {
		// If Types has only one element, use it as Type
		schema.Type = convertMcpType(mcpSchema.Types[0])
	}

	// Convert description
	if mcpSchema.Description != "" {
		schema.Description = mcpSchema.Description
	}

	// Convert properties for object types
	if mcpSchema.Properties != nil {
		schema.Properties = make(map[string]*llms.Schema)
		for key, propSchema := range mcpSchema.Properties {
			schema.Properties[key] = convertMcpSchema(propSchema)
		}
	}

	// Convert required fields
	if len(mcpSchema.Required) > 0 {
		schema.Required = make([]string, len(mcpSchema.Required))
		copy(schema.Required, mcpSchema.Required)
	}

	// Convert array items schema
	if mcpSchema.Items != nil {
		schema.Items = convertMcpSchema(mcpSchema.Items)
	}

	return schema
}

// convertMcpType converts MCP schema type to our schema type
func convertMcpType(mcpType string) llms.SchemaType {
	switch mcpType {
	case "object":
		return llms.TypeObject
	case "array":
		return llms.TypeArray
	case "string":
		return llms.TypeString
	case "boolean":
		return llms.TypeBoolean
	case "number":
		return llms.TypeNumber
	case "integer":
		return llms.TypeInteger
	default:
		// Default to string for unknown types
		return llms.TypeString
	}
}

func newMcpTool(descriptor *llms.ToolDescriptor, session *mcp.ClientSession) (*McpTool, error) {
	if descriptor == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "tool descriptor cannot be nil")
	}
	if session == nil {
		return nil, errors.Errorf(ErrorCodeCreateMcpToolFailed, "client session cannot be nil")
	}
	return &McpTool{
		descriptor: descriptor,
		session:    session,
	}, nil
}

var _ tools.Tool = &McpTool{}

// McpTool is a tool that interacts with the Model Context Protocol (MCP) client.
type McpTool struct {
	descriptor *llms.ToolDescriptor
	session    *mcp.ClientSession
}

func (t *McpTool) Descriptor() *llms.ToolDescriptor {
	return t.descriptor
}

func (t *McpTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Convert tool call to MCP format
	mcpParams := &mcp.CallToolParams{
		Name:      params.Name,
		Arguments: params.Arguments,
	}

	// Call the MCP tool through the session
	result, err := t.session.CallTool(ctx, mcpParams)
	if err != nil {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"error":   err.Error(),
				"success": false,
			},
		}, err
	}

	// Convert MCP result to our ToolCallResult format
	resultData := map[string]any{
		"success": !result.IsError,
	}

	// Add content from MCP result
	if len(result.Content) > 0 {
		var contentTexts []string
		for _, content := range result.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				contentTexts = append(contentTexts, textContent.Text)
			}
		}
		if len(contentTexts) > 0 {
			resultData["content"] = contentTexts
		}
	}

	// Add structured content if available
	if result.StructuredContent != nil {
		resultData["structured_content"] = result.StructuredContent
	}

	// Add error information if the tool failed
	if result.IsError {
		resultData["error"] = "Tool execution failed"
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result:     resultData,
	}, nil
}

// Close closes the underlying MCP session
func (t *McpTool) Close() error {
	if t.session != nil {
		return t.session.Close()
	}
	return nil
}
