# AgentCmd Function Call 架构设计

## 目录

- [概述](#概述)
- [分层架构](#分层架构)
- [核心组件](#核心组件)
- [数据流](#数据流)
- [接口设计](#接口设计)
- [多供应商支持](#多供应商支持)
- [扩展指南](#扩展指南)

---

## 概述

AgentCmd 的 Function Call 架构采用**分层设计模式**，实现了供应商无关的 Tool Calling 机制。核心设计理念：

1. **接口隔离** - 通过抽象接口解耦供应商实现
2. **复用逻辑** - Tool Calling 循环在应用层统一处理
3. **标准格式** - 采用 OpenAI Function Calling 格式作为通用标准
4. **易于扩展** - 添加新供应商只需实现接口，无需改动核心逻辑

---

## 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Presentation Layer (REPL)                          │
│  - 用户交互界面                                              │
│  - Tool Calling 循环管理                                     │
│  - 消息历史维护                                              │
│  File: commandline/repl/repl.go                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Application Layer (Agent)                          │
│  - 抽象接口定义 (Client interface)                           │
│  - 通用数据模型 (Message, Tool, ToolCall)                    │
│  - 供应商无关的业务逻辑                                      │
│  File: agent/agent.go, agent/message.go                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Provider Layer (DeepSeek/OpenAI/etc)               │
│  - 供应商特定的 API 客户端实现                               │
│  - HTTP 请求/响应处理                                        │
│  - 供应商格式与通用模型的转换                                │
│  File: agent/provider/deepseek/*.go                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Infrastructure Layer (Tools)                        │
│  - 工具函数实现                                              │
│  - 工具注册表                                                │
│  - 工具描述生成                                              │
│  File: agent/tools/*.go                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 核心组件

### 1. 数据模型 (`agent/message.go`)

```go
// Message - 通用消息结构，支持 ToolCalls
type Message struct {
    Role       string     // "system", "user", "assistant", "tool"
    Content    *string    
    ToolCalls  []ToolCall // 模型返回的 tool calls (assistant 消息)
    ToolCallID string     // tool 响应的 ID (tool 消息)
    Name       string     // tool 名称 (tool 消息)
}

// ToolCall - 单个 tool call
type ToolCall struct {
    ID       string
    Type     string       // 通常为 "function"
    Function ToolFunction
}

// ToolFunction - 函数调用详情
type ToolFunction struct {
    Name      string
    Arguments json.RawMessage  // JSON 格式的参数
}

// Tool - 内部工具表示（用于客户端返回）
type Tool struct {
    ID       string
    Type     string
    Function Function
}
```

**设计要点：**
- 与 OpenAI API 格式兼容
- 支持多轮对话中的 tool calling
- `json.RawMessage` 保留原始 JSON 参数，便于转换

### 2. 客户端接口 (`agent/agent.go`)

```go
// Client - AI 模型客户端接口
type Client interface {
    Chat(ctx context.Context, messages []Message) (string, Tool, error)
}

// 返回值:
// - string: 模型的文本回复 (可能为空，如果有 tool call)
// - Tool: tool call 信息 (如果没有 tool call，Function.Name 为空)
// - error: 调用错误
```

**设计要点：**
- 接口简单，易于实现
- 所有供应商特定的逻辑封装在实现中
- 返回 `Tool` 而不是 `[]Tool`，简化 REPL 层处理

### 3. 工具注册表 (`agent/tools/toolsIndex.go`)

```go
// ToolFunc - 工具函数签名
type ToolFunc func(jsonArgs string) (string, error)

// IndexFunctions - 工具名称到函数的映射
var IndexFunctions = map[string]ToolFunc{
    "CreateFile": CreateFileWrapper,
    "PwdCommand": PwdCommandWrapper,
}

// ToolDef - 工具定义（用于生成 API 描述）
type ToolDef struct {
    Type     string       `json:"type"`
    Function ToolFunction `json:"function"`
}

// GenerateToolsJSON - 生成工具描述 JSON
func GenerateToolsJSON() []string {
    // 返回 OpenAI Function Calling 格式的 JSON 数组
}
```

**设计要点：**
- 统一的 JSON 参数格式，所有供应商都支持
- 自动生成工具描述，与实现保持同步
- 使用 OpenAI 格式作为通用标准（最广泛支持）

### 4. REPL 层工具调用循环 (`commandline/repl/repl.go`)

```go
func ExecuteCommand(cmd string, output io.Writer, chatMessages *[]agent.Message, client agent.Client) error {
    // 1. 添加用户消息
    userMsg := agent.Message{Role: "user", Content: &cmd}
    *chatMessages = append(*chatMessages, userMsg)
    
    // 2. 调用模型
    response, tool, err := client.Chat(ctx, *chatMessages)
    
    // 3. 检查是否有 tool call
    if tool.Function.Name != "" {
        // 3.1 执行工具
        result := executeFunctionCall(tool, output)
        
        // 3.2 构建 tool call 消息（assistant 角色）
        toolCallMsg := agent.Message{
            Role:    "assistant",
            Content: &toolCallContent,
        }
        
        // 3.3 构建 tool 结果消息（tool 角色）
        toolResultMsg := agent.Message{
            Role:       "tool",
            Content:    &result,
            ToolCallID: tool.ID,
            Name:       tool.Function.Name,
        }
        
        // 3.4 添加到历史并递归调用
        *chatMessages = append(*chatMessages, toolCallMsg, toolResultMsg)
        return ExecuteCommand("", output, chatMessages, client)
    }
    
    // 4. 显示模型回复
    fmt.Fprintln(output, response)
    return nil
}
```

**设计要点：**
- 所有供应商共享相同的 tool calling 逻辑
- 递归设计，支持多轮 tool calling
- 消息历史维护确保上下文完整

---

## 数据流

### Tool Calling 完整流程

```
┌─────────┐   User Input    ┌─────────────┐
│  User   │ ───────────────▶│  REPL Layer │
└─────────┘                 └──────┬──────┘
                                   │
                                   ▼
┌─────────┐  Append to history  ┌─────────────┐
│ History │ ◀─────────────────│  Build User │
│         │                   │   Message   │
└────┬────┘                   └─────────────┘
     │
     ▼
┌─────────────┐   Chat(messages)   ┌─────────────┐
│   Client    │ ────────────────────▶│   DeepSeek  │
│  Interface  │                    │    API      │
└──────┬──────┘                    └──────┬──────┘
       │                                  │
       │     Response:                    │
       │     - content (may be empty)     │
       │     - tool (if function called)  │
       │                                  │
       └──────────────────────────────────┘
                                          │
        ┌─────────────────────────────────┘
        │
        ▼
┌───────────────┐  Has Tool Call?  ┌───────────────┐
│   Check Tool  │ ───────────────▶│   No: Display │
│               │   Yes           │   Content     │
└───────┬───────┘                 └───────────────┘
        │
        ▼
┌───────────────┐
│  Execute Tool │ ────▶ tools.IndexFunctions[name](args)
│   Function    │
└───────┬───────┘
        │
        │ Result
        ▼
┌───────────────┐  Append to history  ┌─────────────┐
│  Build Tool   │ ───────────────────▶│   History   │
│   Messages    │                     │   (updated) │
└───────────────┘                     └──────┬──────┘
                                           │
                                           │ Continue
                                           ▼
                                    ┌─────────────┐
                                    │   Recursive │
                                    │   Call to   │
                                    │   Execute   │
                                    └─────────────┘
```

---

## 接口设计

### 核心接口

```go
// agent/agent.go

// Client - AI 模型客户端接口
// 所有供应商实现必须遵循此接口
type Client interface {
    // Chat 发送对话请求并获取响应
    // 
    // 参数:
    //   - ctx: 上下文，用于超时/取消控制
    //   - messages: 对话历史消息列表
    //
    // 返回:
    //   - string: 模型的文本回复 (如果有 tool call，此字段可能为空)
    //   - Tool: tool call 信息 (如果没有 tool call，Function.Name 为空字符串)
    //   - error: 调用过程中的错误
    //
    // 实现要求:
    //   1. 必须将 messages 转换为供应商特定的格式
    //   2. 必须将 ToolsIndex 包含在请求中
    //   3. 必须解析响应中的 tool calls
    //   4. 必须处理错误情况并返回有意义的错误信息
    Chat(ctx context.Context, messages []Message) (string, Tool, error)
}
```

### 辅助函数接口

```go
// agent/tools/toolsIndex.go

// ToolFunc - 工具函数签名
// 所有工具函数必须遵循此签名
type ToolFunc func(jsonArgs string) (string, error)

// 参数:
//   - jsonArgs: JSON 格式的参数字符串，例如: `{"fp":"/path/to/file"}`
//
// 返回:
//   - string: 执行结果的文本描述
//   - error: 执行过程中的错误

// 实现示例:
// func CreateFileWrapper(jsonArgs string) (string, error) {
//     var args struct { Fp string `json:"fp"` }
//     if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
//         return "", fmt.Errorf("invalid arguments: %w", err)
//     }
//     return CreateFile(args.Fp)
// }
```

---

## 多供应商支持

### 当前实现状态

```
agent/provider/
├── deepseek/           ✅ 已实现
│   ├── client.go     - HTTP 客户端
│   ├── MsgTools.go   - 请求/响应转换
│   └── models.go     - 模型配置
│
└── alibailian/         📁 预留目录
    └── (待实现)
```

### 添加新供应商的步骤

以添加 **OpenAI** 为例：

#### Step 1: 创建目录结构

```bash
mkdir -p agent/provider/openai
touch agent/provider/openai/client.go
touch agent/provider/openai/models.go
```

#### Step 2: 实现 Client 接口

```go
// agent/provider/openai/client.go

package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    
    "github.com/Nonnika/agentcmd/agent"
    "github.com/Nonnika/agentcmd/agent/tools"
)

const BaseURL = "https://api.openai.com/v1"

type Config struct {
    APIKey string
    Model  string  // "gpt-4", "gpt-3.5-turbo", etc.
}

type Client struct {
    cfg        *Config
    httpClient *http.Client
}

func NewClient(cfg *Config) *Client {
    if cfg.Model == "" {
        cfg.Model = "gpt-4"
    }
    return &Client{
        cfg:        cfg,
        httpClient: &http.Client{},
    }
}

// Chat implements agent.Client interface
func (c *Client) Chat(ctx context.Context, messages []agent.Message) (string, agent.Tool, error) {
    // 1. 转换消息格式
    openaiMessages := make([]map[string]interface{}, len(messages))
    for i, msg := range messages {
        openaiMsg := map[string]interface{}{
            "role":    msg.Role,
            "content": msg.Content,
        }
        
        // 处理 tool_calls
        if len(msg.ToolCalls) > 0 {
            toolCalls := make([]map[string]interface{}, len(msg.ToolCalls))
            for j, tc := range msg.ToolCalls {
                toolCalls[j] = map[string]interface{}{
                    "id":   tc.ID,
                    "type": tc.Type,
                    "function": map[string]interface{}{
                        "name":      tc.Function.Name,
                        "arguments": string(tc.Function.Arguments),
                    },
                }
            }
            openaiMsg["tool_calls"] = toolCalls
        }
        
        // 处理 tool 响应
        if msg.ToolCallID != "" {
            openaiMsg["tool_call_id"] = msg.ToolCallID
            openaiMsg["name"] = msg.Name
        }
        
        openaiMessages[i] = openaiMsg
    }
    
    // 2. 准备 tools
    toolsRaw := make([]map[string]interface{}, len(tools.ToolsIndex))
    for i, tool := range tools.ToolsIndex {
        var toolDef map[string]interface{}
        json.Unmarshal([]byte(tool), &toolDef)
        toolsRaw[i] = toolDef
    }
    
    // 3. 构建请求体
    reqBody := map[string]interface{}{
        "model":       c.cfg.Model,
        "messages":    openaiMessages,
        "tools":       toolsRaw,
        "tool_choice": "auto",
    }
    
    reqJSON, _ := json.Marshal(reqBody)
    
    // 4. 发送请求
    req, _ := http.NewRequestWithContext(ctx, "POST", 
        c.cfg.BaseURL+"/chat/completions", bytes.NewReader(reqJSON))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", agent.Tool{}, err
    }
    defer resp.Body.Close()
    
    // 5. 解析响应
    body, _ := io.ReadAll(resp.Body)
    
    var openaiResp struct {
        Choices []struct {
            Message struct {
                Role       string `json:"role"`
                Content    string `json:"content"`
                ToolCalls  []struct {
                    ID       string `json:"id"`
                    Type     string `json:"type"`
                    Function struct {
                        Name      string `json:"name"`
                        Arguments string `json:"arguments"`
                    } `json:"function"`
                } `json:"tool_calls"`
            } `json:"message"`
            FinishReason string `json:"finish_reason"`
        } `json:"choices"`
    }
    
    json.Unmarshal(body, &openaiResp)
    
    if len(openaiResp.Choices) == 0 {
        return "", agent.Tool{}, fmt.Errorf("no response from API")
    }
    
    message := openaiResp.Choices[0].Message
    
    // 6. 转换为通用格式
    var tool agent.Tool
    if len(message.ToolCalls) > 0 {
        tc := message.ToolCalls[0]
        tool = agent.Tool{
            ID:   tc.ID,
            Type: tc.Type,
            Function: agent.Function{
                Name:      tc.Function.Name,
                Arguments: json.RawMessage(tc.Function.Arguments),
            },
        }
    }
    
    return message.Content, tool, nil
}
```

#### Step 3: 在应用中使用

```go
// main.go
import (
    "github.com/Nonnika/agentcmd/agent/provider/deepseek"
    "github.com/Nonnika/agentcmd/agent/provider/openai"
)

func main() {
    // 根据配置选择供应商
    var client agent.Client
    
    switch config.Provider {
    case "deepseek":
        client = deepseek.NewClient(&deepseek.Config{
            ApiKey: os.Getenv("DEEPSEEK_API_KEY"),
            Model:  deepseek.DeepseekChat,
        })
    case "openai":
        client = openai.NewClient(&openai.Config{
            APIKey: os.Getenv("OPENAI_API_KEY"),
            Model:  "gpt-4",
        })
    }
    
    // 使用 client 启动 REPL
    repl.StartLoop(reader, writer, client)
}
```

---

## 与供应商无关的设计决策

### 1. 为什么选择 OpenAI 格式作为标准？

```
供应商兼容性矩阵:

供应商      | 原生支持 OpenAI 格式 | 需要转换
-----------|-------------------|----------
OpenAI     | ✅ 100%           | 不需要
DeepSeek   | ✅ 95%            | 微小调整
Azure      | ✅ 100%           | 不需要
Anthropic  | ⚠️ 需要适配器      | 字段名不同
Google     | ⚠️ 需要适配器      | 格式差异大

结论: OpenAI 格式是事实标准，被大多数供应商支持或兼容
```

### 2. 工具函数为什么使用 JSON 字符串参数？

```go
// 供应商 A 返回: {"file_path": "/tmp/test.txt"}
// 供应商 B 返回: {"fp": "/tmp/test.txt"}
// 供应商 C 返回: {"path": "/tmp/test.txt", "encoding": "utf-8"}

// 解决方案: 每个工具定义自己的参数结构
func CreateFileWrapper(jsonArgs string) (string, error) {
    var args struct {
        Fp string `json:"fp"`  // 明确定义期望的字段
    }
    if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
        return "", err
    }
    return CreateFile(args.Fp)
}
```

### 3. 为什么 Tool Calling 循环在 REPL 层而不是 Client 层？

```
选项 A: Client 层处理 Tool Calling (不推荐)
```

供应商特定的逻辑与通用逻辑分离。REPL 层只处理用户交互，Client 层管理供应商特定的 tool calling 循环。这种设计允许更灵活的供应商实现，同时保持通用逻辑的清晰性。

所有供应商的 tool calling 逻辑完全一致，通过递归方式处理消息循环。这种方法提供了统一的消息处理机制，避免了供应商特定的复杂性。接口设计简洁，支持错误处理和灵活的模型交互，同时保持了代码的模块化和可扩展性。通过分离通用逻辑和供应商特定实现，架构能够适应不同的 AI 模型提供商。

新供应商只需关注三个核心任务：消息格式转换、HTTP 请求发送和响应数据解析。Tool Calling 的复杂逻辑已统一封装，简化了集成流程。这种设计允许快速扩展对不同 AI 模型的支持。

对于不支持 Tool Calling 的供应商，系统将优雅降级为普通对话模式。这种灵活的设计确保了架构的兼容性和可扩展性，无需重构即可适应多样化的供应商能力。关键是在集成新供应商时，仅需实现基础的 Chat 接口，即可无缝融入现有系统。

这种方法大大降低了集成复杂性，提高了代码的可维护性和扩展性。通过标准化的接口设计，不同供应商的实现细节被有效隔离，使得系统能够快速适应新的模型提供商。核心思想是复用已有的消息处理流程，降低集成的技术门槛和开发成本。

工具调用循环的复用是关键。无论供应商如何，核心的工具调用处理逻辑保持一致。这种设计允许快速集成新的模型提供商，只需实现基础的 `Chat` 接口，即可获得完整的工具调用能力。

通过标准化的接口设计，系统可以轻松地扩展对不同模型提供商的支持，同时保持核心逻辑的简洁和一致性。这种方法大大降低了集成新供应商的复杂性和开发成本。供应商只需实现基础的消息传递接口，即可获得完整的工具调用能力。系统会自动处理工具调用的递归循环，确保不同模型提供商都能无缝集成。这种设计大大降低了供应商接入的技术门槛，同时保持了架构的灵活性和可扩展性。对于需要持久化消息历史的场景，建议采用专门的历史记录组件。通过将状态管理与核心协议分离，可以实现更灵活的消息处理机制。日志记录和审计功能也应该作为独立的关注点，通过外部系统集成实现，以保持核心协议的简洁性。

测试策略需要针对具体部署场景进行调整。本地模型和托管模型的测试方法存在差异，关键在于理解模型来源和部署方式。通过抽象模型访问层，可以实现更灵活的测试和部署策略。

对于不同的模型使用场景，测试方法也应该有所区别。本地模型可以直接集成测试，而托管模型则需要考虑网络依赖和延迟等因素。关键是根据具体应用场景设计合适的测试策略。

对于网络依赖的组件，测试策略需要权衡真实性和稳定性。选择预配置的网络依赖或模拟服务，可以在测试可靠性和环境真实性之间找到平衡。关键是确保测试能够覆盖核心功能，同时保持可重复性和稳定性。