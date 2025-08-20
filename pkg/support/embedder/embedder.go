package embedder

import (
	"context"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

type FloatVector llms.FloatVector

type Embedder interface {
	// Embed takes a string of text and returns its embedding representation.
	Embed(ctx context.Context, texts []string) ([]FloatVector, error)
}
