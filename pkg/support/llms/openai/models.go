// Package openai provides OpenAI AI model provider implementations for the agent-go framework.
// This file contains model definitions and registration for OpenAI's GPT and embedding models.
package openai

import (
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	// ModelProviderOpenAI is the provider identifier for OpenAI models.
	ModelProviderOpenAI llms.ModelProvider = "openai"
	// OpenAIDefaultReasoningEffort is the default reasoning effort for OpenAI models.
	OpenAIDefaultReasoningEffort = llms.ReasoningEffortMedium

	// OpenAI model identifiers for completion tasks
	ModelGPT41        = "gpt-4.1"         // GPT 4.1 model
	ModelGPT41Mini    = "gpt-4.1-mini"    // GPT 4.1 mini model
	ModelGPT41Nano    = "gpt-4.1-nano"    // GPT 4.1 nano model
	ModelGPT45Preview = "gpt-4.5-preview" // GPT 4.5 preview model
	ModelGPT4o        = "gpt-4o"          // GPT 4o model
	ModelGPT4oMini    = "gpt-4o-mini"     // GPT 4o mini model
	ModelO1           = "o1"              // O1 model
	ModelO1Pro        = "o1-pro"          // O1 Pro model
	ModelO1Mini       = "o1-mini"         // O1 mini model
	ModelO3           = "o3"              // O3 model
	ModelO3Mini       = "o3-mini"         // O3 mini model
	ModelO4Mini       = "o4-mini"         // O4 mini model

	// OpenAI model identifiers for embedding tasks
	ModelTextEmbedding3Small = "text-embedding-3-small" // Text embedding 3 small model
	ModelTextEmbedding3Large = "text-embedding-3-large" // Text embedding 3 large model
	ModelTextEmbeddingADA002 = "text-embedding-ada-002" // Text embedding Ada 002 model
)

