package behavior_patterns

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCOTResponseParsing(t *testing.T) {
	processor := &thinkStateProcessor{
		responseBuffer: &strings.Builder{},
		cotResponse:    &COTResponse{},
	}

	tests := []struct {
		name             string
		input            string
		expectedValid    bool
		expectedThinking string
		expectedAnswer   string
	}{
		{
			name: "Valid JSON response",
			input: `{
				"thinking": "Let me analyze this step by step. First, I need to understand the problem...",
				"final_answer": "The answer is 42."
			}`,
			expectedValid:    true,
			expectedThinking: "Let me analyze this step by step. First, I need to understand the problem...",
			expectedAnswer:   "The answer is 42.",
		},
		{
			name: "JSON with extra text around",
			input: `Here is my response:
			{
				"thinking": "Breaking down the problem into parts...",
				"final_answer": "Solution found."
			}
			That's my analysis.`,
			expectedValid:    true,
			expectedThinking: "Breaking down the problem into parts...",
			expectedAnswer:   "Solution found.",
		},
		{
			name:          "Invalid JSON",
			input:         "This is just plain text without JSON structure.",
			expectedValid: false,
		},
		{
			name: "Incomplete JSON",
			input: `{
				"thinking": "Partial response...",`,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.tryParseStructuredResponse(tt.input)

			if tt.expectedValid {
				if result == nil {
					t.Errorf("Expected valid response but got nil")
					return
				}

				if result.Thinking != tt.expectedThinking {
					t.Errorf("Expected thinking '%s', got '%s'", tt.expectedThinking, result.Thinking)
				}

				if result.FinalAnswer != tt.expectedAnswer {
					t.Errorf("Expected final answer '%s', got '%s'", tt.expectedAnswer, result.FinalAnswer)
				}
			} else {
				if result != nil {
					t.Errorf("Expected nil response but got %+v", result)
				}
			}
		})
	}
}

func TestCOTResponseMarshaling(t *testing.T) {
	response := COTResponse{
		Thinking:    "Step-by-step analysis here",
		FinalAnswer: "Final conclusion",
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal COTResponse: %v", err)
	}

	var unmarshaled COTResponse
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal COTResponse: %v", err)
	}

	if unmarshaled.Thinking != response.Thinking {
		t.Errorf("Expected thinking '%s', got '%s'", response.Thinking, unmarshaled.Thinking)
	}

	if unmarshaled.FinalAnswer != response.FinalAnswer {
		t.Errorf("Expected final answer '%s', got '%s'", response.FinalAnswer, unmarshaled.FinalAnswer)
	}
}

func TestJSONRegexExtraction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with surrounding text",
			input:    `Here is the JSON: {"thinking": "analysis", "final_answer": "result"} and some more text`,
			expected: `{"thinking": "analysis", "final_answer": "result"}`,
		},
		{
			name: "Multiline JSON",
			input: `{
				"thinking": "multi-line thinking",
				"final_answer": "answer"
			}`,
			expected: `{
				"thinking": "multi-line thinking",
				"final_answer": "answer"
			}`,
		},
		{
			name:     "No JSON",
			input:    `This is just plain text`,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonRegex.FindString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
