package memory

import "context"

// NewSimpleMemory creates a simple memory instance using in-memory storage
func NewSimpleMemory() Memory {
	return NewSimpleMemoryWithStore(NewInMemoryStore())
}

// NewSimpleMemoryWithStore creates a simple memory instance using specified storage
func NewSimpleMemoryWithStore(store MemoryStore) Memory {
	return &SimpleMemory{
		store: store,
	}
}

var _ Memory = &SimpleMemory{}

type SimpleMemory struct {
	store MemoryStore
}

func (m *SimpleMemory) Reset() error {
	return m.store.Clear(context.Background())
}

func (m *SimpleMemory) Add(ctx context.Context, memory MemoryItem) error {
	return m.store.Store(ctx, memory)
}

func (m *SimpleMemory) Retrieve(ctx context.Context, options ...MemoryRetrieveOption) ([]MemoryItem, error) {
	retrieveOptions := NewMemoryRetrieveOptions()
	for _, option := range options {
		option(retrieveOptions)
	}

	items, err := m.store.Load(ctx)
	if err != nil {
		return nil, err
	}

	limit := retrieveOptions.Limit
	if limit < 0 {
		return items, nil
	}

	if len(items) <= limit {
		return items, nil
	}

	if limit == 0 {
		return nil, nil
	}

	return items[:limit], nil
}
