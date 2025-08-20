// Package anthropic provides Anthropic AI model provider implementations for the agent-go framework.
// This file contains model definitions and registration for Anthropic's Claude models.
package anthropic

import (
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	// ModelProviderAnthropic is the provider identifier for Anthropic models.
	ModelProviderAnthropic llms.ModelProvider = "anthropic"

	// Anthropic model identifiers
	ModelClaude35Sonnet = "claude-3.5-sonnet" // Claude 3.5 Sonnet model
	ModelClaude3Haiku   = "claude-3-haiku"    // Claude 3 Haiku model
	ModelClaude37Sonnet = "claude-3.7-sonnet" // Claude 3.7 Sonnet model
	ModelClaude35Haiku  = "claude-3.5-haiku"  // Claude 3.5 Haiku model
	ModelClaude3Opus    = "claude-3-opus"     // Claude 3 Opus model
	ModelClaude4Opus    = "claude-4-opus"     // Claude 4 Opus model
	ModelClaude4Sonnet  = "claude-4-sonnet"   // Claude 4 Sonnet model
)

// AnthropicModels contains the configuration for all available Anthropic models.
// Pricing and capabilities are based on the official Anthropic documentation:
// https://docs.anthropic.com/en/docs/about-claude/models/all-models
var AnthropicModels = map[string]llms.Model{
	ModelClaude35Sonnet: {
		ModelId: llms.ModelId{
			ID:       ModelClaude35Sonnet,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 3.5 Sonnet",
		ApiModelName:       "claude-3-5-sonnet-latest",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  3.75,
		CostPer1MOutCached: 0.30,
		CostPer1MOut:       15.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   5000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude3Haiku: {
		ModelId: llms.ModelId{
			ID:       ModelClaude3Haiku,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 3 Haiku",
		ApiModelName:       "claude-3-haiku-20240307",
		CostPer1MIn:        0.25,
		CostPer1MInCached:  0.30,
		CostPer1MOutCached: 0.03,
		CostPer1MOut:       1.25,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   4096,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude37Sonnet: {
		ModelId: llms.ModelId{
			ID:       ModelClaude37Sonnet,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 3.7 Sonnet",
		ApiModelName:       "claude-3-7-sonnet-latest",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  3.75,
		CostPer1MOutCached: 0.30,
		CostPer1MOut:       15.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude35Haiku: {
		ModelId: llms.ModelId{
			ID:       ModelClaude35Haiku,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 3.5 Haiku",
		ApiModelName:       "claude-3-5-haiku-latest",
		CostPer1MIn:        0.80,
		CostPer1MInCached:  1.0,
		CostPer1MOutCached: 0.08,
		CostPer1MOut:       4.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   4096,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude3Opus: {
		ModelId: llms.ModelId{
			ID:       ModelClaude3Opus,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 3 Opus",
		ApiModelName:       "claude-3-opus-latest",
		CostPer1MIn:        15.0,
		CostPer1MInCached:  18.75,
		CostPer1MOutCached: 1.50,
		CostPer1MOut:       75.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   4096,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude4Sonnet: {
		ModelId: llms.ModelId{
			ID:       ModelClaude4Sonnet,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 4 Sonnet",
		ApiModelName:       "claude-sonnet-4-20250514",
		CostPer1MIn:        3.0,
		CostPer1MInCached:  3.75,
		CostPer1MOutCached: 0.30,
		CostPer1MOut:       15.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelClaude4Opus: {
		ModelId: llms.ModelId{
			ID:       ModelClaude4Opus,
			Provider: ModelProviderAnthropic,
		},
		Name:               "Claude 4 Opus",
		ApiModelName:       "claude-opus-4-20250514",
		CostPer1MIn:        15.0,
		CostPer1MInCached:  18.75,
		CostPer1MOutCached: 1.50,
		CostPer1MOut:       75.0,
		ContextWindowSize:  200000,
		DefaultMaxTokens:   4096,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
}

// init registers all Anthropic models with the global model registry.
// This function is called automatically when the package is imported.
func init() {
	// Register Anthropic models with the model registry
	for _, model := range AnthropicModels {
		if err := llms.RegisterModel(&model); err != nil {
			klog.Warningf("Failed to register Anthropic model %s: %v", model.ModelId.ID, err)
		}
	}
}
