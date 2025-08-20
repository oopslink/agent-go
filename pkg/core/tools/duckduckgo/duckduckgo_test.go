package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestDuckDuckGoTool_Descriptor(t *testing.T) {
	tool := NewDuckDuckGoTool()
	descriptor := tool.Descriptor()

	assert.Equal(t, "duckduckgo_search", descriptor.Name)
	assert.Contains(t, descriptor.Description, "Search DuckDuckGo")
	assert.NotNil(t, descriptor.Parameters)
	assert.Equal(t, llms.TypeObject, descriptor.Parameters.Type)
	assert.Contains(t, descriptor.Parameters.Properties, "query")
	assert.Equal(t, []string{"query"}, descriptor.Parameters.Required)
}

func TestDuckDuckGoTool_Call_EmptyQuery(t *testing.T) {
	tool := NewDuckDuckGoTool()
	params := &llms.ToolCall{
		ToolCallId: "test-id",
		Name:       "duckduckgo_search",
		Arguments: map[string]any{
			"query": "",
		},
	}

	result, err := tool.Call(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Result["success"].(bool))
	assert.Contains(t, result.Result["error"].(string), "cannot be empty")
}

func TestDuckDuckGoTool_Call_MissingQuery(t *testing.T) {
	tool := NewDuckDuckGoTool()
	params := &llms.ToolCall{
		ToolCallId: "test-id",
		Name:       "duckduckgo_search",
		Arguments:  map[string]any{},
	}

	result, err := tool.Call(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Result["success"].(bool))
	assert.Contains(t, result.Result["error"].(string), "query parameter is required")
}

func TestDuckDuckGoTool_Call_InvalidQueryType(t *testing.T) {
	tool := NewDuckDuckGoTool()
	params := &llms.ToolCall{
		ToolCallId: "test-id",
		Name:       "duckduckgo_search",
		Arguments: map[string]any{
			"query": 123,
		},
	}

	result, err := tool.Call(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Result["success"].(bool))
	assert.Contains(t, result.Result["error"].(string), "query parameter must be a string")
}

func TestDuckDuckGoTool_Call_ValidQuery(t *testing.T) {
	tool := NewDuckDuckGoTool()
	params := &llms.ToolCall{
		ToolCallId: "test-id",
		Name:       "duckduckgo_search",
		Arguments: map[string]any{
			"query": "golang programming",
		},
	}

	result, err := tool.Call(context.Background(), params)
	require.NoError(t, err)

	// The search might succeed or fail depending on network/availability
	// We just check the structure is correct
	assert.NotNil(t, result)
	assert.Equal(t, "test-id", result.ToolCallId)
	assert.Equal(t, "duckduckgo_search", result.Name)

	if success, ok := result.Result["success"].(bool); ok && success {
		// If search succeeded, verify the structure
		assert.Contains(t, result.Result, "query")
		assert.Contains(t, result.Result, "results")
		assert.Contains(t, result.Result, "count")

		query := result.Result["query"].(string)
		assert.Equal(t, "golang programming", query)

		count := result.Result["count"].(int)
		assert.GreaterOrEqual(t, count, 0)
	} else {
		// If search failed, there should be an error message
		assert.Contains(t, result.Result, "error")
	}
}

func TestDuckDuckGoTool_CleanURL(t *testing.T) {
	tool := NewDuckDuckGoTool()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "duckduckgo redirect url",
			input:    "/l/?uddg=https%3A//example.com",
			expected: "https://example.com",
		},
		{
			name:     "full url",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "relative path",
			input:    "/some/path",
			expected: "https://duckduckgo.com/some/path",
		},
		{
			name:     "invalid redirect url",
			input:    "/l/?uddg=invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tool.cleanURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
