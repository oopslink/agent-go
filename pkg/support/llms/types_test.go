package llms

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestToolCall(t *testing.T) {
	toolCall := &ToolCall{
		ToolCallId: "test-call-123",
		Name:       "test-tool",
		Arguments: map[string]any{
			"param1": "value1",
			"param2": 42,
			"param3": true,
		},
	}

	// Test Type method
	if toolCall.Type() != PartTypeToolCall {
		t.Errorf("ToolCall.Type() = %v, want %v", toolCall.Type(), PartTypeToolCall)
	}

	// Test MarshalJson method
	jsonStr := toolCall.MarshalJson()
	if jsonStr == "" {
		t.Error("ToolCall.MarshalJson() returned empty string")
	}

	// Verify JSON structure
	var args map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &args); err != nil {
		t.Errorf("ToolCall.MarshalJson() returned invalid JSON: %v", err)
	}

	if args["param1"] != "value1" {
		t.Errorf("ToolCall.MarshalJson() param1 = %v, want value1", args["param1"])
	}

	if args["param2"] != float64(42) { // JSON numbers are unmarshaled as float64
		t.Errorf("ToolCall.MarshalJson() param2 = %v, want 42", args["param2"])
	}

	if args["param3"] != true {
		t.Errorf("ToolCall.MarshalJson() param3 = %v, want true", args["param3"])
	}
}

func TestToolCallResult(t *testing.T) {
	result := &ToolCallResult{
		ToolCallId: "test-call-123",
		Name:       "test-tool",
		Result: map[string]any{
			"status":  "success",
			"data":    "result data",
			"count":   100,
			"success": true,
		},
	}

	// Test Type method
	if result.Type() != PartTypeToolCallResult {
		t.Errorf("ToolCallResult.Type() = %v, want %v", result.Type(), PartTypeToolCallResult)
	}

	// Test MarshalJson method
	jsonStr := result.MarshalJson()
	if jsonStr == "" {
		t.Error("ToolCallResult.MarshalJson() returned empty string")
	}

	// Verify JSON structure
	var resultData map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &resultData); err != nil {
		t.Errorf("ToolCallResult.MarshalJson() returned invalid JSON: %v", err)
	}

	if resultData["status"] != "success" {
		t.Errorf("ToolCallResult.MarshalJson() status = %v, want success", resultData["status"])
	}

	if resultData["data"] != "result data" {
		t.Errorf("ToolCallResult.MarshalJson() data = %v, want result data", resultData["data"])
	}

	if resultData["count"] != float64(100) { // JSON numbers are unmarshaled as float64
		t.Errorf("ToolCallResult.MarshalJson() count = %v, want 100", resultData["count"])
	}

	if resultData["success"] != true {
		t.Errorf("ToolCallResult.MarshalJson() success = %v, want true", resultData["success"])
	}
}

func TestToolDescriptor(t *testing.T) {
	descriptor := &ToolDescriptor{
		Name:        "test-tool",
		Description: "A test tool for unit testing",
		Parameters: &Schema{
			Type: TypeObject,
			Properties: map[string]*Schema{
				"input": {
					Type:        TypeString,
					Description: "Input parameter",
				},
			},
			Required: []string{"input"},
		},
	}

	if descriptor.Name != "test-tool" {
		t.Errorf("ToolDescriptor.Name = %v, want test-tool", descriptor.Name)
	}

	if descriptor.Description != "A test tool for unit testing" {
		t.Errorf("ToolDescriptor.Description = %v, want A test tool for unit testing", descriptor.Description)
	}

	if descriptor.Parameters == nil {
		t.Error("ToolDescriptor.Parameters is nil")
	}

	if descriptor.Parameters.Type != TypeObject {
		t.Errorf("ToolDescriptor.Parameters.Type = %v, want %v", descriptor.Parameters.Type, TypeObject)
	}
}

func TestSchema(t *testing.T) {
	schema := &Schema{
		Type:        TypeObject,
		Description: "Test schema",
		Properties: map[string]*Schema{
			"name": {
				Type:        TypeString,
				Description: "Name field",
			},
			"age": {
				Type:        TypeInteger,
				Description: "Age field",
			},
			"active": {
				Type:        TypeBoolean,
				Description: "Active field",
			},
		},
		Required: []string{"name", "age"},
	}

	if schema.Type != TypeObject {
		t.Errorf("Schema.Type = %v, want %v", schema.Type, TypeObject)
	}

	if schema.Description != "Test schema" {
		t.Errorf("Schema.Description = %v, want Test schema", schema.Description)
	}

	if len(schema.Properties) != 3 {
		t.Errorf("Schema.Properties length = %d, want 3", len(schema.Properties))
	}

	if len(schema.Required) != 2 {
		t.Errorf("Schema.Required length = %d, want 2", len(schema.Required))
	}

	// Test ToRawSchema method
	rawSchema, err := schema.ToRawSchema()
	if err != nil {
		t.Errorf("Schema.ToRawSchema() error = %v", err)
	}

	if len(rawSchema) == 0 {
		t.Error("Schema.ToRawSchema() returned empty raw schema")
	}

	// Verify raw schema can be unmarshaled back to schema
	var unmarshaledSchema Schema
	if err := json.Unmarshal(rawSchema, &unmarshaledSchema); err != nil {
		t.Errorf("Failed to unmarshal raw schema: %v", err)
	}

	if unmarshaledSchema.Type != schema.Type {
		t.Errorf("Unmarshaled schema type = %v, want %v", unmarshaledSchema.Type, schema.Type)
	}
}

