package vectordb

import (
	"context"
	"errors"

	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
)

// VectorDB defines the interface for vector database operations.
// It provides methods for adding documents and searching through them using vector similarity.
type VectorDB interface {
	// AddDocuments adds a collection of documents to the vector database.
	// It returns the document IDs of the added documents and any error that occurred.
	AddDocuments(ctx context.Context, documents []*document.Document, opts ...InsertOption) ([]document.DocumentId, error)

	UpdateDocuments(ctx context.Context, documents []*document.Document, opts ...UpdateOption) error

	// Search performs a similarity search in the vector database using the provided query.
	// It returns the top 'maxDocuments' most similar documents with their similarity scores.
	Search(ctx context.Context, query string, maxDocuments int, opts ...SearchOption) ([]*ScoredDocument, error)

	Get(ctx context.Context, documentId document.DocumentId) (*document.Document, error)

	Delete(ctx context.Context, documentId document.DocumentId) error
}

// Retriever is a vector database that can retrieve documents based on a query.
type Retriever interface {
	// Search performs a similarity search in the vector database using the provided query.
	// It returns the top 'maxDocuments' most similar documents with their similarity scores.
	Search(ctx context.Context, query string) ([]*ScoredDocument, error)
}

// ScoredDocument represents a document with its similarity score from a search operation.
type ScoredDocument struct {
	document.Document

	// Similarity score between the query and this document.
	// The score is a value between 0 and 1, where 1 means the document is exactly the same as the query.
	Score float32
}

// InsertOptions defines the interface for insert operation configuration.
type InsertOptions interface {
	GetCollection() string
	GetEmbedder() embedder.Embedder
}

// UpdateOptions defines the interface for update operation configuration.
type UpdateOptions interface {
	GetCollection() string
	GetEmbedder() embedder.Embedder
}

// SearchOptions defines the interface for search operation configuration.
type SearchOptions interface {
	GetCollection() string
	GetScoreThreshold() float32
	GetFilters() any
	GetEmbedder() embedder.Embedder
}

// InsertOption is a function that configures vector database insert operation behavior.
type InsertOption func(InsertOptions)

// UpdateOption is a function that configures vector database update operation behavior.
type UpdateOption func(UpdateOptions)

// SearchOption is a function that configures vector database search operation behavior.
type SearchOption func(SearchOptions)

// Standard Insert Options

// WithInsertCollection sets the collection name for the insert operation.
func WithInsertCollection(collection string) InsertOption {
	return func(o InsertOptions) {
		if setter, ok := o.(interface{ SetCollection(string) }); ok {
			setter.SetCollection(collection)
		}
	}
}

// WithInsertEmbedder sets the embedder to use for generating vector embeddings during insert.
func WithInsertEmbedder(emb embedder.Embedder) InsertOption {
	return func(o InsertOptions) {
		if setter, ok := o.(interface{ SetEmbedder(embedder.Embedder) }); ok {
			setter.SetEmbedder(emb)
		}
	}
}

// Standard Search Options

// WithSearchCollection sets the collection name for the search operation.
func WithSearchCollection(collection string) SearchOption {
	return func(o SearchOptions) {
		if setter, ok := o.(interface{ SetCollection(string) }); ok {
			setter.SetCollection(collection)
		}
	}
}

// WithScoreThreshold sets the minimum similarity score threshold for search results.
// Only documents with scores above this threshold will be returned.
func WithScoreThreshold(threshold float32) SearchOption {
	return func(o SearchOptions) {
		if setter, ok := o.(interface{ SetScoreThreshold(float32) }); ok {
			setter.SetScoreThreshold(threshold)
		}
	}
}

// WithFilters sets additional filters to apply during search operations.
// The filters can be used to narrow down search results based on metadata or other criteria.
func WithFilters(filters any) SearchOption {
	return func(o SearchOptions) {
		if setter, ok := o.(interface{ SetFilters(any) }); ok {
			setter.SetFilters(filters)
		}
	}
}

// WithSearchEmbedder sets the embedder to use for generating vector embeddings during search.
func WithSearchEmbedder(emb embedder.Embedder) SearchOption {
	return func(o SearchOptions) {
		if setter, ok := o.(interface{ SetEmbedder(embedder.Embedder) }); ok {
			setter.SetEmbedder(emb)
		}
	}
}

// NewRetriever creates a new Retriever instance.
func NewRetriever(vectorDB VectorDB, maxDocuments int, options ...SearchOption) (Retriever, error) {
	if vectorDB == nil {
		return nil, errors.New("vectorDB is nil")
	}
	if maxDocuments <= 0 {
		return nil, errors.New("maxDocuments must be greater than 0")
	}
	return &retriever{
		vectorDB:     vectorDB,
		maxDocuments: maxDocuments,
		options:      options,
	}, nil
}

var _ Retriever = &retriever{}

// retriever is an implementation of the Retriever interface.
type retriever struct {
	vectorDB     VectorDB
	maxDocuments int
	options      []SearchOption
}

func (r *retriever) Search(ctx context.Context, query string) ([]*ScoredDocument, error) {
	return r.vectorDB.Search(ctx, query, r.maxDocuments, r.options...)
}
