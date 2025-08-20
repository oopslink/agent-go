package knowledge

import (
	"context"
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestKnowledgeRetrieverToolWrapper(t *testing.T) {
	// Create mock knowledge bases
	metadata1 := NewKnowledgeBaseMetadata(
		"test_base_1",
		"test knowledge base 1",
		[]string{"domain1", "domain2"},
		map[string]string{"tag1": "value1"},
	)

	metadata2 := NewKnowledgeBaseMetadata(
		"test_base_2",
		"test knowledge base 2",
		[]string{"domain3"},
		map[string]string{"tag2": "value2"},
	)

	// Create mock knowledge bases
	kb1 := &knowledgeBase{metadata: metadata1}
	kb2 := &knowledgeBase{metadata: metadata2}

	// Create tool wrapper
	wrapper := WrapperAsRetrieverTool(kb1, kb2)

	// Test tool name
	if wrapper.toolName() != KNOWLEDGE_RETRIVER_TOOL {
		t.Errorf("Expected tool name 'knowledge_retriever', got '%s'", wrapper.toolName())
	}

	// Test descriptor
	descriptor := wrapper.Descriptor()
	if descriptor.Name != KNOWLEDGE_RETRIVER_TOOL {
		t.Errorf("Expected descriptor name 'knowledge_retriever', got '%s'", descriptor.Name)
	}

	if descriptor.Parameters == nil {
		t.Error("Expected descriptor parameters to be set")
	}

	// Test parameter parsing
	toolCall := &llms.ToolCall{
		ToolCallId: "test_call_1",
		Name:       KNOWLEDGE_RETRIVER_TOOL,
		Arguments: map[string]any{
			"query":           "test query",
			"max_results":     5.0,
			"score_threshold": 0.8,
			"collection":      "test_collection",
			"domains":         []interface{}{"domain1"},
			"max_bases":       2.0,
		},
	}

	params := wrapper.makeKnowledgeRetrieverParams(toolCall)
	if params.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", params.Query)
	}

	if params.MaxResults != 5 {
		t.Errorf("Expected max results 5, got %d", params.MaxResults)
	}

	if params.ScoreThreshold != 0.8 {
		t.Errorf("Expected score threshold 0.8, got %f", params.ScoreThreshold)
	}

	if len(params.Domains) != 1 || params.Domains[0] != "domain1" {
		t.Errorf("Expected domains ['domain1'], got %v", params.Domains)
	}

	if params.MaxBases != 2 {
		t.Errorf("Expected max bases 2, got %d", params.MaxBases)
	}

	if len(params.SearchOptions) == 0 {
		t.Error("Expected search options to be set")
	}
}

