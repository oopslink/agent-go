// Package types provides core data structures and types for the agent-go framework.
// This file contains model-related types for AI model configuration and features.
package llms

import (
	"fmt"
	"slices"
)

// ModelProvider represents the provider of an AI model (e.g., OpenAI, Anthropic).
type ModelProvider string

// ModelFeature represents a capability or feature of an AI model.
type ModelFeature string

const (
	ModelFeatureCompletion ModelFeature = "can_completion"       // Model can perform completions
	ModelFeatureEmbedding  ModelFeature = "can_embedding"        // Model can perform embedding
	ModelFeatureReasoning  ModelFeature = "can_reason"           // Model can perform reasoning tasks
	ModelFeatureAttachment ModelFeature = "supports_attachments" // Model supports file attachments
)

// ModelId uniquely identifies an AI model by its provider and ID.
type ModelId struct {
	Provider ModelProvider `json:"provider"` // The provider of the model (e.g., "openai")
	ID       string        `json:"id"`       // The specific model ID (e.g., "gpt-4")
}

func (m *ModelId) String() string {
	return fmt.Sprintf("%s/%s", m.Provider, m.ID)
}

// Model represents a complete AI model configuration including pricing and capabilities.
// It embeds ModelId to inherit the provider and ID fields.
type Model struct {
	ModelId            `json:"_"`     // Embedded ModelId for provider and ID
	Name               string         `json:"name"`                   // Human-readable name of the model
	ApiModelName       string         `json:"api_model_name"`         // Model name for openapi
	CostPer1MIn        float64        `json:"cost_per_1m_in"`         // Cost per 1M input tokens
	CostPer1MOut       float64        `json:"cost_per_1m_out"`        // Cost per 1M output tokens
	CostPer1MInCached  float64        `json:"cost_per_1m_in_cached"`  // Cost per 1M cached input tokens
	CostPer1MOutCached float64        `json:"cost_per_1m_out_cached"` // Cost per 1M cached output tokens
	ContextWindowSize  int64          `json:"context_window_size"`    // Maximum context window size in tokens
	DefaultMaxTokens   int64          `json:"default_max_tokens"`     // Default maximum tokens for responses
	Features           []ModelFeature `json:"features,omitempty"`     // List of model capabilities
}

// IsSupport checks if the model supports a specific feature.
// Returns true if the feature is in the model's features list.
func (m *Model) IsSupport(f ModelFeature) bool {
	return slices.Contains(m.Features, f)
}
