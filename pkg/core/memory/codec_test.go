package memory

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJsonCodec(t *testing.T) {
	codec := NewJsonCodec()
	assert.NotNil(t, codec)
	assert.IsType(t, &JsonCodec{}, codec)
}

func TestNewDefaultCodec_BackwardCompatibility(t *testing.T) {
	// 测试向后兼容的函数
	codec := NewDefaultCodec()
	assert.NotNil(t, codec)
	assert.IsType(t, &JsonCodec{}, codec)
}

func TestJsonCodec_Encode_NilItem(t *testing.T) {
	codec := &JsonCodec{}
	data, err := codec.Encode(nil)
	
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "item is nil")
}

func TestJsonCodec_Encode_ChatMessageMemoryItem(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建测试用的 Message
	timestamp := time.Now().UTC() // 使用 UTC 时区
	message := &llms.Message{
		MessageId: "test-message-id",
		Creator: llms.MessageCreator{
			Role: llms.MessageRoleUser,
			Name: nil,
		},
		Model: llms.ModelId{
			Provider: "openai",
			ID:       "gpt-4",
		},
		Parts: []llms.Part{
			&llms.TextPart{Text: "Hello, world!"},
		},
		Timestamp: timestamp,
	}
	
	// 创建 ChatMessageMemoryItem
	item := &ChatMessageMemoryItem{
		memoryId: MemoryItemId("test-id"),
		message:  message,
	}
	
	// 编码
	data, err := codec.Encode(item)
	
	require.NoError(t, err)
	assert.NotNil(t, data)
	
	// 验证编码后的数据结构
	var serialized serializedMemoryItem
	err = json.Unmarshal(data, &serialized)
	require.NoError(t, err)
	
	assert.Equal(t, MemoryItemId("test-id"), serialized.ID)
	assert.Equal(t, "chat_message", serialized.Type)
	// 由于 JSON 序列化时间会丢失一些精度，我们检查是否在合理范围内
	timeDiff := timestamp.Sub(serialized.CreatedAt)
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second)
	assert.NotNil(t, serialized.Content)
}

func TestJsonCodec_Encode_GenericMemoryItem(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建通用的 MemoryItem
	content := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	item := NewGenericMemoryItem(content)
	
	// 编码
	data, err := codec.Encode(item)
	
	require.NoError(t, err)
	assert.NotNil(t, data)
	
	// 验证编码后的数据结构
	var serialized serializedMemoryItem
	err = json.Unmarshal(data, &serialized)
	require.NoError(t, err)
	
	assert.Equal(t, item.GetId(), serialized.ID)
	assert.Equal(t, "unknown", serialized.Type)
	assert.NotNil(t, serialized.Content)
}

func TestJsonCodec_Encode_ContentMarshalError(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建一个无法序列化的内容（包含循环引用）
	content := make(map[string]interface{})
	content["self"] = content // 循环引用
	
	item := NewGenericMemoryItem(content)
	
	// 编码应该失败
	data, err := codec.Encode(item)
	
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal content")
}

