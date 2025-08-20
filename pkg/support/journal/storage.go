package journal

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Storage 接口定义存储行为
type Storage interface {
	Write(entry Entry) error
	WriteUsage(sessionId string, usage map[string]float64) error
	Close() error
}

// FileStorage 文件存储实现
type FileStorage struct {
	mu   sync.Mutex
	file *os.File
}

func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStorage{file: file}, nil
}

func (f *FileStorage) Write(entry Entry) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := fmt.Fprintf(f.file, "%s\t%s\t%s\t%s\t%s\n",
		entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Category, entry.Source, entry.Message)

	if entry.Data != nil {
		for k, v := range entry.Data {
			_, err = fmt.Fprintf(f.file, "# %s:\n```\n%s\n```\n", k, v)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (f *FileStorage) WriteUsage(sessionId string, usage map[string]float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, err := fmt.Fprintf(f.file, "%s\tUSAGE\t%s\t%v\n", time.Now().Format(time.RFC3339), sessionId, usage)
	return err
}

func (f *FileStorage) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}

// ConsoleStorage 控制台存储实现
type ConsoleStorage struct {
	mu sync.Mutex
}

func NewConsoleStorage() *ConsoleStorage {
	return &ConsoleStorage{}
}

func (c *ConsoleStorage) Write(entry Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Printf("[%s] %s %s/%s: %s %v\n",
		entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Category, entry.Source, entry.Message, entry.Data)
	return nil
}

func (c *ConsoleStorage) WriteUsage(sessionId string, usage map[string]float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Printf("[%s] USAGE %s: %v\n", time.Now().Format(time.RFC3339), sessionId, usage)
	return nil
}

func (c *ConsoleStorage) Close() error {
	return nil
}

// CompositeStorage 组合存储实现
type CompositeStorage struct {
	storages []Storage
	mu       sync.Mutex
}

func NewCompositeStorage(storages ...Storage) *CompositeStorage {
	return &CompositeStorage{storages: storages}
}

func (c *CompositeStorage) Write(entry Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, storage := range c.storages {
		if err := storage.Write(entry); err != nil {
			return err
		}
	}
	return nil
}

func (c *CompositeStorage) WriteUsage(sessionId string, usage map[string]float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, storage := range c.storages {
		if err := storage.WriteUsage(sessionId, usage); err != nil {
			return err
		}
	}
	return nil
}

func (c *CompositeStorage) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, storage := range c.storages {
		if err := storage.Close(); err != nil {
			return err
		}
	}
	return nil
}
