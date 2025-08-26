package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSimpleStateWithInMemoryStore(t *testing.T) {
	state := NewInMemoryState()

	// 测试Put和Get
	err := state.Put("test_key", "test_value")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := state.Get("test_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// 测试Delete
	err = state.Delete("test_key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	value, err = state.Get("test_key")
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}

	if value != nil {
		t.Errorf("Expected nil after delete, got %v", value)
	}
}

func TestSimpleStateWithFileStore(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "state_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	state, err := NewFileState(tempDir)
	if err != nil {
		t.Fatalf("NewFileState failed: %v", err)
	}

	// 测试Put和Get
	err = state.Put("test_key", "test_value")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := state.Get("test_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// 验证文件是否存在
	expectedFile := filepath.Join(tempDir, "test_key.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedFile)
	}

	// 测试Delete
	err = state.Delete("test_key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证文件已被删除
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Errorf("Expected file %s to be deleted", expectedFile)
	}

	value, err = state.Get("test_key")
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}

	if value != nil {
		t.Errorf("Expected nil after delete, got %v", value)
	}
}

func TestSimpleStateClear(t *testing.T) {
	// 测试内存存储的Clear
	state := NewInMemoryState()

	state.Put("key1", "value1")
	state.Put("key2", "value2")

	err := state.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	value1, _ := state.Get("key1")
	value2, _ := state.Get("key2")

	if value1 != nil || value2 != nil {
		t.Errorf("Expected all values to be cleared")
	}

	// 测试文件存储的Clear
	tempDir, err := os.MkdirTemp("", "state_clear_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileState, err := NewFileState(tempDir)
	if err != nil {
		t.Fatalf("NewFileState failed: %v", err)
	}

	fileState.Put("key1", "value1")
	fileState.Put("key2", "value2")

	err = fileState.Clear()
	if err != nil {
		t.Fatalf("File state Clear failed: %v", err)
	}

	value1, _ = fileState.Get("key1")
	value2, _ = fileState.Get("key2")

	if value1 != nil || value2 != nil {
		t.Errorf("Expected all file values to be cleared")
	}
}

func TestStateFactory(t *testing.T) {
	// 测试内存状态工厂
	memoryState, err := NewState(StateConfig{Type: StateTypeMemory})
	if err != nil {
		t.Fatalf("Failed to create memory state: %v", err)
	}

	err = memoryState.Put("factory_test", "memory_value")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err := memoryState.Get("factory_test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "memory_value" {
		t.Errorf("Expected 'memory_value', got %v", value)
	}

	// 测试文件状态工厂
	tempDir, err := os.MkdirTemp("", "factory_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileState, err := NewState(StateConfig{
		Type:     StateTypeFile,
		FilePath: tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to create file state: %v", err)
	}

	err = fileState.Put("factory_test", "file_value")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	value, err = fileState.Get("factory_test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if value != "file_value" {
		t.Errorf("Expected 'file_value', got %v", value)
	}

	// 测试错误情况
	_, err = NewState(StateConfig{Type: StateTypeFile}) // 缺少FilePath
	if err == nil {
		t.Error("Expected error for file state without path")
	}

	_, err = NewState(StateConfig{Type: "invalid"})
	if err == nil {
		t.Error("Expected error for invalid state type")
	}
}