func TestKnowledgeRetrieverToolWrapper_SelectKnowledgeBases(t *testing.T) {
	// Create mock knowledge bases
	metadata1 := NewKnowledgeBaseMetadata(
		"test_base_1",
		"test knowledge base 1",
		[]string{"domain1", "domain2"},
		map[string]string{"tag1": "value1"},
	)

	metadata2 := NewKnowledgeBaseMetadata(
		"test_base_2",
		"test knowledge base 2",
		[]string{"domain3"},
		map[string]string{"tag2": "value2"},
	)

	kb1 := &knowledgeBase{metadata: metadata1}
	kb2 := &knowledgeBase{metadata: metadata2}

	wrapper := &KnowledgeRetrieverToolWrapper{
		knowledgeBases: []KnowledgeBase{kb1, kb2},
		strategy:       NewMaxCountStrategy(-1), // Select all
	}

	// Test selecting all bases when no domains specified
	params := &KnowledgeRetrieverParams{
		Query:   "test query",
		Domains: []string{},
	}

	bases, err := wrapper.selectKnowledgeBases(context.Background(), params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(bases) != 2 {
		t.Errorf("Expected 2 bases, got %d", len(bases))
	}

	// Test selecting bases by domain
	params.Domains = []string{"domain1"}
	bases, err = wrapper.selectKnowledgeBases(context.Background(), params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(bases) != 1 {
		t.Errorf("Expected 1 base, got %d", len(bases))
	}

	// Test selecting bases by multiple domains
	params.Domains = []string{"domain1", "domain3"}
	bases, err = wrapper.selectKnowledgeBases(context.Background(), params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(bases) != 2 {
		t.Errorf("Expected 2 bases, got %d", len(bases))
	}

	// Test selecting bases by non-existent domain
	params.Domains = []string{"non_existent"}
	bases, err = wrapper.selectKnowledgeBases(context.Background(), params)
	if err == nil {
		t.Error("Expected error for non-existent domain")
	}
}

func TestKnowledgeRetrieverToolWrapper_EmptyBases(t *testing.T) {
	wrapper := &KnowledgeRetrieverToolWrapper{
		knowledgeBases: []KnowledgeBase{},
		strategy:       NewMaxCountStrategy(-1),
	}

	params := &KnowledgeRetrieverParams{
		Query:   "test query",
		Domains: []string{},
	}

	_, err := wrapper.selectKnowledgeBases(context.Background(), params)
	if err == nil {
		t.Error("Expected error for empty knowledge bases")
	}
}

func TestMaxCountStrategy(t *testing.T) {
	// Create mock knowledge bases
	metadata1 := NewKnowledgeBaseMetadata(
		"test_base_1",
		"test knowledge base 1",
		[]string{"domain1"},
		map[string]string{"tag1": "value1"},
	)

	metadata2 := NewKnowledgeBaseMetadata(
		"test_base_2",
		"test knowledge base 2",
		[]string{"domain1"},
		map[string]string{"tag2": "value2"},
	)

	metadata3 := NewKnowledgeBaseMetadata(
		"test_base_3",
		"test knowledge base 3",
		[]string{"domain2"},
		map[string]string{"tag3": "value3"},
	)

	kb1 := &knowledgeBase{metadata: metadata1}
	kb2 := &knowledgeBase{metadata: metadata2}
	kb3 := &knowledgeBase{metadata: metadata3}

	bases := []KnowledgeBase{kb1, kb2, kb3}

	// Test strategy with max count = 2
	strategy := NewMaxCountStrategy(2)
	selected, err := strategy.Select(bases, []string{"domain1"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(selected) != 2 {
		t.Errorf("Expected 2 bases, got %d", len(selected))
	}

	// Test strategy with max count = 1
	strategy = NewMaxCountStrategy(1)
	selected, err = strategy.Select(bases, []string{"domain1"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(selected) != 1 {
		t.Errorf("Expected 1 base, got %d", len(selected))
	}

	// Test strategy with max count = -1 (select all)
	strategy = NewMaxCountStrategy(-1)
	selected, err = strategy.Select(bases, []string{"domain1"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(selected) != 2 {
		t.Errorf("Expected 2 bases, got %d", len(selected))
	}

	// Test strategy with no domain filter
	selected, err = strategy.Select(bases, []string{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(selected) != 3 {
		t.Errorf("Expected 3 bases, got %d", len(selected))
	}
}

func TestWrapperAsRetrieverToolWithStrategy(t *testing.T) {
	// Create mock knowledge bases
	metadata1 := NewKnowledgeBaseMetadata(
		"test_base_1",
		"test knowledge base 1",
		[]string{"domain1"},
		map[string]string{"tag1": "value1"},
	)

	metadata2 := NewKnowledgeBaseMetadata(
		"test_base_2",
		"test knowledge base 2",
		[]string{"domain1"},
		map[string]string{"tag2": "value2"},
	)

	kb1 := &knowledgeBase{metadata: metadata1}
	kb2 := &knowledgeBase{metadata: metadata2}

	// Create wrapper with custom strategy
	strategy := NewMaxCountStrategy(1)
	wrapper := WrapperAsRetrieverToolWithStrategy([]KnowledgeBase{kb1, kb2}, strategy)

	// Test that the strategy is applied
	params := &KnowledgeRetrieverParams{
		Query:   "test query",
		Domains: []string{"domain1"},
	}

	bases, err := wrapper.selectKnowledgeBases(context.Background(), params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(bases) != 1 {
		t.Errorf("Expected 1 base due to max count strategy, got %d", len(bases))
	}
}

func TestKnowledgeRetrieverToolWrapper_DescriptorWithAvailableOptions(t *testing.T) {
	// Create mock knowledge bases with different domains
	metadata1 := NewKnowledgeBaseMetadata(
		"test_base_1",
		"test knowledge base 1",
		[]string{"domain1", "domain2"},
		map[string]string{"tag1": "value1"},
	)

	metadata2 := NewKnowledgeBaseMetadata(
		"test_base_2",
		"test knowledge base 2",
		[]string{"domain2", "domain3"},
		map[string]string{"tag2": "value2"},
	)

	metadata3 := NewKnowledgeBaseMetadata(
		"test_base_3",
		"test knowledge base 3",
		[]string{"domain4"},
		map[string]string{"tag3": "value3"},
	)

	kb1 := &knowledgeBase{metadata: metadata1}
	kb2 := &knowledgeBase{metadata: metadata2}
	kb3 := &knowledgeBase{metadata: metadata3}

	// Create wrapper
	wrapper := WrapperAsRetrieverTool(kb1, kb2, kb3)

	// Get descriptor
	descriptor := wrapper.Descriptor()

	// Verify descriptor has the expected structure
	if descriptor.Name != KNOWLEDGE_RETRIVER_TOOL {
		t.Errorf("Expected descriptor name 'knowledge_retriever', got '%s'", descriptor.Name)
	}

	if descriptor.Parameters == nil {
		t.Error("Expected descriptor parameters to be set")
	}

	// Check that domains parameter description includes available domains
	domainsSchema := descriptor.Parameters.Properties["domains"]
	if domainsSchema == nil {
		t.Error("Expected domains parameter to be defined")
	}

	// The description should include available domains
	expectedDomains := []string{"domain1", "domain2", "domain3", "domain4"}
	for _, domain := range expectedDomains {
		if !contains(domainsSchema.Description, domain) {
			t.Errorf("Expected domains description to contain '%s', got: %s", domain, domainsSchema.Description)
		}
	}

	// Check that collection parameter is defined
	collectionSchema := descriptor.Parameters.Properties["collection"]
	if collectionSchema == nil {
		t.Error("Expected collection parameter to be defined")
	}

	// Check that max_bases parameter is defined
	maxBasesSchema := descriptor.Parameters.Properties["max_bases"]
	if maxBasesSchema == nil {
		t.Error("Expected max_bases parameter to be defined")
	}
}

func TestKnowledgeRetrieverToolWrapper_DescriptorWithNoBases(t *testing.T) {
	// Create wrapper with no knowledge bases
	wrapper := WrapperAsRetrieverTool()

	// Get descriptor
	descriptor := wrapper.Descriptor()

	// Verify descriptor still works with no bases
	if descriptor.Name != KNOWLEDGE_RETRIVER_TOOL {
		t.Errorf("Expected descriptor name 'knowledge_retriever', got '%s'", descriptor.Name)
	}

	// Check that domains parameter has default description
	domainsSchema := descriptor.Parameters.Properties["domains"]
	if domainsSchema == nil {
		t.Error("Expected domains parameter to be defined")
	}

	expectedDescription := "Filter by knowledge base domains (optional)"
	if domainsSchema.Description != expectedDescription {
		t.Errorf("Expected domains description '%s', got '%s'", expectedDescription, domainsSchema.Description)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
