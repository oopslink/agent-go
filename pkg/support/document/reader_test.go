package document

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock chunker for testing
type mockChunker struct {
	shouldError bool
	chunks      []*Document
}

func (m *mockChunker) Chunk(doc *Document) ([]*Document, error) {
	if m.shouldError {
		return nil, errors.New("chunking error")
	}
	return m.chunks, nil
}

// Mock io.Reader that returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestDefaultReader_Read(t *testing.T) {
	reader := &defaultReader{}

	t.Run("successful read without chunker", func(t *testing.T) {
		content := "This is test content"
		stringReader := strings.NewReader(content)

		docs, err := reader.Read("test.txt", stringReader)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(docs) != 1 {
			t.Fatalf("Expected 1 document, got %d", len(docs))
		}

		doc := docs[0]
		if doc.Id != "default" {
			t.Errorf("Expected document ID 'default', got %s", doc.Id)
		}

		if doc.Name != "test.txt" {
			t.Errorf("Expected document name 'test.txt', got %s", doc.Name)
		}

		if doc.Content != content {
			t.Errorf("Expected content '%s', got '%s'", content, doc.Content)
		}

		if doc.Metadata == nil {
			t.Error("Expected metadata to be initialized")
		}
	})

	t.Run("successful read with chunker", func(t *testing.T) {
		content := "This is test content"
		stringReader := strings.NewReader(content)

		expectedChunks := []*Document{
			{Id: "chunk1", Name: "test.txt", Content: "This is"},
			{Id: "chunk2", Name: "test.txt", Content: "test content"},
		}

		chunker := &mockChunker{
			shouldError: false,
			chunks:      expectedChunks,
		}

		docs, err := reader.Read("test.txt", stringReader, WithChunker(chunker))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(docs) != 2 {
			t.Fatalf("Expected 2 documents, got %d", len(docs))
		}

		for i, doc := range docs {
			if doc.Id != expectedChunks[i].Id {
				t.Errorf("Expected document ID '%s', got '%s'", expectedChunks[i].Id, doc.Id)
			}
			if doc.Content != expectedChunks[i].Content {
				t.Errorf("Expected content '%s', got '%s'", expectedChunks[i].Content, doc.Content)
			}
		}
	})

	t.Run("read error from io.Reader", func(t *testing.T) {
		errorReader := &errorReader{}

		docs, err := reader.Read("test.txt", errorReader)

		if err == nil {
			t.Fatal("Expected error from io.Reader, got nil")
		}

		if docs != nil {
			t.Error("Expected nil documents when read fails")
		}

		if err.Error() != "read error" {
			t.Errorf("Expected error message 'read error', got '%s'", err.Error())
		}
	})

	t.Run("chunker error", func(t *testing.T) {
		content := "This is test content"
		stringReader := strings.NewReader(content)

		chunker := &mockChunker{
			shouldError: true,
		}

		docs, err := reader.Read("test.txt", stringReader, WithChunker(chunker))

		if err == nil {
			t.Fatal("Expected error from chunker, got nil")
		}

		if docs != nil {
			t.Error("Expected nil documents when chunking fails")
		}

		if err.Error() != "chunking error" {
			t.Errorf("Expected error message 'chunking error', got '%s'", err.Error())
		}
	})

	t.Run("empty content", func(t *testing.T) {
		stringReader := strings.NewReader("")

		docs, err := reader.Read("empty.txt", stringReader)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(docs) != 1 {
			t.Fatalf("Expected 1 document, got %d", len(docs))
		}

		if docs[0].Content != "" {
			t.Errorf("Expected empty content, got '%s'", docs[0].Content)
		}
	})
}

func TestOfString(t *testing.T) {
	t.Run("successful string reader creation", func(t *testing.T) {
		content := "Test content for string reader"

		reader, err := OfString(content)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if reader == nil {
			t.Fatal("Expected non-nil reader")
		}

		// Test reading from the returned reader
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected content '%s', got '%s'", content, string(data))
		}
	})

	t.Run("empty string", func(t *testing.T) {
		reader, err := OfString("")

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != "" {
			t.Errorf("Expected empty content, got '%s'", string(data))
		}
	})

	t.Run("large string", func(t *testing.T) {
		content := strings.Repeat("Large content test ", 1000)

		reader, err := OfString(content)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != content {
			t.Error("Large string content mismatch")
		}
	})
}

