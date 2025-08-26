package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// StartMcpServer initializes and starts the MCP server with the given transport and tools.
func StartMcpServer(
	ctx context.Context, name, version string,
	transport mcp.Transport, toolList ...tools.Tool) error {
	// Create a new MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	// Register each tool with the MCP server
	for _, tool := range toolList {
		if err := registerToolWithMcpServer(server, tool); err != nil {
			return err
		}
	}

	// Run the server with the provided transport in a goroutine
	go func() {
		if err := server.Run(ctx, transport); err != nil {
			journal.Error("mcp", "StartMcpServer", "MCP server failed to run", "error", err)
		}
	}()

	return nil
}

// registerToolWithMcpServer converts a tools.Tool to MCP format and registers it with the server
func registerToolWithMcpServer(server *mcp.Server, tool tools.Tool) error {
	descriptor := tool.Descriptor()
	if descriptor == nil {
		return nil // Skip tools without descriptors
	}

	// Convert tool descriptor to MCP tool format
	mcpTool := &mcp.Tool{
		Name:        descriptor.Name,
		Description: descriptor.Description,
		InputSchema: convertSchemaToMcpSchema(descriptor.Parameters),
	}

	// Create a handler function that wraps the original tool
	handler := func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
		// Convert MCP call to our tool call format
		var arguments map[string]any
		if params.Arguments != nil {
			if argsMap, ok := params.Arguments.(map[string]any); ok {
				arguments = argsMap
			} else {
				// If it's not a map, try to convert to map through JSON
				if argsBytes, err := json.Marshal(params.Arguments); err == nil {
					if err := json.Unmarshal(argsBytes, &arguments); err != nil {
						arguments = make(map[string]any)
					}
				} else {
					arguments = make(map[string]any)
				}
			}
		} else {
			arguments = make(map[string]any)
		}

		toolCall := &llms.ToolCall{
			Name:      params.Name,
			Arguments: arguments,
		}

		// Call the original tool
		result, err := tool.Call(ctx, toolCall)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: " + err.Error()},
				},
				IsError: true,
			}, err
		}

		// Convert tool result to MCP format
		var content []mcp.Content

		// Try to extract meaningful content from the result
		if result != nil && result.Result != nil {
			// Convert the map result to JSON string for display
			if resultBytes, err := json.Marshal(result.Result); err == nil {
				content = append(content, &mcp.TextContent{Text: string(resultBytes)})
			} else {
				content = append(content, &mcp.TextContent{Text: "Tool executed successfully"})
			}
		} else {
			content = append(content, &mcp.TextContent{Text: "Tool executed successfully"})
		}

		return &mcp.CallToolResult{
			Content: content,
			IsError: false,
		}, nil
	}

	// Register the tool with the MCP server
	mcp.AddTool(server, mcpTool, handler)
	return nil
}

// convertSchemaToMcpSchema converts our llms.Schema to MCP jsonschema.Schema
func convertSchemaToMcpSchema(schema *llms.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	mcpSchema := &jsonschema.Schema{}

	// Convert type
	if schema.Type != "" {
		mcpSchema.Type = convertTypeToMcpType(schema.Type)
	}

	// Convert description
	if schema.Description != "" {
		mcpSchema.Description = schema.Description
	}

	// Convert properties for object types
	if schema.Properties != nil {
		mcpSchema.Properties = make(map[string]*jsonschema.Schema)
		for key, propSchema := range schema.Properties {
			mcpSchema.Properties[key] = convertSchemaToMcpSchema(propSchema)
		}
	}

	// Convert required fields
	if len(schema.Required) > 0 {
		mcpSchema.Required = make([]string, len(schema.Required))
		copy(mcpSchema.Required, schema.Required)
	}

	// Convert array items schema
	if schema.Items != nil {
		mcpSchema.Items = convertSchemaToMcpSchema(schema.Items)
	}

	return mcpSchema
}

// convertTypeToMcpType converts our schema type to MCP schema type
func convertTypeToMcpType(schemaType llms.SchemaType) string {
	switch schemaType {
	case llms.TypeObject:
		return "object"
	case llms.TypeArray:
		return "array"
	case llms.TypeString:
		return "string"
	case llms.TypeBoolean:
		return "boolean"
	case llms.TypeNumber:
		return "number"
	case llms.TypeInteger:
		return "integer"
	default:
		return "string" // Default to string for unknown types
	}
}
