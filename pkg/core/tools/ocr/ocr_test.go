package ocr

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestOCRTool_Descriptor(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	descriptor := tool.Descriptor()
	if descriptor == nil {
		t.Fatal("Expected non-nil descriptor")
	}

	if descriptor.Name != "ocr_recognize_text" {
		t.Errorf("Expected name 'ocr_recognize_text', got '%s'", descriptor.Name)
	}

	if descriptor.Description == "" {
		t.Error("Expected non-empty description")
	}

	if descriptor.Parameters == nil {
		t.Fatal("Expected non-nil parameters")
	}

	// Check required parameters
	if _, ok := descriptor.Parameters.Properties["image_source"]; !ok {
		t.Error("Expected 'image_source' parameter")
	}

	if _, ok := descriptor.Parameters.Properties["language"]; !ok {
		t.Error("Expected 'language' parameter")
	}

	if _, ok := descriptor.Parameters.Properties["image_path"]; !ok {
		t.Error("Expected 'image_path' parameter")
	}
}

func TestOCRTool_Call_MissingImageSource(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-1",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"language": "eng",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	resultMap := result.Result

	if success, ok := resultMap["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if _, ok := resultMap["error"].(string); !ok {
		t.Error("Expected error message")
	}
}

func TestOCRTool_Call_UnsupportedLanguage(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-2",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"image_source": "http://example.com/test.jpg",
			"language":     "unsupported_lang_xyz",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	resultMap := result.Result

	if success, ok := resultMap["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	if errorMsg, ok := resultMap["error"].(string); !ok || !strings.Contains(errorMsg, "not supported") {
		t.Errorf("Expected language not supported error, got: %v", errorMsg)
	}
}

func TestOCRTool_Call_InvalidImagePath(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-3",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"image_path": "/nonexistent/path/to/image.jpg",
			"language":   "eng",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	resultMap := result.Result

	if success, ok := resultMap["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}
}

func TestOCRTool_Call_WithBase64Image(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	// Create a simple test image with text (1x1 white pixel as placeholder)
	// In a real test, you would use an actual image with text
	imageData := []byte{0xFF, 0xFF, 0xFF} // Simple white pixel
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Data)

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-4",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"image_source": dataURL,
			"language":     "eng",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Note: The OCR might fail on this tiny test image, but it should at least
	// process the base64 decoding without crashing
}

func TestOCRTool_Call_WithURL(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	// Create a test HTTP server that serves a simple image
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		// Simple 1x1 white PNG
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		}
		w.Write(pngData)
	}))
	defer server.Close()

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-5",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"image_source": server.URL,
			"language":     "eng",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Note: The OCR might fail on this test image, but it should at least
	// download and process the image without crashing
}

func TestOCRTool_isLanguageSupported(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	// English should always be supported
	if !tool.isLanguageSupported("eng") {
		t.Error("Expected 'eng' to be supported")
	}

	// Invalid language should not be supported
	if tool.isLanguageSupported("invalid_lang_xyz") {
		t.Error("Expected 'invalid_lang_xyz' to not be supported")
	}
}

func TestOCRTool_decodeBase64Image(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	tests := []struct {
		name      string
		dataURL   string
		expectErr bool
	}{
		{
			name:      "Valid PNG",
			dataURL:   "data:image/png;base64,iVBORw0KGgo=",
			expectErr: false,
		},
		{
			name:      "Valid JPEG",
			dataURL:   "data:image/jpeg;base64,/9j/4AAQ",
			expectErr: false,
		},
		{
			name:      "Invalid format",
			dataURL:   "invalid_data_url",
			expectErr: true,
		},
		{
			name:      "Invalid base64",
			dataURL:   "data:image/png;base64,!!!invalid!!!",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup, err := tool.decodeBase64Image(tt.dataURL)
			defer cleanup()

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if path == "" {
					t.Error("Expected non-empty path")
				}
				// Verify file exists
				if _, statErr := os.Stat(path); statErr != nil {
					t.Errorf("Expected file to exist at path: %s", path)
				}
			}
		})
	}
}

func TestOCRTool_Cleanup(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}

	tempDir := tool.tempDir

	// Verify temp directory exists
	if _, err := os.Stat(tempDir); err != nil {
		t.Fatalf("Expected temp directory to exist: %v", err)
	}

	// Cleanup
	if err := tool.Cleanup(); err != nil {
		t.Errorf("Expected no error during cleanup, got: %v", err)
	}

	// Verify temp directory is removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Expected temp directory to be removed after cleanup")
	}
}

func TestOCRTool_Call_WithLocalFile(t *testing.T) {
	tool, err := NewOCRTool()
	if err != nil {
		t.Skipf("Tesseract not available: %v", err)
		return
	}
	defer tool.Cleanup()

	// Create a temporary test image file
	tempFile := filepath.Join(tool.tempDir, "test_image.png")
	// Simple 1x1 white PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
	}
	if err := os.WriteFile(tempFile, pngData, 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	defer os.Remove(tempFile)

	ctx := context.Background()
	params := &llms.ToolCall{
		ToolCallId: "test-call-6",
		Name:       "ocr_recognize_text",
		Arguments: map[string]any{
			"image_path": tempFile,
			"language":   "eng",
		},
	}

	result, err := tool.Call(ctx, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// The test image is too simple for OCR, so we just verify it doesn't crash
}

