package agent

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"sync/atomic"

	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

type Agent interface {
	Run(ctx *RunContext) (ask chan<- *eventbus.Event, response <-chan *eventbus.Event, err error)
}

type BehaviorPattern interface {
	// SystemInstruction returns the system instruction for the behavior pattern.
	SystemInstruction(header string) string
	// NextStep is called to process the next step in the agent's behavior pattern.
	NextStep(ctx *StepContext) error
}

type RunContext struct {
	SessionId string
	Context   context.Context
}

type SessionContext struct {
	*RunContext
	InputChan <-chan *eventbus.Event
}

type StepContext struct {
	Context      context.Context
	AgentContext Context

	UserRequest    *UserRequest
	ToolCallResult *llms.ToolCallResult

	SessionId string
	Session   llms.Chat
	StepIndex uint64

	OutputChan  chan<- *eventbus.Event
	ChatOptions []llms.ChatOption
}

func (c *StepContext) StepId() string {
	return fmt.Sprintf("step:%s:%s:%d", c.AgentContext.AgentId(), c.SessionId, c.StepIndex)
}

var _ Agent = &genericAgent{}

func NewGenericAgent(
	agentContext Context, behavior BehaviorPattern,
	llmProvider llms.ChatProvider, model *llms.Model, chatOptions []llms.ChatOption) (Agent, error) {
	return &genericAgent{
		agentContext: agentContext,
		behavior:     behavior,

		llmProvider: llmProvider,
		model:       model,
		chatOptions: chatOptions,

		stepCounter: &atomic.Uint64{},
	}, nil
}

type genericAgent struct {
	agentContext Context
	behavior     BehaviorPattern

	llmProvider llms.ChatProvider
	model       *llms.Model
	chatOptions []llms.ChatOption

	stepCounter *atomic.Uint64
}

func (a *genericAgent) Run(ctx *RunContext) (input chan<- *eventbus.Event, output <-chan *eventbus.Event, err error) {
	session, err := a.llmProvider.NewChat(a.agentContext.SystemPrompt(), a.model)
	if err != nil {
		return nil, nil, err
	}

	inputChan := make(chan *eventbus.Event, 10)
	outputChan := make(chan *eventbus.Event, 10)

	go a.startLoop(
		&SessionContext{
			RunContext: ctx,
			InputChan:  inputChan,
		},
		session, outputChan)
	return inputChan, outputChan, nil
}

func (a *genericAgent) startLoop(ctx *SessionContext, session llms.Chat, output chan<- *eventbus.Event) {
	for {
		select {
		case <-ctx.Context.Done():
			journal.Info("agent", a.agentContext.AgentId(),
				"Chat cancelled during response processing", "err", ctx.Context.Err())
			output <- NewAgentResponseEndEvent("", &AgentResponseEnd{
				Abort:        true,
				Error:        errors.Errorf(ErrorCodeChatSessionAbort, "canceled: %s", ctx.Context.Err()),
				FinishReason: llms.FinishReasonCanceled,
			})
			return
		case inputEvent := <-ctx.InputChan:
			journal.Info("agent", a.agentContext.AgentId(), "receive an input event", "input", inputEvent)
			if err := a.nextStep(ctx, session, inputEvent, output); err != nil {
				journal.Info("agent", a.agentContext.AgentId(),
					"Failed to process next step", "err", err.Error())
				output <- NewAgentResponseEndEvent("", &AgentResponseEnd{
					Error:        err,
					FinishReason: llms.FinishReasonError,
				})
				continue
			}
		}
	}
}

func (a *genericAgent) nextStep(ctx *SessionContext,
	session llms.Chat, inputEvent *eventbus.Event, output chan<- *eventbus.Event) error {
	if inputEvent == nil {
		return nil
	}

	stepContext := &StepContext{
		Context:      ctx.Context,
		AgentContext: a.agentContext,

		SessionId: ctx.SessionId,
		Session:   session,
		StepIndex: a.stepCounter.Add(1),

		OutputChan:  output,
		ChatOptions: a.chatOptions,
	}

	switch inputEvent.Topic {
	case EventTypeExternalActionResult:
		if result, ok := inputEvent.Data.(*ExternalActionResult); ok {
			stepContext.ToolCallResult = result.ToolCallResult
			if len(result.Message) > 0 {
				stepContext.UserRequest = &UserRequest{
					Message: result.Message,
				}
			}
		}
	case EventTypeUserRequest:
		stepContext.UserRequest = inputEvent.Data.(*UserRequest)
	default:
		return errors.Errorf(ErrorCodeInvalidInputEvent, "invalid input event type: %s", inputEvent.Topic)
	}

	return a.behavior.NextStep(stepContext)
}
