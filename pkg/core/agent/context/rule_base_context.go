package context

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"sync"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/core/memory"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

var _ agent.Context = &ruleBaseContext{}

func NewRuleBaseContext(
	agentId, systemPrompt string,
	llmProvider llms.ChatProvider, model *llms.Model,
	behavior agent.BehaviorPattern,
	memory memory.Memory, state agent.AgentState,
	knowledgeBases []knowledge.KnowledgeBase,
	toolRegistry *tools.ToolCollection,
	rules ContextRules,
) agent.Context {

	if rules.UseKnowledgeAsTool && len(knowledgeBases) > 0 {
		knowledgeTool := knowledge.WrapperAsRetrieverTool(knowledgeBases...)
		if toolRegistry == nil {
			toolRegistry = tools.OfTools(knowledgeTool)
		} else {
			toolRegistry.AddTools(knowledgeTool)
		}
	}

	return &ruleBaseContext{
		agentId:      agentId,
		systemPrompt: systemPrompt,

		llmProvider: llmProvider,
		model:       model,

		behavior: behavior,

		memory:         memory,
		state:          state,
		knowledgeBases: knowledgeBases,
		toolRegistry:   toolRegistry,

		rules: &rules,
	}
}

type ContextRulesUpdateParams struct {
	EnableAutoTools  []string
	DisableAutoTools []string
}

type ContextRuleUpdater interface {
	UpdateContextRules(params *ContextRulesUpdateParams)
}

type ContextRules struct {
	MemoryRetrieveOptions   []memory.MemoryRetrieveOption
	UseKnowledgeAsTool      bool
	AutoAddToolInstructions bool
	AutoTools               []string
}

var _ ContextRuleUpdater = &ruleBaseContext{}

type ruleBaseContext struct {
	agentId      string
	systemPrompt string

	llmProvider llms.ChatProvider
	model       *llms.Model

	behavior agent.BehaviorPattern

	memory         memory.Memory
	state          agent.AgentState
	knowledgeBases []knowledge.KnowledgeBase
	toolRegistry   *tools.ToolCollection

	rulesLock sync.RWMutex
	rules     *ContextRules
}

func (r *ruleBaseContext) GetState() agent.AgentState {
	return r.state
}

func (r *ruleBaseContext) GetModel() *llms.Model {
	return r.model
}

func (r *ruleBaseContext) UpdateContextRules(params *ContextRulesUpdateParams) {
	r.rulesLock.Lock()
	defer r.rulesLock.Unlock()

	var autoTools []string
	if len(params.DisableAutoTools) > 0 {
		for _, t := range r.rules.AutoTools {
			if !utils.ContainString(params.DisableAutoTools, t) {
				autoTools = append(autoTools, t)
			}
		}
	} else {
		autoTools = r.rules.AutoTools
	}
	if len(params.EnableAutoTools) > 0 {
		for _, t := range params.EnableAutoTools {
			if !utils.ContainString(autoTools, t) {
				autoTools = append(autoTools, t)
			}
		}
	}
	r.rules.AutoTools = autoTools
}

func (r *ruleBaseContext) getRules() ContextRules {
	r.rulesLock.RLock()
	defer r.rulesLock.RUnlock()

	rules := *r.rules
	return rules
}

func (r *ruleBaseContext) ValidateToolCall(toolCall *llms.ToolCall) error {
	if toolCall == nil {
		return errors.Errorf(agent.ErrorCodeInvalidToolCall, "empty tool call")
	}
	if !r.toolRegistry.ContainsTool(toolCall.Name) {
		return errors.Errorf(agent.ErrorCodeInvalidToolCall,
			"invalid tool, tool [%s] is not exists", toolCall.Name)
	}
	return nil
}

func (r *ruleBaseContext) CanAutoCall(toolCall *llms.ToolCall) bool {
	rules := r.getRules()
	return utils.ContainString(rules.AutoTools, toolCall.Name)
}