// OpenAIModels contains the configuration for all available OpenAI models.
// Pricing and capabilities are based on the official OpenAI documentation.
var OpenAIModels = map[string]llms.Model{

	// Models for completion tasks

	ModelGPT41: {
		ModelId: llms.ModelId{
			ID:       ModelGPT41,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4.1",
		ApiModelName:       "gpt-4.1",
		CostPer1MIn:        2.00,
		CostPer1MInCached:  0.50,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       8.00,
		ContextWindowSize:  1_047_576,
		DefaultMaxTokens:   20000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGPT41Mini: {
		ModelId: llms.ModelId{
			ID:       ModelGPT41Mini,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4.1 mini",
		ApiModelName:       "gpt-4.1-mini",
		CostPer1MIn:        0.40,
		CostPer1MInCached:  0.10,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       1.60,
		ContextWindowSize:  200_000,
		DefaultMaxTokens:   20000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGPT41Nano: {
		ModelId: llms.ModelId{
			ID:       ModelGPT41Nano,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4.1 nano",
		ApiModelName:       "gpt-4.1-nano",
		CostPer1MIn:        0.10,
		CostPer1MInCached:  0.025,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       0.40,
		ContextWindowSize:  1_047_576,
		DefaultMaxTokens:   20000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGPT45Preview: {
		ModelId: llms.ModelId{
			ID:       ModelGPT45Preview,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4.5 preview",
		ApiModelName:       "gpt-4.5-preview",
		CostPer1MIn:        75.00,
		CostPer1MInCached:  37.50,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       150.00,
		ContextWindowSize:  128_000,
		DefaultMaxTokens:   15000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGPT4o: {
		ModelId: llms.ModelId{
			ID:       ModelGPT4o,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4o",
		ApiModelName:       "gpt-4o",
		CostPer1MIn:        2.50,
		CostPer1MInCached:  1.25,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       10.00,
		ContextWindowSize:  128_000,
		DefaultMaxTokens:   4096,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelGPT4oMini: {
		ModelId: llms.ModelId{
			ID:       ModelGPT4oMini,
			Provider: ModelProviderOpenAI,
		},
		Name:               "GPT 4o mini",
		ApiModelName:       "gpt-4o-mini",
		CostPer1MIn:        0.15,
		CostPer1MInCached:  0.075,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       0.60,
		ContextWindowSize:  128_000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO1: {
		ModelId: llms.ModelId{
			ID:       ModelO1,
			Provider: ModelProviderOpenAI,
		},
		Name:               "O1",
		ApiModelName:       "o1",
		CostPer1MIn:        15.00,
		CostPer1MInCached:  7.50,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       60.00,
		ContextWindowSize:  200_000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO1Pro: {
		ModelId: llms.ModelId{
			ID:       ModelO1Pro,
			Provider: ModelProviderOpenAI,
		},
		Name:               "o1 pro",
		ApiModelName:       "o1-pro",
		CostPer1MIn:        150.00,
		CostPer1MInCached:  0.0,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       600.00,
		ContextWindowSize:  200_000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO1Mini: {
		ModelId: llms.ModelId{
			ID:       ModelO1Mini,
			Provider: ModelProviderOpenAI,
		},
		Name:               "o1 mini",
		ApiModelName:       "o1-mini",
		CostPer1MIn:        1.10,
		CostPer1MInCached:  0.55,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       4.40,
		ContextWindowSize:  128_000,
		DefaultMaxTokens:   50000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO3: {
		ModelId: llms.ModelId{
			ID:       ModelO3,
			Provider: ModelProviderOpenAI,
		},
		Name:               "o3",
		ApiModelName:       "o3",
		CostPer1MIn:        10.00,
		CostPer1MInCached:  2.50,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       40.00,
		ContextWindowSize:  200_000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO3Mini: {
		ModelId: llms.ModelId{
			ID:       ModelO3Mini,
			Provider: ModelProviderOpenAI,
		},
		Name:               "o3 mini",
		ApiModelName:       "o3-mini",
		CostPer1MIn:        2.50,
		CostPer1MInCached:  0.625,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       10.00,
		ContextWindowSize:  128_000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},
	ModelO4Mini: {
		ModelId: llms.ModelId{
			ID:       ModelO4Mini,
			Provider: ModelProviderOpenAI,
		},
		Name:               "o4 mini",
		ApiModelName:       "o4-mini",
		CostPer1MIn:        5.00,
		CostPer1MInCached:  1.25,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       20.00,
		ContextWindowSize:  128_000,
		Features: []llms.ModelFeature{
			llms.ModelFeatureCompletion,
			llms.ModelFeatureReasoning,
			llms.ModelFeatureAttachment,
		},
	},

	// Models for embedding
	//
	// Model                   ~ Pages per dollar    Performance on MTEB eval    Max input
	// text-embedding-3-small    62,500                62.3%                        8192
	// text-embedding-3-large    9,615                64.6%                        8192
	// text-embedding-ada-002    12,500                61.0%                        8192

	ModelTextEmbedding3Small: {
		ModelId: llms.ModelId{
			ID:       ModelTextEmbedding3Small,
			Provider: ModelProviderOpenAI,
		},
		Name:               "Text Embedding 3 Small",
		ApiModelName:       "text-embedding-3-small",
		CostPer1MIn:        1.00,
		CostPer1MInCached:  0.0,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       1.00,
		ContextWindowSize:  8192,
		Features: []llms.ModelFeature{
			llms.ModelFeatureEmbedding,
		},
	},
	ModelTextEmbedding3Large: {
		ModelId: llms.ModelId{
			ID:       ModelTextEmbedding3Large,
			Provider: ModelProviderOpenAI,
		},
		Name:               "Text Embedding 3 large",
		ApiModelName:       "text-embedding-3-large",
		CostPer1MIn:        1.00,
		CostPer1MInCached:  0.0,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       1.00,
		ContextWindowSize:  8192,
		Features: []llms.ModelFeature{
			llms.ModelFeatureEmbedding,
		},
	},
	ModelTextEmbeddingADA002: {
		ModelId: llms.ModelId{
			ID:       ModelTextEmbeddingADA002,
			Provider: ModelProviderOpenAI,
		},
		Name:               "Text Embedding 3 large",
		ApiModelName:       "text-embedding-ada-002",
		CostPer1MIn:        1.00,
		CostPer1MInCached:  0.0,
		CostPer1MOutCached: 0.0,
		CostPer1MOut:       1.00,
		ContextWindowSize:  8192,
		Features: []llms.ModelFeature{
			llms.ModelFeatureEmbedding,
		},
	},
}

func init() {
	// Register OpenAI models with the model registry
	for _, model := range OpenAIModels {
		if err := llms.RegisterModel(&model); err != nil {
			klog.Warningf("Failed to register OpenAI model %s: %v", model.ModelId.ID, err)
		}
	}
}
