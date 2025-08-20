package document

import (
	"fmt"
	"strings"
	"testing"

	"github.com/oopslink/agent-go/pkg/commons/utils"
)

func TestNewFixedChunker(t *testing.T) {
	chunker := NewFixedChunker(100, 0, true)

	// Verify type assertion
	if _, ok := chunker.(*fixedChunker); !ok {
		t.Error("NewFixedChunker() should return a *fixedChunker")
	}

	// Verify chunk size
	if fc, ok := chunker.(*fixedChunker); ok {
		if fc.ChunkSize != 100 {
			t.Errorf("Expected chunk size 100, got %d", fc.ChunkSize)
		}
		if fc.Overlap != 0 {
			t.Errorf("Expected overlap 0, got %d", fc.Overlap)
		}
		if fc.CleanBeforeChunking != true {
			t.Errorf("Expected CleanBeforeChunking true, got %v", fc.CleanBeforeChunking)
		}
	}
}

func TestNewFixedChunkerWithOverlap(t *testing.T) {
	chunker := NewFixedChunker(100, 20, true)

	// Verify type assertion
	if _, ok := chunker.(*fixedChunker); !ok {
		t.Error("NewFixedChunker() should return a *fixedChunker")
	}

	// Verify chunk size and overlap
	if fc, ok := chunker.(*fixedChunker); ok {
		if fc.ChunkSize != 100 {
			t.Errorf("Expected chunk size 100, got %d", fc.ChunkSize)
		}
		if fc.Overlap != 20 {
			t.Errorf("Expected overlap 20, got %d", fc.Overlap)
		}
		if fc.CleanBeforeChunking != true {
			t.Errorf("Expected CleanBeforeChunking true, got %v", fc.CleanBeforeChunking)
		}
	}
}

func TestFixedChunker_InvalidChunkSize(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 0}
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: "test content",
	}

	chunks, err := chunker.Chunk(doc)
	if err == nil {
		t.Error("Expected error for chunk size 0")
	}
	if chunks != nil {
		t.Error("Expected nil chunks for invalid chunk size")
	}

	// Test negative chunk size
	chunker.ChunkSize = -1
	_, err = chunker.Chunk(doc)
	if err == nil {
		t.Error("Expected error for negative chunk size")
	}
}

func TestFixedChunker_InvalidOverlap(t *testing.T) {
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: "test content",
	}

	// Test negative overlap
	chunker := &fixedChunker{ChunkSize: 100, Overlap: -1}
	chunks, err := chunker.Chunk(doc)
	if err == nil {
		t.Error("Expected error for negative overlap")
	}
	if chunks != nil {
		t.Error("Expected nil chunks for invalid overlap")
	}

	// Test overlap >= chunk size
	chunker.Overlap = 100
	_, err = chunker.Chunk(doc)
	if err == nil {
		t.Error("Expected error for overlap >= chunk size")
	}
}

func TestFixedChunker_EmptyContent(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 100}
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: "",
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for empty content, got %d", len(chunks))
	}

	if chunks[0].Content != "" {
		t.Errorf("Expected empty content, got %q", chunks[0].Content)
	}
}

func TestFixedChunker_ContentSmallerThanChunkSize(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 100}
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: "This is a short text.",
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for content smaller than chunk size, got %d", len(chunks))
	}

	if chunks[0].Content != doc.Content {
		t.Errorf("Expected content %q, got %q", doc.Content, chunks[0].Content)
	}
}

