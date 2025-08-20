package mcp

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"time"

	amcp "github.com/oopslink/agent-go/pkg/core/mcp"
	"github.com/oopslink/agent-go/pkg/core/tools"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var ErrorCodeSetupMcpToolsFailed = errors.ErrorCode{
	Code:           93000,
	Name:           "SetupMcpToolsFailed",
	DefaultMessage: "Failed to setup mcp tools",
}

func SetupMcpTools(ctx context.Context) ([]tools.Tool, error) {
	server := newMcpServer()

	clientRef, serverRef, err := amcp.WithInMemory()
	if err != nil {
		return nil, errors.Errorf(ErrorCodeSetupMcpToolsFailed,
			"failed to create in-memory transport: %s", err.Error())
	}

	// run the server and capture the exit error
	go func() {
		fmt.Println("# Starting MCP server for tools ...")
		defer fmt.Println("# MCP server for tools stopped")

		transport, err := serverRef.CreateTransport()
		if err != nil {
			fmt.Printf("< Failed to create transport: %v\n", err)
			return
		}

		if err = server.Run(ctx, transport); err != nil {
			fmt.Printf("< MCP server error: %v\n", err)
		}
		fmt.Println("< MCP server exited")
	}()

	// create a client that connects to the server
	return amcp.WithMcpTools(clientRef)
}

func newMcpServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "demo-mcp-tools", Version: "v0.0.1"}, nil)
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "get_current_time",
			Description: "get current time",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"timezone": {
						Type:        "string",
						Description: "Timezone to get the current time for",
					},
				},
				Required: []string{"timezone"},
			},
		},
		getCurrentTime,
	)
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "get_location",
			Description: "get the location of a city",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"city": {
						Type:        "string",
						Description: "City to get the current location for",
					},
				},
				Required: []string{"city"},
			},
		},
		getLocation,
	)
	return server
}

type getCurrentTimeParams struct {
	TimeZone string `json:"timezone,omitempty"`
}

func getCurrentTime(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[getCurrentTimeParams]) (*mcp.CallToolResultFor[any], error) {
	loc, _ := time.LoadLocation(params.Arguments.TimeZone)

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Current time in " + params.Arguments.TimeZone + " is " + time.Now().In(loc).Format(time.RFC3339)},
		},
	}, nil
}

type getLocationParams struct {
	City string `json:"city,omitempty"`
}

func getLocation(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[getLocationParams]) (*mcp.CallToolResultFor[any], error) {
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Current location in " + params.Arguments.City + " is Latitude: 0, Longitude: 0"},
		},
	}, nil
}
