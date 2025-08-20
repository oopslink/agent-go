package memory

import (
	"time"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func NewChatMessageMemoryItem(message *llms.Message) MemoryItem {
	return &ChatMessageMemoryItem{
		memoryId: MemoryItemId(utils.GenerateUUID()),
		message:  message,
	}
}

type ChatMessageMemoryItem struct {
	memoryId MemoryItemId
	message  *llms.Message
}

func (c *ChatMessageMemoryItem) GetId() MemoryItemId {
	return c.memoryId
}

func (c *ChatMessageMemoryItem) GetContent() any {
	return c.message
}

func (c *ChatMessageMemoryItem) GetCreatedAt() time.Time {
	if c.message == nil {
		return time.Now()
	}
	return c.message.Timestamp
}

// AsMessage implements MemoryItem.
func (c *ChatMessageMemoryItem) AsMessage() (*llms.Message, bool) {
	if c.message == nil {
		return nil, false
	}
	return c.message, true
}
