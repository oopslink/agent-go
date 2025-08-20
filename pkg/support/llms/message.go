// Package types provides core data structures and types for the agent-go framework.
// It defines message structures, part types, and builders for creating structured
// communication between AI agents and external systems.
package llms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MessageRole defines the role of a message participant in a conversation.
// It can be system, user, assistant, or tool.
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"    // System messages provide instructions or context
	MessageRoleUser      MessageRole = "user"      // User messages represent human input
	MessageRoleAssistant MessageRole = "assistant" // Assistant messages represent AI responses
	MessageRoleTool      MessageRole = "tool"      // Tool messages represent tool execution results
)

// FinishReason indicates why a conversation or response generation ended.
type FinishReason string

const (
	FinishReasonNormalEnd FinishReason = "normal_end" // Normal completion
	FinishReasonMaxTokens FinishReason = "max_tokens" // Reached token limit
	FinishReasonToolUse   FinishReason = "tool_use"   // Tool was called
	FinishReasonCanceled  FinishReason = "canceled"   // Operation was canceled
	FinishReasonError     FinishReason = "error"      // An error occurred
	FinishReasonDenied    FinishReason = "denied"     // Request was denied
	FinishReasonUnknown   FinishReason = "unknown"    // Unknown reason
)

// ReasoningEffort indicates the level of reasoning effort required or used.
type ReasoningEffort string

const (
	ReasoningEffortLow    ReasoningEffort = "low"    // Low reasoning effort
	ReasoningEffortMedium ReasoningEffort = "medium" // Medium reasoning effort
	ReasoningEffortHigh   ReasoningEffort = "high"   // High reasoning effort
)

// MessageCreator represents the creator of a message with role and optional name.
type MessageCreator struct {
	Role MessageRole // The role of the message creator
	Name *string     // Optional name identifier for the creator
}

func (c *MessageCreator) String() string {
	var items []string
	items = append(items, string(c.Role))
	if c.Name != nil {
		items = append(items, fmt.Sprintf("(%s)", *c.Name))
	}
	return strings.Join(items, " ")
}

// NewSystemMessage creates a new message for instructions.
// The message is automatically timestamped with the current time.
func NewSystemMessage(content string) *Message {
	return &Message{
		Creator: MessageCreator{
			Role: MessageRoleSystem,
		},
		Parts: []Part{
			NewTextPartBuilder().Text(content).Build(),
		},
		Timestamp: time.Now(),
	}
}

// NewUserMessage creates a new message from a user with the given content.
// The message is automatically timestamped with the current time.
func NewUserMessage(content string) *Message {
	return &Message{
		Creator: MessageCreator{
			Role: MessageRoleUser,
		},
		Parts: []Part{
			NewTextPartBuilder().Text(content).Build(),
		},
		Timestamp: time.Now(),
	}
}

// NewAssistantMessage creates a new message from an assistant with the given content.
// The message is automatically timestamped with the current time.
func NewAssistantMessage(
	messageId string, modelId ModelId, content string, toolCalls ...*ToolCall) *Message {
	var parts []Part
	if len(content) > 0 {
		parts = append(parts, NewTextPartBuilder().Text(content).Build())
	}
	if len(toolCalls) > 0 {
		for _, toolCall := range toolCalls {
			parts = append(parts, toolCall)
		}
	}

	if len(parts) == 0 {
		return nil
	}

	return &Message{
		MessageId: messageId,
		Creator: MessageCreator{
			Role: MessageRoleAssistant,
		},
		Model:     modelId,
		Parts:     parts,
		Timestamp: time.Now(),
	}
}

// NewToolCallResultMessage creates a new message from a tool with the given result.
// The message is automatically timestamped with the current time.
func NewToolCallResultMessage(result *ToolCallResult, timestamp time.Time) *Message {
	return &Message{
		Creator: MessageCreator{
			Role: MessageRoleTool,
		},
		Parts:     []Part{result},
		Timestamp: timestamp,
	}
}

// Message represents a single message in a conversation.
// It contains metadata about the message and its content parts.
type Message struct {
	MessageId string         // Unique identifier for the message
	Creator   MessageCreator // Information about who created the message
	Model     ModelId        // The model used to generate the message
	Parts     []Part         // Content parts of the message
	Timestamp time.Time      // When the message was created
}

// PartType defines the type of content part in a message.
type PartType string

const (
	PartTypeText           PartType = "text"             // Plain text content
	PartTypeData           PartType = "data"             // Structured data content
	PartTypeBinary         PartType = "binary"           // Binary file content
	PartTypeToolCall       PartType = "tool_call"        // Tool call request
	PartTypeToolCallResult PartType = "tool_call_result" // Tool call result
)

// Part is an interface that represents a content part in a message.
// All part types must implement the Type() method.
type Part interface {
	Type() PartType // Returns the type of this part
}

// Interface compliance checks
var _ Part = &TextPart{}
var _ Part = &DataPart{}
var _ Part = &BinaryPart{}

// TextPart represents plain text content in a message.
type TextPart struct {
	Text string // The text content
}

// Type returns the part type as PartTypeText.
func (p *TextPart) Type() PartType {
	return PartTypeText
}

// DataPart represents structured data content in a message.
type DataPart struct {
	Data map[string]any // The structured data as key-value pairs
}

