package knowledge

import (
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/document"
)

func TestKnowledgeItemInterface(t *testing.T) {
	// Create test document
	doc := &document.Document{
		Id:      "test_doc_123",
		Name:    "test_document",
		Content: "This is a test document content.",
		Metadata: map[string]any{
			"category": "test",
			"version":  "1.0",
		},
	}

	// Create factory
	factory := NewBaseKnowledgeItemFactory()

	// Create knowledge item
	item := factory.NewKnowledgeItem(doc)

	// Test basic methods
	if item.GetId() != "test_doc_123" {
		t.Errorf("Expected ID 'test_doc_123', got '%s'", item.GetId())
	}

	// Test ToDocument method
	convertedDoc := item.ToDocument()
	if convertedDoc.Id != "test_doc_123" {
		t.Errorf("Expected converted document ID 'test_doc_123', got '%s'", convertedDoc.Id)
	}

	if convertedDoc.Name != "test_document" {
		t.Errorf("Expected converted document name 'test_document', got '%s'", convertedDoc.Name)
	}

	// Check if metadata contains knowledge item information
	if _, exists := convertedDoc.Metadata["knowledge_item_created_at"]; !exists {
		t.Errorf("Expected knowledge_item_created_at in metadata")
	}

	// Test factory's FromDocument method
	newItem := factory.FromDocument(convertedDoc)

	if newItem.GetId() != "test_doc_123" {
		t.Errorf("Expected new item ID 'test_doc_123', got '%s'", newItem.GetId())
	}

	// Test time information
	createdAt := item.GetCreatedAt()
	updatedAt := item.GetUpdatedAt()

	if createdAt.IsZero() {
		t.Error("Expected createdAt to be set")
	}

	if updatedAt.IsZero() {
		t.Error("Expected updatedAt to be set")
	}

	// Test compatibility functions
	compatItem := NewKnowledgeItem(doc)
	if compatItem.GetId() != "test_doc_123" {
		t.Errorf("Expected compatibility item ID 'test_doc_123', got '%s'", compatItem.GetId())
	}

	staticItem := FromDocument(convertedDoc)
	if staticItem.GetId() != "test_doc_123" {
		t.Errorf("Expected static item ID 'test_doc_123', got '%s'", staticItem.GetId())
	}
}

func TestKnowledgeItemUpdateDocument(t *testing.T) {
	// Create initial document and item
	initialDoc := &document.Document{
		Id:      "update_test_123",
		Name:    "initial_document",
		Content: "Initial content",
		Metadata: map[string]any{
			"category": "initial",
		},
	}

	item := NewKnowledgeItem(initialDoc)

	initialUpdatedAt := item.GetUpdatedAt()

	// Wait a small amount of time to ensure timestamps are different
	time.Sleep(1 * time.Millisecond)

	// Update document
	updatedDoc := &document.Document{
		Id:      "update_test_123",
		Name:    "updated_document",
		Content: "Updated content",
		Metadata: map[string]any{
			"category": "updated",
		},
	}

	item.UpdateDocument(updatedDoc)

	// Verify updates
	if item.ToDocument().Name != "updated_document" {
		t.Errorf("Expected updated document name 'updated_document', got '%s'", item.ToDocument().Name)
	}

	if item.ToDocument().Content != "Updated content" {
		t.Errorf("Expected updated document content 'Updated content', got '%s'", item.ToDocument().Content)
	}

	if item.GetUpdatedAt().Equal(initialUpdatedAt) {
		t.Error("Expected updatedAt to be different after update")
	}
}

func TestKnowledgeBaseMetadata(t *testing.T) {
	// Test metadata creation
	metadata := NewKnowledgeBaseMetadata(
		"test_knowledge_base",
		"test knowledge base description",
		[]string{"domain1", "domain2"},
		map[string]string{"tag1": "value1", "tag2": "value2"},
	)

	if metadata.Name != "test_knowledge_base" {
		t.Errorf("Expected name 'test_knowledge_base', got '%s'", metadata.Name)
	}

	if metadata.Description != "test knowledge base description" {
		t.Errorf("Expected description 'test knowledge base description', got '%s'", metadata.Description)
	}

	if len(metadata.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(metadata.Domains))
	}

	if len(metadata.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(metadata.Tags))
	}

	// Test knowledge base with metadata
	kb := &knowledgeBase{
		metadata: metadata,
	}

	// Test GetMetadata
	retrievedMetadata := kb.GetMetadata()
	if retrievedMetadata.Name != metadata.Name {
		t.Errorf("Expected retrieved name '%s', got '%s'", metadata.Name, retrievedMetadata.Name)
	}

	// Test UpdateMetadata
	newMetadata := NewKnowledgeBaseMetadata(
		"updated_knowledge_base",
		"updated description",
		[]string{"domain3"},
		map[string]string{"tag3": "value3"},
	)

	kb.UpdateMetadata(newMetadata)
	updatedMetadata := kb.GetMetadata()
	if updatedMetadata.Name != "updated_knowledge_base" {
		t.Errorf("Expected updated name 'updated_knowledge_base', got '%s'", updatedMetadata.Name)
	}
}
