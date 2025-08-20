# Makefile for agent-go chat application

BINARY=build/chat-app
CHAT_BINARY=build/chat
CMD_DIR=./cmd
APPS_DIR=./apps

# Default parameters
PROVIDER ?= openai
MODEL ?= gpt-4o-mini
API_KEY ?= 
BASE_URL ?= 
OPENAI_COMPATIBILITY ?= false

.PHONY: build clean run chat install deps fmt test

# Clean build files
clean:
	rm -rf build

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Build main program
build: deps fmt
	@rm -rf $(BINARY)
	@mkdir -p build
	go build -o $(BINARY) .

# Run tests
test:
	go test -v ./...

# Run main program
run: build
	./$(BINARY)

# Quick start - OpenAI
chat: build
	@echo "Starting chat application..."
	./$(BINARY) chat --config ${CONFIG_PATH}

# Help information
help:
	@echo "Available commands:"
	@echo "  build        - Build main program"
	@echo "  build-chat   - Build chat application"
	@echo "  build-all    - Build all applications"
	@echo "  clean        - Clean build files"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  test         - Run tests"
	@echo "  run          - Run main program"
	@echo "  chat         - Run chat application with command line arguments"
	@echo "  help         - Display help information"
	@echo ""
	@echo "Parameters:"
	@echo "  PROVIDER     - AI model provider (openai, anthropic, gemini)"
	@echo "  MODEL        - Model name"
	@echo "  API_KEY      - API key (required)"
	@echo "  BASE_URL     - API base URL (optional)"
	@echo "  OPENAI_COMPATIBILITY - Enable OpenAI compatible mode (true/false)"
	@echo ""
	@echo "Examples:"
	@echo "  make chat API_KEY=sk-xxx"
	@echo "  make openai API_KEY=sk-xxx"
	@echo "  make openai-compat API_KEY=sk-xxx BASE_URL=https://api.xxxproxy.xx/v1"
	@echo "  make anthropic API_KEY=sk-ant-xxx"
	@echo "  make chat PROVIDER=openai MODEL=gpt-4 API_KEY=sk-xxx"

# Default target
.DEFAULT_GOAL := help
