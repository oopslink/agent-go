package memory

import (
	"context"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

// MemoryItemCodec 接口用于序列化和反序列化 MemoryItem
type MemoryItemCodec interface {
	Encode(item MemoryItem) ([]byte, error)
	Decode(data []byte) (MemoryItem, error)
}

// MemoryStore 接口定义存储后端
type MemoryStore interface {
	// 添加一个 MemoryItem 到存储
	Store(ctx context.Context, item MemoryItem) error
	// 检索所有 MemoryItem
	Load(ctx context.Context) ([]MemoryItem, error)
	// 清空存储
	Clear(ctx context.Context) error
	// 关闭存储
	Close() error
}

func NewMemoryRetrieveOptions() *MemoryRetrieveOptions {
	return &MemoryRetrieveOptions{
		Limit: -1,
	}
}

type MemoryRetrieveOptions struct {
	Limit int `json:"limit"`
}

type MemoryRetrieveOption func(o *MemoryRetrieveOptions)

func WithNoLimit() MemoryRetrieveOption {
	return WithMaxLimit(-1)
}

func WithMaxLimit(limit int) MemoryRetrieveOption {
	return func(o *MemoryRetrieveOptions) {
		o.Limit = limit
	}
}

type MemoryRetriever interface {
	Retrieve(ctx context.Context, options ...MemoryRetrieveOption) ([]MemoryItem, error)
}

type Memory interface {
	MemoryRetriever

	Add(ctx context.Context, memory MemoryItem) error

	Reset() error
}

type MemoryItemId string

type MemoryItem interface {
	GetId() MemoryItemId
	GetContent() any
	GetCreatedAt() time.Time

	AsMessage() (*llms.Message, bool)
}

func AsMessages(items []MemoryItem) []*llms.Message {
	var messages []*llms.Message
	for _, item := range items {
		if msg, ok := item.AsMessage(); ok && msg != nil {
			messages = append(messages, msg)
		}
	}
	return messages
}
