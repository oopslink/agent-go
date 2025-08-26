package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// NewFileStore 创建一个文件存储实例
func NewFileStore(filePath string, codec MemoryItemCodec) MemoryStore {
	return &FileStore{
		filePath: filePath,
		codec:    codec,
		mutex:    &sync.RWMutex{},
	}
}

var _ MemoryStore = &FileStore{}

// FileStore 文件存储实现
type FileStore struct {
	filePath string
	codec    MemoryItemCodec
	mutex    *sync.RWMutex
}

// Store 添加一个 MemoryItem 到存储
func (s *FileStore) Store(ctx context.Context, item MemoryItem) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 先加载现有数据
	items, err := s.loadFromFile()
	if err != nil {
		return fmt.Errorf("failed to load existing data: %w", err)
	}

	// 添加新项目
	items = append(items, item)

	// 保存回文件
	return s.saveToFile(items)
}

// Load 检索所有 MemoryItem
func (s *FileStore) Load(ctx context.Context) ([]MemoryItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.loadFromFile()
}

// Clear 清空存储
func (s *FileStore) Clear(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.saveToFile([]MemoryItem{})
}

// Close 关闭存储
func (s *FileStore) Close() error {
	// 文件存储无需特殊关闭操作
	return nil
}

// loadFromFile 从文件加载数据
func (s *FileStore) loadFromFile() ([]MemoryItem, error) {
	// 检查文件是否存在
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []MemoryItem{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return []MemoryItem{}, nil
	}

	// 解析JSON数组
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// 解码每个项目
	items := make([]MemoryItem, 0, len(rawItems))
	for i, rawItem := range rawItems {
		item, err := s.codec.Decode(rawItem)
		if err != nil {
			return nil, fmt.Errorf("failed to decode item %d: %w", i, err)
		}
		items = append(items, item)
	}

	return items, nil
}

// saveToFile 保存数据到文件
func (s *FileStore) saveToFile(items []MemoryItem) error {
	// 确保目录存在
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 编码每个项目
	rawItems := make([]json.RawMessage, 0, len(items))
	for i, item := range items {
		data, err := s.codec.Encode(item)
		if err != nil {
			return fmt.Errorf("failed to encode item %d: %w", i, err)
		}
		rawItems = append(rawItems, data)
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(rawItems, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
