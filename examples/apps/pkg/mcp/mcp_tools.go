package mcp

import (
	"context"
	"fmt"
	"time"

	amcp "github.com/oopslink/agent-go/pkg/core/mcp"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

var ErrorCodeSetupMcpToolsFailed = errors.ErrorCode{
	Code:           93000,
	Name:           "SetupMcpToolsFailed",
	DefaultMessage: "Failed to setup mcp tools",
}

func SetupMcpTools(ctx context.Context) ([]tools.Tool, error) {
	// Create the tools that we want to serve via MCP
	toolList := []tools.Tool{
		newGetCurrentTimeTool(),
		newGetLocationTool(),
	}

	// Create in-memory transport for communication
	clientRef, serverRef, err := amcp.WithInMemory()
	if err != nil {
		return nil, errors.Errorf(ErrorCodeSetupMcpToolsFailed,
			"failed to create in-memory transport: %s", err.Error())
	}

	// Get the server transport
	serverTransport, err := serverRef.CreateTransport()
	if err != nil {
		return nil, errors.Errorf(ErrorCodeSetupMcpToolsFailed,
			"failed to create server transport: %s", err.Error())
	}

	// Start the MCP server with our tools
	fmt.Println("# Starting MCP server for tools ...")
	if err := amcp.StartMcpServer(ctx, "demo-mcp-tools", "v0.0.1", serverTransport, toolList...); err != nil {
		return nil, errors.Errorf(ErrorCodeSetupMcpToolsFailed,
			"failed to start MCP server: %s", err.Error())
	}

	// create a client that connects to the server
	return amcp.WithMcpTools(clientRef)
}

// newGetCurrentTimeTool creates a tool that gets the current time
func newGetCurrentTimeTool() tools.Tool {
	return &getCurrentTimeTool{}
}

type getCurrentTimeTool struct{}

func (t *getCurrentTimeTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "get_current_time",
		Description: "get current time",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"timezone": {
					Type:        llms.TypeString,
					Description: "Timezone to get the current time for",
				},
			},
			Required: []string{"timezone"},
		},
	}
}

func (t *getCurrentTimeTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	timezone, ok := params.Arguments["timezone"].(string)
	if !ok {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	result := "Current time in " + timezone + " is " + time.Now().In(loc).Format(time.RFC3339)

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"time":     time.Now().In(loc).Format(time.RFC3339),
			"timezone": timezone,
			"message":  result,
		},
	}, nil
}

// newGetLocationTool creates a tool that gets the location of a city
func newGetLocationTool() tools.Tool {
	return &getLocationTool{}
}

type getLocationTool struct{}

func (t *getLocationTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "get_location",
		Description: "get the location of a city",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"city": {
					Type:        llms.TypeString,
					Description: "City to get the current location for",
				},
			},
			Required: []string{"city"},
		},
	}
}

func (t *getLocationTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	city, ok := params.Arguments["city"].(string)
	if !ok {
		city = "Unknown"
	}

	result := "Current location in " + city + " is Latitude: 0, Longitude: 0"

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"city":      city,
			"latitude":  0,
			"longitude": 0,
			"message":   result,
		},
	}, nil
}
