package state

import (
	"github.com/oopslink/agent-go/pkg/core/agent"
	"sync"
)

var _ agent.AgentState = &InMemoryState{}

func NewInMemoryState() agent.AgentState {
	return &InMemoryState{
		data: make(map[string]any),
	}
}

type InMemoryState struct {
	mu   sync.Mutex
	data map[string]any
}

func (i *InMemoryState) Get(key string) (any, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.data[key], nil
}

func (i *InMemoryState) Put(key string, value any) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.data[key] = value
	return nil
}
