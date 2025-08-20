package tools

import (
	"context"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentTimeTool_Descriptor(t *testing.T) {
	tool := NewCurrentTimeTool()
	descriptor := tool.Descriptor()

	assert.Equal(t, "current_time", descriptor.Name)
	assert.Contains(t, descriptor.Description, "current date and time")
	assert.Equal(t, llms.TypeObject, descriptor.Parameters.Type)
	assert.Contains(t, descriptor.Parameters.Properties, "timezone")
	assert.Contains(t, descriptor.Parameters.Properties, "format")
	assert.Empty(t, descriptor.Parameters.Required) // No required parameters
}

func TestCurrentTimeTool_Call_Default(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx := context.Background()

	params := &llms.ToolCall{
		ToolCallId: "test-call-1",
		Name:       "current_time",
		Arguments:  map[string]any{},
	}

	result, err := tool.Call(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test-call-1", result.ToolCallId)
	assert.Equal(t, "current_time", result.Name)

	assert.Equal(t, true, result.Result["success"])
	assert.Contains(t, result.Result, "current_time")
	assert.Contains(t, result.Result, "timezone")
	assert.Contains(t, result.Result, "unix_timestamp")
	assert.Contains(t, result.Result, "iso_format")

	// Verify the time is recent (within last minute)
	unixTime, ok := result.Result["unix_timestamp"].(int64)
	require.True(t, ok)
	timeDiff := time.Now().Unix() - unixTime
	assert.True(t, timeDiff >= 0 && timeDiff < 60, "Time should be within the last minute")
}

func TestCurrentTimeTool_Call_WithTimezone(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx := context.Background()

	params := &llms.ToolCall{
		ToolCallId: "test-call-2",
		Name:       "current_time",
		Arguments: map[string]any{
			"timezone": "UTC",
		},
	}

	result, err := tool.Call(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, true, result.Result["success"])
	assert.Contains(t, result.Result["timezone"], "UTC")
}

func TestCurrentTimeTool_Call_WithFormat(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx := context.Background()

	testCases := []struct {
		name           string
		format         string
		expectedFormat string
	}{
		{"RFC822", "RFC822", time.RFC822},
		{"Kitchen", "Kitchen", time.Kitchen},
		{"DateTime", "DateTime", "2006-01-02 15:04:05"},
		{"Date", "Date", "2006-01-02"},
		{"Time", "Time", "15:04:05"},
		{"Custom", "2006/01/02", "2006/01/02"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := &llms.ToolCall{
				ToolCallId: "test-call-format",
				Name:       "current_time",
				Arguments: map[string]any{
					"format": tc.format,
				},
			}

			result, err := tool.Call(ctx, params)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, true, result.Result["success"])
			currentTime, ok := result.Result["current_time"].(string)
			require.True(t, ok)
			assert.NotEmpty(t, currentTime)

			// Try to parse the formatted time to ensure it's valid
			if tc.format == "Date" {
				_, err := time.Parse("2006-01-02", currentTime)
				assert.NoError(t, err)
			} else if tc.format == "Time" {
				_, err := time.Parse("15:04:05", currentTime)
				assert.NoError(t, err)
			}
		})
	}
}

func TestCurrentTimeTool_Call_InvalidTimezone(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx := context.Background()

	params := &llms.ToolCall{
		ToolCallId: "test-call-invalid-tz",
		Name:       "current_time",
		Arguments: map[string]any{
			"timezone": "Invalid/Timezone",
		},
	}

	result, err := tool.Call(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, false, result.Result["success"])
	assert.Contains(t, result.Result["error"], "invalid timezone")
}

func TestCurrentTimeTool_Call_WithComplexTimezone(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx := context.Background()

	// Test with a complex timezone
	params := &llms.ToolCall{
		ToolCallId: "test-call-complex-tz",
		Name:       "current_time",
		Arguments: map[string]any{
			"timezone": "America/New_York",
			"format":   "DateTime",
		},
	}

	result, err := tool.Call(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, true, result.Result["success"])
	assert.Contains(t, result.Result["timezone"], "America/New_York")

	currentTime, ok := result.Result["current_time"].(string)
	require.True(t, ok)

	// Verify the format matches our expected DateTime format
	_, err = time.Parse("2006-01-02 15:04:05", currentTime)
	assert.NoError(t, err)
}

func TestCurrentTimeTool_Call_CancelledContext(t *testing.T) {
	tool := NewCurrentTimeTool()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := &llms.ToolCall{
		ToolCallId: "test-call-cancelled",
		Name:       "current_time",
		Arguments:  map[string]any{},
	}

	// The tool should still work even with cancelled context since it doesn't do any long-running operations
	result, err := tool.Call(ctx, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, true, result.Result["success"])
}
