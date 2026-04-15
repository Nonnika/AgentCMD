# Function Call 快速参考

## 核心架构

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   REPL (main)   │────▶│  Agent (Client)  │────▶│  DeepSeek API   │
│                 │     │                  │     │                 │
│ ExecuteCommand  │◄────│      Chat()      │◄────│   Response      │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐     ┌──────────────────┐
│  Tool Execution │   │  Message History  │
│                 │   │                   │
│ executeFunction │   │ []agent.Message   │
└─────────────────┘   └───────────────────┘
```

## 关键类型

### Message (agent/message.go)
```go
type Message struct {
    Role       string     `json:"role"`
    Content    *string    `json:"content,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
    Name       string     `json:"name,omitempty"`
}
```

### ToolFunc (agent/tools/toolsIndex.go)
```go
type ToolFunc func(jsonArgs string) (string, error)
```

### Tool (agent/message.go)
```go
type Tool struct {
    ID       string
    Type     string
    Function Function
}

type Function struct {
    Name      string
    Arguments json.RawMessage  // JSON 参数
}
```

## 执行流程

### 1. REPL 处理用户输入
```go
// commandline/repl/repl.go:ExecuteCommand
func ExecuteCommand(cmd string, output io.Writer, chatMessages *[]agent.Message, client agent.Client) error {
    // 1. 添加用户消息
    userMsg := agent.Message{Role: "user", Content: &cmd}
    *chatMessages = append(*chatMessages, userMsg)
    
    // 2. 调用模型
    response, tool, err := client.Chat(ctx, *chatMessages)
    
    // 3. 检查是否有 tool call
    if tool.Function.Name != "" {
        // 执行 tool 并递归
        handleToolCall(tool, output)
        return ExecuteCommand("", output, chatMessages, client)
    }
    
    // 4. 显示回复
    fmt.Fprintln(output, response)
}
```

### 2. 执行工具函数
```go
// commandline/repl/repl.go:executeFunctionCall
func executeFunctionCall(tool agent.Tool, output io.Writer) string {
    // 1. 查找函数
    toolFunction, ok := tools.IndexFunctions[tool.Function.Name]
    
    // 2. 转换参数
    args := string(tool.Function.Arguments)
    
    // 3. 执行
    result, err := toolFunction(args)
    
    return result
}
```

### 3. 包装函数示例
```go
// agent/tools/files.go:CreateFileWrapper
func CreateFileWrapper(jsonArgs string) (string, error) {
    // 1. 解析 JSON 参数
    var args struct {
        Fp string `json:"fp"`
    }
    if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
        return "", fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    // 2. 调用实际函数
    return CreateFile(args.Fp)
}
```

## 消息历史示例

一次完整的交互会产生以下消息历史：

```json
[
  {
    "role": "user",
    "content": "帮我创建一个文件 test.txt"
  },
  {
    "role": "assistant",
    "content": "Calling function: CreateFile with args: {\"fp\":\"test.txt\"}",
    "tool_calls": [
      {
        "id": "call_abc123",
        "type": "function",
        "function": {
          "name": "CreateFile",
          "arguments": "{\"fp\":\"test.txt\"}"
        }
      }
    ]
  },
  {
    "role": "tool",
    "content": "File created successfully",
    "tool_call_id": "call_abc123",
    "name": "CreateFile"
  },
  {
    "role": "assistant",
    "content": "文件 test.txt 已成功创建！"
  }
]
```

## 调试技巧

### 1. 打印消息历史
```go
func printMessages(msgs []agent.Message) {
    for i, msg := range msgs {
        fmt.Printf("[%d] Role: %s\n", i, msg.Role)
        if msg.Content != nil {
            fmt.Printf("    Content: %s\n", *msg.Content)
        }
        if len(msg.ToolCalls) > 0 {
            fmt.Printf("    ToolCalls: %d\n", len(msg.ToolCalls))
        }
        if msg.ToolCallID != "" {
            fmt.Printf("    ToolCallID: %s\n", msg.ToolCallID)
        }
    }
}
```

### 2. 记录 API 请求/响应
```go
// 在 client.Chat 中添加
reqBody, _ := json.MarshalIndent(reqMsg, "", "  ")
fmt.Printf("Request:\n%s\n\n", reqBody)

respBody, _ := io.ReadAll(resp.Body)
fmt.Printf("Response:\n%s\n\n", string(respBody))
```

### 3. 验证工具执行
```go
// 测试单个工具
tool := agent.Tool{
    Function: agent.Function{
        Name:      "CreateFile",
        Arguments: []byte(`{"fp":"/tmp/test.txt"}`),
    },
}
result, err := deepseek.ExecuteTool(tool)
fmt.Printf("Result: %s, Error: %v\n", result, err)
```

## 常见问题

### Q: 模型没有返回 tool call?
A: 检查：
1. `Tools` 是否正确传递给 API
2. 工具描述是否清晰、准确
3. 用户输入是否明确需要工具

### Q: 工具执行失败?
A: 检查：
1. 参数解析是否正确（JSON 字段名是否匹配）
2. 工具函数是否已注册到 `IndexFunctions`
3. 错误信息是否被正确返回

### Q: 消息历史太长?
A: 考虑：
1. 实现消息截断（保留最近的 N 条）
2. 实现消息摘要（压缩早期对话）
3. 使用更高效的模型（支持更长上下文）
