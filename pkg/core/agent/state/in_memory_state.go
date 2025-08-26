package state

import (
	"github.com/oopslink/agent-go/pkg/core/agent"
)

var _ agent.AgentState = &SimpleState{}

// NewInMemoryState 创建内存存储的 SimpleState
func NewInMemoryState() agent.AgentState {
	return &SimpleState{
		store: NewInMemoryStore(),
	}
}

// NewFileState 创建文件存储的 SimpleState
func NewFileState(dataDir string) (agent.AgentState, error) {
	store, err := NewFileStore(dataDir)
	if err != nil {
		return nil, err
	}
	return &SimpleState{
		store: store,
	}, nil
}

// SimpleState 简单状态实现，支持多种存储后端
type SimpleState struct {
	store StateStore
}

func (s *SimpleState) Get(key string) (any, error) {
	return s.store.Get(key)
}

func (s *SimpleState) Put(key string, value any) error {
	return s.store.Set(key, value)
}

// Delete 删除指定key的值
func (s *SimpleState) Delete(key string) error {
	return s.store.Delete(key)
}

// Clear 清空所有状态
func (s *SimpleState) Clear() error {
	return s.store.Clear()
}
