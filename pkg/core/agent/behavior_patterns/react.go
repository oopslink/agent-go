package behavior_patterns

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

//go:embed prompts/react.md
var _reactPrompt string

var _ agent.BehaviorPattern = &reactPattern{}

func NewReActPattern(maxIterations int) (agent.BehaviorPattern, error) {
	return &reactPattern{
		maxIterations: maxIterations,
	}, nil
}

type reactPattern struct {
	maxIterations int
	iterations    int
}

func (s *reactPattern) SystemInstruction(header string) string {
	return fmt.Sprintf("%s\n\n%s", header, _reactPrompt)
}

func (s *reactPattern) NextStep(ctx *agent.StepContext) error {
	s.iterations = 0
	return runNextStep(ctx, s.nextStep)
}

func (s *reactPattern) nextStep(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error) {
	generatedContext, err := contextForCurrentStep(ctx)
	if err != nil {
		return nil, err
	}

	reactProcessor := newReActStateProcessor(ctx)

	end, err := askLLM(
		ctx, generatedContext,
		useTextPartHandle(reactProcessor.UpdateReAct),
		handleBeforeEnd(reactProcessor.EndIfGotFinalAnswer))
	if err != nil {
		return nil, err
	}

	if end != nil {
		return end, nil
	}

	s.iterations++
	if s.iterations > s.maxIterations {
		return &agent.AgentResponseEnd{
			FinishReason: llms.FinishReasonCanceled,
			Error:        errors.Errorf(agent.ErrorCodeChatSessionFailed, "max iterations reached, max=%d", s.maxIterations),
		}, nil
	}

	return nil, nil
}

// ReActResponse represents the structured JSON response from the agent
type ReActResponse struct {
	Thought     string           `json:"thought,omitempty"`
	Action      string           `json:"action,omitempty"`
	ToolCalls   []*llms.ToolCall `json:"tool_calls,omitempty"`
	Observation string           `json:"observation,omitempty"`
	Answer      string           `json:"answer,omitempty"`
	Continue    bool             `json:"continue"`
}

func (r *ReActResponse) Formatted() string {
	raw, _ := json.MarshalIndent(r, "", "  ")
	return string(raw)
}

func newReActStateProcessor(ctx *agent.StepContext) *reactStateProcessor {
	return &reactStateProcessor{
		finalAnswer: &strings.Builder{},
		outputChan:  ctx.OutputChan,
		jsonBuffer:  &strings.Builder{},
	}
}

type reactStateProcessor struct {
	finalAnswer *strings.Builder
	outputChan  chan<- *eventbus.Event
	jsonBuffer  *strings.Builder
}

func (s *reactStateProcessor) UpdateReAct(ctx *StepRuntimeContext, part *llms.TextPart) ([]*llms.ToolCall, error) {
	var toolCalls []*llms.ToolCall

	finalAnswer := ""
	responses := s.parseReActResponse(part.Text)
	for idx := range responses {
		response := responses[idx]
		if !response.Continue || response.Answer != "" {
			if response.Answer != "" {
				finalAnswer = response.Answer
			} else {
				finalAnswer = "no answer"
			}
		}
		if len(response.ToolCalls) > 0 {
			toolCalls = append(toolCalls, response.ToolCalls...)
		}

		reactText := response.Formatted()
		if len(reactText) > 0 {
			if message := llms.NewAssistantMessage(ctx.MessageId, ctx.ModelId,
				fmt.Sprintf("# loop index: %s\n%s\n", ctx.StepId, reactText)); message != nil {
				sendEvent(ctx.StepId, "agent react reasoning",
					s.outputChan, agent.NewAgentMessageEvent(ctx.StepId, message))
			}
		}
	}

	if len(toolCalls) > 0 {
		return toolCalls, nil
	}

	if len(finalAnswer) > 0 {
		s.finalAnswer.WriteString(finalAnswer)
	}

	return nil, nil
}

func (s *reactStateProcessor) EndIfGotFinalAnswer(ctx *StepRuntimeContext, fullMessageContent string, toolCalls []*llms.ToolCall) (*agent.AgentResponseEnd, error) {
	if s.finalAnswer.Len() == 0 {
		return nil, nil
	}
	finalText := s.finalAnswer.String()
	if message := llms.NewAssistantMessage(ctx.MessageId, ctx.ModelId, finalText); message != nil {
		sendEvent(ctx.StepId, "agent final answer",
			s.outputChan, agent.NewAgentMessageEvent(ctx.StepId, message))
	}
	return &agent.AgentResponseEnd{
		FinishReason: llms.FinishReasonNormalEnd,
	}, nil
}

func (s *reactStateProcessor) parseReActResponse(text string) []*ReActResponse {
	// Accumulate text in JSON buffer
	s.jsonBuffer.WriteString(text)

	// Try to extract and parse JSON from the accumulated text
	jsonText := s.jsonBuffer.String()

	// Look for JSON blocks in the text
	jsonBlocks := s.extractJSONBlocks(jsonText)

	var responses []*ReActResponse
	for _, block := range jsonBlocks {
		var response ReActResponse
		if err := json.Unmarshal([]byte(block), &response); err != nil {
			journal.Warning("react",
				"failed to parse JSON response", "error", err, "json", block)
			continue
		}
		responses = append(responses, &response)
	}
	return responses
}

func (s *reactStateProcessor) extractJSONBlocks(text string) []string {
	var blocks []string
	var currentBlock strings.Builder
	var braceCount int
	var inJSON bool

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this line starts a JSON block
		if strings.Contains(line, "{") && !inJSON {
			inJSON = true
			braceCount = 0
			currentBlock.Reset()
		}

		if inJSON {
			currentBlock.WriteString(line + "\n")

			// Count braces to track JSON structure
			for _, char := range line {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
				}
			}

			// If braces are balanced, we have a complete JSON block
			if braceCount == 0 {
				blocks = append(blocks, currentBlock.String())
				inJSON = false
			}
		}
	}

	return blocks
}
