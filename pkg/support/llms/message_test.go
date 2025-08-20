package llms

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestMessageRole(t *testing.T) {
	tests := []struct {
		name     string
		role     MessageRole
		expected string
	}{
		{"system", MessageRoleSystem, "system"},
		{"user", MessageRoleUser, "user"},
		{"assistant", MessageRoleAssistant, "assistant"},
		{"tool", MessageRoleTool, "tool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("MessageRole = %v, want %v", tt.role, tt.expected)
			}
		})
	}
}

func TestFinishReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   FinishReason
		expected string
	}{
		{"normal_end", FinishReasonNormalEnd, "normal_end"},
		{"max_tokens", FinishReasonMaxTokens, "max_tokens"},
		{"tool_use", FinishReasonToolUse, "tool_use"},
		{"canceled", FinishReasonCanceled, "canceled"},
		{"error", FinishReasonError, "error"},
		{"denied", FinishReasonDenied, "denied"},
		{"unknown", FinishReasonUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.reason) != tt.expected {
				t.Errorf("FinishReason = %v, want %v", tt.reason, tt.expected)
			}
		})
	}
}

func TestReasoningEffort(t *testing.T) {
	tests := []struct {
		name     string
		effort   ReasoningEffort
		expected string
	}{
		{"low", ReasoningEffortLow, "low"},
		{"medium", ReasoningEffortMedium, "medium"},
		{"high", ReasoningEffortHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.effort) != tt.expected {
				t.Errorf("ReasoningEffort = %v, want %v", tt.effort, tt.expected)
			}
		})
	}
}

func TestNewUserMessage(t *testing.T) {
	content := "Hello, world!"
	msg := NewUserMessage(content)

	if msg.Creator.Role != MessageRoleUser {
		t.Errorf("NewUserMessage() creator role = %v, want %v", msg.Creator.Role, MessageRoleUser)
	}

	if len(msg.Parts) != 1 {
		t.Errorf("NewUserMessage() parts count = %d, want 1", len(msg.Parts))
	}

	if textPart, ok := msg.Parts[0].(*TextPart); !ok {
		t.Error("NewUserMessage() first part is not *TextPart")
	} else if textPart.Text != content {
		t.Errorf("NewUserMessage() text content = %v, want %v", textPart.Text, content)
	}

	if msg.Timestamp.IsZero() {
		t.Error("NewUserMessage() timestamp is zero")
	}
}

func TestNewToolMessage(t *testing.T) {
	result := &ToolCallResult{
		ToolCallId: "test-id",
		Name:       "test-tool",
		Result:     map[string]any{"status": "success"},
	}
	msg := NewToolCallResultMessage(result, time.Now())

	if msg.Creator.Role != MessageRoleTool {
		t.Errorf("NewToolMessage() creator role = %v, want %v", msg.Creator.Role, MessageRoleTool)
	}

	if len(msg.Parts) != 1 {
		t.Errorf("NewToolMessage() parts count = %d, want 1", len(msg.Parts))
	}

	if msg.Parts[0] != result {
		t.Error("NewToolMessage() first part is not the provided result")
	}

	if msg.Timestamp.IsZero() {
		t.Error("NewToolMessage() timestamp is zero")
	}
}

func TestTextPart(t *testing.T) {
	text := "Test text content"
	part := &TextPart{Text: text}

	if part.Type() != PartTypeText {
		t.Errorf("TextPart.Type() = %v, want %v", part.Type(), PartTypeText)
	}

	if part.Text != text {
		t.Errorf("TextPart.Text = %v, want %v", part.Text, text)
	}
}

func TestDataPart(t *testing.T) {
	data := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	part := &DataPart{Data: data}

	if part.Type() != PartTypeData {
		t.Errorf("DataPart.Type() = %v, want %v", part.Type(), PartTypeData)
	}

	if len(part.Data) != len(data) {
		t.Errorf("DataPart.Data length = %d, want %d", len(part.Data), len(data))
	}

	// Test MarshalJson
	jsonStr := part.MarshalJson()
	if jsonStr == "" {
		t.Error("DataPart.MarshalJson() returned empty string")
	}
}

func TestBinaryPart(t *testing.T) {
	name := "test.txt"
	url := "https://example.com/test.txt"
	mimeType := "text/plain"
	content := []byte("test content")
	contentLength := int64(len(content))

	part := &BinaryPart{
		Name:          &name,
		URL:           &url,
		MIMEType:      mimeType,
		Content:       content,
		ContentLength: contentLength,
	}

	if part.Type() != PartTypeBinary {
		t.Errorf("BinaryPart.Type() = %v, want %v", part.Type(), PartTypeBinary)
	}

	if *part.Name != name {
		t.Errorf("BinaryPart.Name = %v, want %v", *part.Name, name)
	}

	if *part.URL != url {
		t.Errorf("BinaryPart.URL = %v, want %v", *part.URL, url)
	}

	if part.MIMEType != mimeType {
		t.Errorf("BinaryPart.MIMEType = %v, want %v", part.MIMEType, mimeType)
	}

	if len(part.Content) != len(content) {
		t.Errorf("BinaryPart.Content length = %d, want %d", len(part.Content), len(content))
	}

	if part.ContentLength != contentLength {
		t.Errorf("BinaryPart.ContentLength = %d, want %d", part.ContentLength, contentLength)
	}

	// Test MarshalBase64
	base64Str := part.MarshalBase64()
	expectedBase64 := base64.StdEncoding.EncodeToString(content)
	if base64Str != expectedBase64 {
		t.Errorf("BinaryPart.MarshalBase64() = %v, want %v", base64Str, expectedBase64)
	}
}

