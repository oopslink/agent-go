package knowledge

import (
	"context"
	"time"

	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/document"
)

// KnowledgeRetriever is the knowledge retrieval interface
type KnowledgeRetriever interface {
	// Search searches for relevant content in the knowledge base
	Search(ctx context.Context, query string, opts ...SearchOption) ([]KnowledgeItem, error)
}

// KnowledgeUpdater is the knowledge update interface
type KnowledgeUpdater interface {
	// AddItem adds a single knowledge base item
	AddItem(ctx context.Context, item KnowledgeItem, opts ...AddOption) error

	// UpdateItem updates an item in the knowledge base
	UpdateItem(ctx context.Context, id document.DocumentId, item KnowledgeItem, opts ...UpdateOption) error
}

// KnowledgeManager is the knowledge management interface
type KnowledgeManager interface {
	// GetMetadata gets knowledge base metadata
	GetMetadata() *KnowledgeBaseMetadata

	// UpdateMetadata updates knowledge base metadata
	UpdateMetadata(metadata *KnowledgeBaseMetadata)

	// GetItem gets a specific knowledge base item by ID
	GetItem(ctx context.Context, id document.DocumentId) (KnowledgeItem, error)

	// DeleteItem deletes a specified knowledge base item
	DeleteItem(ctx context.Context, id document.DocumentId) error
}

// KnowledgeStorage is the knowledge base storage interface, only handles Document
type KnowledgeStorage interface {
	// Add adds document to storage
	Add(ctx context.Context, doc *document.Document, opts ...AddOption) error

	// Update updates document in storage
	Update(ctx context.Context, id document.DocumentId, doc *document.Document, opts ...UpdateOption) error

	// Get gets document
	Get(ctx context.Context, id document.DocumentId) (*document.Document, error)

	// Delete deletes document
	Delete(ctx context.Context, id document.DocumentId) error

	// Search searches documents
	Search(ctx context.Context, query string, opts ...SearchOption) ([]*document.Document, error)
}

// KnowledgeBase is the complete knowledge base interface, combining retrieval, update and management functionality
type KnowledgeBase interface {
	KnowledgeRetriever
	KnowledgeUpdater
	KnowledgeManager
}

// KnowledgeItem is a knowledge base item
type KnowledgeItem interface {
	GetId() document.DocumentId
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time

	// ToDocument converts to Document
	ToDocument() *document.Document
}

// KnowledgeItemFactory is the knowledge item factory interface
type KnowledgeItemFactory interface {
	// FromDocument creates KnowledgeItem from Document
	FromDocument(doc *document.Document) KnowledgeItem
}

func NewKnowledgeBaseMetadata(name, description string, domains []string, tags map[string]string) *KnowledgeBaseMetadata {
	return &KnowledgeBaseMetadata{
		Name:        name,
		Description: description,
		Domains:     domains,
		Tags:        tags,
	}
}

type KnowledgeBaseMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Domains []string          `json:"domains"`
	Tags    map[string]string `json:"tags"`
}

// AddOption is add option
type AddOption func(*AddOptions)

// UpdateOption is update option
type UpdateOption func(*UpdateOptions)

// SearchOption is search option
type SearchOption func(*SearchOptions)

// AddOptions is add configuration
type AddOptions struct {
	Overwrite bool
}

// UpdateOptions is update configuration
type UpdateOptions struct {
}

// SearchOptions is search configuration
type SearchOptions struct {
	MaxResults     int
	ScoreThreshold float32
	Filters        any
}

// GenerateDocumentId generates document ID
func GenerateDocumentId(name string) document.DocumentId {
	if name == "" {
		return document.DocumentId(utils.GenerateUUID())
	}
	return document.DocumentId(name + "_" + utils.GenerateUUID())
}

func WithAddOverwrite(overwrite bool) AddOption {
	return func(opts *AddOptions) {
		opts.Overwrite = overwrite
	}
}

// SearchOption implementations
func WithMaxResults(max int) SearchOption {
	return func(opts *SearchOptions) {
		opts.MaxResults = max
	}
}

func WithScoreThreshold(threshold float32) SearchOption {
	return func(opts *SearchOptions) {
		opts.ScoreThreshold = threshold
	}
}

func WithSearchFilters(filters any) SearchOption {
	return func(opts *SearchOptions) {
		opts.Filters = filters
	}
}
