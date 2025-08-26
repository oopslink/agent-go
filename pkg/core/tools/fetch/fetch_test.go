package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oopslink/agent-go/pkg/support/llms"
)

func TestURLsFetchTool_Descriptor(t *testing.T) {
	tool := NewURLsFetchTool()
	descriptor := tool.Descriptor()

	if descriptor.Name != "urls_fetch" {
		t.Errorf("expected name 'urls_fetch', got '%s'", descriptor.Name)
	}

	if descriptor.Description == "" {
		t.Error("expected non-empty description")
	}

	if descriptor.Parameters == nil {
		t.Error("expected parameters schema")
	}

	if descriptor.Parameters.Type != llms.TypeObject {
		t.Errorf("expected parameters type 'object', got '%s'", descriptor.Parameters.Type)
	}

	// Check required parameters
	if len(descriptor.Parameters.Required) != 1 {
		t.Errorf("expected 1 required parameter, got %d", len(descriptor.Parameters.Required))
	}

	if descriptor.Parameters.Required[0] != "urls" {
		t.Errorf("expected required parameter 'urls', got '%s'", descriptor.Parameters.Required[0])
	}
}

func TestURLsFetchTool_Call_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<h1>Welcome</h1>
				<p>This is a test page with some content.</p>
				<script>console.log('test');</script>
			</body>
			</html>
		`))
	}))
	defer server.Close()

	tool := NewURLsFetchTool()
	
	params := &llms.ToolCall{
		ToolCallId: "test-123",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls":         []any{server.URL},
			"extract_text": true,
			"user_agent":   "test-agent",
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ToolCallId != "test-123" {
		t.Errorf("expected tool call id 'test-123', got '%s'", result.ToolCallId)
	}

	if result.Name != "urls_fetch" {
		t.Errorf("expected name 'urls_fetch', got '%s'", result.Name)
	}

	// Check result structure
	resultMap := result.Result
	if resultMap == nil {
		t.Fatal("expected result to be non-nil")
	}

	if success, ok := resultMap["success"].(bool); !ok || !success {
		t.Error("expected success to be true")
	}

	data, ok := resultMap["data"]
	if !ok {
		t.Fatal("expected data field in result")
	}

	// Convert data to JSON and back to verify structure
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}

	var fetchResult FetchResult
	if err := json.Unmarshal(dataJSON, &fetchResult); err != nil {
		t.Fatalf("failed to unmarshal fetch result: %v", err)
	}

	// Verify summary
	if fetchResult.Summary.Total != 1 {
		t.Errorf("expected total 1, got %d", fetchResult.Summary.Total)
	}

	if fetchResult.Summary.Success != 1 {
		t.Errorf("expected success 1, got %d", fetchResult.Summary.Success)
	}

	if fetchResult.Summary.Failed != 0 {
		t.Errorf("expected failed 0, got %d", fetchResult.Summary.Failed)
	}

	// Verify URL result
	if len(fetchResult.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(fetchResult.Results))
	}

	urlResult := fetchResult.Results[0]
	if urlResult.URL != server.URL {
		t.Errorf("expected URL '%s', got '%s'", server.URL, urlResult.URL)
	}

	if urlResult.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", urlResult.StatusCode)
	}

	if urlResult.Error != "" {
		t.Errorf("expected no error, got '%s'", urlResult.Error)
	}

	if urlResult.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", urlResult.Title)
	}

	if !strings.Contains(urlResult.TextContent, "Welcome") {
		t.Errorf("expected text content to contain 'Welcome', got '%s'", urlResult.TextContent)
	}

	if !strings.Contains(urlResult.TextContent, "test page") {
		t.Errorf("expected text content to contain 'test page', got '%s'", urlResult.TextContent)
	}

	// Script content should be filtered out
	if strings.Contains(urlResult.TextContent, "console.log") {
		t.Errorf("expected script content to be filtered out, but found in text: '%s'", urlResult.TextContent)
	}
}

func TestURLsFetchTool_Call_MultipleURLs(t *testing.T) {
	// Create multiple test servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server 1 response"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Server 2 response"}`))
	}))
	defer server2.Close()

	tool := NewURLsFetchTool()
	
	params := &llms.ToolCall{
		ToolCallId: "test-456",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{server1.URL, server2.URL},
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Convert result for verification
	resultMap := result.Result
	dataJSON, _ := json.Marshal(resultMap["data"])
	var fetchResult FetchResult
	json.Unmarshal(dataJSON, &fetchResult)

	// Verify summary
	if fetchResult.Summary.Total != 2 {
		t.Errorf("expected total 2, got %d", fetchResult.Summary.Total)
	}

	if fetchResult.Summary.Success != 2 {
		t.Errorf("expected success 2, got %d", fetchResult.Summary.Success)
	}

	// Verify both results
	if len(fetchResult.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(fetchResult.Results))
	}

	for _, urlResult := range fetchResult.Results {
		if urlResult.StatusCode != 200 {
			t.Errorf("expected status code 200 for URL %s, got %d", urlResult.URL, urlResult.StatusCode)
		}
		if urlResult.Error != "" {
			t.Errorf("expected no error for URL %s, got '%s'", urlResult.URL, urlResult.Error)
		}
	}
}

