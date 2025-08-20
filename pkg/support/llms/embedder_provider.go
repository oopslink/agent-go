// Package provider provides AI model provider implementations for the agent-go framework.
// This file contains the embedder provider interface and related
package llms

import (
	"context"
)

// EmbeddingResponse represents the response from an embedding operation.
// It contains the model used, the generated vectors, and usage metadata.
type EmbeddingResponse struct {
	Model   *Model        // The model used for embedding
	Vectors []FloatVector // The generated embedding vectors
	Usage   UsageMetadata // Token usage information
}

// EmbedderProvider defines the interface for AI model providers that can generate embeddings.
// Implementations should provide thread-safe embedding generation capabilities.
type EmbedderProvider interface {
	// GetEmbeddings generates embeddings for the provided texts using the configured model.
	// The context can be used to cancel the operation.
	// Returns an EmbeddingResponse containing the vectors and usage information.
	GetEmbeddings(ctx context.Context, texts []string) (*EmbeddingResponse, error)
}
