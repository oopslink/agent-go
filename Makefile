# Makefile for agent-go
BINARY=build/cli

# Get all packages
PACKAGES=$(shell go list ./...)
PACKAGE_DIRS=$(shell go list -f '{{.Dir}}' ./...)

.PHONY: build test test/coverage clean fmt gomock

clean:
	rm -rf build

# Install dependencies
deps:
	go mod tidy

# Code formatting
fmt:
	go fmt ./...

build: deps fmt
	go build ./...

# Generate mock classes
gomock:
	@echo "Installing mockgen..."
	@go install github.com/golang/mock/mockgen@latest
	@echo "Generating mocks..."
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/commons/utils/retry.go -destination=pkg/commons/utils/mock_retry.go -package=utils
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/core/agent/agent.go -destination=pkg/core/agent/mock_agent.go -package=agent
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/core/memory/memory.go -destination=pkg/core/memory/mock_memory.go -package=memory
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/core/knowledge/knowledge.go -destination=pkg/core/knowledge/mock_knowledge.go -package=knowledge
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/llms/chat_provider.go -destination=pkg/support/llms/mock_chat_provider.go -package=llms
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/llms/embedder_provider.go -destination=pkg/support/llms/mock_embedder_provider.go -package=llms
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/document/chunking.go -destination=pkg/support/document/mock_chunking.go -package=document
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/document/reader.go -destination=pkg/support/document/mock_reader.go -package=document
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/embedder/embedder.go -destination=pkg/support/embedder/mock_embedder.go -package=embedder
	@$(shell go env GOPATH)/bin/mockgen -source=pkg/support/vectordb/vectordb.go -destination=pkg/support/vectordb/mock_vectordb.go -package=vectordb
	@echo "Mock generation completed"

# Unit testing
test:
	go test -v ./...

# Run tests with coverage
test/coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Ignore build output
.gitignore:
	@grep -qxF 'build/' .gitignore || echo 'build/' >> .gitignore
