# ReAct (Reasoning and Acting) Example

This example demonstrates how to use the ReAct (Reasoning and Acting) behavior pattern in the agent-go framework. ReAct is a powerful approach that combines reasoning and action-taking to solve complex problems step by step.

## What is ReAct?

ReAct (Reasoning and Acting) is a pattern that enables AI agents to:
- **Reason**: Think through problems step by step
- **Act**: Take actions based on reasoning (use tools, make calculations, search for information)
- **Observe**: Learn from the results of actions
- **Iterate**: Continue the cycle until reaching a solution

The ReAct pattern follows this structured approach:
1. **Thought**: Analyze the current situation and plan the next step
2. **Action**: Execute a specific action (tool usage, calculation, search, etc.)
3. **Tool Calls**: When tools are needed, specify them in the `tool_calls` array
4. **Observation**: Process the result of the action
5. **Continue**: Decide whether to continue the cycle or provide the final answer

## Features

- **Structured Reasoning**: Each step follows a clear thought-action-observation cycle
- **Tool Integration**: Can use various tools and external resources with explicit tool calls
- **JSON Response Format**: Structured responses for easy parsing and processing
- **Iterative Problem Solving**: Breaks down complex problems into manageable steps
- **Tool Call Support**: Explicit tool call specification in JSON format

## Usage

### Prerequisites

1. Set up your API key:
   ```bash
   export API_KEY="your-api-key-here"
   ```

2. (Optional) Configure provider and model:
   ```bash
   export PROVIDER="openai"  # or "anthropic", "gemini"
   export MODEL="gpt-4o-mini"  # or other model names
   ```

### Running the Example

```bash
cd examples/snippets/react
go run main.go "What is the current weather in New York?"
```

### Example Output

The ReAct agent will respond with structured JSON that shows its reasoning process, including tool calls:

```json
{
  "thought": "I need to find the current weather in New York. This requires accessing real-time weather data.",
  "action": "I should use a search tool to get current weather information.",
  "tool_calls": [
    {
      "name": "duckduckgo_search",
      "arguments": {
        "query": "current weather New York"
      }
    }
  ],
  "observation": "Current weather in New York: 72°F, Partly Cloudy, Humidity: 65%",
  "continue": false
}
```

## Tool Call Support

The ReAct pattern now supports explicit tool calls through the `tool_calls` field in the JSON response:

- **Tool Calls Array**: Specify multiple tools to call in a single step
- **Tool Parameters**: Pass arguments to tools in a structured format
- **Tool Integration**: Seamlessly integrate with the existing tool infrastructure
- **Event System**: Tool calls are processed through the event system for proper handling

### Tool Call Format

```json
{
  "tool_calls": [
    {
      "name": "tool_name",
      "arguments": {
        "param1": "value1",
        "param2": "value2"
      }
    }
  ]
}
```

## How It Works

1. **Initialization**: Creates a ReAct behavior pattern and agent instance with tool support
2. **Question Processing**: Takes the user's question and starts the reasoning cycle
3. **ReAct Cycle**: 
   - **Thought**: Agent thinks about what needs to be done
   - **Action**: Agent decides on an action to take
   - **Tool Calls**: Agent specifies tools to call if needed
   - **Observation**: Agent processes the result
   - **Continue**: Agent decides whether to continue or provide final answer
4. **Final Answer**: When the problem is solved, provides the final response

## Customization

You can customize the ReAct agent by:

1. **Changing the System Prompt**: Modify the agent's personality and capabilities
2. **Adding Tools**: Integrate additional tools for specific actions
3. **Using Different Models**: Switch between different LLM providers and models
4. **Adding Knowledge Bases**: Provide additional context and information sources

## Example Questions

Try these questions to see ReAct with tool calls in action:

- "What is the current weather in New York?"
- "Search for the latest news about AI technology"
- "Find information about machine learning algorithms"
- "What are the top restaurants in San Francisco?"

## Architecture

The ReAct implementation consists of:

- **ReAct Pattern**: Defines the reasoning and acting behavior
- **State Processor**: Manages the ReAct cycle and response parsing
- **JSON Parser**: Extracts structured responses from the LLM
- **Tool Call Handler**: Processes tool calls and integrates with the tool system
- **Event System**: Handles communication between components

This example demonstrates how to build intelligent agents that can reason through complex problems, take appropriate actions, and use tools effectively to find solutions. 