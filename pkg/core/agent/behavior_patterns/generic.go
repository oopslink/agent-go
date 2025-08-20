package behavior_patterns

import (
	"github.com/oopslink/agent-go/pkg/core/agent"
)

var _ agent.BehaviorPattern = &genericPattern{}

func NewGenericPattern() (agent.BehaviorPattern, error) {
	return &genericPattern{}, nil
}

type genericPattern struct {
}

func (s *genericPattern) SystemInstruction(header string) string {
	return header
}

func (s *genericPattern) NextStep(ctx *agent.StepContext) error {
	return runNextStep(ctx, s.nextStep)
}

func (s *genericPattern) nextStep(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error) {
	generatedContext, err := contextForCurrentStep(ctx)
	if err != nil {
		return nil, err
	}

	end, err := askLLM(
		ctx, generatedContext,
		useDefaultTextPartHandle, noCustomizeEndHandler)
	if err != nil {
		return nil, err
	}

	return end, nil
}
