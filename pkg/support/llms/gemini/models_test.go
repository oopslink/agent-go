package gemini

import (
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
)

func TestGeminiModels_Registration(t *testing.T) {
	// Test that all models are properly configured
	for modelID, model := range GeminiModels {
		t.Run(modelID, func(t *testing.T) {
			// Verify model ID matches the key
			assert.Equal(t, modelID, model.ModelId.ID)

			// Verify provider is set correctly
			assert.Equal(t, ModelProviderGemini, model.ModelId.Provider)

			// Verify model has required fields
			assert.NotEmpty(t, model.Name)
			assert.NotEmpty(t, model.ApiModelName)
			assert.Greater(t, model.ContextWindowSize, int64(0))
			assert.Greater(t, model.DefaultMaxTokens, int64(0))

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

func TestGeminiModels_Features(t *testing.T) {
	// Test that models have appropriate features
	testCases := []struct {
		modelID       string
		hasCompletion bool
		hasEmbedding  bool
		hasAttachment bool
	}{
		{ModelGemini25Flash, true, false, true},
		{ModelGemini25, true, false, true},
		{ModelGemini20Flash, true, false, true},
		{ModelGemini20FlashLite, true, false, true},
		{ModelGeminiEmbedding001, false, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := GeminiModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.hasCompletion, model.IsSupport(llms.ModelFeatureCompletion))
			assert.Equal(t, tc.hasEmbedding, model.IsSupport(llms.ModelFeatureEmbedding))
			assert.Equal(t, tc.hasAttachment, model.IsSupport(llms.ModelFeatureAttachment))
		})
	}
}

func TestGeminiModels_Pricing(t *testing.T) {
	// Test that pricing is consistent across models
	testCases := []struct {
		modelID         string
		expectedInCost  float64
		expectedOutCost float64
	}{
		{ModelGemini25Flash, 0.15, 0.60},
		{ModelGemini25, 1.25, 10.0},
		{ModelGemini20Flash, 0.10, 0.40},
		{ModelGemini20FlashLite, 0.05, 0.30},
		{ModelGeminiEmbedding001, 0.0001, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := GeminiModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.expectedInCost, model.CostPer1MIn)
			assert.Equal(t, tc.expectedOutCost, model.CostPer1MOut)
		})
	}
}

func TestGeminiModels_ContextWindow(t *testing.T) {
	// Test that completion models have large context windows
	completionModels := []string{
		ModelGemini25Flash,
		ModelGemini25,
		ModelGemini20Flash,
		ModelGemini20FlashLite,
	}

	expectedContextWindow := int64(1000000)

	for _, modelID := range completionModels {
		t.Run(modelID, func(t *testing.T) {
			model, exists := GeminiModels[modelID]
			assert.True(t, exists, "Model should exist")
			assert.Equal(t, expectedContextWindow, model.ContextWindowSize)
		})
	}

	// Test that embedding model has smaller context window
	embeddingModel := GeminiModels[ModelGeminiEmbedding001]
	assert.Equal(t, int64(2048), embeddingModel.ContextWindowSize)
}

func TestGeminiModels_Constants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "gemini", string(ModelProviderGemini))
	assert.Equal(t, "gemini-2.5-flash", ModelGemini25Flash)
	assert.Equal(t, "gemini-2.5", ModelGemini25)
	assert.Equal(t, "gemini-2.0-flash", ModelGemini20Flash)
	assert.Equal(t, "gemini-2.0-flash-lite", ModelGemini20FlashLite)
	assert.Equal(t, "gemini-embedding-001", ModelGeminiEmbedding001)
}

func TestGeminiModels_ModelCount(t *testing.T) {
	// Test that we have the expected number of models
	expectedCount := 5
	assert.Len(t, GeminiModels, expectedCount)
}

func TestGeminiModels_ApiModelNames(t *testing.T) {
	// Test that API model names are properly set
	testCases := []struct {
		modelID         string
		expectedApiName string
	}{
		{ModelGemini25Flash, "gemini-2.5-flash"},
		{ModelGemini25, "gemini-2.5-pro"},
		{ModelGemini20Flash, "gemini-2.0-flash"},
		{ModelGemini20FlashLite, "gemini-2.0-flash-lite"},
		{ModelGeminiEmbedding001, "embedding-001"},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := GeminiModels[tc.modelID]
			assert.True(t, exists, "Model should exist")
			assert.Equal(t, tc.expectedApiName, model.ApiModelName)
		})
	}
}

func TestGeminiModels_DefaultMaxTokens(t *testing.T) {
	// Test that default max tokens are reasonable
	for modelID, model := range GeminiModels {
		t.Run(modelID, func(t *testing.T) {
			assert.Greater(t, model.DefaultMaxTokens, int64(0))
			assert.LessOrEqual(t, model.DefaultMaxTokens, model.ContextWindowSize)
		})
	}
}

func TestGeminiModels_EmbeddingModel(t *testing.T) {
	// Test that embedding model has correct configuration
	embeddingModel := GeminiModels[ModelGeminiEmbedding001]

	// Should only support embedding feature
	assert.True(t, embeddingModel.IsSupport(llms.ModelFeatureEmbedding))
	assert.False(t, embeddingModel.IsSupport(llms.ModelFeatureCompletion))
	assert.False(t, embeddingModel.IsSupport(llms.ModelFeatureAttachment))

	// Should have very low cost for input tokens
	assert.Equal(t, 0.0001, embeddingModel.CostPer1MIn)
	assert.Equal(t, 0.0, embeddingModel.CostPer1MOut)

	// Should have smaller context window
	assert.Equal(t, int64(2048), embeddingModel.ContextWindowSize)
	assert.Equal(t, int64(2048), embeddingModel.DefaultMaxTokens)
}
