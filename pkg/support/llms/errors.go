package llms

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeModelAlreadyExists = errors.ErrorCode{
		Code:           30700,
		Name:           "ModelAlreadyExists ",
		DefaultMessage: "Model already exists",
	}
	ErrorCodeChatProviderAlreadyExists = errors.ErrorCode{
		Code:           30701,
		Name:           "ChatProviderAlreadyExists ",
		DefaultMessage: "Chat provider already exists",
	}
	ErrorCodeChatProviderNotFound = errors.ErrorCode{
		Code:           30702,
		Name:           "ChatProviderNotFound ",
		DefaultMessage: "Chat provider not found",
	}
	ErrorCodeCreateChatProviderFailed = errors.ErrorCode{
		Code:           30703,
		Name:           "CreateChatProviderFailed ",
		DefaultMessage: "Failed to create chat provider",
	}
	ErrorCodeEmbedderProviderAlreadyExists = errors.ErrorCode{
		Code:           30704,
		Name:           "EmbedderProviderAlreadyExists ",
		DefaultMessage: "Embedder provider already exists",
	}
	ErrorCodeEmbedderProviderNotFound = errors.ErrorCode{
		Code:           30705,
		Name:           "EmbedderProviderNotFound ",
		DefaultMessage: "Embedder provider not found",
	}
	ErrorCodeCreateEmbedderProviderFailed = errors.ErrorCode{
		Code:           30706,
		Name:           "CreateEmbedderProviderFailed ",
		DefaultMessage: "Failed to create embedder provider",
	}
	ErrorCodeInvalidSchema = errors.ErrorCode{
		Code:           30707,
		Name:           "InvalidSchema ",
		DefaultMessage: "Schema invalid",
	}
	ErrorCodeModelFeatureNotMatched = errors.ErrorCode{
		Code:           30708,
		Name:           "ModelFeatureNotMatched ",
		DefaultMessage: "Feature of model not matched",
	}
	ErrorCodeChatSessionFailed = errors.ErrorCode{
		Code:           30709,
		Name:           "ChatSessionFailed ",
		DefaultMessage: "Chat session failed",
	}
	ErrorCodeEmbeddingSessionFailed = errors.ErrorCode{
		Code:           30710,
		Name:           "EmbeddingSessionFailed ",
		DefaultMessage: "Embedding session failed",
	}
)
