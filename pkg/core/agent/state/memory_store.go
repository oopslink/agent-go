package state

import "sync"

// InMemoryStore in-memory storage implementation
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewInMemoryStore() StateStore {
	return &InMemoryStore{
		data: make(map[string]any),
	}
}

func (s *InMemoryStore) Get(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key], nil
}

func (s *InMemoryStore) Set(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *InMemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *InMemoryStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
	return nil
}