func (r *ruleBaseContext) CallTool(ctx context.Context, call *llms.ToolCall) (result *llms.ToolCallResult, err error) {
	defer func() {
		if err != nil {
			journal.Error("tool", call.Name,
				"failed to call the tool", "args", call.Arguments, "err", err)
		} else {
			journal.Info("tool", call.Name,
				"tool called", "args", call.Arguments, "result", result)
		}
	}()

	return r.toolRegistry.Call(ctx, call)
}

func (r *ruleBaseContext) AgentId() string {
	return r.agentId
}

func (r *ruleBaseContext) SystemPrompt() string {
	return r.behavior.SystemInstruction(r.systemPrompt)
}

func (r *ruleBaseContext) UpdateMemory(ctx context.Context, messages ...*llms.Message) error {

	if r.memory == nil {
		return nil
	}
	for idx := range messages {
		message := messages[idx]
		memoryItem := memory.NewChatMessageMemoryItem(message)

		//d, _ := json.Marshal(memoryItem.GetContent())
		//fmt.Println(fmt.Sprintf("   :memory: <- %s", string(d)))
		journal.Info("context/memory", r.agentId, "add messages to memory",
			"memory id", memoryItem.GetId(), "content", memoryItem.GetContent())

		if err := r.memory.Add(ctx, memoryItem); err != nil {
			journal.Warning("context/memory", r.agentId,
				fmt.Sprintf("failed to add messages to memory: %v", err))
			return err
		}
	}
	return nil
}

func (r *ruleBaseContext) Generate(ctx context.Context, params *agent.GenerateContextParams) (*agent.GeneratedContext, error) {
	if params == nil {
		return nil, errors.Errorf(agent.ErrorCodeGenerateContextFailed,
			"generate context params cannot be nil")
	}

	// 1. Select tools based on rules
	toolDescriptors := r.selectTools(params)

	// 2. Prepare chat options
	var chatOptions []llms.ChatOption
	if len(params.ChatOptions) > 0 {
		chatOptions = append(chatOptions, params.ChatOptions...)
	}
	if len(toolDescriptors) > 0 {
		chatOptions = append(chatOptions, llms.WithTools(toolDescriptors...))
	}

	// 3. Make messages context
	rules := r.getRules()
	var messageContext []*llms.Message
	if rules.AutoAddToolInstructions {
		if instructionMessage := r.generateToolInstructions(toolDescriptors); instructionMessage != nil {
			messageContext = append(messageContext, instructionMessage)
		}
	}

	history, err := r.retrieveMemory(ctx)
	if err != nil {
		return nil, errors.Wrap(agent.ErrorCodeGenerateContextFailed, err)
	}
	if len(history) > 0 {
		messageContext = append(messageContext, history...)
	}

	messages := params.ToMessages()
	if len(messages) > 0 {
		messageContext = append(messageContext, messages...)
		_ = r.UpdateMemory(ctx, messages...)
	}

	return &agent.GeneratedContext{
		Messages: messageContext,
		Options:  chatOptions,
	}, nil
}

func (r *ruleBaseContext) retrieveMemory(ctx context.Context) ([]*llms.Message, error) {
	var history []*llms.Message
	if r.memory != nil {
		rules := r.getRules()
		memoryItems, err := r.memory.Retrieve(ctx, rules.MemoryRetrieveOptions...)
		if err != nil {
			journal.Warning("context", r.agentId,
				fmt.Sprintf("failed to retrieve messages from memory: %v", err))
			return nil, err
		}

		history = append(history, memory.AsMessages(memoryItems)...)
	}
	return history, nil
}

func (r *ruleBaseContext) generateToolInstructions(toolDescriptors []*llms.ToolDescriptor) *llms.Message {
	if len(toolDescriptors) == 0 {
		return nil
	}

	instructions := tools.MakeToolInstructions(toolDescriptors...)
	if instructions == "" {
		return nil
	}

	return llms.NewSystemMessage(instructions)
}

func (r *ruleBaseContext) selectTools(params *agent.GenerateContextParams) []*llms.ToolDescriptor {
	// current simple return all tools
	if r.toolRegistry == nil {
		return nil
	}
	return r.toolRegistry.Descriptors()
}