func TestIsImagePart(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
		format   string
	}{
		{"png image", "image/png", true, "png"},
		{"jpeg image", "image/jpeg", true, "jpeg"},
		{"gif image", "image/gif", true, "gif"},
		{"text file", "text/plain", false, ""},
		{"audio file", "audio/mp3", false, ""},
		{"empty mime", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &BinaryPart{MIMEType: tt.mimeType}
			format, isImage := IsImagePart(part)

			if isImage != tt.expected {
				t.Errorf("IsImagePart() = %v, want %v", isImage, tt.expected)
			}

			if isImage && format != tt.format {
				t.Errorf("IsImagePart() format = %v, want %v", format, tt.format)
			}
		})
	}
}

func TestIsAudioPart(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
		format   string
	}{
		{"mp3 audio", "audio/mp3", true, "mp3"},
		{"wav audio", "audio/wav", true, "wav"},
		{"ogg audio", "audio/ogg", true, "ogg"},
		{"text file", "text/plain", false, ""},
		{"image file", "image/png", false, ""},
		{"empty mime", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := &BinaryPart{MIMEType: tt.mimeType}
			format, isAudio := IsAudioPart(part)

			if isAudio != tt.expected {
				t.Errorf("IsAudioPart() = %v, want %v", isAudio, tt.expected)
			}

			if isAudio && format != tt.format {
				t.Errorf("IsAudioPart() format = %v, want %v", format, tt.format)
			}
		})
	}
}

func TestTextPartBuilder(t *testing.T) {
	text := "Builder test text"
	builder := NewTextPartBuilder()
	part := builder.Text(text).Build()

	if part.Text != text {
		t.Errorf("TextPartBuilder.Text() = %v, want %v", part.Text, text)
	}

	if part.Type() != PartTypeText {
		t.Errorf("TextPartBuilder.Build() type = %v, want %v", part.Type(), PartTypeText)
	}
}

func TestDataPartBuilder(t *testing.T) {
	builder := NewDataPartBuilder()
	part := builder.
		Add("key1", "value1").
		Add("key2", 42).
		Add("key3", true).
		Build()

	if part.Type() != PartTypeData {
		t.Errorf("DataPartBuilder.Build() type = %v, want %v", part.Type(), PartTypeData)
	}

	if len(part.Data) != 3 {
		t.Errorf("DataPartBuilder.Data length = %d, want 3", len(part.Data))
	}

	if part.Data["key1"] != "value1" {
		t.Errorf("DataPartBuilder.Data[key1] = %v, want value1", part.Data["key1"])
	}

	if part.Data["key2"] != 42 {
		t.Errorf("DataPartBuilder.Data[key2] = %v, want 42", part.Data["key2"])
	}

	if part.Data["key3"] != true {
		t.Errorf("DataPartBuilder.Data[key3] = %v, want true", part.Data["key3"])
	}
}

func TestBinaryPartBuilder(t *testing.T) {
	name := "test.txt"
	url := "https://example.com/test.txt"
	mimeType := "text/plain"
	content := []byte("test content")

	builder := NewBinaryPartBuilder()
	part := builder.
		Name(name).
		URL(url).
		MIMEType(mimeType).
		Content(content).
		Build()

	if part.Type() != PartTypeBinary {
		t.Errorf("BinaryPartBuilder.Build() type = %v, want %v", part.Type(), PartTypeBinary)
	}

	if *part.Name != name {
		t.Errorf("BinaryPartBuilder.Name() = %v, want %v", *part.Name, name)
	}

	if *part.URL != url {
		t.Errorf("BinaryPartBuilder.URL() = %v, want %v", *part.URL, url)
	}

	if part.MIMEType != mimeType {
		t.Errorf("BinaryPartBuilder.MIMEType() = %v, want %v", part.MIMEType, mimeType)
	}

	if len(part.Content) != len(content) {
		t.Errorf("BinaryPartBuilder.Content() length = %d, want %d", len(part.Content), len(content))
	}

	if part.ContentLength != int64(len(content)) {
		t.Errorf("BinaryPartBuilder.ContentLength = %d, want %d", part.ContentLength, len(content))
	}
}

func TestBinaryPartBuilderContentLength(t *testing.T) {
	builder := NewBinaryPartBuilder()
	part := builder.ContentLength(100).Build()

	if part.ContentLength != 100 {
		t.Errorf("BinaryPartBuilder.ContentLength() = %d, want 100", part.ContentLength)
	}
}

func TestPartInterfaceCompliance(t *testing.T) {
	// Test that all part types implement the Part interface
	var parts []Part

	textPart := &TextPart{Text: "test"}
	parts = append(parts, textPart)

	dataPart := &DataPart{Data: map[string]any{"key": "value"}}
	parts = append(parts, dataPart)

	binaryPart := &BinaryPart{MIMEType: "text/plain", Content: []byte("test")}
	parts = append(parts, binaryPart)

	for i, part := range parts {
		if part.Type() == "" {
			t.Errorf("Part %d Type() returned empty string", i)
		}
	}
}
