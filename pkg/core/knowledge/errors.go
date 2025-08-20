package knowledge

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeNoKnowledgeBaseFound = errors.ErrorCode{
		Code:           20100,
		Name:           "NoKnowledgeBaseFound",
		DefaultMessage: "Failed to found knowledge base",
	}
)
