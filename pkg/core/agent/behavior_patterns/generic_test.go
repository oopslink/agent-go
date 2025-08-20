package behavior_patterns

import (
	"testing"

	"github.com/oopslink/agent-go/pkg/core/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenericPattern(t *testing.T) {
	pattern, err := NewGenericPattern()
	assert.NoError(t, err)
	assert.NotNil(t, pattern)
	assert.Implements(t, (*agent.BehaviorPattern)(nil), pattern)
}

func TestGenericPatternSystemInstruction(t *testing.T) {
	pattern, err := NewGenericPattern()
	require.NoError(t, err)

	header := "Test header"
	instruction := pattern.SystemInstruction(header)

	// Generic pattern should return just the header without additional prompt
	assert.Equal(t, header, instruction)
}

func TestGenericPatternStructure(t *testing.T) {
	pattern, err := NewGenericPattern()
	require.NoError(t, err)

	genericPattern, ok := pattern.(*genericPattern)
	assert.True(t, ok)
	assert.NotNil(t, genericPattern)
}

func TestGenericPatternInterface(t *testing.T) {
	pattern, err := NewGenericPattern()
	require.NoError(t, err)

	// Test that it implements the BehaviorPattern interface
	var _ agent.BehaviorPattern = pattern
}

func TestGenericPatternConsistency(t *testing.T) {
	pattern1, err1 := NewGenericPattern()
	pattern2, err2 := NewGenericPattern()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, pattern1)
	assert.NotNil(t, pattern2)

	// Both patterns should behave identically
	header := "Test header"
	assert.Equal(t, pattern1.SystemInstruction(header), pattern2.SystemInstruction(header))
}
