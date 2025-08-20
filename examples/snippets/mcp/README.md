# MCP (Model Context Protocol) Agent Example

This example demonstrates how to create an AI agent that uses MCP (Model Context Protocol) tools to perform specific tasks. In this case, the agent has access to a time tool that can get the current time in various formats through an MCP server.

## Features

- **MCP Server Integration**: Uses a proper MCP server to provide tools
- **Time Tool**: Get current time in various formats (RFC3339, Unix timestamp, ISO8601, Custom)
- **Timezone Support**: Specify different timezones for time retrieval
- **Streaming Responses**: Real-time response generation with streaming output
- **Multiple LLM Providers**: Support for OpenAI, Anthropic, and Gemini models

## Architecture

The example uses a proper MCP server architecture:

1. **MCP Server**: Runs in the background and provides tools via the MCP protocol
2. **MCP Client**: Connects to the server and exposes tools to the agent
3. **Agent Integration**: The agent can call tools through the MCP client
4. **Tool Execution**: Tools are executed on the MCP server and results are returned

## Usage

1. Set your API key:
   ```bash
   export API_KEY="your-api-key-here"
   ```

2. Optionally set the provider and model:
   ```bash
   export PROVIDER="openai"  # or "anthropic", "gemini"
   export MODEL="gpt-4o-mini"  # or any other model
   ```

3. Run the example:
   ```bash
   go run *.go "What time is it?"
   ```

## Example Questions

Try these example questions to see the MCP tool usage in action:

- "What time is it?"
- "Get the current time in Unix format"
- "What's the time in New York?"
- "Show me the time in ISO8601 format"
- "What's the current timestamp?"

## How It Works

1. **MCP Server Setup**: The `SetupMcpTools` function creates an MCP server with time tools
2. **Server Registration**: Tools are registered with the MCP server using proper MCP protocol
3. **Client Connection**: An MCP client connects to the server and exposes tools to the agent
4. **Tool Execution**: When the agent decides to use a tool, it calls the tool through the MCP client
5. **Response Generation**: The tool result is returned through the MCP protocol and used by the agent

## Tool Capabilities

The time tool supports:

- **Multiple Formats**: RFC3339, Unix timestamp, ISO8601, Custom format
- **Timezone Support**: Specify timezones like 'UTC', 'America/New_York', 'Europe/London'
- **Rich Results**: Returns formatted time through the MCP protocol

## MCP Server Architecture

The MCP server follows this workflow:

1. **Server Initialization**: Creates an MCP server with tool definitions
2. **Tool Registration**: Registers tools with proper JSON schemas
3. **Transport Setup**: Uses in-memory transport for client-server communication
4. **Background Execution**: Server runs in the background
5. **Client Connection**: MCP client connects and discovers available tools
6. **Tool Discovery**: Agent discovers tools through the MCP client
7. **Tool Execution**: Agent calls tools through the MCP protocol

## Customization

To add your own MCP tools:

1. Add new tools to the `newMcpServer` function:
   ```go
   mcp.AddTool(server,
       &mcp.Tool{
           Name:        "my_tool",
           Description: "Description of my tool",
           InputSchema: &jsonschema.Schema{
               // Define tool parameters
           },
       },
       myToolHandler,
   )
   ```

2. Implement the tool handler function:
   ```go
   type myToolParams struct {
       Param1 string `json:"param1,omitempty"`
   }

   func myToolHandler(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[myToolParams]) (*mcp.CallToolResultFor[any], error) {
       // Implement tool logic
       return &mcp.CallToolResultFor[any]{
           Content: []mcp.Content{
               &mcp.TextContent{Text: "Tool result"},
           },
       }, nil
   }
   ```

3. The tools will automatically be available to the agent through the MCP client

## Dependencies

- agent-go core packages
- MCP SDK for server and client implementation
- LLM provider SDKs (OpenAI, Anthropic, Gemini)
- Event bus for communication
- Memory system for conversation history 