package tools

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// NewDuckDuckGoTool creates a new DuckDuckGo tool instance
func NewDuckDuckGoTool() *DuckDuckGoTool {
	return &DuckDuckGoTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// SearchResponse represents the response from a DuckDuckGo search
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
}

// Ensure the tool implements the Tool interface
var _ tools.Tool = &DuckDuckGoTool{}

// DuckDuckGoTool represents a tool for searching DuckDuckGo
type DuckDuckGoTool struct {
	client *http.Client
}

// Call implements the Tool interface
func (t *DuckDuckGoTool) Call(ctx context.Context, params *llms.ToolCall) (*llms.ToolCallResult, error) {
	// Extract query from parameters
	queryArg, ok := params.Arguments["query"]
	if !ok {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "query parameter is required",
			},
		}, nil
	}

	query, ok := queryArg.(string)
	if !ok {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "query parameter must be a string",
			},
		}, nil
	}

	if strings.TrimSpace(query) == "" {
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   "query cannot be empty",
			},
		}, nil
	}

	// Perform the search
	results, err := t.search(ctx, query)
	if err != nil {
		klog.Errorf("duckduckgo search failed: %v", err)
		return &llms.ToolCallResult{
			ToolCallId: params.ToolCallId,
			Name:       params.Name,
			Result: map[string]any{
				"success": false,
				"error":   fmt.Sprintf("search failed: %v", err),
			},
		}, nil
	}

	return &llms.ToolCallResult{
		ToolCallId: params.ToolCallId,
		Name:       params.Name,
		Result: map[string]any{
			"success": true,
			"query":   results.Query,
			"results": results.Results,
			"count":   results.Count,
		},
	}, nil
}

// Descriptor implements the Tool interface
func (t *DuckDuckGoTool) Descriptor() *llms.ToolDescriptor {
	return &llms.ToolDescriptor{
		Name:        "duckduckgo_search",
		Description: "Search DuckDuckGo for information. Returns search results with titles, URLs, and descriptions.",
		Parameters: &llms.Schema{
			Type: llms.TypeObject,
			Properties: map[string]*llms.Schema{
				"query": {
					Type:        llms.TypeString,
					Description: "The search query to perform on DuckDuckGo",
				},
			},
			Required: []string{"query"},
		},
	}
}

// search performs the actual DuckDuckGo search and parses the results
func (t *DuckDuckGoTool) search(ctx context.Context, query string) (*SearchResponse, error) {
	// Construct the DuckDuckGo search URL
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to create request: %s", err.Error())
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// Removed Accept-Encoding to avoid compression issues
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make the request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to make request: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"unexpected status code: %d", resp.StatusCode)
	}

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Errorf(tools.ErrorCodeToolCallFailed,
			"failed to parse HTML: %s", err.Error())
	}

	// Extract search results
	var results []SearchResult

	// DuckDuckGo HTML structure: results are in .result elements
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		// Extract title and URL
		titleElement := s.Find(".result__title a")
		title := strings.TrimSpace(titleElement.Text())
		href, exists := titleElement.Attr("href")

		if !exists || title == "" {
			return
		}

		// Extract description
		description := strings.TrimSpace(s.Find(".result__snippet").Text())

		// Clean up the URL (DuckDuckGo uses redirect URLs)
		cleanURL := t.cleanURL(href)

		results = append(results, SearchResult{
			Title:       title,
			URL:         cleanURL,
			Description: description,
		})
	})

	// If no results found with .result, try alternative selectors
	if len(results) == 0 {
		doc.Find(".web-result").Each(func(i int, s *goquery.Selection) {
			titleElement := s.Find(".web-result__title a")
			title := strings.TrimSpace(titleElement.Text())
			href, exists := titleElement.Attr("href")

			if !exists || title == "" {
				return
			}

			description := strings.TrimSpace(s.Find(".web-result__snippet").Text())
			cleanURL := t.cleanURL(href)

			results = append(results, SearchResult{
				Title:       title,
				URL:         cleanURL,
				Description: description,
			})
		})
	}

	return &SearchResponse{
		Query:   query,
		Results: results,
		Count:   len(results),
		Success: true,
	}, nil
}

// cleanURL attempts to extract the actual URL from DuckDuckGo's redirect URL
func (t *DuckDuckGoTool) cleanURL(href string) string {
	// DuckDuckGo uses redirect URLs like: /l/?uddg=...
	// Try to extract the actual URL from the redirect
	if strings.Contains(href, "uddg=") {
		if parsedURL, err := url.Parse(href); err == nil {
			if uddg := parsedURL.Query().Get("uddg"); uddg != "" {
				if decodedURL, err := url.QueryUnescape(uddg); err == nil {
					return decodedURL
				}
			}
		}
	}

	// If it's already a full URL, return as is
	if strings.HasPrefix(href, "http") {
		return href
	}

	// Otherwise, construct the full DuckDuckGo URL
	if strings.HasPrefix(href, "/") {
		return "https://duckduckgo.com" + href
	}

	return href
}
