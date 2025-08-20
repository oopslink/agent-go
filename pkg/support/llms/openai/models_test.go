package openai

import (
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIModels_Registration(t *testing.T) {
	// Test that all models are properly configured
	for modelID, model := range OpenAIModels {
		t.Run(modelID, func(t *testing.T) {
			// Verify model ID matches the key
			assert.Equal(t, modelID, model.ModelId.ID)

			// Verify provider is set correctly
			assert.Equal(t, ModelProviderOpenAI, model.ModelId.Provider)

			// Verify model has required fields
			assert.NotEmpty(t, model.Name)
			assert.NotEmpty(t, model.ApiModelName)
			assert.Greater(t, model.ContextWindowSize, int64(0))
			assert.GreaterOrEqual(t, model.DefaultMaxTokens, int64(0))

			// Verify model has at least one feature
			assert.NotEmpty(t, model.Features)

			// Verify pricing is reasonable
			assert.GreaterOrEqual(t, model.CostPer1MIn, 0.0)
			assert.GreaterOrEqual(t, model.CostPer1MOut, 0.0)
			assert.GreaterOrEqual(t, model.CostPer1MInCached, 0.0)
			assert.GreaterOrEqual(t, model.CostPer1MOutCached, 0.0)
		})
	}
}

func TestOpenAIModels_Features(t *testing.T) {
	// Test that models have appropriate features
	testCases := []struct {
		modelID       string
		hasCompletion bool
		hasEmbedding  bool
		hasAttachment bool
	}{
		{ModelGPT41, true, false, true},
		{ModelGPT41Mini, true, false, true},
		{ModelGPT41Nano, true, false, true},
		{ModelGPT45Preview, true, false, true},
		{ModelGPT4o, true, false, true},
		{ModelGPT4oMini, true, false, true},
		{ModelO1, true, false, true},
		{ModelO1Pro, true, false, true},
		{ModelO1Mini, true, false, true},
		{ModelO3, true, false, true},
		{ModelO3Mini, true, false, true},
		{ModelO4Mini, true, false, true},
		{ModelTextEmbedding3Small, false, true, false},
		{ModelTextEmbedding3Large, false, true, false},
		{ModelTextEmbeddingADA002, false, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := OpenAIModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.hasCompletion, model.IsSupport(llms.ModelFeatureCompletion))
			assert.Equal(t, tc.hasEmbedding, model.IsSupport(llms.ModelFeatureEmbedding))
			assert.Equal(t, tc.hasAttachment, model.IsSupport(llms.ModelFeatureAttachment))
		})
	}
}

func TestOpenAIModels_Pricing(t *testing.T) {
	// Test that pricing is consistent across models
	testCases := []struct {
		modelID         string
		expectedInCost  float64
		expectedOutCost float64
	}{
		{ModelGPT41, 2.00, 8.00},
		{ModelGPT41Mini, 0.40, 1.60},
		{ModelGPT41Nano, 0.10, 0.40},
		{ModelGPT45Preview, 75.00, 150.00},
		{ModelGPT4o, 2.50, 10.00},
		{ModelGPT4oMini, 0.15, 0.60},
		{ModelO1, 15.00, 60.00},
		{ModelO1Pro, 150.00, 600.00},
		{ModelO1Mini, 1.10, 4.40},
		{ModelO3, 10.00, 40.00},
		{ModelO3Mini, 2.50, 10.00},
		{ModelO4Mini, 5.00, 20.00},
		{ModelTextEmbedding3Small, 1.00, 1.00},
		{ModelTextEmbedding3Large, 1.00, 1.00},
		{ModelTextEmbeddingADA002, 1.00, 1.00},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := OpenAIModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.expectedInCost, model.CostPer1MIn)
			assert.Equal(t, tc.expectedOutCost, model.CostPer1MOut)
		})
	}
}

func TestOpenAIModels_Constants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "openai", string(ModelProviderOpenAI))
	assert.Equal(t, llms.ReasoningEffortMedium, OpenAIDefaultReasoningEffort)

	// Test completion model constants
	assert.Equal(t, "gpt-4.1", ModelGPT41)
	assert.Equal(t, "gpt-4.1-mini", ModelGPT41Mini)
	assert.Equal(t, "gpt-4.1-nano", ModelGPT41Nano)
	assert.Equal(t, "gpt-4.5-preview", ModelGPT45Preview)
	assert.Equal(t, "gpt-4o", ModelGPT4o)
	assert.Equal(t, "gpt-4o-mini", ModelGPT4oMini)

	// Test O-series model constants
	assert.Equal(t, "o1", ModelO1)
	assert.Equal(t, "o1-pro", ModelO1Pro)
	assert.Equal(t, "o1-mini", ModelO1Mini)
	assert.Equal(t, "o3", ModelO3)
	assert.Equal(t, "o3-mini", ModelO3Mini)
	assert.Equal(t, "o4-mini", ModelO4Mini)

	// Test embedding model constants
	assert.Equal(t, "text-embedding-3-small", ModelTextEmbedding3Small)
	assert.Equal(t, "text-embedding-3-large", ModelTextEmbedding3Large)
	assert.Equal(t, "text-embedding-ada-002", ModelTextEmbeddingADA002)
}

