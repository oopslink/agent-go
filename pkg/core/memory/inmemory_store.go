package memory

import (
	"context"
	"sync"
)

// NewInMemoryStore creates an in-memory storage instance
func NewInMemoryStore() MemoryStore {
	return &InMemoryStore{
		items: make([]MemoryItem, 0),
		mutex: &sync.RWMutex{},
	}
}

var _ MemoryStore = &InMemoryStore{}

// InMemoryStore in-memory storage implementation
type InMemoryStore struct {
	items []MemoryItem
	mutex *sync.RWMutex
}

// Store adds a MemoryItem to storage
func (s *InMemoryStore) Store(ctx context.Context, item MemoryItem) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = append(s.items, item)
	return nil
}

// Load retrieves all MemoryItem
func (s *InMemoryStore) Load(ctx context.Context) ([]MemoryItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to avoid concurrent modification
	result := make([]MemoryItem, len(s.items))
	copy(result, s.items)
	return result, nil
}

// Clear clears storage
func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = make([]MemoryItem, 0)
	return nil
}

// Close closes storage
func (s *InMemoryStore) Close() error {
	// In-memory storage requires no close operation
	return nil
}
