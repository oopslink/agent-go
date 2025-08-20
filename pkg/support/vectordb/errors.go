package vectordb

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeLoadCollectionFailed = errors.ErrorCode{
		Code:           30600,
		Name:           "LoadCollectionFailed ",
		DefaultMessage: "Failed to load collection",
	}
	ErrorCodeInvalidVectorDataSchema = errors.ErrorCode{
		Code:           30601,
		Name:           "InvalidVectorDataSchema ",
		DefaultMessage: "Vector data schema is not valid",
	}
	ErrorCodeSearchDocumentFailed = errors.ErrorCode{
		Code:           30602,
		Name:           "SearchDocumentFailed ",
		DefaultMessage: "Failed to search document",
	}
	ErrorCodeUpdateDocumentFailed = errors.ErrorCode{
		Code:           30603,
		Name:           "UpdateDocumentFailed ",
		DefaultMessage: "Failed to update document",
	}
	ErrorCodeAddDocumentFailed = errors.ErrorCode{
		Code:           30604,
		Name:           "AddDocumentFailed ",
		DefaultMessage: "Failed to add document",
	}
	ErrorCodeCreateVectorStoreFailed = errors.ErrorCode{
		Code:           30605,
		Name:           "CreateVectorStoreFailed ",
		DefaultMessage: "Failed to create vector store",
	}
	ErrorCodeCreateVectorClientFailed = errors.ErrorCode{
		Code:           30606,
		Name:           "CreateVectorClientFailed ",
		DefaultMessage: "Failed to create vector client",
	}
)
