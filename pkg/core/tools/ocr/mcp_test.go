package ocr

import (
	"context"
	"testing"
	"time"
)

func TestWithOCRMcpServer(t *testing.T) {
	// Skip if tesseract is not available
	_, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tools, err := WithOCRMcpServer(ctx)
	if err != nil {
		t.Fatalf("Failed to create OCR MCP server: %v", err)
	}

	if len(tools) == 0 {
		t.Fatal("Expected at least one tool, got none")
	}

	// Verify the tool descriptor
	for _, tool := range tools {
		descriptor := tool.Descriptor()
		if descriptor == nil {
			t.Error("Expected non-nil descriptor")
			continue
		}

		if descriptor.Name == "ocr_recognize_text" {
			t.Logf("Found OCR tool: %s", descriptor.Name)
			if descriptor.Description == "" {
				t.Error("Expected non-empty description")
			}
			return
		}
	}

	t.Error("OCR tool not found in MCP server tools")
}

func TestNewOCRMcpServer(t *testing.T) {
	server := newOCRMcpServer()
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
}

func TestOCRParams(t *testing.T) {
	// Test that ocrParams struct can be properly instantiated
	params := &ocrParams{
		ImageSource: "http://example.com/test.jpg",
		Language:    "eng",
		ImagePath:   "/path/to/image.jpg",
	}

	if params.ImageSource != "http://example.com/test.jpg" {
		t.Errorf("Expected ImageSource to be 'http://example.com/test.jpg', got '%s'", params.ImageSource)
	}

	if params.Language != "eng" {
		t.Errorf("Expected Language to be 'eng', got '%s'", params.Language)
	}

	if params.ImagePath != "/path/to/image.jpg" {
		t.Errorf("Expected ImagePath to be '/path/to/image.jpg', got '%s'", params.ImagePath)
	}
}

