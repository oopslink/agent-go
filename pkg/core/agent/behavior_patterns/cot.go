package behavior_patterns

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	_ "embed"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/eventbus"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

//go:embed prompts/cot.md
var _cotPrompt string

// COTResponse represents the structured response from Chain of Thought reasoning
type COTResponse struct {
	Thinking    string `json:"thinking"`
	FinalAnswer string `json:"final_answer"`
}

// JSON extraction regex pattern
var jsonRegex = regexp.MustCompile(`(?s)\{.*\}`)

var _ agent.BehaviorPattern = &chainOfThoughtPattern{}

func NewChainOfThoughtPattern() (agent.BehaviorPattern, error) {
	return &chainOfThoughtPattern{}, nil
}

type chainOfThoughtPattern struct {
}

func (s *chainOfThoughtPattern) SystemInstruction(header string) string {
	return fmt.Sprintf("%s\n\n%s", header, _cotPrompt)
}

func (s *chainOfThoughtPattern) NextStep(ctx *agent.StepContext) error {
	return runNextStep(ctx, s.nextStep)
}

func (s *chainOfThoughtPattern) nextStep(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error) {
	generatedContext, err := contextForCurrentStep(ctx)
	if err != nil {
		return nil, err
	}

	thinker := newThinkStateProcessor(ctx.OutputChan)

	end, err := askLLM(
		ctx, generatedContext,
		useTextPartHandle(thinker.UpdateThinking),
		handleBeforeEnd(thinker.EndIfGotFinalAnswer))
	if err != nil {
		return nil, err
	}

	return end, nil
}

func newThinkStateProcessor(outputChan chan<- *eventbus.Event) *thinkStateProcessor {
	return &thinkStateProcessor{
		responseBuffer: &strings.Builder{},
		outputChan:     outputChan,
		cotResponse:    &COTResponse{},
	}
}

type thinkStateProcessor struct {
	responseBuffer *strings.Builder
	cotResponse    *COTResponse
	outputChan     chan<- *eventbus.Event
}

func (s *thinkStateProcessor) UpdateThinking(ctx *StepRuntimeContext, part *llms.TextPart) ([]*llms.ToolCall, error) {
	// Accumulate the response text
	s.responseBuffer.WriteString(part.Text)

	// Try to parse JSON from accumulated response
	response := s.responseBuffer.String()
	if cotResp := s.tryParseStructuredResponse(response); cotResp != nil {
		// Update our COT response with parsed data
		if cotResp.Thinking != "" && cotResp.Thinking != s.cotResponse.Thinking {
			s.cotResponse.Thinking = cotResp.Thinking
			// Send thinking content
			if message := llms.NewAssistantMessage(ctx.MessageId, ctx.ModelId, cotResp.Thinking); message != nil {
				sendEvent(ctx.StepId, "agent thinking step",
					s.outputChan, agent.NewAgentMessageEvent(ctx.StepId, message))
			}
		}

		if cotResp.FinalAnswer != "" {
			s.cotResponse.FinalAnswer = cotResp.FinalAnswer
		}
	}

	return nil, nil
}

func (s *thinkStateProcessor) EndIfGotFinalAnswer(ctx *StepRuntimeContext, fullMessageContent string, toolCalls []*llms.ToolCall) (*agent.AgentResponseEnd, error) {
	// Try to parse the complete response one more time
	response := s.responseBuffer.String()
	if cotResp := s.tryParseStructuredResponse(response); cotResp != nil && cotResp.FinalAnswer != "" {
		// Send final answer
		finalAnswer := fmt.Sprintf("\nFinal Answer: %s", cotResp.FinalAnswer)
		if message := llms.NewAssistantMessage(ctx.MessageId, ctx.ModelId, finalAnswer); message != nil {
			sendEvent(ctx.StepId, "agent final answer",
				s.outputChan, agent.NewAgentMessageEvent(ctx.StepId, message))
		}
		return &agent.AgentResponseEnd{
			FinishReason: llms.FinishReasonNormalEnd,
		}, nil
	}
	return nil, nil
}

// tryParseStructuredResponse attempts to extract and parse JSON from the response
func (s *thinkStateProcessor) tryParseStructuredResponse(response string) *COTResponse {
	// Extract JSON using regex
	matches := jsonRegex.FindString(response)
	if matches == "" {
		return nil
	}

	var cotResp COTResponse
	if err := json.Unmarshal([]byte(matches), &cotResp); err != nil {
		return nil
	}

	return &cotResp
}
