package memory

import "context"

func NewInMemoryMemory() Memory {
	return &InMemoryMemory{
		items: make([]MemoryItem, 0),
	}
}

var _ Memory = &InMemoryMemory{}

type InMemoryMemory struct {
	items []MemoryItem
}

func (m *InMemoryMemory) Reset() error {
	m.items = make([]MemoryItem, 0)
	return nil
}

func (m *InMemoryMemory) Add(ctx context.Context, memory MemoryItem) error {
	m.items = append(m.items, memory)
	return nil
}

func (m *InMemoryMemory) Retrieve(ctx context.Context, options ...MemoryRetrieveOption) ([]MemoryItem, error) {
	retrieveOptions := NewMemoryRetrieveOptions()
	for _, option := range options {
		option(retrieveOptions)
	}

	limit := retrieveOptions.Limit
	if limit < 0 {
		return m.items, nil
	}

	if len(m.items) <= limit {
		return m.items, nil
	}

	if limit == 0 {
		return nil, nil
	}

	return m.items[:limit], nil
}
