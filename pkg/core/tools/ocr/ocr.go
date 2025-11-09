package ocr

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// OCRTool performs optical character recognition on images
type OCRTool struct {
	tempDir        string
	tesseractPath  string
	supportedLangs []string
}

// OCRResult represents the result of an OCR operation
type OCRResult struct {
	Text     string   `json:"text"`
	Language string   `json:"language"`
	Success  bool     `json:"success"`
	Error    string   `json:"error,omitempty"`
	ImageURL string   `json:"image_url,omitempty"`
	Lines    []string `json:"lines,omitempty"`
}

// NewOCRTool creates a new OCR tool instance
func NewOCRTool() (*OCRTool, error) {
	// Find tesseract executable
	tesseractPath, err := exec.LookPath("tesseract")
	if err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"tesseract not found: %s. Please install tesseract-ocr", err.Error())
	}

	// Create temp directory for image processing
	tempDir := filepath.Join(os.TempDir(), "agent-go-ocr")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to create temp directory: %s", err.Error())
	}

	// Get supported languages
	supportedLangs, err := getSupportedLanguages(tesseractPath)
	if err != nil {
		klog.Warningf("Failed to get supported languages: %v", err)
		supportedLangs = []string{"eng"} // Default to English
	}

	return &OCRTool{
		tempDir:        tempDir,
		tesseractPath:  tesseractPath,
		supportedLangs: supportedLangs,
	}, nil
}

// Descriptor implements the Tool interface
func (t *OCRTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "ocr_recognize_text",
		Description: "Extract text from images using Optical Character Recognition (OCR). Supports multiple languages and various image formats (JPEG, PNG, GIF, BMP, TIFF). Can process images from URLs or base64 encoded data.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"image_source": {
					Type:        llms.TypeString,
					Description: "Image source: either a URL (http://... or https://...) or base64 encoded image data with prefix (data:image/...)",
				},
				"language": {
					Type:        llms.TypeString,
					Description: "Language code for OCR (e.g., 'eng' for English, 'chi_sim' for Simplified Chinese, 'jpn' for Japanese). Default is 'eng'",
				},
				"image_path": {
					Type:        llms.TypeString,
					Description: "Local file path to the image (alternative to image_source)",
				},
			},
			Required: []string{},
		},
	}
}

// Call implements the Tool interface
func (t *OCRTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Extract parameters
	imageSource := ""
	imagePath := ""
	language := "eng" // Default to English

	if src, ok := params.Arguments["image_source"].(string); ok {
		imageSource = src
	}
	if path, ok := params.Arguments["image_path"].(string); ok {
		imagePath = path
	}
	if lang, ok := params.Arguments["language"].(string); ok && lang != "" {
		language = lang
	}

	// Validate that at least one image source is provided
	if imageSource == "" && imagePath == "" {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "either image_source or image_path must be provided",
			},
		}, nil
	}

	// Validate language is supported
	if !t.isLanguageSupported(language) {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   fmt.Sprintf("language '%s' is not supported. Supported languages: %v", language, t.supportedLangs),
			},
		}, nil
	}

	// Get the local image path
	localImagePath, cleanup, err := t.getImagePath(ctx, imageSource, imagePath)
	if err != nil {
		klog.Errorf("Failed to get image path: %v", err)
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   fmt.Sprintf("failed to process image: %v", err),
			},
		}, nil
	}
	defer cleanup()

	// Perform OCR
	text, err := t.performOCR(ctx, localImagePath, language)
	if err != nil {
		klog.Errorf("OCR failed: %v", err)
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   fmt.Sprintf("OCR failed: %v", err),
			},
		}, nil
	}

	// Process the result
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			nonEmptyLines = append(nonEmptyLines, trimmed)
		}
	}

	result := map[string]any{
		"success":  true,
		"text":     strings.TrimSpace(text),
		"language": language,
		"lines":    nonEmptyLines,
	}

	if imageSource != "" {
		result["image_url"] = imageSource
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result:     result,
	}, nil
}

