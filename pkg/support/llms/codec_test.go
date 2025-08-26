package llms

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestNewJsonCodec(t *testing.T) {
	codec := NewJsonCodec()
	if codec == nil {
		t.Fatal("NewJsonCodec returned nil")
	}
	_, ok := codec.(*JsonCodec)
	if !ok {
		t.Fatal("NewJsonCodec did not return *JsonCodec")
	}
}

func TestJsonCodec_EncodeDecodeAllPartTypes(t *testing.T) {
	codec := NewJsonCodec()
	
	// Test with all part types
	msg := &Message{
		MessageId: "msg-1",
		Creator:   MessageCreator{Role: MessageRoleUser, Name: strPtr("testuser")},
		Model:     ModelId{Provider: "test", ID: "gpt-test"},
		Timestamp: time.Now().UTC().Truncate(time.Second),
		Parts: []Part{
			&TextPart{Text: "hello world"},
			&DataPart{Data: map[string]interface{}{"foo": "bar", "num": 42}},
			&BinaryPart{
				Name:          strPtr("test.txt"),
				URL:           strPtr("http://example.com/test.txt"),
				MIMEType:      "text/plain",
				Content:       []byte("binary content"),
				ContentLength: 14,
			},
			&ToolCall{
				ToolCallId: "call_123",
				Name:       "test_tool",
				Arguments:  map[string]interface{}{"param": "value"},
			},
			&ToolCallResult{
				ToolCallId: "call_123",
				Name:       "test_tool",
				Result:     map[string]interface{}{"output": "success"},
			},
		},
	}

	data, err := codec.Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Encoded JSON invalid: %v", err)
	}

	msg2, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify basic message fields
	if msg2.MessageId != msg.MessageId {
		t.Errorf("MessageId mismatch: got %s, want %s", msg2.MessageId, msg.MessageId)
	}
	if msg2.Creator.Role != msg.Creator.Role {
		t.Errorf("Creator.Role mismatch: got %s, want %s", msg2.Creator.Role, msg.Creator.Role)
	}
	if msg2.Creator.Name == nil || *msg2.Creator.Name != *msg.Creator.Name {
		t.Errorf("Creator.Name mismatch: got %v, want %v", msg2.Creator.Name, msg.Creator.Name)
	}
	if msg2.Model != msg.Model {
		t.Errorf("Model mismatch: got %s, want %s", msg2.Model, msg.Model)
	}
	if !msg2.Timestamp.Equal(msg.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", msg2.Timestamp, msg.Timestamp)
	}
	if len(msg2.Parts) != len(msg.Parts) {
		t.Fatalf("Parts length mismatch: got %d, want %d", len(msg2.Parts), len(msg.Parts))
	}

	// Verify TextPart
	if tp, ok := msg2.Parts[0].(*TextPart); !ok || tp.Text != "hello world" {
		t.Errorf("TextPart mismatch: got %+v", msg2.Parts[0])
	}

	// Verify DataPart
	if dp, ok := msg2.Parts[1].(*DataPart); !ok || !reflect.DeepEqual(dp.Data, map[string]interface{}{"foo": "bar", "num": float64(42)}) {
		t.Errorf("DataPart mismatch: got %+v", msg2.Parts[1])
	}

	// Verify BinaryPart
	if bp, ok := msg2.Parts[2].(*BinaryPart); !ok {
		t.Errorf("BinaryPart type mismatch: got %T", msg2.Parts[2])
	} else {
		if bp.Name == nil || *bp.Name != "test.txt" {
			t.Errorf("BinaryPart.Name mismatch: got %v, want %s", bp.Name, "test.txt")
		}
		if bp.URL == nil || *bp.URL != "http://example.com/test.txt" {
			t.Errorf("BinaryPart.URL mismatch: got %v, want %s", bp.URL, "http://example.com/test.txt")
		}
		if bp.MIMEType != "text/plain" {
			t.Errorf("BinaryPart.MIMEType mismatch: got %s, want %s", bp.MIMEType, "text/plain")
		}
		if !reflect.DeepEqual(bp.Content, []byte("binary content")) {
			t.Errorf("BinaryPart.Content mismatch: got %v, want %v", bp.Content, []byte("binary content"))
		}
		if bp.ContentLength != 14 {
			t.Errorf("BinaryPart.ContentLength mismatch: got %d, want %d", bp.ContentLength, 14)
		}
	}

	// Verify ToolCall
	if tc, ok := msg2.Parts[3].(*ToolCall); !ok {
		t.Errorf("ToolCall type mismatch: got %T", msg2.Parts[3])
	} else {
		if tc.ToolCallId != "call_123" {
			t.Errorf("ToolCall.ToolCallId mismatch: got %s, want %s", tc.ToolCallId, "call_123")
		}
		if tc.Name != "test_tool" {
			t.Errorf("ToolCall.Name mismatch: got %s, want %s", tc.Name, "test_tool")
		}
		if !reflect.DeepEqual(tc.Arguments, map[string]interface{}{"param": "value"}) {
			t.Errorf("ToolCall.Arguments mismatch: got %v", tc.Arguments)
		}
	}

	// Verify ToolCallResult
	if tcr, ok := msg2.Parts[4].(*ToolCallResult); !ok {
		t.Errorf("ToolCallResult type mismatch: got %T", msg2.Parts[4])
	} else {
		if tcr.ToolCallId != "call_123" {
			t.Errorf("ToolCallResult.ToolCallId mismatch: got %s, want %s", tcr.ToolCallId, "call_123")
		}
		if tcr.Name != "test_tool" {
			t.Errorf("ToolCallResult.Name mismatch: got %s, want %s", tcr.Name, "test_tool")
		}
		if !reflect.DeepEqual(tcr.Result, map[string]interface{}{"output": "success"}) {
			t.Errorf("ToolCallResult.Result mismatch: got %v", tcr.Result)
		}
	}
}

