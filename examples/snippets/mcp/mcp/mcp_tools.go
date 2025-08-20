package mcp

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go-snippets/utils"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"time"

	amcp "github.com/oopslink/agent-go/pkg/core/mcp"
	"github.com/oopslink/agent-go/pkg/core/tools"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func SetupMcpTools(ctx context.Context) ([]tools.Tool, error) {
	server := newMcpServer()

	clientRef, serverRef, err := amcp.WithInMemory()
	if err != nil {
		return nil, errors.Errorf(utils.ErrorCodeSetupMcpToolsFailed,
			"failed to create in-memory transport: %s", err.Error())
	}

	// run the server and capture the exit error
	go func() {
		fmt.Println("* Starting MCP server for time tools")
		defer fmt.Println("MCP server for time tools stopped")

		transport, err := serverRef.CreateTransport()
		if err != nil {
			fmt.Printf("Failed to create transport: %v\n", err)
			return
		}

		if err := server.Run(ctx, transport); err != nil {
			fmt.Printf("MCP server error: %v\n", err)
		}
		fmt.Println("MCP server exited")
	}()

	// create a client that connects to the server
	return amcp.WithMcpTools(clientRef)
}

func newMcpServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "time-mcp-tools", Version: "v0.0.1"}, nil)
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "get_current_time",
			Description: "Get the current time in various formats",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"format": {
						Type:        "string",
						Description: "Time format to return (RFC3339, Unix, ISO8601, Custom)",
					},
					"timezone": {
						Type:        "string",
						Description: "Timezone to use (e.g., 'UTC', 'America/New_York')",
					},
				},
			},
		},
		getCurrentTime,
	)
	return server
}

type getCurrentTimeParams struct {
	Format   string `json:"format,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

func getCurrentTime(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[getCurrentTimeParams]) (*mcp.CallToolResultFor[any], error) {
	// Parse parameters
	format := "RFC3339"
	timezone := "UTC"

	if params.Arguments.Format != "" {
		format = params.Arguments.Format
	}
	if params.Arguments.Timezone != "" {
		timezone = params.Arguments.Timezone
	}

	// Get current time
	now := time.Now()

	// Apply timezone if specified
	if timezone != "UTC" {
		if loc, err := time.LoadLocation(timezone); err == nil {
			now = now.In(loc)
		}
	}

	// Format time based on requested format
	var timeStr string
	switch format {
	case "Unix":
		timeStr = fmt.Sprintf("%d", now.Unix())
	case "ISO8601":
		timeStr = now.Format("2006-01-02T15:04:05Z07:00")
	case "Custom":
		timeStr = now.Format("January 2, 2006 at 3:04 PM MST")
	default: // RFC3339
		timeStr = now.Format(time.RFC3339)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Current time: %s", timeStr)},
		},
	}, nil
}
