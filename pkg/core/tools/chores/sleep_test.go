package tools

import (
	"context"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestSleepTool_Descriptor(t *testing.T) {
	tool := NewSleepTool(time.Second)
	descriptor := tool.Descriptor()

	if descriptor.Name != "sleep" {
		t.Errorf("Expected name 'sleep', got '%s'", descriptor.Name)
	}

	if descriptor.Description == "" {
		t.Error("Expected non-empty description")
	}

	if descriptor.Parameters == nil {
		t.Error("Expected parameters to be defined")
	}

	if descriptor.Parameters.Type != llms.TypeObject {
		t.Errorf("Expected parameters type to be object, got %s", descriptor.Parameters.Type)
	}

	if _, exists := descriptor.Parameters.Properties["duration"]; !exists {
		t.Error("Expected duration property to exist")
	}
}

func TestSleepTool_Call_WithDurationParameter(t *testing.T) {
	tool := NewSleepTool(time.Hour) // Default duration (should be overridden)

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-1",
		Name:       "sleep",
		Arguments: map[string]any{
			"duration": "100ms",
		},
	}

	start := time.Now()
	result, err := tool.Call(ctx, params)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.ToolCallId != params.ToolCallId {
		t.Errorf("Expected tool call ID '%s', got '%s'", params.ToolCallId, result.ToolCallId)
	}

	if result.Name != params.Name {
		t.Errorf("Expected name '%s', got '%s'", params.Name, result.Name)
	}

	success, ok := result.Result["success"].(bool)
	if !ok || !success {
		t.Error("Expected success to be true")
	}

	// Check that it actually slept for approximately the right duration
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Expected to sleep for ~100ms, actual duration: %v", duration)
	}
}

func TestSleepTool_Call_WithDefaultDuration(t *testing.T) {
	tool := NewSleepTool(50 * time.Millisecond)

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-2",
		Name:       "sleep",
		Arguments:  map[string]any{}, // No duration parameter
	}

	start := time.Now()
	result, err := tool.Call(ctx, params)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	success, ok := result.Result["success"].(bool)
	if !ok || !success {
		t.Error("Expected success to be true")
	}

	// Check that it used the default duration
	if duration < 40*time.Millisecond || duration > 100*time.Millisecond {
		t.Errorf("Expected to sleep for ~50ms, actual duration: %v", duration)
	}
}

func TestSleepTool_Call_WithContextCancellation(t *testing.T) {
	tool := NewSleepTool(time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	params := &llms.ToolCall{
		ToolCallId: "test-call-3",
		Name:       "sleep",
		Arguments: map[string]any{
			"duration": "1s",
		},
	}

	// Cancel the context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, err := tool.Call(ctx, params)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	success, ok := result.Result["success"].(bool)
	if !ok || success {
		t.Error("Expected success to be false")
	}

	errorMsg, ok := result.Result["error"].(string)
	if !ok || errorMsg != "sleep cancelled" {
		t.Errorf("Expected error message 'sleep cancelled', got '%s'", errorMsg)
	}

	// Should have been cancelled before completing the full sleep
	if duration > 200*time.Millisecond {
		t.Errorf("Expected to be cancelled quickly, actual duration: %v", duration)
	}
}

func TestSleepTool_Call_WithInvalidDuration(t *testing.T) {
	tool := NewSleepTool(100 * time.Millisecond)

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-4",
		Name:       "sleep",
		Arguments: map[string]any{
			"duration": "invalid-duration",
		},
	}

	// Should fall back to default duration when parsing fails
	start := time.Now()
	result, err := tool.Call(ctx, params)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	success, ok := result.Result["success"].(bool)
	if !ok || !success {
		t.Error("Expected success to be true")
	}

	// Should have used the default duration (100ms)
	if duration < 80*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("Expected to sleep for ~100ms (default), actual duration: %v", duration)
	}
}

func TestNewSleepTool(t *testing.T) {
	duration := 500 * time.Millisecond
	tool := NewSleepTool(duration)

	if tool == nil {
		t.Fatal("Expected tool to be non-nil")
	}

	if tool.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, tool.Duration)
	}
}
