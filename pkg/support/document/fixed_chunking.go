package document

import (
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"regexp"
	"strings"

	"github.com/oopslink/agent-go/pkg/commons/utils"
)

var _ Chunker = (*fixedChunker)(nil)

// NewFixedChunker creates a new fixed-size chunker with overlap support.
func NewFixedChunker(chunkSize int, overlap int, cleanBeforeChunking bool) Chunker {
	if overlap < 0 {
		overlap = 0 // Ensure overlap is non-negative
	}
	return &fixedChunker{ChunkSize: chunkSize, Overlap: overlap, CleanBeforeChunking: cleanBeforeChunking}
}

type fixedChunker struct {
	ChunkSize           int  // Size of each chunk in characters
	Overlap             int  // Number of characters to overlap between chunks
	CleanBeforeChunking bool // Whether to clean the text before chunking
}

func (c *fixedChunker) Chunk(doc *Document) ([]*Document, error) {
	if c.ChunkSize <= 0 {
		return nil, errors.Errorf(ErrorCodeChunkingFailed, "chunk size must be greater than 0")
	}

	if c.Overlap < 0 {
		return nil, errors.Errorf(ErrorCodeChunkingFailed, "overlap must be non-negative")
	}

	if c.Overlap >= c.ChunkSize {
		return nil, errors.Errorf(ErrorCodeChunkingFailed,
			"overlap (%d) must be less than chunk size (%d)", c.Overlap, c.ChunkSize)
	}

	// If cleaning is enabled, clean the text content
	content := doc.Content
	if c.CleanBeforeChunking {
		// Clean the text content
		content = c.cleanText(content)
		if len(content) == 0 {
			// Return document with cleaned empty content
			cleanedDoc := &Document{
				Id:       doc.Id,
				Name:     doc.Name,
				Metadata: doc.Metadata,
				Content:  content,
			}
			return []*Document{cleanedDoc}, nil
		}
	}

	// If content is smaller than chunk size, return as single chunk with cleaned content
	if len(content) <= c.ChunkSize {
		cleanedDoc := &Document{
			Id:       doc.Id,
			Name:     doc.Name,
			Metadata: doc.Metadata,
			Content:  content,
		}
		return []*Document{cleanedDoc}, nil
	}

	var chunks []*Document
	start := 0
	chunkNumber := 1

	for start < len(content) {
		end := start + c.ChunkSize

		// Adjust end position to avoid splitting words
		if end < len(content) {
			end = c.findWordBoundary(content, end)
		} else {
			end = len(content)
		}

		// If we couldn't find a good word boundary and we're at the start,
		// just use the chunk size to avoid infinite loops
		if end == start {
			end = start + c.ChunkSize
			if end > len(content) {
				end = len(content)
			}
		}

		chunk := content[start:end]
		chunk = strings.TrimSpace(chunk)

		if len(chunk) > 0 {
			// Create metadata for this chunk
			chunkMetadata := make(map[string]any)
			if doc.Metadata != nil {
				// Copy original metadata
				for k, v := range doc.Metadata {
					chunkMetadata[k] = v
				}
			}

			// Add chunk-specific metadata
			chunkMetadata["chunk"] = chunkNumber
			chunkMetadata["chunk_size"] = len(chunk)

			// Generate chunk ID
			var chunkId DocumentId
			if doc.Id != "" {
				chunkId = DocumentId(fmt.Sprintf("%s_%d", doc.Id, chunkNumber))
			} else if doc.Name != "" {
				chunkId = DocumentId(fmt.Sprintf("%s_%d", doc.Name, chunkNumber))
			} else {
				chunkId = DocumentId(fmt.Sprintf("chunk_%d", chunkNumber))
			}

			chunks = append(chunks, &Document{
				Id:       chunkId,
				Name:     doc.Name,
				Metadata: chunkMetadata,
				Content:  chunk,
			})

			chunkNumber++
		}

		// Move to next chunk position with overlap
		newStart := end - c.Overlap
		if newStart <= start {
			// Prevent infinite loop by ensuring we make progress
			newStart = start + utils.MaxInt(1, c.ChunkSize/10)
		}
		start = newStart
	}

	return chunks, nil
}

// cleanText cleans the text by normalizing whitespace and removing excessive newlines.
func (c *fixedChunker) cleanText(text string) string {
	// Replace multiple newlines with a single newline
	text = regexp.MustCompile(`\n+`).ReplaceAllString(text, "\n")
	// Replace multiple spaces with a single space
	text = regexp.MustCompile(` +`).ReplaceAllString(text, " ")
	// Replace multiple tabs with a single tab
	text = regexp.MustCompile(`\t+`).ReplaceAllString(text, "\t")
	// Replace multiple carriage returns with a single carriage return
	text = regexp.MustCompile(`\r+`).ReplaceAllString(text, "\r")
	// Replace multiple form feeds with a single form feed
	text = regexp.MustCompile(`\f+`).ReplaceAllString(text, "\f")
	// Replace multiple vertical tabs with a single vertical tab
	text = regexp.MustCompile(`\v+`).ReplaceAllString(text, "\v")

	return strings.TrimSpace(text)
}

// findWordBoundary finds the best word boundary before the given position.
func (c *fixedChunker) findWordBoundary(text string, pos int) int {
	if pos >= len(text) {
		return len(text)
	}

	// Look for word boundaries (spaces, newlines, punctuation)
	for i := pos; i > 0; i-- {
		char := text[i]
		if char == ' ' || char == '\n' || char == '\t' || char == '\r' ||
			char == '.' || char == '!' || char == '?' || char == ';' || char == ':' {
			return i + 1
		}
	}

	// If no word boundary found, return the original position
	return pos
}
