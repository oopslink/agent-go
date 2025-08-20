package behavior_patterns

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

//go:embed prompts/plan_and_execute.md
var _planAndExecutePrompt string

const (
	StateKeyPlan = "plan"
)

// PlanAndExecuteConfig defines the configuration for plan and execute behavior
type PlanAndExecuteConfig struct {
	RequirePlanConfirmation bool // whether plan needs user confirmation
	RequireStepConfirmation bool // whether each step needs user confirmation
}

type TaskState string

const (
	TaskStatePending = "Pending"
	TaskStateRunning = "Running"
	TaskStateSucceed = "Succeed"
	TaskStateFailed  = "Failed"
)

type PlanState string

const (
	PlanStatePending   = "Pending"
	PlanStateExecuting = "Executing"
	PlanStateSucceed   = "Succeed"
	PlanStateFailed    = "Failed"
)

type Task struct {
	ID          string    `json:"id,omitempty"`
	Description string    `json:"description"`
	State       TaskState `json:"state"`

	Result string `json:"result,omitempty"`
}

func (t *Task) Marshal() string {
	s, _ := json.MarshalIndent(t, "", "  ")
	return string(s)
}

type Plan struct {
	State PlanState `json:"state"`
	Tasks []*Task   `json:"tasks,omitempty"`
}

func (p *Plan) Marshal() string {
	s, _ := json.MarshalIndent(p, "", "  ")
	return string(s)
}

func (p *Plan) Update(state PlanState, t *Task) {
	p.State = state
	var tasks []*Task
	for idx := range p.Tasks {
		task := p.Tasks[idx]
		if task.ID != t.ID {
			tasks = append(tasks, task)
		} else {
			tasks = append(tasks, t)
		}
	}
	p.Tasks = tasks
}

type PlanAndExecuteAgentResponse struct {
	// PlanResult result of planning or replanning, not set when executing tasks
	PlanResult *Plan `json:"planResult,omitempty"`
	// CurrentTaskStatus execution status of the current step
	CurrentTaskStatus *Task `json:"currentTaskStatus,omitempty"`
	// ExecuteState set this field to true when completed or cannot continue
	ExecuteState PlanState `json:"executeState,omitempty"`
	// Reason reason for the current response, such as executing tasks, etc.
	Reason string `json:"reason,omitempty"`
	// FinalResult when the plan is successfully executed, output the final answer in this field
	FinalResult string `json:"finalResult,omitempty"`
}

type PlanExecuteResult struct {
	*Plan
	FinalResult string `json:"FinalResult,omitempty"`
}

var _ agent.BehaviorPattern = &planAndExecutePattern{}

func NewPlanExecutePattern(config *PlanAndExecuteConfig) (agent.BehaviorPattern, error) {
	if config == nil {
		config = &PlanAndExecuteConfig{
			RequirePlanConfirmation: true,
			RequireStepConfirmation: true,
		}
	}
	return &planAndExecutePattern{
		config: config,
	}, nil
}

type planAndExecutePattern struct {
	config *PlanAndExecuteConfig
}

func (p *planAndExecutePattern) SystemInstruction(header string) string {
	return fmt.Sprintf("%s\n\n%s", header, _planAndExecutePrompt)
}

func (p *planAndExecutePattern) NextStep(ctx *agent.StepContext) error {
	return runNextStep(ctx, p.nextStep)
}

func (p *planAndExecutePattern) nextStep(ctx *agent.StepContext) (endResponse *agent.AgentResponseEnd, err error) {

	processor := &planAndExecuteProcessor{
		ctx:    ctx,
		config: p.config,
	}

	generatedContext, err := contextForCurrentStep(ctx)
	if err != nil {
		return nil, err
	}

	return askLLM(
		ctx, generatedContext,
		useDefaultTextPartHandle,
		handleBeforeEnd(processor.OnBeforeEnd))
}

type planAndExecuteProcessor struct {
	ctx    *agent.StepContext
	config *PlanAndExecuteConfig
}

func (p *planAndExecuteProcessor) OnBeforeEnd(ctx *StepRuntimeContext, fullMessageContent string, toolCalls []*llms.ToolCall) (*agent.AgentResponseEnd, error) {
	if len(toolCalls) > 0 {
		return nil, nil
	}

	result, err := p.parseAgentResponse(fullMessageContent)
	if err != nil {
		return nil, err
	}

	if result == nil {
		_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
			"agent response nothing", "origin text", fullMessageContent)
		return nil, nil
	}

	_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
		"agent response loop", "result", result)

	if result.PlanResult != nil {
		_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
			"plan made", "plan", result.PlanResult)
		if err = p.savePlan(ctx, result.PlanResult); err != nil {
			return nil, err
		}
		if p.config.RequirePlanConfirmation {
			return p.requireConfirmPlan(ctx, result.PlanResult)
		}
	}
	if result.CurrentTaskStatus != nil {
		_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
			"executing task", "task", result.CurrentTaskStatus)
		if err = p.updatePlan(ctx, result.ExecuteState, result.CurrentTaskStatus); err != nil {
			return nil, err
		}
		if result.CurrentTaskStatus.State == TaskStatePending && p.config.RequireStepConfirmation {
			return p.requireConfirmTask(ctx, result.CurrentTaskStatus)
		}
	}
	if result.ExecuteState == PlanStateFailed {
		_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
			"plan execute failed", "reason", result.Reason)
		return p.planAbort(ctx, fmt.Errorf("%s", result.Reason))
	}
	if result.ExecuteState == PlanStateSucceed {
		_ = journal.Info("plan", p.ctx.AgentContext.AgentId(),
			"plan execute succeed", "reason", result.Reason)
		if len(result.FinalResult) > 0 {
			return p.planDone(ctx, result.FinalResult)
		}
		return nil, nil
	}

	return nil, nil
}