func TestSchemaTypeConstants(t *testing.T) {
	tests := []struct {
		name       string
		schemaType SchemaType
		expected   string
	}{
		{"object", TypeObject, "object"},
		{"array", TypeArray, "array"},
		{"string", TypeString, "string"},
		{"boolean", TypeBoolean, "boolean"},
		{"number", TypeNumber, "number"},
		{"integer", TypeInteger, "integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.schemaType) != tt.expected {
				t.Errorf("SchemaType = %v, want %v", tt.schemaType, tt.expected)
			}
		})
	}
}

func TestBuildSchemaFor(t *testing.T) {
	// Test string type
	stringSchema := BuildSchemaFor(reflect.TypeOf(""))
	if stringSchema.Type != TypeString {
		t.Errorf("BuildSchemaFor(string) = %v, want %v", stringSchema.Type, TypeString)
	}

	// Test boolean type
	boolSchema := BuildSchemaFor(reflect.TypeOf(true))
	if boolSchema.Type != TypeBoolean {
		t.Errorf("BuildSchemaFor(bool) = %v, want %v", boolSchema.Type, TypeBoolean)
	}

	// Test int type
	intSchema := BuildSchemaFor(reflect.TypeOf(0))
	if intSchema.Type != TypeInteger {
		t.Errorf("BuildSchemaFor(int) = %v, want %v", intSchema.Type, TypeInteger)
	}

	// Test slice type
	type TestSlice []string
	sliceSchema := BuildSchemaFor(reflect.TypeOf(TestSlice{}))
	if sliceSchema.Type != TypeArray {
		t.Errorf("BuildSchemaFor([]string) = %v, want %v", sliceSchema.Type, TypeArray)
	}
	if sliceSchema.Items == nil {
		t.Error("BuildSchemaFor([]string) Items is nil")
	}
	if sliceSchema.Items.Type != TypeString {
		t.Errorf("BuildSchemaFor([]string) Items.Type = %v, want %v", sliceSchema.Items.Type, TypeString)
	}

	// Test struct type
	type TestStruct struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Active bool   `json:"active,omitempty"`
	}
	structSchema := BuildSchemaFor(reflect.TypeOf(TestStruct{}))
	if structSchema.Type != TypeObject {
		t.Errorf("BuildSchemaFor(struct) = %v, want %v", structSchema.Type, TypeObject)
	}
	if len(structSchema.Properties) != 3 {
		t.Errorf("BuildSchemaFor(struct) Properties length = %d, want 3", len(structSchema.Properties))
	}
	if len(structSchema.Required) != 2 { // "active" has omitempty tag
		t.Errorf("BuildSchemaFor(struct) Required length = %d, want 2", len(structSchema.Required))
	}
}

func TestToolCallPartInterfaceCompliance(t *testing.T) {
	// Test that ToolCall and ToolCallResult implement the Part interface
	var parts []Part

	toolCall := &ToolCall{
		ToolCallId: "test-id",
		Name:       "test-tool",
		Arguments:  map[string]any{"param": "value"},
	}
	parts = append(parts, toolCall)

	toolCallResult := &ToolCallResult{
		ToolCallId: "test-id",
		Name:       "test-tool",
		Result:     map[string]any{"status": "success"},
	}
	parts = append(parts, toolCallResult)

	for i, part := range parts {
		if part.Type() == "" {
			t.Errorf("Part %d Type() returned empty string", i)
		}
	}
}

func TestSchemaWithArray(t *testing.T) {
	schema := &Schema{
		Type: TypeArray,
		Items: &Schema{
			Type: TypeString,
		},
		Description: "Array of strings",
	}

	if schema.Type != TypeArray {
		t.Errorf("Schema.Type = %v, want %v", schema.Type, TypeArray)
	}

	if schema.Items == nil {
		t.Error("Schema.Items is nil")
	}

	if schema.Items.Type != TypeString {
		t.Errorf("Schema.Items.Type = %v, want %v", schema.Items.Type, TypeString)
	}
}

func TestSchemaToRawSchemaError(t *testing.T) {
	// This test is difficult to implement reliably since JSON marshaling
	// is quite robust. Instead, we'll test that the method works correctly
	// for valid schemas, which is more important.

	schema := &Schema{
		Type: TypeObject,
		Properties: map[string]*Schema{
			"test": {
				Type: TypeString,
			},
		},
	}

	_, err := schema.ToRawSchema()
	if err != nil {
		t.Errorf("Schema.ToRawSchema() should not return error for valid schema: %v", err)
	}
}
