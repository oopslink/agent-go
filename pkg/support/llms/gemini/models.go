// Package gemini provides Google Gemini AI model provider implementations for the agent-go framework.
// This file contains model definitions and registration for Google's Gemini models.
package gemini

import (
	"github.com/oopslink/agent-go/pkg/support/llms"

	"k8s.io/klog/v2"
)

const (
	// ModelProviderGemini is the provider identifier for Google Gemini models.
	ModelProviderGemini llms.ModelProvider = "gemini"

	// Gemini model identifiers for completion tasks
	ModelGemini25Flash     = "gemini-2.5-flash"      // Gemini 2.5 Flash model
	ModelGemini25          = "gemini-2.5"            // Gemini 2.5 Pro model
	ModelGemini20Flash     = "gemini-2.0-flash"      // Gemini 2.0 Flash model
	ModelGemini20FlashLite = "gemini-2.0-flash-lite" // Gemini 2.0 Flash Lite model

	// Gemini model identifiers for embedding tasks
	ModelGeminiEmbedding001 = "gemini-embedding-001" // Gemini embedding model
)

// GeminiModels contains the configuration for all available Google Gemini models.
// Pricing and capabilities are based on the official Google AI documentation.
var GeminiModels = map[string]llms.Model{
	ModelGemini25Flash: {
		ModelId: llms.ModelId{
			ID:       ModelGemini25Flash,
			Provider: ModelProviderGemini,
		},
		Name:               "Gemini 2.5 Flash",
		ApiModelName:       "gemini-2.5-flash",
		CostPer1MIn:        0.15,
		CostPer1MInCached:  0,
		CostPer1MOutCached: 0,
		CostPer1MOut:       0.60,
		ContextWindowSize:  1000000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGemini25: {
		ModelId: llms.ModelId{
			ID:       ModelGemini25,
			Provider: ModelProviderGemini,
		},
		Name:               "Gemini 2.5 Pro",
		ApiModelName:       "gemini-2.5-pro",
		CostPer1MIn:        1.25,
		CostPer1MInCached:  0,
		CostPer1MOutCached: 0,
		CostPer1MOut:       10,
		ContextWindowSize:  1000000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGemini20Flash: {
		ModelId: llms.ModelId{
			ID:       ModelGemini20Flash,
			Provider: ModelProviderGemini,
		},
		Name:               "Gemini 2.0 Flash",
		ApiModelName:       "gemini-2.0-flash",
		CostPer1MIn:        0.10,
		CostPer1MInCached:  0,
		CostPer1MOutCached: 0,
		CostPer1MOut:       0.40,
		ContextWindowSize:  1000000,
		DefaultMaxTokens:   6000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGemini20FlashLite: {
		ModelId: llms.ModelId{
			ID:       ModelGemini20FlashLite,
			Provider: ModelProviderGemini,
		},
		Name:               "Gemini 2.0 Flash Lite",
		ApiModelName:       "gemini-2.0-flash-lite",
		CostPer1MIn:        0.05,
		CostPer1MInCached:  0,
		CostPer1MOutCached: 0,
		CostPer1MOut:       0.30,
		ContextWindowSize:  1000000,
		DefaultMaxTokens:   6000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},

	// Embedding models
	ModelGeminiEmbedding001: {
		ModelId: llms.ModelId{
			ID:       ModelGeminiEmbedding001,
			Provider: ModelProviderGemini,
		},
		Name:               "Gemini Embedding 001",
		ApiModelName:       "embedding-001",
		CostPer1MIn:        0.0001,
		CostPer1MInCached:  0,
		CostPer1MOutCached: 0,
		CostPer1MOut:       0,
		ContextWindowSize:  2048,
		DefaultMaxTokens:   2048,
		Features: []llms.ModelFeature{
			llms.ModelFeatureEmbedding,
		},
	},
}

// init registers all Gemini models with the global model registry.
// This function is called automatically when the package is imported.
func init() {
	// Register Gemini models with the model registry
	for _, model := range GeminiModels {
		if err := llms.RegisterModel(&model); err != nil {
			klog.Warningf("Failed to register Gemini model %s: %v", model.ModelId.ID, err)
		}
	}
}
