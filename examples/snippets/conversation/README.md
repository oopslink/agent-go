# Conversation Example

This example demonstrates how to use the `Conversation` class to simplify interaction with a single agent, providing a request-response pattern.

## Overview

The `Conversation` class provides a simplified way to interact with agents by:

1. **Automatic Session Management**: Handles agent lifecycle and session management
2. **Event Processing**: Processes agent events and calls the handler appropriately
3. **Simplified Interface**: Provides a simple request-response pattern instead of dealing with low-level events

## Key Components

### Conversation Class

- `NewConversation(agent)`: Creates a new conversation instance
- `Run(ctx, message, handler)`: Runs a single conversation turn with the agent

### ConversationHandler Interface

```go
type ConversationHandler interface {
    OnResponse(ctx *ConversationContext, agentResponse *AgentResponse) error
}
```

The handler receives two types of responses:
- **Message**: Text content from the agent
- **ToolCall**: When the agent wants to call a tool

### ConversationContext

Provides access to:
- Session information (SessionId, Context)
- Agent input channel for sending additional messages
- Agent context for accessing agent capabilities

## Usage

```go
// Create agent
agent, err := CreateAgent(...)

// Create conversation
conversation := chat.NewConversation(agent)

// Create handler
handler := &MyConversationHandler{}

// Run conversation
err := conversation.Run(ctx, "Hello!", handler)
```

## Example Handler

The example includes a `SimpleConversationHandler` that:
- Prints text messages from the agent
- Logs tool calls (and auto-approves them)
- Demonstrates the basic pattern

## Running

Make sure you have the required environment variables set for your LLM provider:

```bash
# For OpenAI
export OPENAI_API_KEY=your_key_here

# For Anthropic  
export ANTHROPIC_API_KEY=your_key_here

# For Gemini
export GEMINI_API_KEY=your_key_here
```

Then run:

```bash
go run main.go "Your question here"
```

## Advanced Usage

The conversation handler can:
- Collect and process agent responses
- Send additional messages through the context
- Handle tool calls with custom logic
- Maintain conversation state
- Implement custom UI interactions
