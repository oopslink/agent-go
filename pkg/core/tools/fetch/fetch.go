package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
	"golang.org/x/net/html"
)

// URLsFetchTool provides URL fetching capabilities with content extraction
type URLsFetchTool struct {
	client  *http.Client
	timeout time.Duration
}

// NewURLsFetchTool creates a new URLs fetch tool instance
func NewURLsFetchTool() *URLsFetchTool {
	return &URLsFetchTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// WithTimeout sets the timeout for HTTP requests
func (t *URLsFetchTool) WithTimeout(timeout time.Duration) *URLsFetchTool {
	t.timeout = timeout
	t.client.Timeout = timeout
	return t
}

var _ tools.Tool = &URLsFetchTool{}

// FetchParams defines the parameters for URL fetching
type FetchParams struct {
	URLs         []string `json:"urls"`                    // List of URLs to fetch
	ExtractText  bool     `json:"extract_text,omitempty"`  // Whether to extract text content from HTML
	UserAgent    string   `json:"user_agent,omitempty"`    // Custom User-Agent header
	MaxBodySize  int64    `json:"max_body_size,omitempty"` // Maximum response body size in bytes (default: 1MB)
	FollowRedirect bool   `json:"follow_redirect,omitempty"` // Whether to follow redirects (default: true)
	MaxConcurrency int    `json:"max_concurrency,omitempty"` // Maximum concurrent requests (default: 10)
}

// URLResult represents the result of fetching a single URL
type URLResult struct {
	URL          string            `json:"url"`                    // The original URL
	StatusCode   int               `json:"status_code"`            // HTTP status code
	Headers      map[string]string `json:"headers,omitempty"`      // Response headers
	ContentType  string            `json:"content_type,omitempty"` // Content-Type header
	ContentLength int64            `json:"content_length"`         // Content length in bytes
	Content      string            `json:"content,omitempty"`      // Raw content
	TextContent  string            `json:"text_content,omitempty"` // Extracted text content (if extract_text=true)
	Title        string            `json:"title,omitempty"`        // Page title (if HTML)
	Error        string            `json:"error,omitempty"`        // Error message if fetching failed
	FetchTime    int64             `json:"fetch_time_ms"`          // Time taken to fetch in milliseconds
}

// FetchResult represents the overall result of the fetch operation
type FetchResult struct {
	Results []URLResult `json:"results"`
	Summary struct {
		Total     int `json:"total"`
		Success   int `json:"success"`
		Failed    int `json:"failed"`
		TotalTime int64 `json:"total_time_ms"`
	} `json:"summary"`
}

func (t *URLsFetchTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "urls_fetch",
		Description: "Fetch content from multiple URLs concurrently. Can extract text content from HTML pages and return structured information about each URL.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"urls": {
					Type:        llms.TypeArray,
					Description: "List of URLs to fetch",
					Items: &llms.Schema{
						Type: llms.TypeString,
					},
				},
				"extract_text": {
					Type:        llms.TypeBoolean,
					Description: "Whether to extract plain text content from HTML pages (default: false)",
				},
				"user_agent": {
					Type:        llms.TypeString,
					Description: "Custom User-Agent header for requests (default: agent-go/1.0 URLsFetchTool)",
				},
				"max_body_size": {
					Type:        llms.TypeInteger,
					Description: "Maximum response body size in bytes (default: 1048576 = 1MB)",
				},
				"follow_redirect": {
					Type:        llms.TypeBoolean,
					Description: "Whether to follow HTTP redirects (default: true)",
				},
				"max_concurrency": {
					Type:        llms.TypeInteger,
					Description: "Maximum number of concurrent requests (default: 10)",
				},
			},
			Required: []string{"urls"},
		},
	}
}

