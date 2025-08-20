package tools

import (
	"context"
	"time"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func NewSleepTool(duration time.Duration) *SleepTool {
	return &SleepTool{
		Duration: duration,
	}
}

var _ tools.Tool = &SleepTool{}

type SleepTool struct {
	Duration time.Duration
}

// NewToolSleep creates a new sleep tool with the specified default duration.
func NewToolSleep(duration time.Duration) *SleepTool {
	return &SleepTool{
		Duration: duration,
	}
}

// Call implements Tool.
func (t *SleepTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Extract duration from parameters, with fallback to struct field
	duration := t.Duration
	if durationArg, ok := params.Arguments["duration"]; ok {
		if durationStr, ok := durationArg.(string); ok {
			if parsedDuration, err := time.ParseDuration(durationStr); err == nil {
				duration = parsedDuration
			}
		}
	}

	// Sleep for the specified duration or until context is cancelled
	select {
	case <-ctx.Done():
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "sleep cancelled",
			},
		}, ctx.Err()
	case <-time.After(duration):
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success":  true,
				"duration": duration.String(),
			},
		}, nil
	}
}

// Descriptor implements Tool.
func (t *SleepTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "sleep",
		Description: "Sleep for a specified duration. Useful for introducing delays in agent workflows.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"duration": {
					Type:        llms.TypeString,
					Description: "Duration to sleep in Go duration format (e.g., '1s', '500ms', '2m')",
				},
			},
			Required: []string{"duration"},
		},
	}
}
