package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FileStore filesystem storage implementation
type FileStore struct {
	mu      sync.RWMutex
	dataDir string
}

func NewFileStore(dataDir string) (StateStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{
		dataDir: dataDir,
	}, nil
}

func (s *FileStore) getFilePath(key string) string {
	// Replace unsafe filename characters
	safeKey := strings.ReplaceAll(key, "/", "_")
	safeKey = strings.ReplaceAll(safeKey, "\\", "_")
	safeKey = strings.ReplaceAll(safeKey, ":", "_")
	return filepath.Join(s.dataDir, safeKey+".json")
}

func (s *FileStore) Get(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := s.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *FileStore) Set(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	filePath := s.getFilePath(key)
	return os.WriteFile(filePath, data, 0644)
}

func (s *FileStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := s.getFilePath(key)
	err := os.Remove(filePath)
	if os.IsNotExist(err) {
		return nil // File doesn't exist, treat as successful deletion
	}
	return err
}

func (s *FileStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(s.dataDir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
