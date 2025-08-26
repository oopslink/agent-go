package chat

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// MockAgent for testing
type mockAgent struct {
	messages []string
	delay    time.Duration
}

func (m *mockAgent) Run(ctx *agent.RunContext) (chan<- *eventbus.Event, <-chan *eventbus.Event, error) {
	inputChan := make(chan *eventbus.Event, 10)
	outputChan := make(chan *eventbus.Event, 10)

	go func() {
		defer close(outputChan)
		for {
			select {
			case <-ctx.Context.Done():
				return
			case event := <-inputChan:
				if event.Topic == agent.EventTypeUserRequest {
					userRequest := agent.GetUserRequestEventData(event)
					
					// Simulate agent processing delay
					if m.delay > 0 {
						time.Sleep(m.delay)
					}
					
					// Send back a message
					message := llms.NewAssistantMessage("Response to: " + userRequest.Message)
					outputChan <- agent.NewAgentMessageEvent("test-trace", message)
					
					// Send end event
					outputChan <- agent.NewAgentResponseEndEvent("test-trace", &agent.AgentResponseEnd{
						FinishReason: llms.FinishReasonStop,
					})
				}
			}
		}
	}()

	return inputChan, outputChan, nil
}

// MockHandler for testing
type mockHandler struct {
	responses []*AgentResponse
	mu        sync.Mutex
}

func (h *mockHandler) OnResponse(ctx *ConversationContext, agentResponse *AgentResponse) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.responses = append(h.responses, agentResponse)
	return nil
}

func (h *mockHandler) getResponses() []*AgentResponse {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]*AgentResponse, len(h.responses))
	copy(result, h.responses)
	return result
}

func TestConversation_Run(t *testing.T) {
	// Create mock agent
	mockAgent := &mockAgent{
		delay: 10 * time.Millisecond,
	}

	// Create conversation
	conversation := NewConversation(mockAgent)

	// Create mock handler
	handler := &mockHandler{}

	// Run conversation
	ctx := context.Background()
	testMessage := "Hello, how are you?"

	err := conversation.Ask(ctx, testMessage, handler)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check responses
	responses := handler.getResponses()
	if len(responses) != 1 {
		t.Fatalf("Expected 1 response, got %d", len(responses))
	}

	if responses[0].Message == nil {
		t.Fatalf("Expected message response, got nil")
	}

	// Check message content
	if len(responses[0].Message.Parts) == 0 {
		t.Fatalf("Expected message parts, got empty")
	}

	textPart, ok := responses[0].Message.Parts[0].(*llms.TextPart)
	if !ok {
		t.Fatalf("Expected text part, got different type")
	}

	expectedContent := "Response to: " + testMessage
	if textPart.Text != expectedContent {
		t.Fatalf("Expected content '%s', got '%s'", expectedContent, textPart.Text)
	}
}

func TestConversation_ContextCancellation(t *testing.T) {
	// Create mock agent with longer delay
	mockAgent := &mockAgent{
		delay: 100 * time.Millisecond,
	}

	// Create conversation
	conversation := NewConversation(mockAgent)

	// Create mock handler
	handler := &mockHandler{}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Run conversation
	err := conversation.Ask(ctx, "Test message", handler)
	
	// Should return context cancellation error
	if err == nil {
		t.Fatalf("Expected context cancellation error, got nil")
	}

	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled, got: %v", err)
	}
}
