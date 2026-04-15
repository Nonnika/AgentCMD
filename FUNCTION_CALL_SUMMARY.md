# Function Call 实现总结

## 实现概述

已完成完整的 Function Call 调用机制，支持从模型获取 tool call，执行本地函数，并将结果返回给模型。

## 核心组件

### 1. 数据模型 (`agent/message.go`)

- `Message` - 扩展了标准消息格式，支持 ToolCalls
- `ToolCall` - 模型返回的 tool call 结构
- `Tool` - 内部工具表示
- `ToolFunction` / `Function` - 函数调用详情

### 2. 工具索引 (`agent/tools/toolsIndex.go`)

- `ToolFunc` - 工具函数签名：`func(jsonArgs string) (string, error)`
- `IndexFunctions` - 工具名称到函数的映射
- `ToolDef` / `ToolFunction` - 工具定义结构
- `GenerateToolsJSON()` - 自动生成工具描述 JSON

### 3. 消息构建 (`agent/provider/deepseek/MsgTools.go`)

- `BuildReqMsg()` - 构建 API 请求
- `ConvertRespMessageToAgent()` - 转换响应为内部格式
- `BuildToolCallMessage()` - 构建 tool call 消息
- `ExecuteTool()` - 执行工具函数

### 4. REPL 集成 (`commandline/repl/repl.go`)

- `ExecuteCommand()` - 处理用户输入，调用模型
- `executeFunctionCall()` - 执行具体的工具函数
- `handleToolCall()` - 处理 tool call 全流程

## 使用流程

```
用户输入
    ↓
添加到消息历史
    ↓
调用模型 (client.Chat)
    ↓
模型返回: 内容 + 可能的 tool call
    ↓
如果有 tool call:
    ├── 显示"正在调用函数 X"
    ├── 执行函数 (executeFunctionCall)
    ├── 将 tool call 和结果添加到消息历史
    └── 递归调用 ExecuteCommand (空输入) 让模型继续处理
    ↓
显示模型回复
    ↓
添加到消息历史
```

## 示例交互

```
User: 帮我创建一个文件 test.txt
Broith: 正在调用函数 CreateFile
Arguments: {"fp":"test.txt"}

Broith: 文件 test.txt 已成功创建！
```

## 扩展工具

要添加新工具，需要:

1. 在 `agent/tools/files.go` 或其他文件中实现函数
2. 在 `toolsIndex.go` 的 `IndexFunctions` 中添加映射
3. 在 `GenerateToolsJSON()` 中添加工具描述

示例:

```go
// 实现函数
func ReadFileWrapper(jsonArgs string) (string, error) {
    var args struct {
        Path string `json:"path"`
    }
    if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
        return "", err
    }
    return ReadFile(args.Path)
}

// 添加到索引
var IndexFunctions = map[string]ToolFunc{
    "CreateFile": CreateFileWrapper,
    "ReadFile":   ReadFileWrapper,  // 新添加
}

// 添加到工具描述
func GenerateToolsJSON() []string {
    tools := []ToolDef{
        // ... 已有工具 ...
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "ReadFile",
                Description: "Read the contents of a file.",
                Parameters: map[string]interface{}{
                    "type":     "object",
                    "required": []string{"path"},
                    "properties": map[string]interface{}{
                        "path": map[string]interface{}{
                            "type":        "string",
                            "description": "The path to the file to read.",
                        },
                    },
                },
            },
        },
    }
    // ...
}
```
