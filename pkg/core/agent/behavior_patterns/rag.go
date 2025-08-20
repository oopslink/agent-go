package behavior_patterns

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

//go:embed prompts/rag.md
var _ragPrompt string

var _ agent.BehaviorPattern = &ragPattern{}

func NewRAGPattern() (agent.BehaviorPattern, error) {
	return &ragPattern{
		knowledgeBases: []knowledge.KnowledgeBase{},
	}, nil
}

func NewRAGPatternWithKnowledgeBases(knowledgeBases []knowledge.KnowledgeBase) (agent.BehaviorPattern, error) {
	return &ragPattern{
		knowledgeBases: knowledgeBases,
	}, nil
}

type ragPattern struct {
	knowledgeBases []knowledge.KnowledgeBase
}

func (s *ragPattern) SystemInstruction(header string) string {
	return fmt.Sprintf("%s\n\n%s", header, _ragPrompt)
}

func (s *ragPattern) NextStep(ctx *agent.StepContext) error {
	return runNextStep(ctx, s.nextStep)
}

func (s *ragPattern) nextStep(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error) {
	stepId := ctx.StepId()
	if err = s.rag(ctx, stepId); err != nil {
		return nil, err
	}

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

func (s *ragPattern) rag(ctx *agent.StepContext, stepId string) error {
	// Retrieve relevant knowledge based on user request
	knowledgeItems, err := s.retrieveKnowledge(ctx, stepId)
	if err != nil {
		journal.Warning("step", stepId,
			fmt.Sprintf("failed to retrieve knowledge: %v", err))
		return err
	}

	if len(knowledgeItems) == 0 {
		return nil
	}

	knowledgeMessage := llms.NewUserMessage(s.makeKnowledgeText(knowledgeItems))
	return ctx.AgentContext.UpdateMemory(ctx.Context, knowledgeMessage)
}

func (s *ragPattern) retrieveKnowledge(ctx *agent.StepContext, stepId string) ([]knowledge.KnowledgeItem, error) {
	if ctx.UserRequest == nil || ctx.UserRequest.Message == "" {
		return []knowledge.KnowledgeItem{}, nil
	}

	if len(s.knowledgeBases) == 0 {
		journal.Info("step", stepId, "no knowledge bases configured")
		return []knowledge.KnowledgeItem{}, nil
	}

	var allItems []knowledge.KnowledgeItem
	for i, kb := range s.knowledgeBases {
		items, err := kb.Search(ctx.Context, ctx.UserRequest.Message,
			knowledge.WithMaxResults(3),
			knowledge.WithScoreThreshold(0.7))
		if err != nil {
			journal.Warning("step", stepId,
				fmt.Sprintf("failed to search knowledge base %d: %v", i, err))
			continue
		}
		allItems = append(allItems, items...)
	}

	journal.Info("step", stepId, fmt.Sprintf("retrieved %d knowledge items from %d knowledge bases", len(allItems), len(s.knowledgeBases)))
	return allItems, nil
}

func (s *ragPattern) makeKnowledgeText(knowledgeItems []knowledge.KnowledgeItem) string {
	var prompt strings.Builder
	if len(knowledgeItems) > 0 {
		prompt.WriteString("Retrieved Knowledge:\n")
		for i, item := range knowledgeItems {
			doc := item.ToDocument()
			prompt.WriteString(fmt.Sprintf("\n--- Knowledge Item %d ---\n", i+1))
			prompt.WriteString(fmt.Sprintf("ID: %s\n", doc.Id))
			prompt.WriteString(fmt.Sprintf("Content: %s\n", doc.Content))
			if len(doc.Metadata) > 0 {
				prompt.WriteString("Metadata: ")
				for k, v := range doc.Metadata {
					prompt.WriteString(fmt.Sprintf("%s=%v, ", k, v))
				}
				prompt.WriteString("\n")
			}
		}
		prompt.WriteString("\n--- End of Retrieved Knowledge ---\n\n")
	} else {
		prompt.WriteString("No relevant knowledge was retrieved from the knowledge base.\n\n")
	}

	return prompt.String()
}
