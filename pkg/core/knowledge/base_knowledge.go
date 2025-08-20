package knowledge

import (
	"context"
	"sync"
	"time"

	"github.com/oopslink/agent-go/pkg/support/document"
)

// NewKnowledgeBase creates a new knowledge base instance
func NewKnowledgeBase(
	storage KnowledgeStorage,
	factory KnowledgeItemFactory,
	metadata *KnowledgeBaseMetadata,
) KnowledgeBase {
	return &knowledgeBase{
		storage:  storage,
		factory:  factory,
		metadata: metadata,
	}
}

// knowledgeBase is the base implementation of knowledge base, combining retrieval, update and management functionality
type knowledgeBase struct {
	mu sync.RWMutex

	storage  KnowledgeStorage
	factory  KnowledgeItemFactory
	metadata *KnowledgeBaseMetadata
}

// Implement KnowledgeRetriever interface
func (kb *knowledgeBase) Search(ctx context.Context, query string, opts ...SearchOption) ([]KnowledgeItem, error) {
	result, err := kb.storage.Search(ctx, query, opts...)
	if err != nil {
		return nil, err
	}
	var items []KnowledgeItem
	for idx := range result {
		doc := result[idx]
		item := kb.factory.FromDocument(doc)
		if item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// Implement KnowledgeUpdater interface
func (kb *knowledgeBase) AddItem(ctx context.Context, item KnowledgeItem, opts ...AddOption) error {
	// Convert to Document
	doc := item.ToDocument()
	// Add to storage
	return kb.storage.Add(ctx, doc, opts...)
}

func (kb *knowledgeBase) UpdateItem(ctx context.Context, id document.DocumentId, item KnowledgeItem, opts ...UpdateOption) error {
	// Convert to Document
	doc := item.ToDocument()

	// Update storage
	return kb.storage.Update(ctx, id, doc, opts...)
}

// Implement KnowledgeManager interface
func (kb *knowledgeBase) GetItem(ctx context.Context, id document.DocumentId) (KnowledgeItem, error) {
	doc, err := kb.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return kb.factory.FromDocument(doc), nil
}

func (kb *knowledgeBase) DeleteItem(ctx context.Context, id document.DocumentId) error {
	return kb.storage.Delete(ctx, id)
}

func (kb *knowledgeBase) GetMetadata() *KnowledgeBaseMetadata {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return kb.metadata
}

func (kb *knowledgeBase) UpdateMetadata(metadata *KnowledgeBaseMetadata) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.metadata = metadata
}

// Private methods
func (kb *knowledgeBase) parseAddOptions(opts ...AddOption) *AddOptions {
	options := &AddOptions{
		Overwrite: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// BaseKnowledgeItemFactory is the base implementation of knowledge item factory
type BaseKnowledgeItemFactory struct{}

// NewBaseKnowledgeItemFactory creates a base knowledge item factory
func NewBaseKnowledgeItemFactory() *BaseKnowledgeItemFactory {
	return &BaseKnowledgeItemFactory{}
}

// FromDocument creates KnowledgeItem from Document
func (f *BaseKnowledgeItemFactory) FromDocument(doc *document.Document) KnowledgeItem {
	item := &baseKnowledgeItem{
		id:       doc.Id,
		document: doc,
	}

	// Try to restore time information from metadata
	if doc.Metadata != nil {
		if createdAtStr, ok := doc.Metadata["knowledge_item_created_at"].(string); ok {
			if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				item.createdAt = createdAt
			}
		}
		if updatedAtStr, ok := doc.Metadata["knowledge_item_updated_at"].(string); ok {
			if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
				item.updatedAt = updatedAt
			}
		}
	}

	// If no time information, use current time
	if item.createdAt.IsZero() {
		item.createdAt = time.Now()
	}
	if item.updatedAt.IsZero() {
		item.updatedAt = time.Now()
	}

	return item
}

// NewKnowledgeItem creates a new knowledge item
func (f *BaseKnowledgeItemFactory) NewKnowledgeItem(doc *document.Document) KnowledgeItem {
	now := time.Now()
	return &baseKnowledgeItem{
		id:        doc.Id,
		document:  doc,
		createdAt: now,
		updatedAt: now,
	}
}

// baseKnowledgeItem is the base implementation of knowledge item
type baseKnowledgeItem struct {
	id        document.DocumentId
	document  *document.Document
	createdAt time.Time
	updatedAt time.Time
}

// NewKnowledgeItem creates a new knowledge item (compatibility function)
func NewKnowledgeItem(doc *document.Document) *baseKnowledgeItem {
	now := time.Now()
	return &baseKnowledgeItem{
		id:        doc.Id,
		document:  doc,
		createdAt: now,
		updatedAt: now,
	}
}

// NewFromDocument creates a new knowledge item from Document (compatibility function)
func NewFromDocument(doc *document.Document) *baseKnowledgeItem {
	factory := NewBaseKnowledgeItemFactory()
	item := factory.FromDocument(doc).(*baseKnowledgeItem)
	return item
}

// FromDocument creates KnowledgeItem from Document (compatibility function)
func FromDocument(doc *document.Document) KnowledgeItem {
	factory := NewBaseKnowledgeItemFactory()
	item := factory.FromDocument(doc).(*baseKnowledgeItem)
	return item
}

func (k *baseKnowledgeItem) GetId() document.DocumentId {
	return k.id
}

func (k *baseKnowledgeItem) GetCreatedAt() time.Time {
	return k.createdAt
}

func (k *baseKnowledgeItem) GetUpdatedAt() time.Time {
	return k.updatedAt
}

// UpdateDocument updates document content
func (k *baseKnowledgeItem) UpdateDocument(doc *document.Document) {
	k.document = doc
	k.updatedAt = time.Now()
}

// ToDocument converts to Document
func (k *baseKnowledgeItem) ToDocument() *document.Document {
	// Merge KnowledgeItem metadata into Document Metadata
	metadata := make(map[string]any)
	if k.document.Metadata != nil {
		for key, value := range k.document.Metadata {
			metadata[key] = value
		}
	}

	// Add KnowledgeItem specific metadata
	metadata["knowledge_item_created_at"] = k.createdAt.Format(time.RFC3339)
	metadata["knowledge_item_updated_at"] = k.updatedAt.Format(time.RFC3339)

	return &document.Document{
		Id:       k.id,
		Name:     k.document.Name,
		Content:  k.document.Content,
		Metadata: metadata,
	}
}
