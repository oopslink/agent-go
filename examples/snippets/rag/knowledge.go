package main

import (
	"context"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"strings"

	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/oopslink/agent-go/pkg/support/document"
)

// createSampleKnowledgeBase creates a sample knowledge base with some educational content
func createSampleKnowledgeBase() (knowledge.KnowledgeBase, error) {
	// For this example, we'll create a simple mock knowledge base
	// In a real application, you would use a proper vector database with embeddings

	// Create a simple in-memory knowledge base implementation
	kb := &simpleKnowledgeBase{
		documents: make(map[document.DocumentId]*document.Document),
		metadata: knowledge.NewKnowledgeBaseMetadata(
			"sample-education",
			"Sample educational knowledge base",
			[]string{"education", "technology"},
			map[string]string{"type": "educational"},
		),
	}

	// Add some sample documents
	sampleDocs := []*document.Document{
		{
			Id:      "doc-1",
			Content: "Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed. It uses algorithms to identify patterns in data and make predictions or decisions.",
			Metadata: map[string]any{
				"topic": "machine learning",
				"type":  "definition",
			},
		},
		{
			Id:      "doc-2",
			Content: "Deep learning is a subset of machine learning that uses neural networks with multiple layers to model and understand complex patterns. It's particularly effective for tasks like image recognition, natural language processing, and speech recognition.",
			Metadata: map[string]any{
				"topic": "deep learning",
				"type":  "definition",
			},
		},
		{
			Id:      "doc-3",
			Content: "Natural Language Processing (NLP) is a field of artificial intelligence that focuses on the interaction between computers and human language. It enables machines to understand, interpret, and generate human language.",
			Metadata: map[string]any{
				"topic": "nlp",
				"type":  "definition",
			},
		},
	}

	for _, doc := range sampleDocs {
		kb.documents[doc.Id] = doc
	}

	return kb, nil
}

// simpleKnowledgeBase is a simple in-memory implementation for the example
type simpleKnowledgeBase struct {
	documents map[document.DocumentId]*document.Document
	metadata  *knowledge.KnowledgeBaseMetadata
}

func (kb *simpleKnowledgeBase) Search(ctx context.Context, query string, opts ...knowledge.SearchOption) ([]knowledge.KnowledgeItem, error) {
	// Simple keyword-based search for demonstration
	query = strings.ToLower(query)
	var results []knowledge.KnowledgeItem

	for _, doc := range kb.documents {
		content := strings.ToLower(doc.Content)
		if strings.Contains(content, query) {
			item := knowledge.NewKnowledgeItem(doc)
			results = append(results, item)
		}
	}

	return results, nil
}

func (kb *simpleKnowledgeBase) AddItem(ctx context.Context, item knowledge.KnowledgeItem, opts ...knowledge.AddOption) error {
	doc := item.ToDocument()
	kb.documents[doc.Id] = doc
	return nil
}

func (kb *simpleKnowledgeBase) UpdateItem(ctx context.Context, id document.DocumentId, item knowledge.KnowledgeItem, opts ...knowledge.UpdateOption) error {
	doc := item.ToDocument()
	kb.documents[id] = doc
	return nil
}

func (kb *simpleKnowledgeBase) GetMetadata() *knowledge.KnowledgeBaseMetadata {
	return kb.metadata
}

func (kb *simpleKnowledgeBase) UpdateMetadata(metadata *knowledge.KnowledgeBaseMetadata) {
	kb.metadata = metadata
}

func (kb *simpleKnowledgeBase) GetItem(ctx context.Context, id document.DocumentId) (knowledge.KnowledgeItem, error) {
	doc, exists := kb.documents[id]
	if !exists {
		return nil, errors.Errorf(errors.InternalError, "document not found: %s", id)
	}
	return knowledge.NewKnowledgeItem(doc), nil
}

func (kb *simpleKnowledgeBase) DeleteItem(ctx context.Context, id document.DocumentId) error {
	delete(kb.documents, id)
	return nil
}
