package deepseek

import (
	"encoding/json"

	"github.com/Nonnika/agentcmd/agent/tools"
)

// This package is used for DeepSeek Model
// Include create chat request, response, and build request, response from config, messages

type Message struct {
	Role    string  `json:"role"`
	Content *string `json:"content,omitempty"`
}

type RespMessage struct {
	Role             string     `json:"role"`
	Content          *string    `json:"content,omitempty"`
	ReasoningContent *string    `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Id       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// strcut for decode response message
type RespMsg struct {
	Id                string    `json:"id"`
	Object            string    `json:"object"`
	Created           int64     `json:"created"`
	Model             string    `json:"model"`
	Choices           []Choices `json:"choices"`
	Usage             Usage     `json:"usage"`
	SystemFingerprint string    `json:"system_fingerprint"`
}

type ReqMsg struct {
	Model     string            `json:"model"`
	Messages  []Message         `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
	Tools     []json.RawMessage `json:"tools"`
}

type Choices struct {
	Index        int         `json:"index"`
	Message      RespMessage `json:"message"`
	Logprobs     string      `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	TotalTokens           int                `json:"total_tokens"`
	PromptTokens          int                `json:"prompt_tokens"`
	CompletionTokens      int                `json:"completion_tokens"`
	PromptTokenDetails    PromptTokenDetails `json:"prompt_token_details"`
	PromptCacheHitTokens  int                `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int                `json:"prompt_cache_miss_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens int64 `json:"cached_tokens"`
}

func BuildReqMsg(cfg *Config, Messages []Message) ReqMsg {
	toolsRaw := make([]json.RawMessage, len(tools.ToolsIndex))
	for i, tool := range tools.ToolsIndex {
		toolsRaw[i] = json.RawMessage(tool)
	}
	return ReqMsg{
		Model:     models[cfg.Model],
		Messages:  Messages,
		MaxTokens: 1024,
		Tools:     toolsRaw,
	}
}
