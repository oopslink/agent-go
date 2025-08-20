package anthropic

import (
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
)

func TestAnthropicModels_Registration(t *testing.T) {
	// Test that all models are properly configured
	for modelID, model := range AnthropicModels {
		t.Run(modelID, func(t *testing.T) {
			// Verify model ID matches the key
			assert.Equal(t, modelID, model.ModelId.ID)

			// Verify provider is set correctly
			assert.Equal(t, ModelProviderAnthropic, model.ModelId.Provider)

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

func TestAnthropicModels_Features(t *testing.T) {
	// Test that models have appropriate features
	testCases := []struct {
		modelID       string
		hasCompletion bool
		hasReasoning  bool
		hasAttachment bool
	}{
		{ModelClaude35Sonnet, true, false, true},
		{ModelClaude3Haiku, true, false, true},
		{ModelClaude37Sonnet, true, true, true},
		{ModelClaude35Haiku, true, false, true},
		{ModelClaude3Opus, true, false, true},
		{ModelClaude4Sonnet, true, true, true},
		{ModelClaude4Opus, true, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := AnthropicModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.hasCompletion, model.IsSupport(llms.ModelFeatureCompletion))
			assert.Equal(t, tc.hasReasoning, model.IsSupport(llms.ModelFeatureReasoning))
			assert.Equal(t, tc.hasAttachment, model.IsSupport(llms.ModelFeatureAttachment))
		})
	}
}

func TestAnthropicModels_Pricing(t *testing.T) {
	// Test that pricing is consistent across models
	testCases := []struct {
		modelID         string
		expectedInCost  float64
		expectedOutCost float64
	}{
		{ModelClaude35Sonnet, 3.0, 15.0},
		{ModelClaude3Haiku, 0.25, 1.25},
		{ModelClaude37Sonnet, 3.0, 15.0},
		{ModelClaude35Haiku, 0.80, 4.0},
		{ModelClaude3Opus, 15.0, 75.0},
		{ModelClaude4Sonnet, 3.0, 15.0},
		{ModelClaude4Opus, 15.0, 75.0},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := AnthropicModels[tc.modelID]
			assert.True(t, exists, "Model should exist")

			assert.Equal(t, tc.expectedInCost, model.CostPer1MIn)
			assert.Equal(t, tc.expectedOutCost, model.CostPer1MOut)
		})
	}
}

func TestAnthropicModels_ContextWindow(t *testing.T) {
	// Test that all models have the same context window size
	expectedContextWindow := int64(200000)

	for modelID, model := range AnthropicModels {
		t.Run(modelID, func(t *testing.T) {
			assert.Equal(t, expectedContextWindow, model.ContextWindowSize)
		})
	}
}

func TestAnthropicModels_Constants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "anthropic", string(ModelProviderAnthropic))
	assert.Equal(t, "claude-3.5-sonnet", ModelClaude35Sonnet)
	assert.Equal(t, "claude-3-haiku", ModelClaude3Haiku)
	assert.Equal(t, "claude-3.7-sonnet", ModelClaude37Sonnet)
	assert.Equal(t, "claude-3.5-haiku", ModelClaude35Haiku)
	assert.Equal(t, "claude-3-opus", ModelClaude3Opus)
	assert.Equal(t, "claude-4-opus", ModelClaude4Opus)
	assert.Equal(t, "claude-4-sonnet", ModelClaude4Sonnet)
}

func TestAnthropicModels_ModelCount(t *testing.T) {
	// Test that we have the expected number of models
	expectedCount := 7
	assert.Len(t, AnthropicModels, expectedCount)
}

func TestAnthropicModels_ApiModelNames(t *testing.T) {
	// Test that API model names are properly set
	testCases := []struct {
		modelID         string
		expectedApiName string
	}{
		{ModelClaude35Sonnet, "claude-3-5-sonnet-latest"},
		{ModelClaude3Haiku, "claude-3-haiku-20240307"},
		{ModelClaude37Sonnet, "claude-3-7-sonnet-latest"},
		{ModelClaude35Haiku, "claude-3-5-haiku-latest"},
		{ModelClaude3Opus, "claude-3-opus-latest"},
		{ModelClaude4Sonnet, "claude-sonnet-4-20250514"},
		{ModelClaude4Opus, "claude-opus-4-20250514"},
	}

	for _, tc := range testCases {
		t.Run(tc.modelID, func(t *testing.T) {
			model, exists := AnthropicModels[tc.modelID]
			assert.True(t, exists, "Model should exist")
			assert.Equal(t, tc.expectedApiName, model.ApiModelName)
		})
	}
}

func TestAnthropicModels_DefaultMaxTokens(t *testing.T) {
	// Test that default max tokens are reasonable
	for modelID, model := range AnthropicModels {
		t.Run(modelID, func(t *testing.T) {
			assert.Greater(t, model.DefaultMaxTokens, int64(0))
			assert.LessOrEqual(t, model.DefaultMaxTokens, model.ContextWindowSize)
		})
	}
}