func (t *URLsFetchTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	var fetchParams FetchParams
	if err := mapToStruct(params.Arguments, &fetchParams); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate parameters
	if len(fetchParams.URLs) == 0 {
		return nil, fmt.Errorf("urls parameter is required and cannot be empty")
	}

	// Set defaults
	if fetchParams.MaxBodySize <= 0 {
		fetchParams.MaxBodySize = 1024 * 1024 // 1MB default
	}
	if fetchParams.UserAgent == "" {
		fetchParams.UserAgent = "agent-go/1.0 URLsFetchTool"
	}
	if fetchParams.MaxConcurrency <= 0 {
		fetchParams.MaxConcurrency = 10 // Default to 10 concurrent requests
	}

	// Configure HTTP client
	client := &http.Client{
		Timeout: t.timeout,
	}
	if !fetchParams.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	startTime := time.Now()
	results := t.fetchURLsConcurrently(ctx, client, fetchParams)
	totalTime := time.Since(startTime).Milliseconds()

	// Calculate summary
	summary := struct {
		Total     int   `json:"total"`
		Success   int   `json:"success"`
		Failed    int   `json:"failed"`
		TotalTime int64 `json:"total_time_ms"`
	}{
		Total:     len(results),
		TotalTime: totalTime,
	}

	for _, result := range results {
		if result.Error == "" {
			summary.Success++
		} else {
			summary.Failed++
		}
	}

	fetchResult := FetchResult{
		Results: results,
		Summary: summary,
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"data":    fetchResult,
		},
	}, nil
}

// fetchURLsConcurrently fetches multiple URLs concurrently
func (t *URLsFetchTool) fetchURLsConcurrently(ctx context.Context, client *http.Client, params FetchParams) []URLResult {
	results := make([]URLResult, len(params.URLs))
	var wg sync.WaitGroup
	
	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, params.MaxConcurrency)

	for i, urlStr := range params.URLs {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			results[index] = t.fetchSingleURL(ctx, client, url, params)
		}(i, urlStr)
	}

	wg.Wait()
	return results
}

// fetchSingleURL fetches content from a single URL
func (t *URLsFetchTool) fetchSingleURL(ctx context.Context, client *http.Client, urlStr string, params FetchParams) URLResult {
	startTime := time.Now()
	result := URLResult{
		URL: urlStr,
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		result.Error = fmt.Sprintf("invalid URL: %v", err)
		result.FetchTime = time.Since(startTime).Milliseconds()
		return result
	}

	// Only allow HTTP and HTTPS schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.Error = "only HTTP and HTTPS URLs are supported"
		result.FetchTime = time.Since(startTime).Milliseconds()
		return result
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		result.FetchTime = time.Since(startTime).Milliseconds()
		return result
	}

	// Set User-Agent
	req.Header.Set("User-Agent", params.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		result.FetchTime = time.Since(startTime).Milliseconds()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")

	// Extract headers
	result.Headers = make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	// Read body with size limit
	limitedReader := io.LimitReader(resp.Body, params.MaxBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read response body: %v", err)
		result.FetchTime = time.Since(startTime).Milliseconds()
		return result
	}

	result.ContentLength = int64(len(body))
	result.Content = string(body)

	// Extract text content if requested and content is HTML
	if params.ExtractText && isHTMLContent(result.ContentType) {
		textContent, title := extractTextFromHTML(result.Content)
		result.TextContent = textContent
		result.Title = title
	}

	result.FetchTime = time.Since(startTime).Milliseconds()
	return result
}

// isHTMLContent checks if the content type indicates HTML
func isHTMLContent(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html") ||
		   strings.Contains(strings.ToLower(contentType), "application/xhtml")
}

// extractTextFromHTML extracts plain text and title from HTML content
func extractTextFromHTML(htmlContent string) (text, title string) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent, "" // Return original content if parsing fails
	}

	var textBuilder strings.Builder
	var titleText string

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "title":
				if titleText == "" { // Get the first title
					titleText = getTextContent(n)
				}
			case "script", "style", "noscript":
				return // Skip these elements
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				if textBuilder.Len() > 0 {
					textBuilder.WriteString(" ")
				}
				textBuilder.WriteString(text)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Clean up extracted text
	text = cleanText(textBuilder.String())
	title = strings.TrimSpace(titleText)

	return text, title
}

// getTextContent gets the text content of an HTML node
func getTextContent(n *html.Node) string {
	var textBuilder strings.Builder
	
	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			textBuilder.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(n)
	return textBuilder.String()
}

// cleanText cleans up extracted text by removing extra whitespace
func cleanText(text string) string {
	// Replace multiple whitespace with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	// Trim leading and trailing whitespace
	text = strings.TrimSpace(text)
	
	return text
}

// mapToStruct converts a map[string]any to a struct using JSON marshaling/unmarshaling
func mapToStruct(m map[string]any, target interface{}) error {
	if m == nil {
		return nil
	}

	// Convert map to JSON bytes
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	// Convert JSON bytes to target struct
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return nil
}