// Type returns the part type as PartTypeData.
func (p *DataPart) Type() PartType {
	return PartTypeData
}

// MarshalJson converts the data to a JSON string.
// Returns an empty string if marshaling fails.
func (p *DataPart) MarshalJson() string {
	data, _ := json.Marshal(p.Data)
	return string(data)
}

// BinaryPart represents binary file content in a message.
type BinaryPart struct {
	Name     *string // Optional file name
	URL      *string // Optional file URL
	MIMEType string  // MIME type of the file
	Content  []byte  // Binary content of the file
	// ContentLength >= 0: means Content is loaded (and it can be empty content)
	// ContentLength < 0:  means Content is not loaded yet
	ContentLength int64 // Length of the content in bytes
}

// Type returns the part type as PartTypeBinary.
func (p *BinaryPart) Type() PartType {
	return PartTypeBinary
}

// MarshalBase64 converts the binary content to a base64-encoded string.
func (p *BinaryPart) MarshalBase64() string {
	return base64.StdEncoding.EncodeToString(p.Content)
}

// IsImagePart checks if the binary part is an image file.
// Returns the image format and true if it's an image, empty string and false otherwise.
func IsImagePart(p *BinaryPart) (format string, is bool) {
	if strings.HasPrefix(p.MIMEType, "image/") {
		return p.MIMEType[6:], true
	}
	return "", false
}

// IsAudioPart checks if the binary part is an audio file.
// Returns the audio format and true if it's audio, empty string and false otherwise.
func IsAudioPart(p *BinaryPart) (format string, is bool) {
	if strings.HasPrefix(p.MIMEType, "audio/") {
		return p.MIMEType[6:], true
	}
	return "", false
}

func IsPlainTextPart(p *BinaryPart) bool {
	if strings.HasSuffix(p.MIMEType, "/txt") {
		return true
	}
	if strings.HasSuffix(p.MIMEType, "/text") {
		return true
	}
	return false
}

func IsPDFPart(p *BinaryPart) bool {
	if strings.HasSuffix(p.MIMEType, "/pdf") {
		return true
	}
	return false
}

// TextPartBuilder provides a fluent interface for building TextPart instances
type TextPartBuilder struct {
	text string // The text content to be built
}

// NewTextPartBuilder creates a new TextPartBuilder
func NewTextPartBuilder() *TextPartBuilder {
	return &TextPartBuilder{}
}

// Text sets the text content for the TextPart
func (b *TextPartBuilder) Text(text string) *TextPartBuilder {
	b.text = text
	return b
}

// Build creates and returns a new TextPart instance
func (b *TextPartBuilder) Build() *TextPart {
	return &TextPart{
		Text: b.text,
	}
}

// DataPartBuilder provides a fluent interface for building DataPart instances
type DataPartBuilder struct {
	data map[string]any // The data map to be built
}

// NewDataPartBuilder creates a new DataPartBuilder
func NewDataPartBuilder() *DataPartBuilder {
	return &DataPartBuilder{
		data: make(map[string]any),
	}
}

// Add adds a key-value pair to the data map
func (b *DataPartBuilder) Add(key string, value any) *DataPartBuilder {
	b.data[key] = value
	return b
}

// SetData sets the entire data map
func (b *DataPartBuilder) SetData(data map[string]any) *DataPartBuilder {
	b.data = data
	return b
}

// Build creates and returns a new DataPart instance
func (b *DataPartBuilder) Build() *DataPart {
	return &DataPart{
		Data: b.data,
	}
}

// BinaryPartBuilder provides a fluent interface for building BinaryPart instances
type BinaryPartBuilder struct {
	name          *string // Optional file name
	url           *string // Optional file URL
	mimeType      string  // MIME type of the file
	content       []byte  // Binary content
	contentLength int64   // Content length in bytes
}

// NewBinaryPartBuilder creates a new BinaryPartBuilder
func NewBinaryPartBuilder() *BinaryPartBuilder {
	return &BinaryPartBuilder{
		contentLength: -1, // Default to not loaded
	}
}

// Name sets the file name
func (b *BinaryPartBuilder) Name(name string) *BinaryPartBuilder {
	b.name = &name
	return b
}

// URL sets the file URL
func (b *BinaryPartBuilder) URL(url string) *BinaryPartBuilder {
	b.url = &url
	return b
}

// MIMEType sets the MIME type of the file
func (b *BinaryPartBuilder) MIMEType(mimeType string) *BinaryPartBuilder {
	b.mimeType = mimeType
	return b
}

// Content sets the file content and automatically sets contentLength
func (b *BinaryPartBuilder) Content(content []byte) *BinaryPartBuilder {
	b.content = content
	b.contentLength = int64(len(content))
	return b
}

// ContentLength sets the content length explicitly
func (b *BinaryPartBuilder) ContentLength(length int64) *BinaryPartBuilder {
	b.contentLength = length
	return b
}

// Build creates and returns a new BinaryPart instance
func (b *BinaryPartBuilder) Build() *BinaryPart {
	return &BinaryPart{
		Name:          b.name,
		URL:           b.url,
		MIMEType:      b.mimeType,
		Content:       b.content,
		ContentLength: b.contentLength,
	}
}
