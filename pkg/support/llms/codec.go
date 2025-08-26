package llms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Codec defines the interface for encoding and decoding messages
type Codec interface {
	// Encode converts a Message to its encoded representation
	Encode(message *Message) ([]byte, error)
	
	// Decode converts encoded data back to a Message
	Decode(data []byte) (*Message, error)
}

// JsonCodec implements Codec for JSON serialization
type JsonCodec struct{}

// NewJsonCodec creates a new JsonCodec instance
func NewJsonCodec() Codec {
	return &JsonCodec{}
}

// serializableMessage represents the JSON-serializable version of Message
type serializableMessage struct {
	MessageId string                 `json:"message_id"`
	Creator   MessageCreator         `json:"creator"`
	Model     ModelId                `json:"model"`
	Parts     []serializablePart     `json:"parts"`
	Timestamp time.Time              `json:"timestamp"`
}

// serializablePart represents the JSON-serializable version of Part
type serializablePart struct {
	Type    PartType    `json:"type"`
	Content interface{} `json:"content"`
}

// Encode converts a Message to JSON bytes
func (c *JsonCodec) Encode(message *Message) ([]byte, error) {
	if message == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	serializable := serializableMessage{
		MessageId: message.MessageId,
		Creator:   message.Creator,
		Model:     message.Model,
		Timestamp: message.Timestamp,
		Parts:     make([]serializablePart, len(message.Parts)),
	}

	for i, part := range message.Parts {
		serializablePart := serializablePart{
			Type: part.Type(),
		}

		switch p := part.(type) {
		case *TextPart:
			serializablePart.Content = map[string]interface{}{
				"text": p.Text,
			}
		case *DataPart:
			serializablePart.Content = map[string]interface{}{
				"data": p.Data,
			}
		case *BinaryPart:
			serializablePart.Content = map[string]interface{}{
				"name":           p.Name,
				"url":            p.URL,
				"mime_type":      p.MIMEType,
				"content":        p.Content,
				"content_length": p.ContentLength,
			}
		case *ToolCall:
			serializablePart.Content = map[string]interface{}{
				"tool_call_id": p.ToolCallId,
				"name":         p.Name,
				"arguments":    p.Arguments,
			}
		case *ToolCallResult:
			serializablePart.Content = map[string]interface{}{
				"tool_call_id": p.ToolCallId,
				"name":         p.Name,
				"result":       p.Result,
			}
		default:
			return nil, fmt.Errorf("unsupported part type: %T", part)
		}

		serializable.Parts[i] = serializablePart
	}

	return json.Marshal(serializable)
}

// Decode converts JSON bytes back to a Message
func (c *JsonCodec) Decode(data []byte) (*Message, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var serializable serializableMessage
	if err := json.Unmarshal(data, &serializable); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	message := &Message{
		MessageId: serializable.MessageId,
		Creator:   serializable.Creator,
		Model:     serializable.Model,
		Timestamp: serializable.Timestamp,
		Parts:     make([]Part, len(serializable.Parts)),
	}

	for i, serializedPart := range serializable.Parts {
		part, err := c.deserializePart(serializedPart)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize part %d: %w", i, err)
		}
		message.Parts[i] = part
	}

	return message, nil
}

// deserializePart converts a serializablePart back to a Part interface
func (c *JsonCodec) deserializePart(sp serializablePart) (Part, error) {
	content, ok := sp.Content.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid part content format")
	}

	switch sp.Type {
	case PartTypeText:
		text, ok := content["text"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid text part content")
		}
		return &TextPart{Text: text}, nil

	case PartTypeData:
		data, ok := content["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid data part content")
		}
		return &DataPart{Data: data}, nil

	case PartTypeBinary:
		binaryPart := &BinaryPart{}
		
		if name, exists := content["name"]; exists && name != nil {
			if nameStr, ok := name.(string); ok {
				binaryPart.Name = &nameStr
			}
		}
		
		if url, exists := content["url"]; exists && url != nil {
			if urlStr, ok := url.(string); ok {
				binaryPart.URL = &urlStr
			}
		}
		
		if mimeType, ok := content["mime_type"].(string); ok {
			binaryPart.MIMEType = mimeType
		}
		
		if contentData, exists := content["content"]; exists && contentData != nil {
			// JSON marshals []byte as base64 string, but when unmarshaling back to interface{},
			// it might come back as a string or as an array of numbers
			switch v := contentData.(type) {
			case []byte:
				binaryPart.Content = v
			case string:
				// It's a base64 string, decode it
				if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
					binaryPart.Content = decoded
				}
			case []interface{}:
				// It's an array of numbers, convert to []byte
				bytes := make([]byte, len(v))
				for i, num := range v {
					if f, ok := num.(float64); ok {
						bytes[i] = byte(f)
					}
				}
				binaryPart.Content = bytes
			}
		}
		
		if contentLength, ok := content["content_length"].(float64); ok {
			binaryPart.ContentLength = int64(contentLength)
		}
		
		return binaryPart, nil

	case PartTypeToolCall:
		toolCallId, _ := content["tool_call_id"].(string)
		name, _ := content["name"].(string)
		arguments, _ := content["arguments"].(map[string]interface{})
		
		return &ToolCall{
			ToolCallId: toolCallId,
			Name:       name,
			Arguments:  arguments,
		}, nil

	case PartTypeToolCallResult:
		toolCallId, _ := content["tool_call_id"].(string)
		name, _ := content["name"].(string)
		result, _ := content["result"].(map[string]interface{})
		
		return &ToolCallResult{
			ToolCallId: toolCallId,
			Name:       name,
			Result:     result,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported part type: %s", sp.Type)
	}
}
