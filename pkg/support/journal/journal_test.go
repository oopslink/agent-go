package journal

import (
	"os"
	"strings"
	"testing"
)

func TestJournal_LogAndUsage(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "journal_test_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	journal, err := NewFileJournal(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create FileJournal: %v", err)
	}

	// 测试日志写入
	err = journal.Info("message", "test", "hello world", "user", "alice", "id", 123)
	if err != nil {
		t.Errorf("Info log failed: %v", err)
	}
	err = journal.Error("system", "test", "something wrong", "code", 500)
	if err != nil {
		t.Errorf("Error log failed: %v", err)
	}

	// 测试用量累加
	journal.AccumulateUsage("session-1", map[string]float64{"input_tokens": 10, "output_tokens": 5})
	journal.AccumulateUsage("session-1", map[string]float64{"input_tokens": 2, "cost": 0.01})
	journal.AccumulateUsage("session-2", map[string]float64{"input_tokens": 7})

	// 检查文件内容
	_ = tmpfile.Sync()
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logStr := string(content)

	if !strings.Contains(logStr, "hello world") {
		t.Errorf("log file missing Info message")
	}
	if !strings.Contains(logStr, "something wrong") {
		t.Errorf("log file missing Error message")
	}
	if !strings.Contains(logStr, "session-1") || !strings.Contains(logStr, "input_tokens") {
		t.Errorf("log file missing usage accumulation")
	}
}

func TestUsageAccumulator(t *testing.T) {
	accumulator := NewUsageAccumulator()

	// 测试累加
	result1 := accumulator.Accumulate("session-1", map[string]float64{"input_tokens": 10, "output_tokens": 5})
	if result1["input_tokens"] != 10 || result1["output_tokens"] != 5 {
		t.Errorf("first accumulation failed: %v", result1)
	}

	result2 := accumulator.Accumulate("session-1", map[string]float64{"input_tokens": 2, "cost": 0.01})
	if result2["input_tokens"] != 12 || result2["output_tokens"] != 5 || result2["cost"] != 0.01 {
		t.Errorf("second accumulation failed: %v", result2)
	}

	// 测试获取用量
	usage := accumulator.GetUsage("session-1")
	if usage["input_tokens"] != 12 {
		t.Errorf("get usage failed: %v", usage)
	}

	// 测试重置
	accumulator.Reset("session-1")
	usage = accumulator.GetUsage("session-1")
	if usage != nil {
		t.Errorf("reset failed: %v", usage)
	}
}

func TestCompositeJournal(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "composite_test_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	fileStorage, err := NewFileStorage(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	consoleStorage := NewConsoleStorage()
	journal := NewJournal(fileStorage, consoleStorage)

	// 测试组合日志
	err = journal.Info("test", "composite", "testing composite journal")
	if err != nil {
		t.Errorf("composite journal failed: %v", err)
	}

	// 测试组合用量累加
	journal.AccumulateUsage("composite-session", map[string]float64{"tokens": 100})

	// 检查文件内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	logStr := string(content)

	if !strings.Contains(logStr, "testing composite journal") {
		t.Errorf("composite journal missing log entry")
	}
	if !strings.Contains(logStr, "composite-session") {
		t.Errorf("composite journal missing usage entry")
	}
}

func TestNewJournalWithNoStorage(t *testing.T) {
	// 测试无参数调用，应该默认使用控制台存储
	journal := NewJournal()

	err := journal.Info("test", "default", "testing default console storage")
	if err != nil {
		t.Errorf("default journal failed: %v", err)
	}
}

func TestNewJournalWithSingleStorage(t *testing.T) {
	// 测试单个存储
	consoleStorage := NewConsoleStorage()
	journal := NewJournal(consoleStorage)

	err := journal.Info("test", "single", "testing single storage")
	if err != nil {
		t.Errorf("single storage journal failed: %v", err)
	}
}