func TestOfFile(t *testing.T) {
	t.Run("successful file reading", func(t *testing.T) {
		// Create a temporary file for testing
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		content := "Test file content"

		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		reader, err := OfFile(testFile)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if reader == nil {
			t.Fatal("Expected non-nil reader")
		}

		// Test reading from the file
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected content '%s', got '%s'", content, string(data))
		}

		// Close the file reader
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := OfFile("/non/existent/file.txt")

		if err == nil {
			t.Fatal("Expected error for non-existent file, got nil")
		}

		// Note: os.Open may return non-nil reader even on error, so we don't check reader != nil
	})

	t.Run("empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.txt")

		err := os.WriteFile(testFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		reader, err := OfFile(testFile)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != "" {
			t.Errorf("Expected empty content, got '%s'", string(data))
		}

		// Close the file reader
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})

	t.Run("directory instead of file", func(t *testing.T) {
		tempDir := t.TempDir()

		reader, err := OfFile(tempDir)

		// os.Open can successfully open directories, so we don't expect an error
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if reader == nil {
			t.Fatal("Expected non-nil reader")
		}

		// Try to read from directory (should fail)
		_, readErr := io.ReadAll(reader)
		if readErr == nil {
			t.Error("Expected error when reading from directory")
		}

		// Close the reader
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})
}

func TestOfUrl(t *testing.T) {
	t.Run("successful URL reading", func(t *testing.T) {
		content := "Test HTTP content"

		// Create a test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
		}))
		defer server.Close()

		reader, err := OfUrl(server.URL)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if reader == nil {
			t.Fatal("Expected non-nil reader")
		}

		// Test reading from the URL
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != content {
			t.Errorf("Expected content '%s', got '%s'", content, string(data))
		}

		// Close the response body
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})

	t.Run("HTTP 404 error", func(t *testing.T) {
		// Create a test HTTP server that returns 404
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		reader, err := OfUrl(server.URL)

		if err != nil {
			t.Fatalf("Expected no error for 404 (HTTP client doesn't error on 404), got: %v", err)
		}

		// HTTP client doesn't return error for 404, but we can read the response
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != "Not Found" {
			t.Errorf("Expected content 'Not Found', got '%s'", string(data))
		}

		// Close the response body
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		reader, err := OfUrl("invalid-url")

		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}

		if reader != nil {
			t.Error("Expected nil reader for invalid URL")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		// Create a test HTTP server that returns empty content
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		}))
		defer server.Close()

		reader, err := OfUrl(server.URL)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Expected no error reading, got: %v", err)
		}

		if string(data) != "" {
			t.Errorf("Expected empty content, got '%s'", string(data))
		}

		// Close the response body
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})
}

func TestReaderIntegration(t *testing.T) {
	t.Run("integration with mock chunker", func(t *testing.T) {
		content := "This is a long piece of content that should be chunked into smaller pieces for testing purposes."

		// Create a mock chunker that splits content into fixed chunks
		expectedChunks := []*Document{
			{Id: "chunk1", Name: "test.txt", Content: "This is a long piece of content"},
			{Id: "chunk2", Name: "test.txt", Content: "that should be chunked into smaller"},
			{Id: "chunk3", Name: "test.txt", Content: "pieces for testing purposes."},
		}

		chunker := &mockChunker{
			shouldError: false,
			chunks:      expectedChunks,
		}

		// Create string reader
		stringReader := strings.NewReader(content)

		// Create default reader
		reader := &defaultReader{}

		// Read and chunk the content
		docs, err := reader.Read("test.txt", stringReader, WithChunker(chunker))

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(docs) != 3 {
			t.Fatalf("Expected 3 document chunks, got %d", len(docs))
		}

		// Verify that all chunks have the correct document name
		for i, doc := range docs {
			if doc.Name != "test.txt" {
				t.Errorf("Chunk %d: Expected document name 'test.txt', got '%s'", i, doc.Name)
			}

			if doc.Content == "" {
				t.Errorf("Chunk %d: Expected non-empty content", i)
			}

			if doc.Content != expectedChunks[i].Content {
				t.Errorf("Chunk %d: Expected content '%s', got '%s'", i, expectedChunks[i].Content, doc.Content)
			}
		}
	})
}

func TestReaderInterface(t *testing.T) {
	t.Run("defaultReader implements Reader interface", func(t *testing.T) {
		var reader Reader = &defaultReader{}

		// Test that the interface method is callable
		stringReader := strings.NewReader("test")
		docs, err := reader.Read("test.txt", stringReader)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(docs) != 1 {
			t.Fatalf("Expected 1 document, got %d", len(docs))
		}
	})
}
