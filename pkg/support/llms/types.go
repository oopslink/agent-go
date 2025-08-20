// Package provider provides AI model provider implementations for the agent-go framework.
// This file contains common types used across the provider package.

package llms

import (
	"encoding/json"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"k8s.io/klog/v2"
	"reflect"
	"strings"
)

// FloatVector represents a vector of floating-point numbers.
// This type is commonly used for embedding vectors in AI applications.
type FloatVector []float64

// UsageMetadata contains detailed information about token usage for AI model responses.
// This metadata helps track costs and performance of AI operations.
type UsageMetadata struct {
	InputTokens         int64 // Number of input tokens used in the request
	OutputTokens        int64 // Number of output tokens generated in the response
	CacheCreationTokens int64 // Number of tokens used for cache creation (if applicable)
	CacheReadTokens     int64 // Number of tokens read from cache (if applicable)
}

func (u *UsageMetadata) AsMap() map[string]float64 {
	return map[string]float64{
		"input_tokens":          float64(u.InputTokens),
		"output_tokens":         float64(u.OutputTokens),
		"cache_creation_tokens": float64(u.CacheCreationTokens),
		"cache_read_tokens":     float64(u.CacheReadTokens),
	}
}

// Interface compliance checks
var _ Part = &ToolCall{}
var _ Part = &ToolCallResult{}

// ToolDescriptor describes a tool that can be called by an AI agent.
// It contains the tool's name, description, and parameter schema.
type ToolDescriptor struct {
	Name        string  `json:"name,omitempty"`        // The name of the tool
	Description string  `json:"description,omitempty"` // Human-readable description of the tool
	Parameters  *Schema `json:"parameters,omitempty"`  // JSON schema for the tool's parameters
}

func (t *ToolDescriptor) DeepCopy() *ToolDescriptor {
	if t == nil {
		return nil
	}
	return &ToolDescriptor{
		Name:        t.Name,
		Description: t.Description,
		Parameters:  t.Parameters, // Assuming Schema is immutable or deep copied elsewhere
	}
}

func (t *ToolDescriptor) MarshalJson() string {
	data, err := json.Marshal(t)
	if err != nil {
		klog.Errorf("failed to marshal tool descriptor: %v", err)
		return ""
	}
	return string(data)
}

// ToolCall represents a request to execute a tool.
// It contains the tool call ID, tool name, and arguments to pass to the tool.
type ToolCall struct {
	ToolCallId string         `json:"id,omitempty"`        // Unique identifier for this tool call
	Name       string         `json:"name,omitempty"`      // Name of the tool to call
	Arguments  map[string]any `json:"arguments,omitempty"` // Arguments to pass to the tool
}

// Type returns the part type as PartTypeToolCall.
func (t *ToolCall) Type() PartType {
	return PartTypeToolCall
}

// MarshalJson converts the tool call arguments to a JSON string.
// Returns an empty string if marshaling fails.
func (t *ToolCall) MarshalJson() string {
	data, _ := json.Marshal(t.Arguments)
	return string(data)
}

// ToolCallResult represents the result of a tool execution.
// It contains the tool call ID, tool name, and the result data.
type ToolCallResult struct {
	ToolCallId string         `json:"id,omitempty"`     // Unique identifier for the tool call this result belongs to
	Name       string         `json:"name,omitempty"`   // Name of the tool that was called
	Result     map[string]any `json:"result,omitempty"` // The result data from the tool execution
}

// Type returns the part type as PartTypeToolCallResult.
func (t *ToolCallResult) Type() PartType {
	return PartTypeToolCallResult
}

// MarshalJson converts the tool call result to a JSON string.
// Returns an empty string if marshaling fails.
func (t *ToolCallResult) MarshalJson() string {
	data, _ := json.Marshal(t.Result)
	return string(data)
}

// Schema represents a JSON schema for validating data structures.
// It can describe objects, arrays, and primitive llms.
type Schema struct {
	Type        SchemaType         `json:"type,omitempty"`        // The type of the schema (object, array, string, etc.)
	Properties  map[string]*Schema `json:"properties,omitempty"`  // For objects: property schemas
	Items       *Schema            `json:"items,omitempty"`       // For arrays: schema of array items
	Description string             `json:"description,omitempty"` // Human-readable description
	Required    []string           `json:"required,omitempty"`    // For objects: list of required properties
}

// ToRawSchema converts the schema to a json.RawMessage.
// This is useful for embedding schemas in JSON responses.
func (s *Schema) ToRawSchema() (json.RawMessage, error) {
	jsonSchema, err := json.Marshal(s)
	if err != nil {
		return nil, errors.Errorf(ErrorCodeInvalidSchema,
			"converting tool schema to json: %s", err.Error())
	}
	var rawSchema json.RawMessage
	if err = json.Unmarshal(jsonSchema, &rawSchema); err != nil {
		return nil, errors.Errorf(ErrorCodeInvalidSchema,
			"converting tool schema to json.RawMessage: %s", err.Error())
	}
	return rawSchema, nil
}

// SchemaType defines the type of a JSON schema.
type SchemaType string

const (
	TypeObject SchemaType = "object" // Object type with properties
	TypeArray  SchemaType = "array"  // Array type with items

	TypeString  SchemaType = "string"  // String type
	TypeBoolean SchemaType = "boolean" // Boolean type
	TypeNumber  SchemaType = "number"  // Number type (float)
	TypeInteger SchemaType = "integer" // Integer type
)

// BuildSchemaFor generates a JSON schema for a Go type using reflection.
// It recursively builds schemas for structs, arrays, and primitive llms.
func BuildSchemaFor(t reflect.Type) *Schema {
	out := &Schema{}

	switch t.Kind() {
	case reflect.String:
		out.Type = TypeString
	case reflect.Bool:
		out.Type = TypeBoolean
	case reflect.Int:
		out.Type = TypeInteger
	case reflect.Struct:
		out.Type = TypeObject
		out.Properties = make(map[string]*Schema)
		numFields := t.NumField()
		required := []string{}
		for i := 0; i < numFields; i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				continue
			}
			if strings.HasSuffix(jsonTag, ",omitempty") {
				jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			} else {
				required = append(required, jsonTag)
			}

			fieldType := field.Type

			fieldSchema := BuildSchemaFor(fieldType)
			out.Properties[jsonTag] = fieldSchema
		}

		if len(required) != 0 {
			out.Required = required
		}
	case reflect.Slice:
		out.Type = TypeArray
		out.Items = BuildSchemaFor(t.Elem())
	default:
		klog.Fatalf("unhandled kind %v", t.Kind())
	}

	return out
}

func MakeSystemInstruction(systemPrompt string, messages []*Message) string {
	parts := []string{systemPrompt}
	for _, msg := range messages {
		if msg.Creator.Role == MessageRoleSystem {
			for _, p := range msg.Parts {
				if t, ok := p.(*TextPart); ok {
					parts = append(parts, t.Text)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}