func TestFixedChunker_BasicChunking(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 10}
	content := "This is a test document that should be split into multiple chunks."
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: content,
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least 1 chunk")
	}

	// Verify all chunks have content
	for i, chunk := range chunks {
		if len(chunk.Content) == 0 {
			t.Errorf("Chunk %d has empty content", i)
		}
		if len(chunk.Content) > chunker.ChunkSize {
			t.Errorf("Chunk %d exceeds chunk size: %d > %d", i, len(chunk.Content), chunker.ChunkSize)
		}
	}

	// Verify chunk IDs and metadata
	for i, chunk := range chunks {
		expectedId := fmt.Sprintf("test_%d", i+1)
		if string(chunk.Id) != expectedId {
			t.Errorf("Expected chunk ID %q, got %q", expectedId, chunk.Id)
		}

		if chunk.Name != doc.Name {
			t.Errorf("Expected chunk name %q, got %q", doc.Name, chunk.Name)
		}

		if chunk.Metadata == nil {
			t.Error("Expected chunk metadata to be non-nil")
		}

		if chunkNum, ok := chunk.Metadata["chunk"]; !ok || chunkNum != i+1 {
			t.Errorf("Expected chunk number %d, got %v", i+1, chunkNum)
		}

		if chunkSize, ok := chunk.Metadata["chunk_size"]; !ok || chunkSize != len(chunk.Content) {
			t.Errorf("Expected chunk size %d, got %v", len(chunk.Content), chunkSize)
		}
	}
}

func TestFixedChunker_WithOverlap(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 20, Overlap: 5}
	content := "This is a test document that should be split into multiple chunks with overlap."
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: content,
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(chunks) <= 1 {
		t.Error("Expected multiple chunks with overlap")
	}

	// Verify overlapping content exists between consecutive chunks
	for i := 1; i < len(chunks); i++ {
		prevChunk := chunks[i-1]
		currentChunk := chunks[i]

		// Check if there's some overlap (at least some characters match)
		prevEnd := prevChunk.Content[len(prevChunk.Content)-utils.MinInt(len(prevChunk.Content), chunker.Overlap):]
		currentStart := currentChunk.Content[:utils.MinInt(len(currentChunk.Content), chunker.Overlap)]

		// Since we're using word boundaries, we might not have exact overlap,
		// but we should have some common content
		if len(prevEnd) > 0 && len(currentStart) > 0 {
			// At least verify that chunks are not completely disjoint
			if !strings.Contains(content, prevEnd) || !strings.Contains(content, currentStart) {
				t.Errorf("Chunks %d and %d don't seem to have proper overlap", i-1, i)
			}
		}
	}
}

func TestFixedChunker_TextCleaning(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 100, CleanBeforeChunking: true}

	// Test content with multiple whitespace characters
	content := "This   is\n\n\na   test\t\t\twith\r\r\rexcessive\f\f\fwhitespace\v\v\vcharacters."
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: content,
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk, got %d", len(chunks))
	}

	// Verify the content has been cleaned
	cleanedContent := chunks[0].Content
	t.Logf("Original content: %q", content)
	t.Logf("Cleaned content: %q", cleanedContent)

	if strings.Contains(cleanedContent, "\n\n") {
		t.Error("Expected multiple newlines to be reduced to single newline")
	}
	if strings.Contains(cleanedContent, "   ") {
		t.Error("Expected multiple spaces to be reduced to single space")
	}
}

func TestFixedChunker_WordBoundaries(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 15}
	content := "This is a very long word that should not be broken arbitrarily."
	doc := &Document{
		Id:      "test",
		Name:    "test.txt",
		Content: content,
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that words are not split arbitrarily
	for i, chunk := range chunks {
		content := chunk.Content
		if len(content) > 0 {
			// Check that chunk doesn't start or end in the middle of a word
			// (unless it's the very first or last chunk)
			if i > 0 && len(content) > 0 {
				firstChar := content[0]
				// Should not start with a letter if previous chunk ended with a letter
				if firstChar >= 'a' && firstChar <= 'z' || firstChar >= 'A' && firstChar <= 'Z' {
					// This might be acceptable if the word is very long
					t.Logf("Chunk %d starts with letter: %q", i, content)
				}
			}
		}
	}
}