// getImagePath returns a local file path for the image, downloading if necessary
func (t *OCRTool) getImagePath(ctx context.Context, imageSource, imagePath string) (string, func(), error) {
	cleanup := func() {}

	// If local path is provided, use it directly
	if imagePath != "" {
		if _, err := os.Stat(imagePath); err != nil {
			return "", cleanup, errors.Errorf(tools.ErrorCodeToolCallFailed,
				"image file not found: %s", imagePath)
		}
		return imagePath, cleanup, nil
	}

	// Handle base64 encoded images
	if strings.HasPrefix(imageSource, "data:image/") {
		return t.decodeBase64Image(imageSource)
	}

	// Handle URL
	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		return t.downloadImage(ctx, imageSource)
	}

	return "", cleanup, errors.Errorf(tools.ErrorCodeToolCallFailed,
		"invalid image source format")
}

// decodeBase64Image decodes a base64 encoded image and saves it to a temp file
func (t *OCRTool) decodeBase64Image(dataURL string) (string, func(), error) {
	// Extract base64 data
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"invalid base64 image format")
	}

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to decode base64 image: %s", err.Error())
	}

	// Determine file extension from MIME type
	ext := ".png"
	if strings.Contains(parts[0], "jpeg") || strings.Contains(parts[0], "jpg") {
		ext = ".jpg"
	} else if strings.Contains(parts[0], "gif") {
		ext = ".gif"
	} else if strings.Contains(parts[0], "bmp") {
		ext = ".bmp"
	}

	// Save to temp file
	tempFile := filepath.Join(t.tempDir, fmt.Sprintf("ocr_%s%s", uuid.New().String(), ext))
	if err := os.WriteFile(tempFile, imageData, 0644); err != nil {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to save image: %s", err.Error())
	}

	cleanup := func() {
		os.Remove(tempFile)
	}

	return tempFile, cleanup, nil
}

// downloadImage downloads an image from a URL and saves it to a temp file
func (t *OCRTool) downloadImage(ctx context.Context, url string) (string, func(), error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to create request: %s", err.Error())
	}

	// Download image
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to download image: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to download image: status code %d", resp.StatusCode)
	}

	// Determine file extension from content type
	ext := ".png"
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
		ext = ".jpg"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	} else if strings.Contains(contentType, "bmp") {
		ext = ".bmp"
	}

	// Save to temp file
	tempFile := filepath.Join(t.tempDir, fmt.Sprintf("ocr_%s%s", uuid.New().String(), ext))
	file, err := os.Create(tempFile)
	if err != nil {
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to create temp file: %s", err.Error())
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(tempFile)
		return "", func() {}, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to save image: %s", err.Error())
	}

	cleanup := func() {
		os.Remove(tempFile)
	}

	return tempFile, cleanup, nil
}

// performOCR runs tesseract on the image
func (t *OCRTool) performOCR(ctx context.Context, imagePath, language string) (string, error) {
	// Create output file path (without extension, tesseract adds .txt)
	outputBase := filepath.Join(t.tempDir, fmt.Sprintf("ocr_output_%s", uuid.New().String()))
	outputFile := outputBase + ".txt"
	defer os.Remove(outputFile)

	// Run tesseract
	cmd := exec.CommandContext(ctx, t.tesseractPath, imagePath, outputBase, "-l", language)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Errorf(tools.ErrorCodeToolCallFailed,
			"tesseract failed: %s, output: %s", err.Error(), string(output))
	}

	// Read the output file
	textData, err := os.ReadFile(outputFile)
	if err != nil {
		return "", errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to read OCR output: %s", err.Error())
	}

	return string(textData), nil
}

// isLanguageSupported checks if a language is supported
func (t *OCRTool) isLanguageSupported(lang string) bool {
	for _, supported := range t.supportedLangs {
		if supported == lang {
			return true
		}
	}
	return false
}

// getSupportedLanguages queries tesseract for supported languages
func getSupportedLanguages(tesseractPath string) ([]string, error) {
	cmd := exec.Command(tesseractPath, "--list-langs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var langs []string
	// Skip the first line (header)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		langs = append(langs, strings.TrimSpace(line))
	}

	return langs, nil
}

// Cleanup removes temporary files
func (t *OCRTool) Cleanup() error {
	return os.RemoveAll(t.tempDir)
}

// Ensure OCRTool implements the Tool interface
var _ tools.Tool = &OCRTool{}