func TestJsonCodec_Encode_ChatMessage_NilMessage(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建包含 nil message 的 ChatMessageMemoryItem
	item := &ChatMessageMemoryItem{
		memoryId: MemoryItemId("test-id"),
		message:  nil, // nil message
	}
	
	// 编码应该失败
	data, err := codec.Encode(item)
	
	assert.Nil(t, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal chat message")
}

func TestJsonCodec_Decode_InvalidJSON(t *testing.T) {
	codec := &JsonCodec{}
	
	// 无效的 JSON 数据
	invalidData := []byte("invalid json")
	
	item, err := codec.Decode(invalidData)
	
	assert.Nil(t, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal serialized item")
}

func TestJsonCodec_Decode_ChatMessage(t *testing.T) {
	codec := &JsonCodec{}
	
	// 准备测试数据
	timestamp := time.Now().UTC()
	message := &llms.Message{
		MessageId: "test-message-id",
		Creator: llms.MessageCreator{
			Role: llms.MessageRoleUser,
			Name: nil,
		},
		Model: llms.ModelId{
			Provider: "openai",
			ID:       "gpt-4",
		},
		Parts: []llms.Part{
			&llms.TextPart{Text: "Hello, world!"},
		},
		Timestamp: timestamp,
	}
	
	// 使用 llms.JsonCodec 来序列化消息
	llmsCodec := llms.NewJsonCodec()
	messageData, err := llmsCodec.Encode(message)
	require.NoError(t, err)
	
	serialized := serializedMemoryItem{
		ID:        MemoryItemId("test-id"),
		Type:      "chat_message",
		CreatedAt: timestamp,
		Content:   json.RawMessage(messageData),
	}
	
	data, err := json.Marshal(serialized)
	require.NoError(t, err)
	
	// 解码
	item, err := codec.Decode(data)
	
	require.NoError(t, err)
	assert.NotNil(t, item)
	
	// 验证类型
	chatItem, ok := item.(*ChatMessageMemoryItem)
	require.True(t, ok)
	
	// 验证数据
	assert.Equal(t, MemoryItemId("test-id"), chatItem.GetId())
	// 由于 JSON 序列化时间会丢失一些精度，我们检查是否在合理范围内
	timeDiff := timestamp.Sub(chatItem.GetCreatedAt())
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second)
	
	// 验证消息内容
	msg, ok := chatItem.AsMessage()
	require.True(t, ok)
	assert.Equal(t, "test-message-id", msg.MessageId)
	assert.Equal(t, llms.MessageRoleUser, msg.Creator.Role)
}

func TestJsonCodec_Decode_ChatMessage_InvalidContent(t *testing.T) {
	codec := &JsonCodec{}
	
	// 手动构造包含无效消息内容的 JSON 数据
	invalidJsonData := `{
		"id": "test-id",
		"type": "chat_message",
		"created_at": "2025-08-26T06:00:00Z",
		"content": "invalid json content"
	}`
	
	// 解码应该失败
	item, err := codec.Decode([]byte(invalidJsonData))
	
	assert.Nil(t, item)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal chat message")
}

func TestJsonCodec_Decode_UnknownType(t *testing.T) {
	codec := &JsonCodec{}
	
	// 准备未知类型的数据
	content := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	contentData, err := json.Marshal(content)
	require.NoError(t, err)
	
	timestamp := time.Now().UTC()
	serialized := serializedMemoryItem{
		ID:        MemoryItemId("test-id"),
		Type:      "unknown_type",
		CreatedAt: timestamp,
		Content:   contentData,
	}
	
	data, err := json.Marshal(serialized)
	require.NoError(t, err)
	
	// 解码
	item, err := codec.Decode(data)
	
	require.NoError(t, err)
	assert.NotNil(t, item)
	
	// 验证类型
	genericItem, ok := item.(*GenericMemoryItem)
	require.True(t, ok)
	
	// 验证数据
	assert.Equal(t, MemoryItemId("test-id"), genericItem.GetId())
	// 由于 JSON 序列化时间会丢失一些精度，我们检查是否在合理范围内
	timeDiff := timestamp.Sub(genericItem.GetCreatedAt())
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second)
	
	// 验证内容是 json.RawMessage 类型
	assert.IsType(t, json.RawMessage{}, genericItem.GetContent())
	
	// 验证不能转换为消息
	msg, ok := genericItem.AsMessage()
	assert.Nil(t, msg)
	assert.False(t, ok)
}

func TestJsonCodec_EncodeDecodeRoundTrip_ChatMessage(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建原始数据
	timestamp := time.Now().UTC()
	message := &llms.Message{
		MessageId: "test-message-id",
		Creator: llms.MessageCreator{
			Role: llms.MessageRoleAssistant,
			Name: nil,
		},
		Model: llms.ModelId{
			Provider: "openai",
			ID:       "gpt-4",
		},
		Parts: []llms.Part{
			&llms.TextPart{Text: "Test response"},
		},
		Timestamp: timestamp,
	}
	
	originalItem := &ChatMessageMemoryItem{
		memoryId: MemoryItemId("test-id"),
		message:  message,
	}
	
	// 编码
	data, err := codec.Encode(originalItem)
	require.NoError(t, err)
	
	// 解码
	decodedItem, err := codec.Decode(data)
	require.NoError(t, err)
	
	// 验证类型
	chatItem, ok := decodedItem.(*ChatMessageMemoryItem)
	require.True(t, ok)
	
	// 验证数据一致性
	assert.Equal(t, originalItem.GetId(), chatItem.GetId())
	// 由于 JSON 序列化时间会丢失一些精度，我们检查是否在合理范围内
	timeDiff := originalItem.GetCreatedAt().Sub(chatItem.GetCreatedAt())
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second)
	
	// 验证消息内容一致性
	originalMsg, ok1 := originalItem.AsMessage()
	decodedMsg, ok2 := chatItem.AsMessage()
	require.True(t, ok1)
	require.True(t, ok2)
	
	assert.Equal(t, originalMsg.MessageId, decodedMsg.MessageId)
	assert.Equal(t, originalMsg.Creator.Role, decodedMsg.Creator.Role)
	assert.Equal(t, originalMsg.Model.Provider, decodedMsg.Model.Provider)
	assert.Equal(t, originalMsg.Model.ID, decodedMsg.Model.ID)
}

