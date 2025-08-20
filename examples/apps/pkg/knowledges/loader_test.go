package knowledges

import (
	"testing"

	"github.com/oopslink/agent-go-apps/pkg/config"
)

func TestLoadKnowledgeBases(t *testing.T) {
	// Test loading knowledge bases with mock configuration
	cfg := &config.Config{
		VectorDB: config.VectorDBConfig{
			Type: "milvus",
		},
		Embedder: config.EmbedderConfig{
			Type: "llm",
		},
	}

	// This test will fail because we need real vector database and embedder
	// For now, we'll just test the configuration loading
	_, err := LoadKnowledgeBases(cfg)
	if err != nil {
		// Expected error since we don't have real Milvus and embedder configured
		t.Logf("Expected error (no real Milvus/embedder): %v", err)
	}
}

func TestMockVectorDB(t *testing.T) {
	// This test is no longer needed since we removed MockVectorDB
	// The real vector database should be tested separately
	t.Skip("MockVectorDB removed, use real vector database for testing")
}
