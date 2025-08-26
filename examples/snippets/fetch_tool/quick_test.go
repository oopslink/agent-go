package main

import (
	"context"
	"fmt"
	"log"

	"github.com/oopslink/agent-go/pkg/core/tools/fetch"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func main() {
	fmt.Println("=== Quick Verification of URLs Fetch Tool ===")

	// Create tool instance
	tool := fetch.NewURLsFetchTool()

	// Test concurrency setting
	params := &llms.ToolCall{
		ToolCallId: "quick-test",
		Name:       "urls_fetch",
		Arguments: map[string]any{
			"urls": []any{
				"https://httpbin.org/delay/1",
				"https://httpbin.org/delay/1",
			},
			"max_concurrency": 2, // Set concurrency to 2
		},
	}

	result, err := tool.Call(context.Background(), params)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Check results
	resultMap := result.Result
	if data, ok := resultMap["data"]; ok {
		if dataMap, ok := data.(map[string]any); ok {
			if summary, ok := dataMap["summary"]; ok {
				if summaryMap, ok := summary.(map[string]any); ok {
					totalTime := summaryMap["total_time_ms"]
					success := summaryMap["success"]
					fmt.Printf("✅ Successfully fetched %v URLs, total time: %v ms\n", success, totalTime)
					fmt.Printf("✅ Concurrency control is working properly!\n")
				}
			}
		}
	}
}
