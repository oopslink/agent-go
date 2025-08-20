package embedder

import (
	"context"
	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

var _ Embedder = &LlmEmbedder{}

type LlmEmbedder struct {
	embedderProvider llms.EmbedderProvider
}

func (l *LlmEmbedder) Embed(ctx context.Context, texts []string) ([]FloatVector, error) {
	response, err := l.embedderProvider.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, errors.Wrap(ErrorCodeEmbeddingFailed, err)
	}
	var vectors []FloatVector
	for idx := range response.Vectors {
		vector := FloatVector(response.Vectors[idx])
		vectors = append(vectors, vector)
	}
	return vectors, nil
}
