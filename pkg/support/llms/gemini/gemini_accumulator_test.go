package gemini

import (
	"testing"

	"google.golang.org/genai"
)

func TestGeminiChatCompletionAccumulator(t *testing.T) {
	acc := newGeminiChatCompletionAccumulator()

	// Test adding first chunk with text
	chunk1 := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Role: genai.RoleModel,
					Parts: []*genai.Part{
						{Text: "Hello"},
					},
				},
			},
		},
	}
	acc.AddChunk(chunk1)

	// Verify first chunk
	if len(acc.Candidates) != 1 {
		t.Errorf("Expected 1 candidate, got %d", len(acc.Candidates))
	}
	if acc.Candidates[0].Content.Role != genai.RoleModel {
		t.Errorf("Expected role 'model', got '%s'", acc.Candidates[0].Content.Role)
	}
	if len(acc.Candidates[0].Content.Parts) != 1 {
		t.Errorf("Expected 1 part, got %d", len(acc.Candidates[0].Content.Parts))
	}
	if acc.Candidates[0].Content.Parts[0].Text != "Hello" {
		t.Errorf("Expected text 'Hello', got '%s'", acc.Candidates[0].Content.Parts[0].Text)
	}

	// Test adding second chunk with more text (should concatenate)
	chunk2 := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: " world"},
					},
				},
			},
		},
	}
	acc.AddChunk(chunk2)

	// Verify concatenation
	if len(acc.Candidates[0].Content.Parts) != 1 {
		t.Errorf("Expected 1 part after concatenation, got %d", len(acc.Candidates[0].Content.Parts))
	}
	if acc.Candidates[0].Content.Parts[0].Text != "Hello world" {
		t.Errorf("Expected text 'Hello world', got '%s'", acc.Candidates[0].Content.Parts[0].Text)
	}

	// Test adding chunk with function call
	chunk3 := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							FunctionCall: &genai.FunctionCall{
								ID:   "call_123",
								Name: "test_function",
								Args: map[string]interface{}{"arg1": "value1"},
							},
						},
					},
				},
			},
		},
	}
	acc.AddChunk(chunk3)

	// Verify function call was added as new part
	if len(acc.Candidates[0].Content.Parts) != 2 {
		t.Errorf("Expected 2 parts after function call, got %d", len(acc.Candidates[0].Content.Parts))
	}
	if acc.Candidates[0].Content.Parts[1].FunctionCall == nil {
		t.Error("Expected function call part")
	}
	if acc.Candidates[0].Content.Parts[1].FunctionCall.Name != "test_function" {
		t.Errorf("Expected function name 'test_function', got '%s'", acc.Candidates[0].Content.Parts[1].FunctionCall.Name)
	}

	// Test adding chunk with finish reason and usage metadata
	chunk4 := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				FinishReason: genai.FinishReasonStop,
			},
		},
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 15,
			TotalTokenCount:      25,
		},
	}
	acc.AddChunk(chunk4)

	// Test adding another chunk with more tokens (should accumulate)
	chunk5 := &genai.GenerateContentResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			CandidatesTokenCount: 5,  // Additional tokens
			TotalTokenCount:      30, // This will be recalculated
		},
	}
	acc.AddChunk(chunk5)

	// Verify finish reason and accumulated usage metadata
	if acc.Candidates[0].FinishReason != genai.FinishReasonStop {
		t.Errorf("Expected finish reason 'STOP', got '%s'", acc.Candidates[0].FinishReason)
	}
	if acc.UsageMetadata == nil {
		t.Error("Expected usage metadata")
	}
	if acc.UsageMetadata.PromptTokenCount != 10 {
		t.Errorf("Expected prompt token count 10, got %d", acc.UsageMetadata.PromptTokenCount)
	}
	// Should accumulate: 15 + 5 = 20
	if acc.UsageMetadata.CandidatesTokenCount != 20 {
		t.Errorf("Expected candidates token count 20 (accumulated), got %d", acc.UsageMetadata.CandidatesTokenCount)
	}
	// Should be recalculated: 10 + 20 = 30
	if acc.UsageMetadata.TotalTokenCount != 30 {
		t.Errorf("Expected total token count 30 (recalculated), got %d", acc.UsageMetadata.TotalTokenCount)
	}
}

