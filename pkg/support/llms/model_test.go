package llms

import (
	"testing"
)

func TestModelProvider(t *testing.T) {
	provider := ModelProvider("openai")
	if string(provider) != "openai" {
		t.Errorf("ModelProvider = %v, want openai", provider)
	}
}

func TestModelFeature(t *testing.T) {
	tests := []struct {
		name     string
		feature  ModelFeature
		expected string
	}{
		{"reasoning", ModelFeatureReasoning, "can_reason"},
		{"attachment", ModelFeatureAttachment, "supports_attachments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.feature) != tt.expected {
				t.Errorf("ModelFeature = %v, want %v", tt.feature, tt.expected)
			}
		})
	}
}

func TestModelId(t *testing.T) {
	modelId := ModelId{
		Provider: "openai",
		ID:       "gpt-4",
	}

	if modelId.Provider != "openai" {
		t.Errorf("ModelId.Provider = %v, want openai", modelId.Provider)
	}

	if modelId.ID != "gpt-4" {
		t.Errorf("ModelId.ID = %v, want gpt-4", modelId.ID)
	}
}

func TestModel(t *testing.T) {
	model := &Model{
		ModelId: ModelId{
			Provider: "openai",
			ID:       "gpt-4",
		},
		Name:               "GPT-4",
		CostPer1MIn:        0.03,
		CostPer1MOut:       0.06,
		CostPer1MInCached:  0.01,
		CostPer1MOutCached: 0.03,
		ContextWindowSize:  8192,
		DefaultMaxTokens:   4096,
		Features: []ModelFeature{
			ModelFeatureReasoning,
			ModelFeatureAttachment,
		},
	}

	// Test embedded ModelId fields
	if model.Provider != "openai" {
		t.Errorf("Model.Provider = %v, want openai", model.Provider)
	}

	if model.ID != "gpt-4" {
		t.Errorf("Model.ID = %v, want gpt-4", model.ID)
	}

	// Test Model fields
	if model.Name != "GPT-4" {
		t.Errorf("Model.Name = %v, want GPT-4", model.Name)
	}

	if model.CostPer1MIn != 0.03 {
		t.Errorf("Model.CostPer1MIn = %v, want 0.03", model.CostPer1MIn)
	}

	if model.CostPer1MOut != 0.06 {
		t.Errorf("Model.CostPer1MOut = %v, want 0.06", model.CostPer1MOut)
	}

	if model.CostPer1MInCached != 0.01 {
		t.Errorf("Model.CostPer1MInCached = %v, want 0.01", model.CostPer1MInCached)
	}

	if model.CostPer1MOutCached != 0.03 {
		t.Errorf("Model.CostPer1MOutCached = %v, want 0.03", model.CostPer1MOutCached)
	}

	if model.ContextWindowSize != 8192 {
		t.Errorf("Model.ContextWindowSize = %v, want 8192", model.ContextWindowSize)
	}

	if model.DefaultMaxTokens != 4096 {
		t.Errorf("Model.DefaultMaxTokens = %v, want 4096", model.DefaultMaxTokens)
	}

	if len(model.Features) != 2 {
		t.Errorf("Model.Features length = %d, want 2", len(model.Features))
	}
}

func TestModelIsSupport(t *testing.T) {
	model := &Model{
		Features: []ModelFeature{
			ModelFeatureReasoning,
			ModelFeatureAttachment,
		},
	}

	// Test supported features
	if !model.IsSupport(ModelFeatureReasoning) {
		t.Error("Model.IsSupport(ModelFeatureReasoning) = false, want true")
	}

	if !model.IsSupport(ModelFeatureAttachment) {
		t.Error("Model.IsSupport(ModelFeatureAttachment) = false, want true")
	}

	// Test unsupported feature
	if model.IsSupport("unsupported_feature") {
		t.Error("Model.IsSupport(unsupported_feature) = true, want false")
	}
}

func TestModelIsSupportEmptyFeatures(t *testing.T) {
	model := &Model{
		Features: []ModelFeature{},
	}

	// Test with empty features list
	if model.IsSupport(ModelFeatureReasoning) {
		t.Error("Model.IsSupport(ModelFeatureReasoning) = true, want false for empty features")
	}

	if model.IsSupport(ModelFeatureAttachment) {
		t.Error("Model.IsSupport(ModelFeatureAttachment) = true, want false for empty features")
	}
}

func TestModelIsSupportNilFeatures(t *testing.T) {
	model := &Model{
		Features: nil,
	}

	// Test with nil features list
	if model.IsSupport(ModelFeatureReasoning) {
		t.Error("Model.IsSupport(ModelFeatureReasoning) = true, want false for nil features")
	}

	if model.IsSupport(ModelFeatureAttachment) {
		t.Error("Model.IsSupport(ModelFeatureAttachment) = true, want false for nil features")
	}
}

func TestModelWithPartialFeatures(t *testing.T) {
	model := &Model{
		Features: []ModelFeature{
			ModelFeatureReasoning,
		},
	}

	// Test supported feature
	if !model.IsSupport(ModelFeatureReasoning) {
		t.Error("Model.IsSupport(ModelFeatureReasoning) = false, want true")
	}

	// Test unsupported feature
	if model.IsSupport(ModelFeatureAttachment) {
		t.Error("Model.IsSupport(ModelFeatureAttachment) = true, want false")
	}
}

func TestModelCostCalculation(t *testing.T) {
	model := &Model{
		CostPer1MIn:  0.03,
		CostPer1MOut: 0.06,
	}

	// Test cost calculations (these are just basic field access tests)
	if model.CostPer1MIn != 0.03 {
		t.Errorf("Model.CostPer1MIn = %v, want 0.03", model.CostPer1MIn)
	}

	if model.CostPer1MOut != 0.06 {
		t.Errorf("Model.CostPer1MOut = %v, want 0.06", model.CostPer1MOut)
	}

	// Test that cached costs can be different from regular costs
	if model.CostPer1MInCached == model.CostPer1MIn {
		t.Log("Model.CostPer1MInCached equals Model.CostPer1MIn, which is valid but worth noting")
	}

	if model.CostPer1MOutCached == model.CostPer1MOut {
		t.Log("Model.CostPer1MOutCached equals Model.CostPer1MOut, which is valid but worth noting")
	}
}

func TestModelContextWindow(t *testing.T) {
	model := &Model{
		ContextWindowSize: 8192,
		DefaultMaxTokens:  4096,
	}

	if model.ContextWindowSize != 8192 {
		t.Errorf("Model.ContextWindowSize = %v, want 8192", model.ContextWindowSize)
	}

	if model.DefaultMaxTokens != 4096 {
		t.Errorf("Model.DefaultMaxTokens = %v, want 4096", model.DefaultMaxTokens)
	}

	// Test that default max tokens is reasonable compared to context window
	if model.DefaultMaxTokens > model.ContextWindowSize {
		t.Errorf("Model.DefaultMaxTokens (%d) > Model.ContextWindowSize (%d), which is invalid",
			model.DefaultMaxTokens, model.ContextWindowSize)
	}
}
