package llms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddingResponse(t *testing.T) {
	// Test creating an EmbeddingResponse
	model := &Model{
		ModelId: ModelId{
			Provider: "test-provider",
			ID:       "test-model",
		},
		Name: "Test Model",
	}

	vectors := []FloatVector{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
	}

	usage := UsageMetadata{
		InputTokens:         10,
		OutputTokens:        20,
		CacheCreationTokens: 5,
		CacheReadTokens:     2,
	}

	response := &EmbeddingResponse{
		Model:   model,
		Vectors: vectors,
		Usage:   usage,
	}

	// Verify the response fields
	assert.Equal(t, model, response.Model)
	assert.Equal(t, vectors, response.Vectors)
	assert.Equal(t, usage, response.Usage)
	assert.Len(t, response.Vectors, 2)
	assert.Len(t, response.Vectors[0], 3)
	assert.Len(t, response.Vectors[1], 3)
}

func TestFloatVector(t *testing.T) {
	// Test creating a FloatVector
	vector := FloatVector{0.1, 0.2, 0.3, 0.4, 0.5}

	// Verify the vector
	assert.Len(t, vector, 5)
	assert.Equal(t, 0.1, vector[0])
	assert.Equal(t, 0.5, vector[4])

	// Test appending to vector
	vector = append(vector, 0.6)
	assert.Len(t, vector, 6)
	assert.Equal(t, 0.6, vector[5])
}

func TestUsageMetadata_EmbedderProvider(t *testing.T) {
	// Test creating UsageMetadata
	usage := UsageMetadata{
		InputTokens:         100,
		OutputTokens:        200,
		CacheCreationTokens: 50,
		CacheReadTokens:     25,
	}

	// Verify the usage metadata
	assert.Equal(t, int64(100), usage.InputTokens)
	assert.Equal(t, int64(200), usage.OutputTokens)
	assert.Equal(t, int64(50), usage.CacheCreationTokens)
	assert.Equal(t, int64(25), usage.CacheReadTokens)

	// Test zero values
	zeroUsage := UsageMetadata{}
	assert.Equal(t, int64(0), zeroUsage.InputTokens)
	assert.Equal(t, int64(0), zeroUsage.OutputTokens)
	assert.Equal(t, int64(0), zeroUsage.CacheCreationTokens)
	assert.Equal(t, int64(0), zeroUsage.CacheReadTokens)
}

func TestEmbeddingResponse_EmptyVectors(t *testing.T) {
	// Test EmbeddingResponse with empty vectors
	response := &EmbeddingResponse{
		Model:   &Model{},
		Vectors: []FloatVector{},
		Usage:   UsageMetadata{},
	}

	assert.NotNil(t, response.Model)
	assert.Empty(t, response.Vectors)
	assert.Equal(t, UsageMetadata{}, response.Usage)
}

func TestEmbeddingResponse_NilModel(t *testing.T) {
	// Test EmbeddingResponse with nil model
	response := &EmbeddingResponse{
		Model:   nil,
		Vectors: []FloatVector{{0.1, 0.2}},
		Usage:   UsageMetadata{InputTokens: 10},
	}

	assert.Nil(t, response.Model)
	assert.Len(t, response.Vectors, 1)
	assert.Equal(t, int64(10), response.Usage.InputTokens)
}
