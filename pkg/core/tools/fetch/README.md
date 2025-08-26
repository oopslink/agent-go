# URLs Fetch Tool

A powerful URL batch fetching tool designed for AI Agents, supporting concurrent requests, HTML text extraction, and error handling.

## Features

- **Batch concurrent fetching**: Process multiple URLs simultaneously for improved efficiency
- **HTML text extraction**: Automatically extract webpage titles and plain text content
- **Error handling**: Gracefully handle various network errors and invalid URLs
- **Configurable options**: Support custom User-Agent, timeout, redirects, etc.
- **Safety limits**: Limit response body size to prevent memory overflow
- **Rich metadata**: Return HTTP status codes, response headers, content types, etc.

## Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/oopslink/agent-go/pkg/core/tools/fetch"
    "github.com/oopslink/agent-go/pkg/support/llms"
)

func main() {
    // Create tool instance
    tool := fetch.NewURLsFetchTool()
    
    // Configure tool parameters
    params := &llms.ToolCall{
        ToolCallId: "fetch-001",
        Name:       "urls_fetch",
        Arguments: map[string]any{
            "urls": []any{
                "https://example.com",
                "https://httpbin.org/json",
            },
            "extract_text": true,
            "user_agent":   "My Agent 1.0",
        },
    }
    
    // Execute fetch
    result, err := tool.Call(context.Background(), params)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", result.Result)
}
```

## 参数说明

### 必需参数

- `urls` (array of strings): 要获取的 URL 列表

### 可选参数

- `extract_text` (boolean): 是否从 HTML 页面提取纯文本内容，默认 `false`
- `user_agent` (string): 自定义 User-Agent 头，默认 `"agent-go/1.0 URLsFetchTool"`
- `max_body_size` (integer): 响应体最大大小（字节），默认 `1048576` (1MB)
- `follow_redirect` (boolean): 是否跟随 HTTP 重定向，默认 `true`

## 返回结果

工具返回一个包含以下结构的结果：

```json
{
  "success": true,
  "data": {
    "results": [
      {
        "url": "https://example.com",
        "status_code": 200,
        "headers": {
          "Content-Type": "text/html; charset=utf-8",
          "Content-Length": "1234"
        },
        "content_type": "text/html; charset=utf-8",
        "content_length": 1234,
        "content": "<!DOCTYPE html>...",
        "text_content": "页面的纯文本内容...",
        "title": "页面标题",
        "fetch_time_ms": 150
      }
    ],
    "summary": {
      "total": 1,
      "success": 1,
      "failed": 0,
      "total_time_ms": 200
    }
  }
}
```

### 结果字段说明

#### URLResult 字段

- `url`: 原始 URL
- `status_code`: HTTP 状态码
- `headers`: HTTP 响应头（部分）
- `content_type`: Content-Type 响应头
- `content_length`: 内容长度（字节）
- `content`: 原始响应内容
- `text_content`: 提取的纯文本内容（仅当 `extract_text=true` 且为 HTML 时）
- `title`: HTML 页面标题（仅当 `extract_text=true` 且为 HTML 时）
- `error`: 错误信息（如果获取失败）
- `fetch_time_ms`: 单个 URL 获取耗时（毫秒）

#### Summary 字段

- `total`: 总 URL 数量
- `success`: 成功获取的数量
- `failed`: 失败的数量
- `total_time_ms`: 总耗时（毫秒）

## 高级用法

### 自定义超时时间

```go
tool := fetch.NewURLsFetchTool().WithTimeout(10 * time.Second)
```

### 处理大量 URL

工具内部使用信号量限制并发数（最多 10 个），可以安全地处理大量 URL：

```go
urls := make([]any, 100)
for i := 0; i < 100; i++ {
    urls[i] = fmt.Sprintf("https://httpbin.org/delay/%d", i%5)
}

params := &llms.ToolCall{
    ToolCallId: "batch-fetch",
    Name:       "urls_fetch",
    Arguments: map[string]any{
        "urls": urls,
    },
}
```

### 错误处理

工具会为每个 URL 单独处理错误，即使某些 URL 失败也不会影响其他 URL：

```go
params := &llms.ToolCall{
    Arguments: map[string]any{
        "urls": []any{
            "https://valid-site.com",
            "invalid-url",
            "https://404-site.com/not-found",
        },
    },
}

result, _ := tool.Call(context.Background(), params)
// 结果中会包含成功和失败的 URL 信息
```

## 支持的 URL 类型

- HTTP URLs (`http://`)
- HTTPS URLs (`https://`)

其他协议（如 `ftp://`, `file://`）不被支持。

## 安全考虑

1. **大小限制**: 默认限制响应体大小为 1MB，防止内存溢出
2. **并发限制**: 最多同时处理 10 个请求，防止资源耗尽
3. **超时保护**: 每个请求都有超时限制
4. **协议限制**: 只支持 HTTP/HTTPS 协议
5. **URL 验证**: 对输入的 URL 进行基本验证

## 在 AI Agent 中的应用场景

1. **网页内容分析**: 获取网页内容供 AI 分析
2. **实时信息获取**: 获取最新的新闻、股价、天气等信息
3. **批量内容处理**: 同时处理多个网页或 API 端点
4. **网站监控**: 检查多个网站的可用性
5. **数据收集**: 从多个来源收集结构化数据

## 测试

运行测试：

```bash
go test ./pkg/core/tools/fetch/...
```

测试覆盖了以下场景：
- 基本 URL 获取
- HTML 文本提取
- 多 URL 并发处理
- 错误处理
- 参数验证
- 工具描述符

## 依赖

- `golang.org/x/net/html`: HTML 解析
- 标准库: `net/http`, `context`, `sync` 等
