package ocr

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/core/tools"
	amcp "github.com/oopslink/agent-go/pkg/core/mcp"
	"k8s.io/klog/v2"
)

// WithOCRMcpServer creates an MCP server with OCR tools
// It returns the tools wrapped in the MCP protocol
func WithOCRMcpServer(ctx context.Context) ([]tools.Tool, error) {
	server := newOCRMcpServer()

	clientRef, serverRef, err := amcp.WithInMemory()
	if err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to create in-memory transport: %s", err.Error())
	}

	// Run the server in the background
	go func() {
		klog.Info("Starting MCP server for OCR tools")
		defer klog.Info("MCP server for OCR tools stopped")

		transport, err := serverRef.CreateTransport()
		if err != nil {
			klog.Errorf("Failed to create transport: %v", err)
			return
		}

		if err := server.Run(ctx, transport); err != nil {
			klog.Errorf("MCP server error: %v", err)
		}
	}()

	// Create a client that connects to the server
	return amcp.WithMcpTools(clientRef)
}

// newOCRMcpServer creates a new MCP server with OCR tools
func newOCRMcpServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ocr-mcp-tools",
		Version: "v0.0.1",
	}, nil)

	// Add OCR tool
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "ocr_recognize_text",
			Description: "Extract text from images using Optical Character Recognition (OCR). Supports multiple languages and various image formats (JPEG, PNG, GIF, BMP, TIFF). Can process images from URLs or base64 encoded data.",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"image_source": {
						Type:        "string",
						Description: "Image source: either a URL (http://... or https://...) or base64 encoded image data with prefix (data:image/...)",
					},
					"language": {
						Type:        "string",
						Description: "Language code for OCR (e.g., 'eng' for English, 'chi_sim' for Simplified Chinese, 'jpn' for Japanese). Default is 'eng'",
					},
					"image_path": {
						Type:        "string",
						Description: "Local file path to the image (alternative to image_source)",
					},
				},
			},
		},
		handleOCR,
	)

	return server
}

// ocrParams defines the parameters for OCR tool
type ocrParams struct {
	ImageSource string `json:"image_source,omitempty"`
	Language    string `json:"language,omitempty"`
	ImagePath   string `json:"image_path,omitempty"`
}

// handleOCR handles OCR requests from the MCP client
func handleOCR(ctx context.Context, session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[ocrParams]) (*mcp.CallToolResultFor[any], error) {

	// Create OCR tool instance
	tool, err := NewOCRTool()
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to initialize OCR tool: %v. Please ensure tesseract-ocr is installed.", err),
				},
			},
			IsError: true,
		}, nil
	}
	defer tool.Cleanup()

	// Set defaults
	language := "eng"
	if params.Arguments.Language != "" {
		language = params.Arguments.Language
	}

	// Validate that at least one image source is provided
	if params.Arguments.ImageSource == "" && params.Arguments.ImagePath == "" {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Error: either image_source or image_path must be provided",
				},
			},
			IsError: true,
		}, nil
	}

	// Validate language is supported
	if !tool.isLanguageSupported(language) {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Error: language '%s' is not supported. Supported languages: %v",
						language, tool.supportedLangs),
				},
			},
			IsError: true,
		}, nil
	}

	// Get the local image path
	localImagePath, cleanup, err := tool.getImagePath(ctx, params.Arguments.ImageSource, params.Arguments.ImagePath)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to process image: %v", err),
				},
			},
			IsError: true,
		}, nil
	}
	defer cleanup()

	// Perform OCR
	text, err := tool.performOCR(ctx, localImagePath, language)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("OCR failed: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Format the result
	resultText := fmt.Sprintf("OCR Result (Language: %s):\n\n%s", language, text)

	if params.Arguments.ImageSource != "" {
		resultText = fmt.Sprintf("Image URL: %s\n\n%s", params.Arguments.ImageSource, resultText)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: resultText,
			},
		},
	}, nil
}