func TestURLsFetchTool_Call_ErrorHandling(t *testing.T) {
	tool := NewURLsFetchTool()
	
	params := &llms.ToolCall{
		ToolCallId: "test-error",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{"invalid-url", "http://non-existent-domain-12345.com"},
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Convert result for verification
	resultMap := result.Result
	dataJSON, _ := json.Marshal(resultMap["data"])
	var fetchResult FetchResult
	json.Unmarshal(dataJSON, &fetchResult)

	// Verify summary shows failures
	if fetchResult.Summary.Total != 2 {
		t.Errorf("expected total 2, got %d", fetchResult.Summary.Total)
	}

	if fetchResult.Summary.Failed != 2 {
		t.Errorf("expected failed 2, got %d", fetchResult.Summary.Failed)
	}

	if fetchResult.Summary.Success != 0 {
		t.Errorf("expected success 0, got %d", fetchResult.Summary.Success)
	}

	// Verify both results have errors
	for _, urlResult := range fetchResult.Results {
		if urlResult.Error == "" {
			t.Errorf("expected error for URL %s, but got none", urlResult.URL)
		}
	}
}

func TestURLsFetchTool_Call_InvalidParameters(t *testing.T) {
	tool := NewURLsFetchTool()
	
	// Test missing URLs
	params := &llms.ToolCall{
		ToolCallId: "test-invalid",
		Name:       "urls_fetch",
		Arguments:  map[string]any{},
	}

	_, err := tool.Call(context.Background(), params)
	if err == nil {
		t.Error("expected error for missing URLs parameter")
	}

	// Test empty URLs array
	params.Arguments = map[string]any{
		"urls": []any{},
	}

	_, err = tool.Call(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty URLs array")
	}
}

func TestURLsFetchTool_WithTimeout(t *testing.T) {
	tool := NewURLsFetchTool().WithTimeout(1 * time.Second)
	
	if tool.timeout != 1*time.Second {
		t.Errorf("expected timeout 1s, got %v", tool.timeout)
	}

	if tool.client.Timeout != 1*time.Second {
		t.Errorf("expected client timeout 1s, got %v", tool.client.Timeout)
	}
}

func TestExtractTextFromHTML(t *testing.T) {
	htmlContent := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Title</title>
			<style>body { color: red; }</style>
		</head>
		<body>
			<h1>Main Heading</h1>
			<p>This is a paragraph with <strong>bold</strong> text.</p>
			<script>alert('test');</script>
			<div>
				<span>Nested content</span>
			</div>
		</body>
		</html>
	`

	text, title := extractTextFromHTML(htmlContent)

	if title != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", title)
	}

	expectedParts := []string{"Main Heading", "This is a paragraph", "bold text", "Nested content"}
	for _, part := range expectedParts {
		if !strings.Contains(text, part) {
			t.Errorf("expected text to contain '%s', got: %s", part, text)
		}
	}

	// Should not contain script or style content
	if strings.Contains(text, "alert") || strings.Contains(text, "color: red") {
		t.Errorf("text should not contain script or style content, got: %s", text)
	}
}

func TestIsHTMLContent(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"text/html", true},
		{"text/html; charset=utf-8", true},
		{"application/xhtml+xml", true},
		{"text/plain", false},
		{"application/json", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isHTMLContent(tt.contentType)
		if result != tt.expected {
			t.Errorf("isHTMLContent(%q) = %v, expected %v", tt.contentType, result, tt.expected)
		}
	}
}

func TestURLsFetchTool_Call_CustomConcurrency(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add a small delay to test concurrency
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	}))
	defer server.Close()

	tool := NewURLsFetchTool()
	
	// Test with custom concurrency
	params := &llms.ToolCall{
		ToolCallId: "test-concurrency",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{
				server.URL + "/1",
				server.URL + "/2",
				server.URL + "/3",
				server.URL + "/4",
			},
			"max_concurrency": 2, // Set concurrency to 2
		},
	}

	startTime := time.Now()
	result, err := tool.Call(context.Background(), params)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Convert result for verification
	resultMap := result.Result
	dataJSON, _ := json.Marshal(resultMap["data"])
	var fetchResult FetchResult
	json.Unmarshal(dataJSON, &fetchResult)

	// Verify all requests succeeded
	if fetchResult.Summary.Success != 4 {
		t.Errorf("expected 4 successful requests, got %d", fetchResult.Summary.Success)
	}

	// With concurrency=2 and 4 URLs with 100ms delay each,
	// it should take roughly 200ms (2 batches) rather than 400ms (sequential)
	// Allow some buffer for processing time
	if duration > 350*time.Millisecond {
		t.Errorf("requests took too long (%v), concurrency may not be working", duration)
	}

	fmt.Printf("Completed 4 requests with concurrency=2 in %v\n", duration)
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello   world  ", "hello world"},
		{"line1\n\nline2\t\tline3", "line1 line2 line3"},
		{"", ""},
		{"   ", ""},
		{"normal text", "normal text"},
	}

	for _, tt := range tests {
		result := cleanText(tt.input)
		if result != tt.expected {
			t.Errorf("cleanText(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
