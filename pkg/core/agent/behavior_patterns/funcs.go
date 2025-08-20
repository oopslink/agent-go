package behavior_patterns

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"time"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

type nextStepFunc func(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error)

func runNextStep(ctx *agent.StepContext, nextStep nextStepFunc) error {
	stepId := ctx.StepId()

	agentResponseStart(stepId, ctx.OutputChan)

	for {
		end, err := nextStep(ctx)
		if err != nil {
			return err
		}
		if end != nil {
			agentResponseEnd(stepId, ctx.OutputChan, end)
			break
		}
	}

	return nil
}

func contextForCurrentStep(ctx *agent.StepContext) (*agent.GeneratedContext, error) {
	params := &agent.GenerateContextParams{
		UserRequest:    ctx.UserRequest,
		ToolCallResult: ctx.ToolCallResult,
		ChatOptions:    ctx.ChatOptions,
	}
	generatedContext, err := ctx.AgentContext.Generate(ctx.Context, params)
	if err != nil {
		return nil, err
	}
	return generatedContext, nil
}

func autoCallTool(ctx *agent.StepContext, toolCall *llms.ToolCall) error {
	traceId := "agent/" + ctx.AgentContext.AgentId()
	toolCallResult, err := ctx.AgentContext.CallTool(ctx.Context, toolCall)
	if err != nil {
		journal.Warning("step/auto_call_tool", traceId,
			fmt.Sprintf("failed to call tool %s: %v", toolCall.Name, err))
		return err
	}

	if toolCallResult == nil {
		toolCallResult = &llms.ToolCallResult{
			ToolCallId: toolCall.ToolCallId,
			Name:       toolCall.Name,
		}
	}

	return ctx.AgentContext.UpdateMemory(
		ctx.Context, llms.NewToolCallResultMessage(toolCallResult, time.Now()))
}

type StepRuntimeContext struct {
	StepId    string
	MessageId string
	ModelId   llms.ModelId
}

type FnHandleStepTextPart func(ctx *StepRuntimeContext, part *llms.TextPart) ([]*llms.ToolCall, error)
type FnHandleBeforeEnd func(ctx *StepRuntimeContext, fullMessageContent string, toolCalls []*llms.ToolCall) (*agent.AgentResponseEnd, error)
type AskLLMOption func(*askLLMOptions)

type askLLMOptions struct {
	textPartHandleFn  FnHandleStepTextPart
	handleBeforeEndFn FnHandleBeforeEnd
}

var (
	useDefaultTextPartHandle = func(options *askLLMOptions) {
		options.textPartHandleFn = nil
	}

	noCustomizeEndHandler = func(options *askLLMOptions) {
		options.handleBeforeEndFn = nil
	}

	useTextPartHandle = func(textPartHandleFn FnHandleStepTextPart) AskLLMOption {
		return func(options *askLLMOptions) {
			options.textPartHandleFn = textPartHandleFn
		}
	}

	handleBeforeEnd = func(handleBeforeEndFn FnHandleBeforeEnd) AskLLMOption {
		return func(options *askLLMOptions) {
			options.handleBeforeEndFn = handleBeforeEndFn
		}
	}
)

