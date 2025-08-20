package memory

import (
	"time"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func NewToolCallResultMemoryItem(result *llms.ToolCallResult) MemoryItem {
	return &ToolCallResultMemoryItem{
		memoryId:  MemoryItemId(utils.GenerateUUID()),
		result:    result,
		timestamp: time.Now(),
	}
}

type ToolCallResultMemoryItem struct {
	memoryId  MemoryItemId
	result    *llms.ToolCallResult
	timestamp time.Time
}

// GetId implements MemoryItem.
func (t *ToolCallResultMemoryItem) GetId() MemoryItemId {
	return t.memoryId
}

// GetContent implements MemoryItem.
func (t *ToolCallResultMemoryItem) GetContent() any {
	return t.result
}

// GetCreatedAt implements MemoryItem.
func (t *ToolCallResultMemoryItem) GetCreatedAt() time.Time {
	return t.timestamp
}

// AsMessage implements MemoryItem.
func (c *ToolCallResultMemoryItem) AsMessage() (*llms.Message, bool) {
	if c.result == nil {
		return nil, false
	}
	return llms.NewToolCallResultMessage(c.result, c.timestamp), true
}
