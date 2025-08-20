package journal

import (
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Journal interface {
	Debug(category, source, message string, data ...any) error
	Info(category, source, message string, data ...any) error
	Warning(category, source, message string, data ...any) error
	Error(category, source, message string, data ...any) error

	// 用于累加用量（如 LLM token/cost 统计等）
	AccumulateUsage(usageSessionId string, usage map[string]float64)
	GetUsage(usageSessionId string) map[string]float64
}

type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     Level                  `json:"level"`
	Category  string                 `json:"category"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Journal 实现，使用 Storage 进行存储
type journal struct {
	storage          Storage
	usageAccumulator *UsageAccumulator
	mu               sync.Mutex
}

func NewJournal(storages ...Storage) Journal {
	var storage Storage

	switch len(storages) {
	case 0:
		// 默认使用控制台存储
		storage = NewConsoleStorage()
	case 1:
		// 单个存储
		storage = storages[0]
	default:
		// 多个存储，创建组合存储
		storage = NewCompositeStorage(storages...)
	}

	return &journal{
		storage:          storage,
		usageAccumulator: NewUsageAccumulator(),
	}
}

func (j *journal) log(level Level, category, source, message string, data ...any) error {
	entry := Entry{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Source:    source,
		Message:   message,
		Data:      parseData(data...),
	}
	return j.storage.Write(entry)
}

func (j *journal) Debug(category, source, message string, data ...any) error {
	return j.log(LevelDebug, category, source, message, data...)
}

func (j *journal) Info(category, source, message string, data ...any) error {
	return j.log(LevelInfo, category, source, message, data...)
}

func (j *journal) Warning(category, source, message string, data ...any) error {
	return j.log(LevelWarning, category, source, message, data...)
}

func (j *journal) Error(category, source, message string, data ...any) error {
	return j.log(LevelError, category, source, message, data...)
}

func (j *journal) AccumulateUsage(usageSessionId string, usage map[string]float64) {
	// 使用 UsageAccumulator 进行累加
	accumulated := j.usageAccumulator.Accumulate(usageSessionId, usage)
	// 将累加后的结果写入存储
	_ = j.storage.WriteUsage(usageSessionId, accumulated)
}

func (j *journal) GetUsage(usageSessionId string) map[string]float64 {
	return j.usageAccumulator.GetUsage(usageSessionId)
}

// 便捷构造函数
func NewFileJournal(path string) (Journal, error) {
	storage, err := NewFileStorage(path)
	if err != nil {
		return nil, err
	}
	return NewJournal(storage), nil
}

func NewConsoleJournal() Journal {
	return NewJournal(NewConsoleStorage())
}

func parseData(data ...any) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i+1 < len(data); i += 2 {
		if key, ok := data[i].(string); ok {
			obj := data[i+1]
			if s, ok := obj.(string); ok {
				result[key] = s
				continue
			}
			if err, ok := obj.(error); ok {
				result[key] = err.Error()
				continue
			}
			yamlData, _ := yaml.Marshal(obj)
			result[key] = string(yamlData)
		}
	}
	return result
}
