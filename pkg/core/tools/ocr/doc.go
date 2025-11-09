/*
Package ocr provides Optical Character Recognition (OCR) tools for extracting text from images.

The package supports the Model Context Protocol (MCP) for integration with AI agents and LLM applications.

Features:
  - Extract text from images using Tesseract OCR
  - Support for multiple languages (eng, chi_sim, jpn, etc.)
  - Multiple image sources: URLs, base64 encoded data, local file paths
  - Support for various image formats: JPEG, PNG, GIF, BMP, TIFF
  - MCP protocol integration for seamless agent integration

Requirements:
  - Tesseract OCR must be installed on the system
  - Install on macOS: brew install tesseract
  - Install on Ubuntu/Debian: apt-get install tesseract-ocr
  - Install on Windows: Download from https://github.com/tesseract-ocr/tesseract

Usage as a standalone tool:

	// Create an OCR tool instance
	tool, err := ocr.NewOCRTool()
	if err != nil {
		log.Fatal(err)
	}
	defer tool.Cleanup()

	// Call the tool
	result, err := tool.Call(ctx, &llms.ToolCall{
		Name: "ocr_recognize_text",
		Arguments: map[string]any{
			"image_source": "https://example.com/image.jpg",
			"language":     "eng",
		},
	})

Usage with MCP protocol:

	// Setup MCP server with OCR tools
	ctx := context.Background()
	tools, err := ocr.WithOCRMcpServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Use the tools with an agent
	agent, err := agent.NewAgent(
		agent.WithTools(tools...),
		// ... other agent options
	)

Supported Languages:
  - English (eng)
  - Simplified Chinese (chi_sim)
  - Traditional Chinese (chi_tra)
  - Japanese (jpn)
  - Korean (kor)
  - And many more (depends on installed Tesseract language packs)

Image Sources:
  - HTTP/HTTPS URLs: "https://example.com/image.jpg"
  - Base64 encoded: "data:image/png;base64,iVBORw0KGgo..."
  - Local file paths: "/path/to/image.jpg"

Error Handling:
The tool gracefully handles errors and returns structured error messages when:
  - Tesseract is not installed
  - Image cannot be downloaded or decoded
  - Unsupported language is specified
  - OCR processing fails

For more information about MCP protocol integration, see the mcp.go file.
*/
package ocr