func (p *planAndExecuteProcessor) parseAgentResponse(fullMessageContent string) (*PlanAndExecuteAgentResponse, error) {
	jsonText := jsonRegex.FindString(fullMessageContent)
	if jsonText == "" {
		return nil, nil
	}
	result := &PlanAndExecuteAgentResponse{}
	err := json.Unmarshal([]byte(jsonText), result)
	if err != nil {
		return nil, err
	}
	return result, err
}

func (p *planAndExecuteProcessor) savePlan(ctx *StepRuntimeContext, plan *Plan) error {
	return p.ctx.AgentContext.GetState().Put(StateKeyPlan, plan)
}

func (p *planAndExecuteProcessor) updatePlan(ctx *StepRuntimeContext, state PlanState, task *Task) error {
	plan, err := p.loadPlan(ctx)
	if err != nil {
		return err
	}
	plan.Update(state, task)
	return nil
}

func (p *planAndExecuteProcessor) loadPlan(ctx *StepRuntimeContext) (*Plan, error) {
	state := p.ctx.AgentContext.GetState()
	value, err := state.Get(StateKeyPlan)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errors.Errorf(agent.ErrorCodeLoadPlanFailed, "plan not found")
	}
	plan, ok := value.(*Plan)
	if !ok {
		return nil, errors.Errorf(agent.ErrorCodeLoadPlanFailed, "invaid plan")
	}
	return plan, nil
}

func (p *planAndExecuteProcessor) requireConfirmPlan(ctx *StepRuntimeContext, plan *Plan) (*agent.AgentResponseEnd, error) {
	stepId := p.ctx.StepId()
	event := agent.NewExternalActionEvent(
		fmt.Sprintf("plan made, please confirm:\n\n*** plan ***\n%s", plan.Marshal()))
	sendEvent(stepId, "agent made the plan",
		p.ctx.OutputChan, event)
	return &agent.AgentResponseEnd{
		TraceId:      stepId,
		FinishReason: llms.FinishReasonNormalEnd,
	}, nil
}

func (p *planAndExecuteProcessor) requireConfirmTask(ctx *StepRuntimeContext, taskStatus *Task) (*agent.AgentResponseEnd, error) {
	stepId := p.ctx.StepId()
	event := agent.NewExternalActionEvent(
		fmt.Sprintf("please confirm:\n\n*** task ***\n%s", taskStatus.Marshal()))
	sendEvent(stepId, "agent is going to execute the task",
		p.ctx.OutputChan, event)
	return &agent.AgentResponseEnd{
		TraceId:      stepId,
		FinishReason: llms.FinishReasonNormalEnd,
	}, nil
}

func (p *planAndExecuteProcessor) planAbort(ctx *StepRuntimeContext, err error) (*agent.AgentResponseEnd, error) {
	stepId := p.ctx.StepId()

	finalText := fmt.Sprintf(`# plan failed

reason:

%s
`, utils.WrapString(err.Error(), "```"))

	planStatus := ""
	if plan, _ := p.loadPlan(ctx); plan != nil {
		planStatus = fmt.Sprintf(`# plan execute status

%s
`, utils.WrapString(plan.Marshal(), "```"))
	}
	p.sendMessage(ctx, finalText+"\n"+planStatus)

	return &agent.AgentResponseEnd{
		TraceId:      stepId,
		Error:        err,
		Abort:        true,
		FinishReason: llms.FinishReasonError,
	}, nil
}

func (p *planAndExecuteProcessor) planDone(ctx *StepRuntimeContext, finalResult string) (*agent.AgentResponseEnd, error) {
	stepId := p.ctx.StepId()
	agentContext := p.ctx.AgentContext
	model := agentContext.GetModel()

	if assistantMessage := llms.NewAssistantMessage(ctx.MessageId, model.ModelId, finalResult); assistantMessage != nil {
		if err := agentContext.UpdateMemory(p.ctx.Context, assistantMessage); err != nil {
			journal.Warning("step", stepId,
				fmt.Sprintf("failed to add assistant message to memory: %v", err))
		}
	}

	finalText := fmt.Sprintf(`# final result
%s
`, utils.WrapString(finalResult, "```"))

	planStatus := ""
	if plan, _ := p.loadPlan(ctx); plan != nil {
		planStatus = fmt.Sprintf(`# plan execute status

%s
`, utils.WrapString(plan.Marshal(), "```"))
	}
	p.sendMessage(ctx, finalText+"\n"+planStatus)

	return &agent.AgentResponseEnd{
		TraceId:      stepId,
		Abort:        true,
		FinishReason: llms.FinishReasonNormalEnd,
	}, nil
}

func (p *planAndExecuteProcessor) sendMessage(ctx *StepRuntimeContext, text string) {
	stepId := p.ctx.StepId()
	agentContext := p.ctx.AgentContext
	model := agentContext.GetModel()

	if message := llms.NewAssistantMessage(ctx.MessageId, model.ModelId, text); message != nil {
		sendEvent(stepId, "agent get final result",
			p.ctx.OutputChan, agent.NewAgentMessageEvent(stepId, message))
	}
}