func TestJsonCodec_EncodeDecodeRoundTrip_GenericItem(t *testing.T) {
	codec := &JsonCodec{}
	
	// 创建原始数据
	content := map[string]interface{}{
		"string_field": "test",
		"int_field":    123,
		"bool_field":   true,
		"nested": map[string]interface{}{
			"nested_field": "nested_value",
		},
	}
	
	originalItem := NewGenericMemoryItem(content)
	
	// 编码
	data, err := codec.Encode(originalItem)
	require.NoError(t, err)
	
	// 解码
	decodedItem, err := codec.Decode(data)
	require.NoError(t, err)
	
	// 验证类型
	genericItem, ok := decodedItem.(*GenericMemoryItem)
	require.True(t, ok)
	
	// 验证数据一致性
	assert.Equal(t, originalItem.GetId(), genericItem.GetId())
	// 注意：由于通过 JSON 序列化，时间精度可能会有差异，所以我们检查是否在合理范围内
	timeDiff := originalItem.GetCreatedAt().Sub(genericItem.GetCreatedAt())
	assert.True(t, timeDiff >= -time.Second && timeDiff <= time.Second)
}

func TestGenericMemoryItem_Methods(t *testing.T) {
	content := "test content"
	item := NewGenericMemoryItem(content)
	
	// 验证 GetId 返回非空值
	assert.NotEmpty(t, item.GetId())
	
	// 验证 GetContent 返回正确内容
	assert.Equal(t, content, item.GetContent())
	
	// 验证 GetCreatedAt 返回合理时间
	now := time.Now()
	createdAt := item.GetCreatedAt()
	assert.True(t, createdAt.Before(now.Add(time.Second)) && createdAt.After(now.Add(-time.Second)))
	
	// 验证 AsMessage 返回 false
	msg, ok := item.AsMessage()
	assert.Nil(t, msg)
	assert.False(t, ok)
}

func TestGenericMemoryItem_DirectCreation(t *testing.T) {
	// 直接创建 GenericMemoryItem 进行测试
	id := MemoryItemId("direct-test-id")
	content := "direct test content"
	createdAt := time.Now().Add(-time.Hour)
	
	item := &GenericMemoryItem{
		id:        id,
		content:   content,
		createdAt: createdAt,
	}
	
	assert.Equal(t, id, item.GetId())
	assert.Equal(t, content, item.GetContent())
	assert.Equal(t, createdAt, item.GetCreatedAt())
	
	msg, ok := item.AsMessage()
	assert.Nil(t, msg)
	assert.False(t, ok)
}

func TestChatMessageMemoryItem_Methods(t *testing.T) {
	// 测试 ChatMessageMemoryItem 的所有方法
	timestamp := time.Now().UTC()
	message := &llms.Message{
		MessageId: "test-message-id",
		Creator: llms.MessageCreator{
			Role: llms.MessageRoleUser,
			Name: nil,
		},
		Model: llms.ModelId{
			Provider: "openai",
			ID:       "gpt-4",
		},
		Parts: []llms.Part{
			&llms.TextPart{Text: "Hello, world!"},
		},
		Timestamp: timestamp,
	}
	
	// 使用 NewChatMessageMemoryItem 函数
	item := NewChatMessageMemoryItem(message)
	
	// 验证 GetId 返回非空值
	assert.NotEmpty(t, item.GetId())
	
	// 验证 GetContent 返回消息对象
	assert.Equal(t, message, item.GetContent())
	
	// 验证 GetCreatedAt 返回消息的时间戳
	assert.Equal(t, timestamp, item.GetCreatedAt())
	
	// 验证 AsMessage 返回消息对象
	msg, ok := item.AsMessage()
	assert.True(t, ok)
	assert.Equal(t, message, msg)
}

func TestChatMessageMemoryItem_NilMessage(t *testing.T) {
	// 测试当 message 为 nil 时的行为
	item := &ChatMessageMemoryItem{
		memoryId: MemoryItemId("test-id"),
		message:  nil,
	}
	
	// GetContent 返回 nil
	assert.Nil(t, item.GetContent())
	
	// GetCreatedAt 返回当前时间（会有一些容差）
	createdAt := item.GetCreatedAt()
	now := time.Now()
	assert.True(t, createdAt.Before(now.Add(time.Second)) && createdAt.After(now.Add(-time.Second)))
	
	// AsMessage 返回 nil, false
	msg, ok := item.AsMessage()
	assert.Nil(t, msg)
	assert.False(t, ok)
}
