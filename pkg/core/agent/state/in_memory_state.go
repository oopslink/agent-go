package state

import (
	"github.com/oopslink/agent-go/pkg/core/agent"
)

var _ agent.AgentState = &SimpleState{}

// NewInMemoryState creates a SimpleState with in-memory storage
func NewInMemoryState() agent.AgentState {
	return &SimpleState{
		store: NewInMemoryStore(),
	}
}

// NewFileState creates a SimpleState with file storage
func NewFileState(dataDir string) (agent.AgentState, error) {
	store, err := NewFileStore(dataDir)
	if err != nil {
		return nil, err
	}
	return &SimpleState{
		store: store,
	}, nil
}

// SimpleState simple state implementation, supports multiple storage backends
type SimpleState struct {
	store StateStore
}

func (s *SimpleState) Get(key string) (any, error) {
	return s.store.Get(key)
}

func (s *SimpleState) Put(key string, value any) error {
	return s.store.Set(key, value)
}

// Delete deletes the value for the specified key
func (s *SimpleState) Delete(key string) error {
	return s.store.Delete(key)
}

// Clear clears all state
func (s *SimpleState) Clear() error {
	return s.store.Clear()
}
