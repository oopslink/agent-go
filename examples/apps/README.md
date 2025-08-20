# Agent-Go Demo Application

This directory contains demo applications for the agent-go framework, showcasing how to use the framework to build intelligent conversational applications.

## Directory Structure

```
demo/
├── README.md           # This document
├── Makefile           # Build and run scripts
├── go.mod             # Go module file
├── main.go            # Main program entry
├── cmd/               # Command line interface
│   ├── root.go        # Root command
│   └── chat.go        # Chat command
├── apps/              # Demo applications
│   └── chat/          # Chat application
│       └── chat.go    # Chat application implementation
└── ui/                # UI module
    └── ui.go          # UI utilities
```

## Quick Start

### 1. Install Dependencies

```bash
cd demo
make deps
```

### 2. Build Application

```bash
make build
```

### 3. Run Chat Demo

#### Method 1: Using Makefile Shortcuts

```bash
# OpenAI
make openai API_KEY=sk-xxx

# OpenAI Compatible Mode
make openai-compat API_KEY=sk-xxx BASE_URL=https://api.xxxproxy.xx/v1

# Anthropic
make anthropic API_KEY=sk-ant-xxx

# Gemini
make gemini API_KEY=xxx
```

#### Method 2: Manual Command Line

```bash
# Build first
make build

# Run with different providers
./demo chat --provider openai --model gpt-4o-mini --api-key sk-xxx
./demo chat --provider anthropic --model claude-3.5-sonnet --api-key sk-ant-xxx
./demo chat --provider gemini --model gemini-2.0-flash --api-key xxx
```

## Configuration

### Command Line Parameters

| Parameter | Short | Description | Required |
|-----------|-------|-------------|----------|
| `--provider` | `-p` | AI model provider (openai, anthropic, gemini) | Yes |
| `--model` | `-m` | Specific model name | Yes |
| `--api-key` | `-k` | API authentication key | Yes |
| `--base-url` | `-u` | Custom API endpoint URL | No |
| `--openai-compatibility` | | Enable OpenAI-compatible mode | No |
| `--system-prompt` | `-s` | Custom system prompt | No |

### Environment Variables

You can also use environment variables instead of command line parameters:

```bash
export OPENAI_API_KEY=sk-xxx
export ANTHROPIC_API_KEY=sk-ant-xxx
export GEMINI_API_KEY=xxx
```

### Supported Models

#### OpenAI
- gpt-4.1-turbo
- gpt-4.1-mini
- gpt-4o
- gpt-4o-mini
- gpt-4
- gpt-3.5-turbo
- o1-preview
- o1-mini
- o3-mini
- o4-mini

#### Anthropic
- claude-3.5-sonnet
- claude-3.5-haiku
- claude-3.7-sonnet
- claude-3-haiku
- claude-3-opus
- claude-3-sonnet
- claude-4-turbo

#### Gemini
- gemini-2.5-flash
- gemini-2.5-pro
- gemini-2.0-flash
- gemini-2.0-flash-lite
- gemini-pro (legacy)
- gemini-pro-vision (legacy)

#### OpenAI Compatible
Any third-party API that follows the OpenAI API format.

## Usage Examples

### Basic Chat

```bash
# Start a basic chat session
./demo chat --provider openai --model gpt-4o-mini --api-key sk-xxx
```

### Custom System Prompt

```bash
./demo chat --provider openai --model gpt-4o-mini --api-key sk-xxx \
  --system-prompt "You are a helpful coding assistant"
```

### Using Custom Base URL

```bash
./demo chat --provider openai --model gpt-4o-mini --api-key sk-xxx \
  --base-url https://api.example.com/v1
```

### OpenAI Compatible Mode

```bash
./demo chat --provider openai --model gpt-4o-mini --api-key sk-xxx \
  --base-url https://api.example.com/v1 --openai-compatibility
```

## Interactive Commands

Once the chat session starts, you can use the following commands:

| Command | Description |
|---------|-------------|
| `:quit` or `:exit` | Exit the program |
| `:clear` | Clear conversation history |
| `:info` | Show current configuration |
| `:history` | Display conversation history in structured format |
| `:help` | Show help message |

## Features

- **Multi-provider Support**: Supports OpenAI, Anthropic, Gemini, and OpenAI-compatible APIs
- **Flexible Configuration**: Command line parameters and environment variables
- **Conversation Memory**: Maintains conversation history throughout the session
- **Enhanced Streaming UI**: Improved real-time streaming with better visual feedback
- **Structured History Display**: Clean, organized conversation history format
- **Markdown Rendering**: Beautiful output formatting with syntax highlighting
- **Simple UI**: Clean command-line interface with color coding
- **Error Handling**: Comprehensive error handling and user feedback
- **Tool Integration**: Support for function calling and tool execution

## Recent Improvements

### Enhanced User Experience
- **Improved Streaming**: Better visual feedback during streaming responses
- **Delayed Assistant Label**: Assistant label appears only when content starts streaming
- **Structured History**: Conversation history now displays in a clean, organized format
- **Better Error Recovery**: More robust error handling with graceful degradation

### Technical Enhancements
- **Package Restructuring**: Reorganized codebase for better maintainability
- **Enhanced Testing**: Comprehensive unit tests for all components
- **Memory Management**: Improved memory operations with better error recovery
- **Concurrent Safety**: Thread-safe operations for concurrent access

## Development

### Adding New Providers

To add support for a new AI provider:

1. Implement the provider interface in `pkg/support/provider/`
2. Register the provider in the factory
3. Add model definitions
4. Update documentation

### Customizing UI

The UI module is located in `demo/ui/ui.go`. You can customize:

- Color schemes
- Text formatting
- Layout and spacing
- Interactive elements
- Streaming behavior

### Building from Source

```bash
# Install dependencies
go mod download

# Build the application
go build -o demo .

# Run tests
go test ./...
```

## Troubleshooting

### Common Issues

1. **API Key Issues**
   - Ensure your API key is valid and has sufficient credits
   - Check if the API key format is correct for your provider

2. **Network Issues**
   - Verify your internet connection
   - Check if you're behind a proxy or firewall
   - Try using a different base URL if available

3. **Model Not Found**
   - Ensure the model name is correct
   - Check if the model is available for your API key
   - Refer to the provider's documentation for available models

4. **Build Issues**
   - Ensure you have Go 1.21+ installed
   - Run `go mod tidy` to clean up dependencies
   - Check for any missing dependencies

5. **Streaming Issues**
   - Check your network connection stability
   - Verify the provider supports streaming for your model
   - Try disabling streaming if available

### Getting Help

- Check the main project documentation
- Review the source code for implementation details
- Open an issue on GitHub for bugs or feature requests

## License

This demo application is part of the agent-go project and follows the same license terms.
