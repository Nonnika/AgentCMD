# Function Call 实现演示

## 完整流程说明

### 1. 用户输入
```
User: 帮我创建一个文件 test.txt
```

### 2. 系统处理流程

```
ExecuteCommand(cmd, output, chatMessages, client)
    │
    ├── 添加用户消息到历史
    │   Message{Role: "user", Content: "帮我创建一个文件 test.txt"}
    │
    ├── 调用模型
    │   client.Chat(ctx, chatMessages)
    │   
    │   发送给模型的请求:
    │   {
    │       "model": "deepseek-chat",
    │       "messages": [
    │           {"role": "user", "content": "帮我创建一个文件 test.txt"}
    │       ],
    │       "tools": [
    │           {
    │               "type": "function",
    │               "function": {
    │                   "name": "CreateFile",
    │                   "description": "Create a file at the specified path.",
    │                   "parameters": {
    │                       "type": "object",
    │                       "required": ["fp"],
    │                       "properties": {
    │                           "fp": {"type": "string"}
    │                       }
    │                   }
    │               }
    │           }
    │       ]
    │   }
    │
    ├── 模型返回:
    │   {
    │       "choices": [{
    │           "message": {
    │               "role": "assistant",
    │               "content": null,
    │               "tool_calls": [{
    │                   "id": "call_abc123",
    │                   "type": "function",
    │                   "function": {
    │                       "name": "CreateFile",
    │                       "arguments": "{\"fp\":\"test.txt\"}"
    │                   }
    │               }]
    │           },
    │           "finish_reason": "tool_calls"
    │       }]
    │   }
    │
    ├── 检测到 tool call:
    │   tool.Function.Name = "CreateFile"
    │   tool.Function.Arguments = `{"fp":"test.txt"}`
    │
    ├── 执行函数:
    │   executeFunctionCall(tool, output)
    │   ├── toolFunction = IndexFunctions["CreateFile"]
    │   ├── args = string(tool.Function.Arguments) = `{"fp":"test.txt"}`
    │   ├── result, err = CreateFileWrapper(`{"fp":"test.txt"}`)
    │   │   ├── 解析 JSON: args.Fp = "test.txt"
    │   │   ├── CreateFile("test.txt")
    │   │   │   ├── os.MkdirAll(dir, os.ModePerm)
    │   │   │   ├── os.Stat(fp)
    │   │   │   └── os.Create(fp)
    │   │   └── return "File created successfully", nil
    │   └── return "File created successfully"
    │
    ├── 构建消息历史:
    │   toolCallMsg = Message{
    │       Role: "assistant",
    │       Content: "Calling function: CreateFile..."
    │   }
    │   
    │   toolResultMsg = Message{
    │       Role: "tool",
    │       Content: "File created successfully",
    │       ToolCallID: "call_abc123",
    │       Name: "CreateFile"
    │   }
    │
    ├── 添加到历史:
    │   chatMessages = [
    │       {Role: "user", Content: "帮我创建一个文件 test.txt"},
    │       {Role: "assistant", Content: "Calling function: CreateFile..."},
    │       {Role: "tool", Content: "File created successfully", ToolCallID: "call_abc123", Name: "CreateFile"}
    │   ]
    │
    └── 递归调用 ExecuteCommand("", ...)
        └── 再次调用模型，但这次包含工具结果
            
            模型看到:
            - 用户要求创建文件
            - 工具已被调用并返回成功
            
            模型生成最终回复:
            "文件 test.txt 已成功创建！"
```

### 3. 最终输出
```
User: 帮我创建一个文件 test.txt
Broith: 正在调用函数 CreateFile
Arguments: {"fp":"test.txt"}

Broith: 文件 test.txt 已成功创建！
```

## 关键设计决策

### 1. 统一 JSON 参数格式
所有工具函数接收 JSON 字符串参数：
```go
type ToolFunc func(jsonArgs string) (string, error)
```

这使得与模型返回的 JSON 参数天然对齐，无需复杂的类型转换。

### 2. 递归处理 Tool Calls
使用递归而非循环来处理 tool calls，使得代码逻辑更清晰：
- 调用模型
- 检查是否有 tool call
- 如果有：执行 tool，添加结果到历史，递归调用
- 如果没有：显示结果，完成

### 3. 消息历史的完整性
确保消息历史包含所有必要的上下文：
- 用户消息
- 助手消息（可能包含 tool call）
- Tool 结果消息

这使得模型能够在下一次调用时看到完整的执行上下文。

## 扩展点

### 添加新工具
只需三步：
1. 实现 `XXXWrapper(jsonArgs string) (string, error)`
2. 添加到 `IndexFunctions` 映射
3. 在 `GenerateToolsJSON()` 中添加描述

### 支持多 Tool Calls
当前实现只处理单个 tool call。要支持多个：
1. 修改 `Tool` 类型为 `[]Tool`
2. 在 `handleToolCall` 中遍历处理每个 tool
3. 将所有结果添加到消息历史

### 添加 Tool 执行超时
为长时间运行的工具添加超时机制：
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := toolFunctionWithContext(ctx, args)
```