func TestGeminiChatCompletionAccumulatorNilChunk(t *testing.T) {
	acc := newGeminiChatCompletionAccumulator()

	// Test that nil chunk doesn't crash
	acc.AddChunk(nil)

	// Should have no candidates
	if len(acc.Candidates) != 0 {
		t.Errorf("Expected 0 candidates after nil chunk, got %d", len(acc.Candidates))
	}
}

func TestGeminiChatCompletionAccumulatorEmptyChunk(t *testing.T) {
	acc := newGeminiChatCompletionAccumulator()

	// Test empty chunk
	acc.AddChunk(&genai.GenerateContentResponse{})

	// Should have no candidates
	if len(acc.Candidates) != 0 {
		t.Errorf("Expected 0 candidates after empty chunk, got %d", len(acc.Candidates))
	}
}

func TestGeminiChatCompletionAccumulatorUsageMetadata(t *testing.T) {
	acc := newGeminiChatCompletionAccumulator()

	// First chunk with prompt tokens and some candidate tokens
	chunk1 := &genai.GenerateContentResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:        10,
			CandidatesTokenCount:    3,
			TotalTokenCount:         13,
			CachedContentTokenCount: 5,
		},
	}
	acc.AddChunk(chunk1)

	// Verify first chunk
	if acc.UsageMetadata.PromptTokenCount != 10 {
		t.Errorf("Expected prompt token count 10, got %d", acc.UsageMetadata.PromptTokenCount)
	}
	if acc.UsageMetadata.CandidatesTokenCount != 3 {
		t.Errorf("Expected candidates token count 3, got %d", acc.UsageMetadata.CandidatesTokenCount)
	}
	if acc.UsageMetadata.TotalTokenCount != 13 {
		t.Errorf("Expected total token count 13, got %d", acc.UsageMetadata.TotalTokenCount)
	}
	if acc.UsageMetadata.CachedContentTokenCount != 5 {
		t.Errorf("Expected cached content token count 5, got %d", acc.UsageMetadata.CachedContentTokenCount)
	}

	// Second chunk with more candidate tokens
	chunk2 := &genai.GenerateContentResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			CandidatesTokenCount: 5, // Additional tokens
		},
	}
	acc.AddChunk(chunk2)

	// Verify accumulation
	if acc.UsageMetadata.PromptTokenCount != 10 {
		t.Errorf("Expected prompt token count 10 (unchanged), got %d", acc.UsageMetadata.PromptTokenCount)
	}
	if acc.UsageMetadata.CandidatesTokenCount != 8 { // 3 + 5
		t.Errorf("Expected candidates token count 8 (accumulated), got %d", acc.UsageMetadata.CandidatesTokenCount)
	}
	if acc.UsageMetadata.TotalTokenCount != 18 { // 10 + 8
		t.Errorf("Expected total token count 18 (recalculated), got %d", acc.UsageMetadata.TotalTokenCount)
	}
	if acc.UsageMetadata.CachedContentTokenCount != 5 {
		t.Errorf("Expected cached content token count 5 (unchanged), got %d", acc.UsageMetadata.CachedContentTokenCount)
	}

	// Third chunk with more candidate tokens
	chunk3 := &genai.GenerateContentResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			CandidatesTokenCount: 2, // Additional tokens
		},
	}
	acc.AddChunk(chunk3)

	// Verify final accumulation
	if acc.UsageMetadata.CandidatesTokenCount != 10 { // 8 + 2
		t.Errorf("Expected candidates token count 10 (accumulated), got %d", acc.UsageMetadata.CandidatesTokenCount)
	}
	if acc.UsageMetadata.TotalTokenCount != 20 { // 10 + 10
		t.Errorf("Expected total token count 20 (recalculated), got %d", acc.UsageMetadata.TotalTokenCount)
	}
}
