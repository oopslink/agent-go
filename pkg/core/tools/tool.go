package tools

import (
	"context"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

type Tool interface {
	Descriptor() *llms.ToolDescriptor
	Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error)
}

func WithExtraInstruction(tool Tool, extraInstruction string, replace bool) Tool {
	if tool == nil {
		return nil
	}
	return &withExtraInstruction{
		tool:             tool,
		extraInstruction: extraInstruction,
		replace:          replace,
	}
}

var _ Tool = &withExtraInstruction{}

type withExtraInstruction struct {
	tool             Tool
	extraInstruction string
	replace          bool
}

func (w *withExtraInstruction) Descriptor() *llms.ToolDescriptor {
	descriptor := w.tool.Descriptor()
	if descriptor == nil {
		return nil
	}
	descriptor = descriptor.DeepCopy()
	if w.replace {
		descriptor.Description = w.extraInstruction
	} else {
		descriptor.Description += "\n" + w.extraInstruction
	}
	return descriptor
}

func (w *withExtraInstruction) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	return w.tool.Call(ctx, params)
}
