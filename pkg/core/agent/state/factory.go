package state

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/core/agent"
)

// StateType 状态存储类型
type StateType string

const (
	StateTypeMemory StateType = "memory"
	StateTypeFile   StateType = "file"
)

// StateConfig 状态配置
type StateConfig struct {
	Type     StateType
	FilePath string // 仅对文件类型有效
}

// NewState 创建状态存储实例
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
