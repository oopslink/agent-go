package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/oopslink/agent-go/pkg/core/tools/fetch"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func main() {
	fmt.Println("=== URLs Fetch Tool Demo ===\n")

	// Create tool instance
	tool := fetch.NewURLsFetchTool()

	// Demo 1: Fetch single HTML page and extract text
	fmt.Println("Demo 1: Fetch single webpage and extract text content")
	demo1(tool)

	// Demo 2: Batch fetch multiple URLs with custom concurrency
	fmt.Println("\nDemo 2: Batch fetch multiple URLs (custom concurrency)")
	demo2(tool)

	// Demo 3: Error handling demonstration
	fmt.Println("\nDemo 3: Error handling demonstration")
	demo3(tool)
}

func demo1(tool *fetch.URLsFetchTool) {
	params := &llms.ToolCall{
		ToolCallId: "demo-1",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls":         []any{"https://httpbin.org/html"},
			"extract_text": true,
			"user_agent":   "URLs Fetch Tool Demo",
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Print result summary
	resultMap := result.Result
	if data, ok := resultMap["data"]; ok {
		if dataMap, ok := data.(map[string]any); ok {
			if summary, ok := dataMap["summary"]; ok {
				summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
				fmt.Printf("Summary:\n%s\n", string(summaryJSON))
			}
			
			if results, ok := dataMap["results"].([]any); ok && len(results) > 0 {
				if resultMap, ok := results[0].(map[string]any); ok {
					title := resultMap["title"]
					textContent := resultMap["text_content"]
					if title != nil {
						fmt.Printf("Page title: %s\n", title)
					}
					if textContent != nil {
						text := textContent.(string)
						if len(text) > 200 {
							text = text[:200] + "..."
						}
						fmt.Printf("Text content: %s\n", text)
					}
				}
			}
		}
	}
}

func demo2(tool *fetch.URLsFetchTool) {
	params := &llms.ToolCall{
		ToolCallId: "demo-2",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{
				"https://httpbin.org/json",
				"https://httpbin.org/xml",
				"https://httpbin.org/user-agent",
				"https://httpbin.org/headers",
			},
			"user_agent":      "Batch Fetch Demo",
			"max_concurrency": 2, // Set max concurrency to 2
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Print summary information only
	resultMap := result.Result
	if data, ok := resultMap["data"]; ok {
		if dataMap, ok := data.(map[string]any); ok {
			if summary, ok := dataMap["summary"]; ok {
				summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
				fmt.Printf("Batch fetch summary (concurrency=2):\n%s\n", string(summaryJSON))
			}
			
			if results, ok := dataMap["results"].([]any); ok {
				fmt.Printf("Fetched %d URLs:\n", len(results))
				for i, result := range results {
					if resultMap, ok := result.(map[string]any); ok {
						url := resultMap["url"]
						statusCode := resultMap["status_code"]
						contentLength := resultMap["content_length"]
						fetchTime := resultMap["fetch_time_ms"]
						fmt.Printf("  %d. %s - Status: %v, Size: %v bytes, Time: %vms\n", 
							i+1, url, statusCode, contentLength, fetchTime)
					}
				}
			}
		}
	}
}

func demo3(tool *fetch.URLsFetchTool) {
	params := &llms.ToolCall{
		ToolCallId: "demo-3",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{
				"https://httpbin.org/status/200",  // Success
				"invalid-url",                      // Invalid URL
				"https://httpbin.org/status/404",  // 404 error
			},
			"max_concurrency": 3, // Set concurrency to 3
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	resultMap := result.Result
	if data, ok := resultMap["data"]; ok {
		if dataMap, ok := data.(map[string]any); ok {
			if summary, ok := dataMap["summary"]; ok {
				summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
				fmt.Printf("Mixed results summary:\n%s\n", string(summaryJSON))
			}
			
			if results, ok := dataMap["results"].([]any); ok {
				fmt.Printf("Detailed results:\n")
				for i, result := range results {
					if resultMap, ok := result.(map[string]any); ok {
						url := resultMap["url"]
						statusCode := resultMap["status_code"]
						errorMsg := resultMap["error"]
						
						if errorMsg != nil && errorMsg != "" {
							fmt.Printf("  %d. %s - ❌ Error: %v\n", i+1, url, errorMsg)
						} else {
							fmt.Printf("  %d. %s - ✅ Success: Status %v\n", i+1, url, statusCode)
						}
					}
				}
			}
		}
	}
}
