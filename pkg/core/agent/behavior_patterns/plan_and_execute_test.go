package behavior_patterns

import (
	"encoding/json"
	"testing"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlanExecutePattern(t *testing.T) {
	pattern, err := NewPlanExecutePattern(nil)
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)
}

func TestPlanAndExecutePatternSystemInstruction(t *testing.T) {
	pattern, err := NewPlanExecutePattern(nil)
	require.NoError(t, err)

	header := "Test header"
	instruction := pattern.SystemInstruction(header)

	assert.Contains(t, instruction, header)
	assert.Contains(t, instruction, _planAndExecutePrompt)
}

func TestTaskMarshaling(t *testing.T) {
	task := &Task{
		ID:          "task-1",
		Description: "Test task description",
		State:       TaskStatePending,
		Result:      "Task completed successfully",
	}

	marshaled := task.Marshal()
	assert.NotEmpty(t, marshaled)

	var unmarshaled Task
	err := json.Unmarshal([]byte(marshaled), &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, task.ID, unmarshaled.ID)
	assert.Equal(t, task.Description, unmarshaled.Description)
	assert.Equal(t, task.State, unmarshaled.State)
	assert.Equal(t, task.Result, unmarshaled.Result)
}

func TestPlanMarshaling(t *testing.T) {
	plan := &Plan{
		State: PlanStatePending,
		Tasks: []*Task{
			{
				ID:          "task-1",
				Description: "First task",
				State:       TaskStatePending,
			},
		},
	}

	marshaled := plan.Marshal()
	assert.NotEmpty(t, marshaled)

	var unmarshaled Plan
	err := json.Unmarshal([]byte(marshaled), &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, plan.State, unmarshaled.State)
	assert.Len(t, unmarshaled.Tasks, 1)
}

func TestPlanUpdate(t *testing.T) {
	plan := &Plan{
		State: PlanStatePending,
		Tasks: []*Task{
			{
				ID:          "task-1",
				Description: "First task",
				State:       TaskStatePending,
			},
		},
	}

	updatedTask := &Task{
		ID:          "task-1",
		Description: "First task",
		State:       TaskStateRunning,
		Result:      "In progress",
	}

	plan.Update(PlanStateExecuting, updatedTask)

	assert.Equal(t, PlanState("Executing"), plan.State)
	assert.Len(t, plan.Tasks, 1)
	assert.Equal(t, TaskState("Running"), plan.Tasks[0].State)
	assert.Equal(t, "In progress", plan.Tasks[0].Result)
}

func TestPlanAndExecuteAgentResponseMarshaling(t *testing.T) {
	response := &PlanAndExecuteAgentResponse{
		PlanResult: &Plan{
			State: PlanStatePending,
			Tasks: []*Task{},
		},
		ExecuteState: PlanStatePending,
		Reason:       "Planning phase",
	}

	jsonBytes, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	var unmarshaled PlanAndExecuteAgentResponse
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	assert.NoError(t, err)

	assert.NotNil(t, unmarshaled.PlanResult)
	assert.Equal(t, response.ExecuteState, unmarshaled.ExecuteState)
	assert.Equal(t, response.Reason, unmarshaled.Reason)
}

func TestPlanAndExecuteProcessorParseAgentResponse(t *testing.T) {
	processor := &planAndExecuteProcessor{}

	// Test valid JSON response
	input := `{
		"planResult": {
			"state": "Pending",
			"tasks": []
		},
		"executeState": "Pending",
		"reason": "Planning phase"
	}`

	result, err := processor.parseAgentResponse(input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, PlanState("Pending"), result.ExecuteState)
	assert.Equal(t, "Planning phase", result.Reason)

	// Test invalid JSON
	result, err = processor.parseAgentResponse("Invalid JSON")
	assert.Nil(t, result)
}

func TestTaskStateConstants(t *testing.T) {
	assert.Equal(t, "Pending", string(TaskStatePending))
	assert.Equal(t, "Running", string(TaskStateRunning))
	assert.Equal(t, "Succeed", string(TaskStateSucceed))
	assert.Equal(t, "Failed", string(TaskStateFailed))
}

func TestPlanStateConstants(t *testing.T) {
	assert.Equal(t, "Pending", string(PlanStatePending))
	assert.Equal(t, "Executing", string(PlanStateExecuting))
	assert.Equal(t, "Succeed", string(PlanStateSucceed))
	assert.Equal(t, "Failed", string(PlanStateFailed))
}

func TestPlanAndExecuteConfigDefaults(t *testing.T) {
	pattern, err := NewPlanExecutePattern(nil)
	require.NoError(t, err)

	planPattern, ok := pattern.(*planAndExecutePattern)
	assert.True(t, ok)
	assert.NotNil(t, planPattern.config)
	assert.True(t, planPattern.config.RequirePlanConfirmation)
	assert.True(t, planPattern.config.RequireStepConfirmation)
}
