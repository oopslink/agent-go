package knowledge

import (
	"context"
	"errors"

	"github.com/oopslink/agent-go/pkg/support/document"
	"github.com/oopslink/agent-go/pkg/support/embedder"
	"github.com/oopslink/agent-go/pkg/support/vectordb"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrStorageRequired  = errors.New("storage is required")
)

var _ KnowledgeStorage = &VectorDBStorage{}

// VectorDBStorage is the vector database based storage implementation
type VectorDBStorage struct {
	Collection string
	vectorDB   vectordb.VectorDB
	embedder   embedder.Embedder
}

// NewVectorDBStorage creates vector storage
func NewVectorDBStorage(collectionName string,
	vectorDB vectordb.VectorDB, embedder embedder.Embedder) *VectorDBStorage {
	return &VectorDBStorage{
		Collection: collectionName,
		vectorDB:   vectorDB,
		embedder:   embedder,
	}
}

func (v *VectorDBStorage) Add(ctx context.Context, doc *document.Document, opts ...AddOption) error {

	insertOptions := v.makeInsertOptions(opts...)

	// Add to vector database
	_, err := v.vectorDB.AddDocuments(ctx, []*document.Document{doc}, insertOptions...)

	return err
}

func (v *VectorDBStorage) Update(ctx context.Context, id document.DocumentId, doc *document.Document, opts ...UpdateOption) error {
	updateOptions := v.makeUpdateOptions(opts...)
	_, err := v.vectorDB.AddDocuments(ctx, []*document.Document{doc}, updateOptions...)
	return err
}

func (v *VectorDBStorage) Search(ctx context.Context, query string, opts ...SearchOption) ([]*document.Document, error) {
	maxDocuments, options := v.makeSearchOptions(opts...)
	list, err := v.vectorDB.Search(ctx, query, maxDocuments, options...)
	if err != nil {
		return nil, err
	}
	var result []*document.Document
	for idx := range list {
		scoredDoc := list[idx]
		result = append(result, &scoredDoc.Document)
	}
	return result, nil
}

func (v *VectorDBStorage) Get(ctx context.Context, documentId document.DocumentId) (*document.Document, error) {
	return v.vectorDB.Get(ctx, documentId)
}

func (v *VectorDBStorage) Delete(ctx context.Context, documentId document.DocumentId) error {
	return v.vectorDB.Delete(ctx, documentId)
}

func (v *VectorDBStorage) makeSearchOptions(opts ...SearchOption) (int, []vectordb.SearchOption) {
	options := &SearchOptions{
		MaxResults:     10, // default value
		ScoreThreshold: 0.0,
	}

	// Apply knowledge search options
	for _, opt := range opts {
		opt(options)
	}

	// Convert to vectordb search options
	vectordbOpts := []vectordb.SearchOption{vectordb.WithSearchCollection(v.Collection)}

	// Set score threshold if specified
	if options.ScoreThreshold > 0 {
		vectordbOpts = append(vectordbOpts, vectordb.WithScoreThreshold(options.ScoreThreshold))
	}

	// Set filters if specified
	if options.Filters != nil {
		vectordbOpts = append(vectordbOpts, vectordb.WithFilters(options.Filters))
	}

	// Set embedder if specified (use the storage embedder as default)
	if v.embedder != nil {
		vectordbOpts = append(vectordbOpts, vectordb.WithSearchEmbedder(v.embedder))
	}

	return options.MaxResults, vectordbOpts
}

func (v *VectorDBStorage) makeInsertOptions(opts ...AddOption) []vectordb.InsertOption {
	options := &AddOptions{
		Overwrite: false, // default value
	}

	// Apply knowledge add options
	for _, opt := range opts {
		opt(options)
	}

	// Convert to vectordb insert options
	vectordbOpts := []vectordb.InsertOption{vectordb.WithInsertCollection(v.Collection)}

	// Set embedder (use the storage embedder)
	if v.embedder != nil {
		vectordbOpts = append(vectordbOpts, vectordb.WithInsertEmbedder(v.embedder))
	}

	return vectordbOpts
}

func (v *VectorDBStorage) makeUpdateOptions(opts ...UpdateOption) []vectordb.InsertOption {
	options := &UpdateOptions{}

	// Apply knowledge update options
	for _, opt := range opts {
		opt(options)
	}

	// Convert to vectordb insert options (for update, we use insert with overwrite)
	vectordbOpts := []vectordb.InsertOption{vectordb.WithInsertCollection(v.Collection)}

	// Set embedder
	if v.embedder != nil {
		vectordbOpts = append(vectordbOpts, vectordb.WithInsertEmbedder(v.embedder))
	}

	return vectordbOpts
}
