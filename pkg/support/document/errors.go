package document

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeChunkingFailed = errors.ErrorCode{
		Code:           30100,
		Name:           "ChunkingFailed",
		DefaultMessage: "Failed to chunking",
	}
)
