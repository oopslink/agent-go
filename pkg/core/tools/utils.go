package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func OfTools(tools ...Tool) *ToolCollection {
	return &ToolCollection{
		Tools: tools,
	}
}

type ToolCollection struct {
	Tools []Tool
}

func (tc *ToolCollection) AddTools(tools ...Tool) {
	tc.Tools = append(tc.Tools, tools...)
}

func (tc *ToolCollection) ContainsTool(toolName string) bool {
	tool := tc.findTool(toolName)
	return tool != nil
}

func (tc *ToolCollection) Descriptors() []*llms.ToolDescriptor {
	descriptors := make([]*llms.ToolDescriptor, len(tc.Tools))
	for i, tool := range tc.Tools {
		descriptors[i] = tool.Descriptor()
	}
	return descriptors
}

func (tc *ToolCollection) Call(ctx context.Context, toolCall *llms.ToolCall) (*llms.ToolCallResult, error) {
	tool := tc.findTool(toolCall.Name)
	if tool == nil {
		return nil, errors.Permanent(errors.Errorf(ErrorCodeToolNotFound, "tool %s not found", toolCall.Name))
	}
	return tool.Call(ctx, toolCall)
}

func (tc *ToolCollection) findTool(toolName string) Tool {
	for _, tool := range tc.Tools {
		descriptor := tool.Descriptor()
		if descriptor == nil {
			continue // Skip tools without a valid descriptor
		}
		if descriptor.Name == toolName {
			return tool
		}
	}
	return nil
}

func MakeToolInstructions(tools ...*llms.ToolDescriptor) string {
	if len(tools) == 0 {
		return ""
	}
	instructions := []string{
		"# Available tools:",
		"* ** Important **: if decide to use a tool, ensure it is exists in the tool list below.",
		"<tools>",
	}
	for _, descriptor := range tools {
		instructions = append(instructions,
			fmt.Sprintf(`  <tool>
    <name>%s</name>
    <description>%s</description>
  </tool>`, descriptor.Name, descriptor.Description))
	}
	instructions = append(instructions, "</tools>")
	return strings.Join(instructions, "\n")
}