func TestJsonCodec_BinaryPartWithNilFields(t *testing.T) {
	codec := NewJsonCodec()
	
	msg := &Message{
		MessageId: "msg-1",
		Creator:   MessageCreator{Role: MessageRoleUser},
		Model:     ModelId{Provider: "test", ID: "gpt-test"},
		Timestamp: time.Now().UTC().Truncate(time.Second),
		Parts: []Part{
			&BinaryPart{
				Name:          nil, // Test nil name
				URL:           nil, // Test nil URL
				MIMEType:      "application/octet-stream",
				Content:       nil, // Test nil content
				ContentLength: 0,
			},
		},
	}

	data, err := codec.Encode(msg)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	msg2, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if bp, ok := msg2.Parts[0].(*BinaryPart); !ok {
		t.Errorf("BinaryPart type mismatch: got %T", msg2.Parts[0])
	} else {
		if bp.Name != nil {
			t.Errorf("BinaryPart.Name should be nil, got %v", bp.Name)
		}
		if bp.URL != nil {
			t.Errorf("BinaryPart.URL should be nil, got %v", bp.URL)
		}
		if bp.MIMEType != "application/octet-stream" {
			t.Errorf("BinaryPart.MIMEType mismatch: got %s", bp.MIMEType)
		}
		if bp.Content != nil {
			t.Errorf("BinaryPart.Content should be nil, got %v", bp.Content)
		}
	}
}

// UnsupportedPart is a test part type that's not supported by the codec
type UnsupportedPart struct{}

func (u *UnsupportedPart) Type() PartType { return PartType("unsupported") }

func TestJsonCodec_EncodeErrors(t *testing.T) {
	codec := &JsonCodec{}

	// Test nil message
	_, err := codec.Encode(nil)
	if err == nil {
		t.Error("Expected error for nil message")
	}

	// Test unsupported part type
	msg := &Message{
		MessageId: "msg-1",
		Creator:   MessageCreator{Role: MessageRoleUser},
		Model:     ModelId{Provider: "test", ID: "gpt-test"},
		Timestamp: time.Now(),
		Parts:     []Part{&UnsupportedPart{}},
	}

	_, err = codec.Encode(msg)
	if err == nil {
		t.Error("Expected error for unsupported part type")
	}
}

func TestJsonCodec_DecodeErrors(t *testing.T) {
	codec := &JsonCodec{}

	// Test empty data
	_, err := codec.Decode([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Test invalid JSON
	_, err = codec.Decode([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test invalid part content format
	invalidJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "text",
				"content": "not a map"
			}
		]
	}`
	_, err = codec.Decode([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid part content format")
	}

	// Test unsupported part type
	unsupportedJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "unsupported",
				"content": {}
			}
		]
	}`
	_, err = codec.Decode([]byte(unsupportedJSON))
	if err == nil {
		t.Error("Expected error for unsupported part type")
	}

	// Test invalid text part content
	invalidTextJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "text",
				"content": {"text": 123}
			}
		]
	}`
	_, err = codec.Decode([]byte(invalidTextJSON))
	if err == nil {
		t.Error("Expected error for invalid text part content")
	}

	// Test invalid data part content
	invalidDataJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "data",
				"content": {"data": "not a map"}
			}
		]
	}`
	_, err = codec.Decode([]byte(invalidDataJSON))
	if err == nil {
		t.Error("Expected error for invalid data part content")
	}
}

func TestJsonCodec_DecodeBinaryPartVariations(t *testing.T) {
	codec := &JsonCodec{}

	// Test binary part with array content (numbers)
	arrayContentJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "binary",
				"content": {
					"mime_type": "application/octet-stream",
					"content": [72, 101, 108, 108, 111],
					"content_length": 5
				}
			}
		]
	}`
	
	msg, err := codec.Decode([]byte(arrayContentJSON))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	bp, ok := msg.Parts[0].(*BinaryPart)
	if !ok {
		t.Fatal("Expected BinaryPart")
	}
	
	expected := []byte("Hello")
	if !reflect.DeepEqual(bp.Content, expected) {
		t.Errorf("BinaryPart.Content mismatch: got %v, want %v", bp.Content, expected)
	}

	// Test binary part with base64 string content
	base64ContentJSON := `{
		"message_id": "msg-1",
		"creator": {"role": "user"},
		"model": {"provider": "test", "id": "gpt-test"},
		"timestamp": "2023-01-01T00:00:00Z",
		"parts": [
			{
				"type": "binary",
				"content": {
					"mime_type": "application/octet-stream",
					"content": "SGVsbG8=",
					"content_length": 5
				}
			}
		]
	}`
	
	msg, err = codec.Decode([]byte(base64ContentJSON))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	bp, ok = msg.Parts[0].(*BinaryPart)
	if !ok {
		t.Fatal("Expected BinaryPart")
	}
	
	if !reflect.DeepEqual(bp.Content, expected) {
		t.Errorf("BinaryPart.Content mismatch: got %v, want %v", bp.Content, expected)
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
