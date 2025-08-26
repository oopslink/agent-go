package memory

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// NewJsonCodec creates a JSON encoder/decoder
func NewJsonCodec() MemoryItemCodec {
	return &JsonCodec{}
}

var _ MemoryItemCodec = &JsonCodec{}

// JsonCodec JSON encoder/decoder implementation
type JsonCodec struct{}

// serializedMemoryItem internal structure for serialization
type serializedMemoryItem struct {
	ID        MemoryItemId    `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Content   json.RawMessage `json:"content"`
}

// Encode encodes a MemoryItem
func (c *JsonCodec) Encode(item MemoryItem) ([]byte, error) {
	if item == nil {
		return nil, fmt.Errorf("item is nil")
	}

	var itemType string
	var contentData []byte
	var err error

	// Process based on specific type
	switch v := item.(type) {
	case *ChatMessageMemoryItem:
		itemType = "chat_message"
		// Use llms.JsonCodec to serialize the message
		llmsCodec := llms.NewJsonCodec()
		contentData, err = llmsCodec.Encode(v.message)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chat message: %w", err)
		}
	default:
		itemType = "unknown"
		contentData, err = json.Marshal(v.GetContent())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content: %w", err)
		}
	}

	serialized := serializedMemoryItem{
		ID:        item.GetId(),
		Type:      itemType,
		CreatedAt: item.GetCreatedAt(),
		Content:   contentData,
	}

	return json.Marshal(serialized)
}

// Decode decodes a MemoryItem
func (c *JsonCodec) Decode(data []byte) (MemoryItem, error) {
	var serialized serializedMemoryItem
	if err := json.Unmarshal(data, &serialized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal serialized item: %w", err)
	}

	switch serialized.Type {
	case "chat_message":
		// Use llms.JsonCodec to deserialize the message
		llmsCodec := llms.NewJsonCodec()
		message, err := llmsCodec.Decode([]byte(serialized.Content))
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat message: %w", err)
		}

		item := &ChatMessageMemoryItem{
			memoryId: serialized.ID,
			message:  message,
		}
		return item, nil

	default:
		// For unknown types, create a generic MemoryItem
		return &GenericMemoryItem{
			id:        serialized.ID,
			content:   serialized.Content,
			createdAt: serialized.CreatedAt,
		}, nil
	}
}

// GenericMemoryItem generic MemoryItem implementation for handling unknown types
type GenericMemoryItem struct {
	id        MemoryItemId
	content   interface{}
	createdAt time.Time
}

func (g *GenericMemoryItem) GetId() MemoryItemId {
	return g.id
}

func (g *GenericMemoryItem) GetContent() any {
	return g.content
}

func (g *GenericMemoryItem) GetCreatedAt() time.Time {
	return g.createdAt
}

func (g *GenericMemoryItem) AsMessage() (*llms.Message, bool) {
	return nil, false
}

// NewGenericMemoryItem creates a generic MemoryItem
func NewGenericMemoryItem(content interface{}) MemoryItem {
	return &GenericMemoryItem{
		id:        MemoryItemId(utils.GenerateUUID()),
		content:   content,
		createdAt: time.Now(),
	}
}
