package behavior_patterns

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReActPattern(t *testing.T) {
	pattern, err := NewReActPattern(5)
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)

	reactPattern, ok := pattern.(*reactPattern)
	assert.True(t, ok)
	assert.Equal(t, 5, reactPattern.maxIterations)
}

func TestReActPatternSystemInstruction(t *testing.T) {
	pattern, err := NewReActPattern(5)
	require.NoError(t, err)

	header := "Test header"
	instruction := pattern.SystemInstruction(header)

	assert.Contains(t, instruction, header)
	assert.Contains(t, instruction, _reactPrompt)
}

func TestReActResponseMarshaling(t *testing.T) {
	response := ReActResponse{
		Thought:     "I need to think about this",
		Action:      "search",
		ToolCalls:   []*llms.ToolCall{{Name: "search", ToolCallId: "1"}},
		Observation: "Found some information",
		Answer:      "The answer is 42",
		Continue:    true,
	}

	jsonBytes, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	var unmarshaled ReActResponse
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, response.Thought, unmarshaled.Thought)
	assert.Equal(t, response.Action, unmarshaled.Action)
	assert.Equal(t, response.Answer, unmarshaled.Answer)
	assert.Equal(t, response.Continue, unmarshaled.Continue)
}

func TestReActResponseFormatted(t *testing.T) {
	response := ReActResponse{
		Thought:  "Test thought",
		Action:   "test_action",
		Continue: true,
	}

	formatted := response.Formatted()
	assert.NotEmpty(t, formatted)

	var parsed ReActResponse
	err := json.Unmarshal([]byte(formatted), &parsed)
	assert.NoError(t, err)

	assert.Equal(t, response.Thought, parsed.Thought)
	assert.Equal(t, response.Action, parsed.Action)
	assert.Equal(t, response.Continue, parsed.Continue)
}

func TestReActStateProcessorParseReActResponse(t *testing.T) {
	processor := &reactStateProcessor{
		finalAnswer: &strings.Builder{},
		jsonBuffer:  &strings.Builder{},
	}

	tests := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name: "Single valid JSON response",
			input: `{
				"thought": "I need to search for information",
				"action": "search",
				"continue": true
			}`,
			expectedCount: 1,
		},
		{
			name:          "Invalid JSON",
			input:         "This is not valid JSON",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor.jsonBuffer.Reset()
			responses := processor.parseReActResponse(tt.input)
			assert.Len(t, responses, tt.expectedCount)
		})
	}
}

func TestReActStateProcessorExtractJSONBlocks(t *testing.T) {
	processor := &reactStateProcessor{}

	tests := []struct {
		name           string
		input          string
		expectedBlocks int
	}{
		{
			name:           "Single JSON block",
			input:          `{"key": "value"}`,
			expectedBlocks: 1,
		},
		{
			name:           "No JSON blocks",
			input:          "Just plain text",
			expectedBlocks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := processor.extractJSONBlocks(tt.input)
			assert.Len(t, blocks, tt.expectedBlocks)
		})
	}
}

func TestReActStateProcessorUpdateReAct(t *testing.T) {
	// Create a buffered channel to prevent blocking
	outputChan := make(chan *eventbus.Event, 10)

	processor := &reactStateProcessor{
		finalAnswer: &strings.Builder{},
		jsonBuffer:  &strings.Builder{},
		outputChan:  outputChan,
	}

	ctx := &StepRuntimeContext{
		StepId:    "test-step",
		MessageId: "test-message",
		ModelId:   llms.ModelId{Provider: "test", ID: "model"},
	}

	// Test with tool calls
	text := `{
		"thought": "I need to search",
		"tool_calls": [{"name": "search", "tool_call_id": "1"}],
		"continue": true
	}`

	part := &llms.TextPart{Text: text}
	toolCalls, err := processor.UpdateReAct(ctx, part)

	assert.NoError(t, err)
	assert.Len(t, toolCalls, 1)

	// Clean up
	close(outputChan)
}

func TestReActStateProcessorEndIfGotFinalAnswer(t *testing.T) {
	// Create a buffered channel to prevent blocking
	outputChan := make(chan *eventbus.Event, 10)

	processor := &reactStateProcessor{
		finalAnswer: &strings.Builder{},
		outputChan:  outputChan,
	}

	ctx := &StepRuntimeContext{
		StepId:    "test-step",
		MessageId: "test-message",
		ModelId:   llms.ModelId{Provider: "test", ID: "model"},
	}

	// Test with no final answer
	end, err := processor.EndIfGotFinalAnswer(ctx, "full message", nil)
	assert.NoError(t, err)
	assert.Nil(t, end)

	// Test with final answer
	processor.finalAnswer.WriteString("The answer is 42")
	end, err = processor.EndIfGotFinalAnswer(ctx, "full message", nil)
	assert.NoError(t, err)
	assert.NotNil(t, end)
	assert.Equal(t, llms.FinishReasonNormalEnd, end.FinishReason)

	// Clean up
	close(outputChan)
}
