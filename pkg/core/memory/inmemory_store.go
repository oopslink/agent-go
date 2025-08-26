package memory

import (
	"context"
	"sync"
)

// NewInMemoryStore 创建一个内存存储实例
func NewInMemoryStore() MemoryStore {
	return &InMemoryStore{
		items: make([]MemoryItem, 0),
		mutex: &sync.RWMutex{},
	}
}

var _ MemoryStore = &InMemoryStore{}

// InMemoryStore 内存存储实现
type InMemoryStore struct {
	items []MemoryItem
	mutex *sync.RWMutex
}

// Store 添加一个 MemoryItem 到存储
func (s *InMemoryStore) Store(ctx context.Context, item MemoryItem) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = append(s.items, item)
	return nil
}

// Load 检索所有 MemoryItem
func (s *InMemoryStore) Load(ctx context.Context) ([]MemoryItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回副本以避免并发修改
	result := make([]MemoryItem, len(s.items))
	copy(result, s.items)
	return result, nil
}

// Clear 清空存储
func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items = make([]MemoryItem, 0)
	return nil
}

// Close 关闭存储
func (s *InMemoryStore) Close() error {
	// 内存存储无需关闭操作
	return nil
}
