package state

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/core/agent"
)

// StateType state storage type
type StateType string

const (
	StateTypeMemory StateType = "memory"
	StateTypeFile   StateType = "file"
)

// StateConfig state configuration
type StateConfig struct {
	Type     StateType
	FilePath string // Only valid for file type
}

// NewState creates a state storage instance
func NewState(config StateConfig) (agent.AgentState, error) {
	switch config.Type {
	case StateTypeMemory:
		return NewInMemoryState(), nil
	case StateTypeFile:
		if config.FilePath == "" {
			return nil, fmt.Errorf("file path is required for file state")
		}
		return NewFileState(config.FilePath)
	default:
		return nil, fmt.Errorf("unsupported state type: %s", config.Type)
	}
}
