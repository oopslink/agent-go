package behavior_patterns

import (
	"strings"
	"testing"

	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestCOTIntegration(t *testing.T) {
	// Test the complete COT flow with structured JSON response
	outputChan := make(chan *eventbus.Event, 10)
	processor := newThinkStateProcessor(outputChan)

	// Create a mock context
	ctx := &StepRuntimeContext{
		StepId:    "test-step",
		MessageId: "test-msg",
		ModelId:   llms.ModelId{Provider: "openai", ID: "gpt-4"},
	}

	// Process the response in chunks to simulate streaming
	chunks := []string{
		"Sure, I'll think through this step by step:\n\n{",
		"\n  \"thinking\": \"Let me analyze this problem carefully. First, I need to understand what is being asked. The user wants to know about the benefits of structured output in COT mode. Let me break this down:\\n\\n1. Structured output provides clear separation between reasoning and conclusion\\n2. It makes parsing more reliable than using text indicators\\n3. It enables better automation and downstream processing\\n4. JSON format is widely supported and machine-readable\",",
		"\n  \"final_answer\": \"Structured output in COT mode is highly recommended because it provides clear separation between reasoning process and final conclusions, eliminates dependency on unstable text indicators, and enables reliable automated processing.\"",
		"\n}\n\nThat's my complete analysis.",
	}

	// Process each chunk
	for _, chunk := range chunks {
		part := &llms.TextPart{Text: chunk}
		processor.UpdateThinking(ctx, part)
	}

	// Check if we can extract the final answer
	fullResponse := processor.responseBuffer.String()
	end, err := processor.EndIfGotFinalAnswer(ctx, fullResponse, nil)
	if err != nil {
		t.Fatalf("Error getting final answer: %v", err)
	}

	if end == nil {
		t.Fatal("Expected to get a final answer but got nil")
	}

	if end.FinishReason != llms.FinishReasonNormalEnd {
		t.Errorf("Expected FinishReasonNormalEnd, got %v", end.FinishReason)
	}

	// Verify the response buffer contains the full response
	if !strings.Contains(fullResponse, "thinking") || !strings.Contains(fullResponse, "final_answer") {
		t.Error("Response buffer should contain the complete structured response")
	}

	// Verify the parsed COT response
	cotResp := processor.tryParseStructuredResponse(fullResponse)
	if cotResp == nil {
		t.Fatal("Should be able to parse structured response")
	}

	expectedThinking := "Let me analyze this problem carefully. First, I need to understand what is being asked. The user wants to know about the benefits of structured output in COT mode. Let me break this down:\n\n1. Structured output provides clear separation between reasoning and conclusion\n2. It makes parsing more reliable than using text indicators\n3. It enables better automation and downstream processing\n4. JSON format is widely supported and machine-readable"

	if cotResp.Thinking != expectedThinking {
		t.Errorf("Thinking content mismatch.\nExpected: %q\nGot: %q", expectedThinking, cotResp.Thinking)
	}

	expectedAnswer := "Structured output in COT mode is highly recommended because it provides clear separation between reasoning process and final conclusions, eliminates dependency on unstable text indicators, and enables reliable automated processing."

	if cotResp.FinalAnswer != expectedAnswer {
		t.Errorf("Final answer mismatch.\nExpected: %q\nGot: %q", expectedAnswer, cotResp.FinalAnswer)
	}
}
