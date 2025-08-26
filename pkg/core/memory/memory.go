package memory

import (
	"context"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

// MemoryItemCodec interface for serializing and deserializing MemoryItem
type MemoryItemCodec interface {
	Encode(item MemoryItem) ([]byte, error)
	Decode(data []byte) (MemoryItem, error)
}

// MemoryStore interface defines storage backend
type MemoryStore interface {
	// Add a MemoryItem to storage
	Store(ctx context.Context, item MemoryItem) error
	// Retrieve all MemoryItem
	Load(ctx context.Context) ([]MemoryItem, error)
	// Clear storage
	Clear(ctx context.Context) error
	// Close storage
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