func askLLM(ctx *agent.StepContext,
	generatedContext *agent.GeneratedContext, opts ...AskLLMOption) (*agent.AgentResponseEnd, error) {
	stepId := ctx.StepId()
	options := &askLLMOptions{}
	for _, opt := range opts {
		opt(options)
	}

	recordCurrentContext(ctx, generatedContext.Messages)

	responseIterator, err := ctx.Session.Send(ctx.Context,
		generatedContext.Messages, generatedContext.Options...)
	if err != nil {
		journal.Warning("step", stepId,
			fmt.Sprintf("failed to send messages to provider: %v", err))
		return nil, err
	}

	var messageId string
	var modelId llms.ModelId
	var toolCalls []*llms.ToolCall
	var fullMessageContent string
	var stepRuntimeContext *StepRuntimeContext

	for response, iterErr := range responseIterator {
		// Check for context cancellation before processing each response
		select {
		case <-ctx.Context.Done():
			return &agent.AgentResponseEnd{
				Abort:        true,
				Error:        errors.Errorf(agent.ErrorCodeChatSessionAbort, "canceled: %s", ctx.Context.Err()),
				FinishReason: llms.FinishReasonCanceled,
			}, nil
		default:
			// Continue with response processing
		}

		if iterErr != nil {
			return &agent.AgentResponseEnd{
				Abort:        true,
				Error:        errors.Errorf(agent.ErrorCodeChatSessionFailed, "iteration error: %v", iterErr),
				FinishReason: llms.FinishReasonError,
			}, nil
		}

		// If the response is empty, continue to the next iteration
		if response == nil {
			journal.Warning("step", stepId, "received empty response, skipping")
			continue
		}

		if len(messageId) == 0 {
			messageId = response.MessageId
			modelId = response.Model
			stepRuntimeContext = &StepRuntimeContext{
				StepId:    stepId,
				ModelId:   modelId,
				MessageId: messageId,
			}
		}

		// process the response
		for _, part := range response.Parts {
			if textPart, ok := part.(*llms.TextPart); ok {
				fullMessageContent = fullMessageContent + textPart.Text
				if options.textPartHandleFn != nil {
					toolCallList, err := options.textPartHandleFn(stepRuntimeContext, textPart)
					if err != nil {
						return nil, err
					}
					if len(toolCallList) > 0 {
						toolCalls = append(toolCalls, toolCallList...)
					}
				} else {
					// Send content as it comes
					if textPart.Text != "" {
						if message := llms.NewAssistantMessage(messageId, modelId, textPart.Text); message != nil {
							sendEvent(stepId, "agent response a text part",
								ctx.OutputChan, agent.NewAgentMessageEvent(stepId, message))
						}
					}
				}
			} else if toolCall, ok := part.(*llms.ToolCall); ok {
				toolCalls = append(toolCalls, toolCall)
			} else {
				journal.Warning("step", stepId,
					fmt.Sprintf("unknown response part type: %T, discard.", part))
			}
		}
	}

	agentContext := ctx.AgentContext

	if assistantMessage := llms.NewAssistantMessage(messageId, modelId, fullMessageContent, toolCalls...); assistantMessage != nil {
		if err = agentContext.UpdateMemory(ctx.Context, assistantMessage); err != nil {
			journal.Warning("step", stepId,
				fmt.Sprintf("failed to add assistant message to memory: %v", err))
		}
	}

	if options.handleBeforeEndFn != nil {
		end, err := options.handleBeforeEndFn(stepRuntimeContext, fullMessageContent, toolCalls)
		if err != nil {
			return nil, err
		}
		if end != nil {
			return end, nil
		}
		if len(toolCalls) == 0 {
			return nil, nil
		}
	}

	toolsToConfirm := 0
	invalidCalls := 0
	autoCalls := 0
	for idx := range toolCalls {
		toolCall := toolCalls[idx]
		if err = agentContext.ValidateToolCall(toolCall); err != nil {
			invalidCalls++
			_ = ctx.AgentContext.UpdateMemory(
				ctx.Context, llms.NewToolCallResultMessage(
					&llms.ToolCallResult{
						ToolCallId: toolCall.ToolCallId,
						Name:       toolCall.Name,
						Result: map[string]any{
							"state": fmt.Sprintf("invalid tool, reason: [%s]", err.Error()),
						},
					}, time.Now()))
		} else {
			if agentContext.CanAutoCall(toolCall) {
				autoCalls++
				_ = autoCallTool(ctx, toolCall)
			} else {
				toolsToConfirm++
				sendEvent(stepId, "received tool call",
					ctx.OutputChan, agent.NewToolCallEvent(toolCall))
			}
		}
	}

	if invalidCalls > 0 || autoCalls > 0 {
		return nil, nil
	} else if toolsToConfirm > 0 {
		return &agent.AgentResponseEnd{
			FinishReason: llms.FinishReasonToolUse,
		}, nil
	} else {
		return &agent.AgentResponseEnd{
			FinishReason: llms.FinishReasonNormalEnd,
		}, nil
	}
}

func recordCurrentContext(ctx *agent.StepContext, messages []*llms.Message) {
	_ = journal.Info("context", ctx.AgentContext.AgentId(),
		fmt.Sprintf("context for %s", ctx.StepId()), "instruction", ctx.AgentContext.SystemPrompt(), "context", messages)
}

func sendEvent(
	source, comment string, output chan<- *eventbus.Event, event *eventbus.Event) {
	_ = journal.Info("send_event", source, comment, "event", event)
	output <- event
}

func agentResponseStart(stepId string, output chan<- *eventbus.Event) {
	sendEvent(stepId, "agent response start",
		output, agent.NewAgentResponseStartEvent(stepId))
}

func agentResponseEnd(
	traceId string, output chan<- *eventbus.Event, end *agent.AgentResponseEnd) {
	sendEvent(traceId, "agent response end",
		output, agent.NewAgentResponseEndEvent(traceId, end))
}
