package chat

import (
	"context"
	"fmt"
	
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/oopslink/agent-go/pkg/support/journal"
)

// SessionContext represents the session context for a conversation
type SessionContext struct {
	SessionId string
	Context   context.Context
}

type ConversationContext struct {
	SessionContext

	AgentContext agent.Context
	AgentInput   chan<- *eventbus.Event
}

type AgentResponse struct {
	Messages  []*llms.Message
	ToolCalls []*llms.ToolCall
}

type ConversationHandler interface {
	OnResponse(ctx *ConversationContext, agentResponse *AgentResponse) error
}

type Conversation struct {
	theAgent agent.Agent
	
	// Temporary storage for collecting events until ResponseEnd
	currentMessages  []*llms.Message
	currentToolCalls []*llms.ToolCall
}

// NewConversation creates a new conversation instance
func NewConversation(agent agent.Agent) *Conversation {
	return &Conversation{
		theAgent: agent,
	}
}

func (c *Conversation) Ask(ctx context.Context, question string, handler ConversationHandler) error {
	// Generate a unique session ID
	sessionId := fmt.Sprintf("conversation-%s", utils.GenerateUUID())
	
	// Start the agent
	inputChan, outputChan, err := c.theAgent.Run(&agent.RunContext{
		SessionId: sessionId,
		Context:   ctx,
	})
	if err != nil {
		return errors.Errorf(errors.InternalError, "failed to start agent: %v", err)
	}

	// Create conversation context
	conversationCtx := &ConversationContext{
		SessionContext: SessionContext{
			SessionId: sessionId,
			Context:   ctx,
		},
		AgentInput: inputChan,
	}

	// Send the user message
	inputChan <- agent.NewUserRequestEvent(&agent.UserRequest{
		Message: question,
	})

	// Process agent responses
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-outputChan:
			if !ok {
				// Channel closed, conversation ended
				return nil
			}
			
			done, err := c.handleAgentEvent(conversationCtx, event, handler)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

func (c *Conversation) handleAgentEvent(conversationCtx *ConversationContext, event *eventbus.Event, handler ConversationHandler) (bool, error) {
	switch event.Topic {
	case agent.EventTypeAgentMessage:
		if messageEvent := agent.GetAgentMessageEventData(event); messageEvent != nil {
			c.currentMessages = append(c.currentMessages, messageEvent.Message)
		}
		
	case agent.EventTypeExternalAction:
		if externalAction := agent.GetExternalActionEventData(event); externalAction != nil {
			if externalAction.ToolCall != nil {
				c.currentToolCalls = append(c.currentToolCalls, externalAction.ToolCall)
			}
		}
		
	case agent.EventTypeAgentResponseEnd:
		if endEvent := agent.GetAgentResponseEndEventData(event); endEvent != nil {
			if endEvent.Error != nil && !endEvent.Abort {
				return false, endEvent.Error
			}
			
			// Now call the handler with all collected responses
			if len(c.currentMessages) > 0 || len(c.currentToolCalls) > 0 {
				agentResponse := &AgentResponse{
					Messages:  c.currentMessages,
					ToolCalls: c.currentToolCalls,
				}
				if err := handler.OnResponse(conversationCtx, agentResponse); err != nil {
					return false, err
				}
			}
			
			// Clear the collected events for next round
			c.currentMessages = nil
			c.currentToolCalls = nil
			
			return true, nil // Signal to stop processing
		}
		
	default:
		// Log and ignore other event types
		journal.Info("conversation", "agent",
			fmt.Sprintf("ignoring event type: %s", event.Topic))
	}
	
	return false, nil
}