func TestOpenAIModels_ModelCount(t *testing.T) {
	// Test that we have the expected number of models
	expectedCount := 15
	assert.Len(t, OpenAIModels, expectedCount)
}

func TestOpenAIModels_ContextWindows(t *testing.T) {
	// Test that models have appropriate context window sizes
	testCases := []struct {
		modelID               string
		expectedContextWindow int64
	}{
		{ModelGPT41, 1_047_576},
		{ModelGPT41Mini, 200_000},
		{ModelGPT41Nano, 1_047_576},
		{ModelGPT45Preview, 128_000},
		{ModelGPT4o, 128_000},
		{ModelGPT4oMini, 128_000},
		{ModelO1, 200_000},
		{ModelO1Pro, 200_000},
		{ModelO1Mini, 128_000},
		{ModelO3, 200_000},
		{ModelO3Mini, 128_000},
		{ModelO4Mini, 128_000},
		{ModelTextEmbedding3Small, 8192},
		{ModelTextEmbedding3Large, 8192},
		{ModelTextEmbeddingADA002, 8192},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := OpenAIModels[tc.modelID]
			assert.True(t, exists, "Model should exist")
			assert.Equal(t, tc.expectedContextWindow, model.ContextWindowSize)
		})
	}
}

func TestOpenAIModels_EmbeddingModels(t *testing.T) {
	// Test that embedding models have correct configuration
	embeddingModels := []string{
		ModelTextEmbedding3Small,
		ModelTextEmbedding3Large,
		ModelTextEmbeddingADA002,
	}

	for _, modelID := range embeddingModels {
		t.Run(modelID, func(t *testing.T) {
			model := OpenAIModels[modelID]

			// Should only support embedding feature
			assert.True(t, model.IsSupport(llms.ModelFeatureEmbedding))
			assert.False(t, model.IsSupport(llms.ModelFeatureCompletion))
			assert.False(t, model.IsSupport(llms.ModelFeatureAttachment))

			// Should have reasonable cost for input tokens
			assert.GreaterOrEqual(t, model.CostPer1MIn, 0.0)
			assert.GreaterOrEqual(t, model.CostPer1MOut, 0.0)

			// Should have smaller context window
			assert.Equal(t, int64(8192), model.ContextWindowSize)
		})
	}
}

func TestOpenAIModels_DefaultMaxTokens(t *testing.T) {
	// Test that default max tokens are reasonable
	for modelID, model := range OpenAIModels {
		t.Run(modelID, func(t *testing.T) {
			// Some models may have 0 as default max tokens, which is valid
			assert.GreaterOrEqual(t, model.DefaultMaxTokens, int64(0))
			if model.DefaultMaxTokens > 0 {
				assert.LessOrEqual(t, model.DefaultMaxTokens, model.ContextWindowSize)
			}
		})
	}
}

func TestOpenAIModels_ApiModelNames(t *testing.T) {
	// Test that API model names are properly set
	testCases := []struct {
		modelID         string
		expectedApiName string
	}{
		{ModelGPT41, "gpt-4.1"},
		{ModelGPT41Mini, "gpt-4.1-mini"},
		{ModelGPT41Nano, "gpt-4.1-nano"},
		{ModelGPT45Preview, "gpt-4.5-preview"},
		{ModelGPT4o, "gpt-4o"},
		{ModelGPT4oMini, "gpt-4o-mini"},
		{ModelO1, "o1"},
		{ModelO1Pro, "o1-pro"},
		{ModelO1Mini, "o1-mini"},
		{ModelO3, "o3"},
		{ModelO3Mini, "o3-mini"},
		{ModelO4Mini, "o4-mini"},
		{ModelTextEmbedding3Small, "text-embedding-3-small"},
		{ModelTextEmbedding3Large, "text-embedding-3-large"},
		{ModelTextEmbeddingADA002, "text-embedding-ada-002"},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := OpenAIModels[tc.modelID]
			assert.True(t, exists, "Model should exist")
			assert.Equal(t, tc.expectedApiName, model.ApiModelName)
		})
	}
}
