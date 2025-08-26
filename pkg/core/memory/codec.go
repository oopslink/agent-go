package memory

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// NewJsonCodec 创建JSON编解码器
func NewJsonCodec() MemoryItemCodec {
	return &JsonCodec{}
}

// NewDefaultCodec 创建JSON编解码器（向后兼容性）
// Deprecated: 使用 NewJsonCodec 代替
func NewDefaultCodec() MemoryItemCodec {
	return NewJsonCodec()
}

var _ MemoryItemCodec = &JsonCodec{}

// JsonCodec JSON编解码器实现
type JsonCodec struct{}

// serializedMemoryItem 用于序列化的内部结构
type serializedMemoryItem struct {
	ID        MemoryItemId    `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Content   json.RawMessage `json:"content"`
}

// Encode 编码 MemoryItem
func (c *JsonCodec) Encode(item MemoryItem) ([]byte, error) {
	if item == nil {
		return nil, fmt.Errorf("item is nil")
	}

	var itemType string
	var contentData []byte
	var err error

	// 根据具体类型处理
	switch v := item.(type) {
	case *ChatMessageMemoryItem:
		itemType = "chat_message"
		// 使用 llms.JsonCodec 来序列化消息
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

// Decode 解码 MemoryItem
func (c *JsonCodec) Decode(data []byte) (MemoryItem, error) {
	var serialized serializedMemoryItem
	if err := json.Unmarshal(data, &serialized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal serialized item: %w", err)
	}

	switch serialized.Type {
	case "chat_message":
		// 使用 llms.JsonCodec 来反序列化消息
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
		// 对于未知类型，创建一个通用的 MemoryItem
		return &GenericMemoryItem{
			id:        serialized.ID,
			content:   serialized.Content,
			createdAt: serialized.CreatedAt,
		}, nil
	}
}

// GenericMemoryItem 通用的 MemoryItem 实现，用于处理未知类型
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

// NewGenericMemoryItem 创建一个通用的 MemoryItem
func NewGenericMemoryItem(content interface{}) MemoryItem {
	return &GenericMemoryItem{
		id:        MemoryItemId(utils.GenerateUUID()),
		content:   content,
		createdAt: time.Now(),
	}
}