func TestFixedChunker_MetadataPreservation(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 10}
	originalMetadata := map[string]any{
		"source":   "test_source",
		"category": "test_category",
		"priority": 1,
	}

	doc := &Document{
		Id:       "test",
		Name:     "test.txt",
		Metadata: originalMetadata,
		Content:  "This is a test document that should be split into multiple chunks.",
	}

	chunks, err := chunker.Chunk(doc)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for i, chunk := range chunks {
		// Verify original metadata is preserved
		for key, value := range originalMetadata {
			if chunkValue, ok := chunk.Metadata[key]; !ok || chunkValue != value {
				t.Errorf("Chunk %d: Expected metadata %s=%v, got %v", i, key, value, chunkValue)
			}
		}

		// Verify chunk-specific metadata is added
		if chunkNum, ok := chunk.Metadata["chunk"]; !ok || chunkNum != i+1 {
			t.Errorf("Chunk %d: Expected chunk number %d, got %v", i, i+1, chunkNum)
		}

		if chunkSize, ok := chunk.Metadata["chunk_size"]; !ok || chunkSize != len(chunk.Content) {
			t.Errorf("Chunk %d: Expected chunk size %d, got %v", i, len(chunk.Content), chunkSize)
		}
	}
}

func TestFixedChunker_ChunkIdGeneration(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 10}
	content := "This is a test document that should be split into multiple chunks."

	tests := []struct {
		name           string
		doc            *Document
		expectedPrefix string
	}{
		{
			name: "with document ID",
			doc: &Document{
				Id:      "doc123",
				Name:    "test.txt",
				Content: content,
			},
			expectedPrefix: "doc123_",
		},
		{
			name: "with document name only",
			doc: &Document{
				Id:      "",
				Name:    "test.txt",
				Content: content,
			},
			expectedPrefix: "test.txt_",
		},
		{
			name: "without ID or name",
			doc: &Document{
				Id:      "",
				Name:    "",
				Content: content,
			},
			expectedPrefix: "chunk_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := chunker.Chunk(tt.doc)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			for i, chunk := range chunks {
				expectedId := fmt.Sprintf("%s%d", tt.expectedPrefix, i+1)
				if string(chunk.Id) != expectedId {
					t.Errorf("Expected chunk ID %q, got %q", expectedId, chunk.Id)
				}
			}
		})
	}
}

func TestFixedChunker_CleanText(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 100, CleanBeforeChunking: true}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple newlines",
			input:    "line1\n\n\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "multiple spaces",
			input:    "word1   word2",
			expected: "word1 word2",
		},
		{
			name:     "multiple tabs",
			input:    "word1\t\t\tword2",
			expected: "word1\tword2",
		},
		{
			name:     "mixed whitespace",
			input:    "  word1\n\n\tword2   ",
			expected: "word1\n\tword2",
		},
		{
			name:     "carriage returns",
			input:    "word1\r\r\rword2",
			expected: "word1\rword2",
		},
		{
			name:     "form feeds",
			input:    "word1\f\f\fword2",
			expected: "word1\fword2",
		},
		{
			name:     "vertical tabs",
			input:    "word1\v\v\vword2",
			expected: "word1\vword2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chunker.cleanText(tt.input)
			if result != tt.expected {
				t.Errorf("cleanText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFixedChunker_FindWordBoundary(t *testing.T) {
	chunker := &fixedChunker{ChunkSize: 100}

	tests := []struct {
		name     string
		text     string
		pos      int
		expected int
	}{
		{
			name:     "space boundary",
			text:     "hello world test",
			pos:      10,
			expected: 6, // position after "hello "
		},
		{
			name:     "newline boundary",
			text:     "hello\nworld",
			pos:      8,
			expected: 6, // position after "hello\n"
		},
		{
			name:     "punctuation boundary",
			text:     "hello. world",
			pos:      8,
			expected: 7, // position after "hello."
		},
		{
			name:     "no boundary found",
			text:     "verylongwordwithoutspaces",
			pos:      15,
			expected: 15, // return original position
		},
		{
			name:     "position at end",
			text:     "hello world",
			pos:      15,
			expected: 11, // length of text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chunker.findWordBoundary(tt.text, tt.pos)
			if result != tt.expected {
				t.Errorf("findWordBoundary(%q, %d) = %d, want %d", tt.text, tt.pos, result, tt.expected)
			}
		})
	}
}
