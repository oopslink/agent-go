# URLs Fetch Tool Example

This example demonstrates how to use the URLs Fetch Tool to batch fetch web content.

## Features Demonstrated

1. **Single webpage fetching**: Fetch HTML pages and extract text content and titles
2. **Batch concurrent fetching**: Fetch multiple URLs simultaneously with custom concurrency settings
3. **Error handling**: Demonstrate how to handle invalid URLs and network errors

## Running the Example

```bash
cd /workspaces/agent-go/examples/snippets/fetch_tool
go run main.go
```

## Key Features

- **Configurable concurrency**: Control the number of simultaneous requests via `max_concurrency` parameter (default 10)
- **HTML text extraction**: Automatically extract plain text and titles from HTML pages
- **Error recovery**: Single URL failures don't affect other URL processing
- **Detailed metadata**: Returns HTTP status codes, response headers, timing information, etc.

## Sample Output

The program demonstrates three different usage scenarios:

1. Fetch a single HTML page's title and text content
2. Batch fetch multiple URLs using custom concurrency settings
3. Mixed success and failure requests to showcase error handling
