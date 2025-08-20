package behavior_patterns

import (
	"testing"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	"github.com/stretchr/testify/assert"
)

func TestNewRAGPattern(t *testing.T) {
	pattern, err := NewRAGPattern()
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)
}

func TestNewRAGPatternWithKnowledgeBases(t *testing.T) {
	// Test with empty knowledge bases
	pattern, err := NewRAGPatternWithKnowledgeBases([]knowledge.KnowledgeBase{})
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)

	// Test with nil knowledge bases
	pattern, err = NewRAGPatternWithKnowledgeBases(nil)
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)
}
