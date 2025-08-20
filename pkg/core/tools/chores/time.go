// Package tools provides various utility tools for agent workflows, including time-related operations.
package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// NewCurrentTimeTool creates a new current time tool.
// This tool provides current date and time information with support for:
// - Multiple timezones (UTC, America/New_York, Asia/Shanghai, etc.)
// - Various output formats (RFC3339, DateTime, custom formats)
// - Rich output including Unix timestamp and ISO format
func NewCurrentTimeTool() *CurrentTimeTool {
	return &CurrentTimeTool{}
}

var _ tools.Tool = &CurrentTimeTool{}

// CurrentTimeTool is a tool that provides current date and time information.
// It supports different timezones and output formats, making it useful for
// agent workflows that need to work with time-sensitive data or scheduling.
type CurrentTimeTool struct{}

// Call implements Tool.
func (t *CurrentTimeTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Get current time
	now := time.Now()

	// Extract timezone from parameters (default to local)
	timezone := "Local"
	if timezoneArg, ok := params.Arguments["timezone"]; ok {
		if timezoneStr, ok := timezoneArg.(string); ok {
			timezone = timezoneStr
		}
	}

	// Extract format from parameters (default to RFC3339)
	format := time.RFC3339
	if formatArg, ok := params.Arguments["format"]; ok {
		if formatStr, ok := formatArg.(string); ok {
			format = formatStr
		}
	}

	// Load timezone if specified
	var loc *time.Location = now.Location()
	if timezone != "Local" {
		if tz, err := time.LoadLocation(timezone); err == nil {
			loc = tz
			now = now.In(loc)
		} else {
			return &llms.ToolCallResult{
				ToolCallId: params.ToolCallId,
				Name:       params.Name,
				Result: map[string]any{
					"success": false,
					"error":   fmt.Sprintf("invalid timezone: %s", timezone),
				},
			}, nil
		}
	}

	// Format the time
	var formattedTime string
	switch format {
	case "RFC3339":
		formattedTime = now.Format(time.RFC3339)
	case "RFC822":
		formattedTime = now.Format(time.RFC822)
	case "Kitchen":
		formattedTime = now.Format(time.Kitchen)
	case "Stamp":
		formattedTime = now.Format(time.Stamp)
	case "DateTime":
		formattedTime = now.Format("2006-01-02 15:04:05")
	case "Date":
		formattedTime = now.Format("2006-01-02")
	case "Time":
		formattedTime = now.Format("15:04:05")
	default:
		// Use custom format
		formattedTime = now.Format(format)
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success":        true,
			"current_time":   formattedTime,
			"timezone":       loc.String(),
			"unix_timestamp": now.Unix(),
			"iso_format":     now.Format(time.RFC3339),
		},
	}, nil
}

// Descriptor implements Tool.
func (t *CurrentTimeTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "current_time",
		Description: "Get the current date and time. Supports different timezones and output formats.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"timezone": {
					Type:        llms.TypeString,
					Description: "Timezone to display the time in (e.g., 'UTC', 'America/New_York', 'Asia/Shanghai'). Defaults to local timezone.",
				},
				"format": {
					Type:        llms.TypeString,
					Description: "Output format. Can be 'RFC3339', 'RFC822', 'Kitchen', 'Stamp', 'DateTime', 'Date', 'Time', or a custom Go time format string. Defaults to 'RFC3339'.",
				},
			},
			Required: []string{},
		},
	}
}
