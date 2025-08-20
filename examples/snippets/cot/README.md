# Chain of Thought (CoT) Agent Example

This example demonstrates a Chain of Thought (CoT) agent that thinks through problems step by step before providing final answers.

## Features

- **Step-by-step reasoning**: The agent breaks down complex problems into smaller parts
- **Transparent thinking**: Shows the thinking process before giving final answers
- **Multiple provider support**: Works with OpenAI, Anthropic, and Gemini
- **Single conversation**: Simple one-shot question and answer

## Usage

### Environment Setup

Set the required environment variables:

```bash
# Required
export API_KEY="your-api-key-here"

# Optional (defaults shown)
export PROVIDER="openai"           # openai, anthropic, gemini
export MODEL="gpt-4o-mini"         # model name for the provider
```

### Running the Example

```bash
# Run with default settings (OpenAI)
API_KEY=xxx go run *.go "What is 15 * 23?"

# Run with specific provider
PROVIDER=anthropic MODEL=claude-3-5-sonnet-20241022 API_KEY=xxx go run *.go "What is 15 * 23?"

# Run with Gemini
PROVIDER=gemini MODEL=gemini-2.5-flash API_KEY=xxx go run *.go "What is 15 * 23?"
```

### Example Output

```
Question: What is 15 * 23?

🤖 Agent thinking...
Let me think through this step by step:

1. First, I need to understand what 15 * 23 means - it's 15 multiplied by 23
2. I can break this down using the distributive property: 15 * 23 = 15 * (20 + 3)
3. This gives me: 15 * 20 + 15 * 3
4. 15 * 20 = 300
5. 15 * 3 = 45
6. So 15 * 23 = 300 + 45 = 345

Therefore, 15 * 23 = 345

✅ Agent finished with reason: normal_end
```

## How It Works

The CoT agent uses the `ChainOfThoughtPattern` behavior pattern which:

1. **Analyzes the problem** - Understands what needs to be solved
2. **Breaks it down** - Divides complex problems into manageable parts
3. **Considers approaches** - Evaluates different solution methods
4. **Shows reasoning** - Displays the thinking process in real-time
5. **Provides final answer** - Gives the conclusion after thorough analysis

The agent is configured with a system prompt that encourages step-by-step thinking and transparent reasoning.

## Configuration

The agent can be configured through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `API_KEY` | Required | Your API key for the chosen provider |
| `PROVIDER` | `openai` | AI provider (openai, anthropic, gemini) |
| `MODEL` | Provider-specific | Model name for the chosen provider |

### Default Models by Provider

- **OpenAI**: `gpt-4o-mini`
- **Anthropic**: `claude-3-5-sonnet-20241022`
- **Gemini**: `gemini-2.5-flash` 
