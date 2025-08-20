package embedder

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeEmbeddingFailed = errors.ErrorCode{
		Code:           30400,
		Name:           "EmbeddingFailed",
		DefaultMessage: "Failed to embedding",
	}
)
